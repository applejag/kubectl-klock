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
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jilleJr/kubectl-klock/pkg/table"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
)

type Options struct {
	ConfigFlags *genericclioptions.ConfigFlags

	LabelSelector string
	FieldSelector string
	AllNamespaces bool

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
	configLoader := o.ConfigFlags.ToRawKubeConfigLoader()
	//kubeconfigFiles := configLoader.ConfigAccess().GetLoadingPrecedence()

	ns, _, err := configLoader.Namespace()
	if err != nil {
		return fmt.Errorf("read namespace: %w", err)
	}
	if ns == "" {
		return fmt.Errorf("no namespace selected")
	}

	t := table.New()
	printer := Printer{Table: t, WideOutput: o.Output == "wide"}
	w := NewWatcher(o, ns, printer)
	p := tea.NewProgram(t)
	t.StartSpinner()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watchAndPrint := func() error {
		if err := w.Watch(ctx, args...); err != nil {
			return err
		}
		for {
			select {
			case event, ok := <-w.EventsChan():
				if !ok {
					return nil
				}
				cmd, err := printer.PrintObj(event.Object, event.Type)
				if err != nil {
					return err
				}
				p.Send(cmd())
			case err := <-w.ErrorChan():
				t.SetError(err)
				p.Send(nil)
			}
		}
	}

	go func() {
		if err := watchAndPrint(); err != nil {
			p.Quit()
			fmt.Fprintf(os.Stderr, "err: %s\n", err)
		}
	}()

	_, err = p.Run()
	return err
}

func NewWatcher(options Options, namespace string, printer Printer) *Watcher {
	return &Watcher{
		Options:   options,
		Namespace: namespace,
		Printer:   printer,

		eventChan: make(chan watch.Event, 3),
		errorChan: make(chan error, 3),
	}
}

type Watcher struct {
	Options
	Namespace string
	Printer   Printer

	eventChan chan watch.Event
	errorChan chan error
}

func (w *Watcher) EventsChan() <-chan watch.Event {
	return w.eventChan
}

func (w *Watcher) ErrorChan() <-chan error {
	return w.errorChan
}

func (w *Watcher) Watch(ctx context.Context, args ...string) error {
	r := resource.NewBuilder(w.ConfigFlags).
		Unstructured().
		NamespaceParam(w.Namespace).DefaultNamespace().AllNamespaces(w.AllNamespaces).
		//FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelectorParam(w.LabelSelector).
		FieldSelectorParam(w.FieldSelector).
		//RequestChunksOf(o.ChunkSize).
		ResourceTypeOrNameArgs(true, args...).
		SingleResourceType().
		Latest().
		TransformRequests(transformRequests).
		Do()
	if err := r.Err(); err != nil {
		return err
	}

	obj, err := r.Object()
	if err != nil {
		return err
	}

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

	for _, objToPrint := range objsToPrint {
		if _, err := w.Printer.PrintObj(objToPrint, watch.Added); err != nil {
			return err
		}
	}

	go w.watchLoop(ctx, r, resVersion)
	return nil
}

func (w *Watcher) watchLoop(ctx context.Context, r *resource.Result, resVersion string) {
	for {
		err := w.pipeEvents(ctx, r, resVersion)
		if ctx.Err() != nil {
			close(w.eventChan)
			return
		}
		if err != nil {
			w.errorChan <- fmt.Errorf("retry in 5s: %w", err)
		}
		time.Sleep(5 * time.Second)
	}
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
			w.eventChan <- event
		}
	}
}

type Printer struct {
	Table      *table.Model
	WideOutput bool
	colDefs    []metav1.TableColumnDefinition
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

	headers := make([]string, 0, len(objTable.ColumnDefinitions))
	for _, colDef := range objTable.ColumnDefinitions {
		if colDef.Priority == 0 || p.WideOutput {
			headers = append(headers, strings.ToUpper(colDef.Name))
		}
	}
	p.Table.SetHeaders(headers)
	p.colDefs = objTable.ColumnDefinitions
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
		creationTimestamp, ok := metadata["creationTimestamp"].(string)
		if !ok {
			return nil, fmt.Errorf("metadata.creationTimestamp: want string, got %T", metadata["creationTimestamp"])
		}
		creationTime, err := time.Parse(time.RFC3339, creationTimestamp)
		if err != nil {
			return nil, fmt.Errorf("metadata.creationTimestamp: %w", err)
		}
		tableRow := table.Row{
			ID:     uid,
			Fields: make([]any, 0, len(p.colDefs)),
		}
		for i, cell := range row.Cells {
			if i >= len(p.colDefs) {
				return nil, fmt.Errorf("cant find index %d (%v) in column defs: %v", i, cell, p.colDefs)
			}
			colDef := p.colDefs[i]
			if colDef.Priority != 0 && !p.WideOutput {
				continue
			}
			cellStr := fmt.Sprint(cell)
			switch strings.ToLower(colDef.Name) {
			case "age":
				tableRow.Fields = append(tableRow.Fields, creationTime)
			case "status":
				if eventType == watch.Deleted {
					cell = "Deleted"
				} else {
					style := ParseStatusStyle(cellStr)
					cell = table.StyledColumn{
						Value: cell,
						Style: style,
					}
				}
				tableRow.Fields = append(tableRow.Fields, cell)
			case "restarts":
				if eventType != watch.Deleted && cellStr != "0" {
					cell = table.StyledColumn{
						Value: cell,
						Style: StyleFractionWarning,
					}
				}
				tableRow.Fields = append(tableRow.Fields, cell)
			default:
				if eventType != watch.Deleted {
					fractionStyle, ok := ParseFractionStyle(cellStr)
					if ok {
						cell = table.StyledColumn{
							Value: cell,
							Style: fractionStyle,
						}
					}
				}
				tableRow.Fields = append(tableRow.Fields, cell)
			}
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
