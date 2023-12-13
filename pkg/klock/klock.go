// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package klock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"

	"github.com/applejag/kubectl-klock/pkg/table"
)

type Options struct {
	ConfigFlags *genericclioptions.ConfigFlags

	LabelSelector   string
	FieldSelector   string
	AllNamespaces   bool
	WatchKubeconfig bool
	LabelColumns    []string

	Output string
}

func (o Options) Validate() error {
	const allowedFormats = "wide"
	switch o.Output {
	case "", "wide":
		// Valid
		return nil
	case "custom-columns", "custom-columns-file", "go-template",
		"go-template-file", "json", "jsonpath", "jsonpath-as-json",
		"jsonpath-file", "name", "template", "templatefile", "yaml":
		return fmt.Errorf("unsupported output format: %q, allowed formats are: %s", o.Output, allowedFormats)

	default:
		return fmt.Errorf("unknown output format: %q, allowed formats are: %s", o.Output, allowedFormats)
	}
}

func Execute(o Options, args []string) error {
	if err := o.Validate(); err != nil {
		return err
	}
	var fileEvents chan fsnotify.Event
	if o.WatchKubeconfig {
		if fileWatcher, err := fsnotify.NewWatcher(); err == nil {
			configLoader := o.ConfigFlags.ToRawKubeConfigLoader()
			kubeconfigFiles := configLoader.ConfigAccess().GetLoadingPrecedence()
			for _, file := range kubeconfigFiles {
				fileWatcher.Add(file)
			}
			fileEvents = fileWatcher.Events
			defer fileWatcher.Close()
		}
	}

	t := table.New()
	printer := Printer{
		Table:      t,
		WideOutput: o.Output == "wide",
		LabelCols:  o.LabelColumns,
	}
	p := tea.NewProgram(t)
	w := NewWatcher(o, p, printer, args)
	t.StartSpinner()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	restartChan := make(chan struct{})
	defer close(restartChan)

	go func() {
		for {
			select {
			case event, ok := <-fileEvents:
				if !ok {
					fileEvents = nil
					continue
				}
				if event.Op != fsnotify.Write {
					continue
				}
				restartChan <- struct{}{}

			case err := <-w.ErrorChan():
				t.SetError(err)
				p.Send(nil)
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		w.WatchLoop(ctx, restartChan)
	}()

	_, err := p.Run()
	return err
}

func NewWatcher(options Options, program *tea.Program, printer Printer, args []string) *Watcher {
	return &Watcher{
		Options: options,
		Program: program,
		Printer: printer,
		Args:    args,

		errorChan: make(chan error, 3),
	}
}

type Watcher struct {
	Options
	Program *tea.Program
	Printer Printer
	Args    []string

	errorChan chan error
}

func (w *Watcher) ErrorChan() <-chan error {
	return w.errorChan
}

func (w *Watcher) WatchLoop(ctx context.Context, restartChan <-chan struct{}) error {
	clearBeforePrinting := false
	watchErrChan := make(chan error, 1)
	for {
		var wg sync.WaitGroup
		wg.Add(1)
		watchCtx, cancel := context.WithCancel(ctx)
		go func(clearBeforePrinting bool) {
			defer wg.Done()
			err := w.watch(watchCtx, clearBeforePrinting)
			if watchCtx.Err() == nil {
				watchErrChan <- err
			}
		}(clearBeforePrinting)
		clearBeforePrinting = true
		select {
		case err := <-watchErrChan:
			if cmd := w.Printer.Table.StartSpinner(); cmd != nil {
				w.Program.Send(cmd())
			}
			w.errorChan <- fmt.Errorf("restart in 5s: %w", err)
			time.Sleep(5 * time.Second)
			w.Printer.Table.StopSpinner()
			w.Printer.Table.SetError(nil)
		case <-restartChan:
			if cmd := w.Printer.Table.StartSpinner(); cmd != nil {
				w.Program.Send(cmd())
			}
			// Prevent it from restarting too eagerly when we're told to restart
			// so the filesystem has time to flush, such as in case of
			// "kubectx" on bigger kubeconfigs.
			// https://github.com/applejag/kubectl-klock/issues/62
			slidingSleep(150*time.Millisecond, restartChan)
			cancel()
		case <-ctx.Done():
			cancel()
			return ctx.Err()
		}
		wg.Wait()
	}
}

func slidingSleep(dur time.Duration, ch <-chan struct{}) {
	timer := time.NewTimer(dur)
	for {
		select {
		case <-timer.C:
			return
		case <-ch:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(dur)
		}
	}
}

func (w *Watcher) Watch(ctx context.Context) error {
	return w.watch(ctx, false)
}

func (w *Watcher) watch(ctx context.Context, clearBeforePrinting bool) error {
	ns, _, err := w.ConfigFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return fmt.Errorf("read namespace: %w", err)
	}
	if ns == "" {
		return fmt.Errorf("no namespace selected")
	}

	r := resource.NewBuilder(w.ConfigFlags).
		Unstructured().
		NamespaceParam(ns).DefaultNamespace().AllNamespaces(w.AllNamespaces).
		//FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelectorParam(w.LabelSelector).
		FieldSelectorParam(w.FieldSelector).
		//RequestChunksOf(o.ChunkSize).
		ResourceTypeOrNameArgs(true, w.Args...).
		SingleResourceType().
		Latest().
		TransformRequests(transformRequests).
		Do()
	if err := r.Err(); err != nil {
		return err
	}

	infos, err := r.Infos()
	if err != nil {
		return err
	}

	if len(infos) != 1 {
		return fmt.Errorf("expected a single resource info, but got %d", len(infos))
	}
	info := infos[0]
	obj := info.Object
	mapping := info.Mapping

	printNamespace := w.Options.AllNamespaces
	if mapping != nil && mapping.Scope.Name() == meta.RESTScopeNameRoot {
		// Resource isn't namespaced
		printNamespace = false
	}
	w.Printer.Configure(mapping.GroupVersionKind, printNamespace)

	// watching from resourceVersion 0, starts the watch at ~now and
	// will return an initial watch event.  Starting form ~now, rather
	// the resVersion of the object will insure that we start the watch from
	// inside the watch window, which the resVersion of the object might not be.
	resVersion := "0"
	isList := meta.IsListType(obj)
	var objsToPrint []runtime.Object
	if isList {
		// the resourceVersion of list objects is ~now but won't return
		// an initial watch event
		resVersion, err = meta.NewAccessor().ResourceVersion(obj)
		if err != nil {
			return err
		}
		objsToPrint, _ = meta.ExtractList(obj)
	} else {
		objsToPrint = []runtime.Object{obj}
	}

	if clearBeforePrinting {
		w.Printer.Clear()
	}

	for _, objToPrint := range objsToPrint {
		if _, err := w.Printer.PrintObj(objToPrint, watch.Added); err != nil {
			return err
		}
	}

	w.Printer.Table.StopSpinner()

	return w.pipeEvents(ctx, r, resVersion)
}

func (w *Watcher) pipeEvents(ctx context.Context, r *resource.Result, resVersion string) error {
	watch, err := r.Watch(resVersion)
	if err != nil {
		return err
	}
	defer watch.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watch.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}
			cmd, err := w.Printer.PrintObj(event.Object, event.Type)
			if err != nil {
				return err
			}
			w.Program.Send(cmd())
		}
	}
}

type Printer struct {
	Table      *table.Model
	WideOutput bool
	colDefs    []metav1.TableColumnDefinition
	LabelCols  []string

	info           schema.GroupVersionKind
	apiVersion     string
	kind           string
	printNamespace bool
}

func (p *Printer) Configure(info schema.GroupVersionKind, printNamespace bool) {
	p.info = info
	p.apiVersion, p.kind = info.ToAPIVersionAndKind()
	p.printNamespace = printNamespace
}

func (p *Printer) Clear() {
	p.Table.SetRows(nil)
}

func (p *Printer) PrintObj(obj runtime.Object, eventType watch.EventType) (tea.Cmd, error) {
	objTable, err := decodeIntoTable(obj)
	if err != nil {
		return nil, err
	}
	p.updateColDefHeaders(objTable)
	return p.addObjectToTable(objTable, eventType)
}

func (p *Printer) updateColDefHeaders(objTable *metav1.Table) {
	if len(objTable.ColumnDefinitions) == 0 {
		return
	}

	numColumns := len(objTable.ColumnDefinitions)
	if p.printNamespace {
		numColumns++
	}

	headers := make([]string, 0, numColumns)

	if p.printNamespace {
		headers = append(headers, "NAMESPACE")
	}
	for _, colDef := range objTable.ColumnDefinitions {
		if colDef.Priority == 0 || p.WideOutput {
			headers = append(headers, strings.ToUpper(colDef.Name))
		}
	}
	for _, label := range p.LabelCols {
		headers = append(headers, labelColumnHeader(label))
	}
	p.Table.SetHeaders(headers)
	p.colDefs = objTable.ColumnDefinitions
}

func labelColumnHeader(label string) string {
	label = strings.ToUpper(label)
	index := strings.LastIndexByte(label, '/')
	if index == -1 {
		return label
	}
	return label[index+1:]
}

func (p *Printer) addObjectToTable(objTable *metav1.Table, eventType watch.EventType) (tea.Cmd, error) {
	var cmd tea.Cmd
	for _, row := range objTable.Rows {
		unstrucObj, ok := row.Object.Object.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("want *unstructured.Unstructured, got %T", row.Object.Object)
		}
		metadata, ok := unstrucObj.Object["metadata"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("metadata: want map[string]any, got %T", unstrucObj.Object["metadata"])
		}
		uid, ok := metadata["uid"].(string)
		if !ok {
			return nil, fmt.Errorf("metadata.uid: want string, got %T", metadata["uid"])
		}
		name, ok := metadata["name"].(string)
		if !ok {
			return nil, fmt.Errorf("metadata.name: want string, got %T", metadata["name"])
		}
		creationTimestamp, ok := metadata["creationTimestamp"].(string)
		if !ok {
			return nil, fmt.Errorf("metadata.creationTimestamp: want string, got %T", metadata["creationTimestamp"])
		}
		creationTime, err := time.Parse(time.RFC3339, creationTimestamp)
		if err != nil {
			return nil, fmt.Errorf("metadata.creationTimestamp: %w", err)
		}
		tableRow := table.Row{
			ID:        uid,
			Fields:    make([]any, 0, len(p.colDefs)),
			SortField: name,
		}
		if p.apiVersion == "v1" && p.kind == "Event" {
			tableRow.SortField = creationTimestamp
		}
		if p.printNamespace {
			namespace := metadata["namespace"]
			tableRow.Fields = append(tableRow.Fields, namespace)
			tableRow.SortField = fmt.Sprintf("%s/%s", namespace, tableRow.SortField)
		}
		for i, cell := range row.Cells {
			if i >= len(p.colDefs) {
				return nil, fmt.Errorf("cant find index %d (%v) in column defs: %v", i, cell, p.colDefs)
			}
			colDef := p.colDefs[i]
			if colDef.Priority != 0 && !p.WideOutput {
				continue
			}
			tableRow.Fields = append(tableRow.Fields, p.parseCell(cell, row, eventType, unstrucObj.Object, colDef, creationTime))
		}
		for _, label := range p.LabelCols {
			labelValue := unstrucObj.GetLabels()[label]
			tableRow.Fields = append(tableRow.Fields, labelValue)
		}
		switch eventType {
		case watch.Error:
			tableRow.Status = table.StatusError
		case watch.Deleted:
			tableRow.Status = table.StatusDeleted
		}
		// it's fine to only use the latest returned cmd, because of how
		// [table.AddRow] is implemented

		cmd = p.Table.AddRow(tableRow)
	}
	return cmd, nil
}

func (p *Printer) parseCell(cell any, row metav1.TableRow, eventType watch.EventType, object map[string]any, colDef metav1.TableColumnDefinition, creationTime time.Time) any {
	cellStr := fmt.Sprint(cell)
	columnNameLower := strings.ToLower(colDef.Name)
	switch {
	case columnNameLower == "age",
		// some non-namespaced resources (e.g Role) gives timestamp instead of age
		columnNameLower == "created at":
		return creationTime
	case columnNameLower == "status":
		if eventType == watch.Deleted {
			return table.AgoColumn{
				Value: "Deleted",
				Time:  time.Now(),
			}
		}
		return StatusColumn(cellStr)
	case p.apiVersion == "v1" && p.kind == "Event" && columnNameLower == "last seen",
		p.apiVersion == "batch/v1" && p.kind == "CronJob" && columnNameLower == "last schedule":

		dur, ok := parseHumanDuration(cellStr)
		if !ok {
			return cell
		}
		return time.Now().Add(-dur)
	case p.apiVersion == "batch/v1" && p.kind == "Job" && columnNameLower == "duration":
		var completionsCell any
		for i, otherCell := range row.Cells {
			if i >= len(p.colDefs) {
				continue
			}
			def := p.colDefs[i]
			if strings.EqualFold(def.Name, "completions") {
				completionsCell = otherCell
				break
			}
		}
		if completionsCell == nil {
			return cell
		}
		f, ok := ParseFraction(fmt.Sprint(completionsCell))
		if !ok {
			return cell
		}
		if f.Count >= f.Total {
			return cell
		}
		dur, ok := parseHumanDuration(cellStr)
		if !ok {
			return cell
		}
		return time.Now().Add(-dur)
	case p.apiVersion == "v1" && p.kind == "Event" && columnNameLower == "reason":
		return StatusColumn(cellStr)
	case p.apiVersion == "v1" && p.kind == "Pod" && columnNameLower == "restarts":
		// 0, the most common case
		if cellStr == "0" {
			return cell
		}
		countStr, dur, ok := parsePodRestarts(cellStr)
		if ok {
			cell = table.AgoColumn{
				Value: countStr,
				Time:  time.Now().Add(-dur),
			}
		}
		// Only add styling if not deleted, to not add excess coloring
		if eventType != watch.Deleted {
			cell = table.StyledColumn{
				Value: cell,
				Style: StyleFractionWarning,
			}
		}
		return cell
	case p.apiVersion == "storage.k8s.io/v1" && p.kind == "StorageClass" && columnNameLower == "reclaimpolicy":
		return StatusColumn(cellStr)
	// Only parse fraction (e.g "1/2") if the resources was not deleted,
	// so we don't have colored fraction on a grayed-out row.
	case eventType != watch.Deleted:
		fractionStyle, ok := FractionStyle(cellStr)
		if ok {
			cell = table.StyledColumn{
				Value: cell,
				Style: fractionStyle,
			}
		}
		return cell
	default:
		return cell
	}
}

func transformRequests(req *rest.Request) {
	// TODO: Skip if custom column output mode

	//if !o.ServerPrint || !o.IsHumanReadablePrinter {
	//	return
	//}

	req.SetHeader("Accept", strings.Join([]string{
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName),
		"application/json",
	}, ","))
}
