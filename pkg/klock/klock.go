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

	LabelSelector     string
	FieldSelector     string
	AllNamespaces     bool
	OutputWatchEvents bool
}

func Execute(o Options, args []string) error {
	ns, _, err := o.ConfigFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return fmt.Errorf("read namespace: %w", err)
	}
	if ns == "" {
		return fmt.Errorf("no namespace selected")
	}

	r := resource.NewBuilder(o.ConfigFlags).
		Unstructured().
		NamespaceParam(ns).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		//FilenameParam(o.ExplicitNamespace, &o.FilenameOptions).
		LabelSelectorParam(o.LabelSelector).
		FieldSelectorParam(o.FieldSelector).
		//RequestChunksOf(o.ChunkSize).
		ResourceTypeOrNameArgs(true, args...).
		SingleResourceType().
		Latest().
		TransformRequests(transformRequests).
		Do()
	if err := r.Err(); err != nil {
		return err
	}

	t := table.New()
	p := tea.NewProgram(t)
	go p.Send(t.StartSpinner()())
	printer := Printer{Table: t}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watchAndPrint := func() error {
		obj, err := r.Object()
		if err != nil {
			return err
		}

		// watching from resourceVersion 0, starts the watch at ~now and
		// will return an initial watch event.  Starting form ~now, rather
		// the rv of the object will insure that we start the watch from
		// inside the watch window, which the rv of the object might not be.
		rv := "0"
		isList := meta.IsListType(obj)
		var objsToPrint []runtime.Object
		if isList {
			// the resourceVersion of list objects is ~now but won't return
			// an initial watch event
			rv, err = meta.NewAccessor().ResourceVersion(obj)
			if err != nil {
				return err
			}
			objsToPrint, _ = meta.ExtractList(obj)
		} else {
			objsToPrint = []runtime.Object{obj}
		}

		var cmd tea.Cmd
		for _, objToPrint := range objsToPrint {
			var err error
			cmd, err = printer.PrintObj(objToPrint, watch.Added)
			if err != nil {
				return err
			}
		}
		t.StopSpinner()
		p.Send(cmd())

		watch, err := r.Watch(rv)
		if err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				watch.Stop()
				return nil
			case event := <-watch.ResultChan():
				cmd, err := printer.PrintObj(event.Object, event.Type)
				if err != nil {
					return err
				}
				p.Send(cmd())
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

type Printer struct {
	Table   *table.Model
	colDefs []metav1.TableColumnDefinition
}

func (p *Printer) PrintObj(obj runtime.Object, eventType watch.EventType) (tea.Cmd, error) {
	objTable, err := decodeIntoTable(obj)
	if err != nil {
		return nil, err
	}
	p.colDefs = updateColDefHeaders(p.Table, p.colDefs, objTable)
	return addObjectToTable(p.Table, p.colDefs, objTable, eventType)
}

func updateColDefHeaders(t *table.Model, oldColDefs []metav1.TableColumnDefinition, objTable *metav1.Table) []metav1.TableColumnDefinition {
	if len(objTable.ColumnDefinitions) == 0 {
		return oldColDefs
	}

	headers := make([]string, 0, len(objTable.ColumnDefinitions))
	for _, colDef := range objTable.ColumnDefinitions {
		if colDef.Priority == 0 {
			headers = append(headers, strings.ToUpper(colDef.Name))
		}
	}
	t.SetHeaders(headers)
	return objTable.ColumnDefinitions
}

func addObjectToTable(t *table.Model, colDefs []metav1.TableColumnDefinition, objTable *metav1.Table, eventType watch.EventType) (tea.Cmd, error) {
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
			Fields: make([]any, 0, len(colDefs)),
		}
		for i, cell := range row.Cells {
			if i >= len(colDefs) {
				return nil, fmt.Errorf("cant find index %d (%v) in column defs: %v", i, cell, colDefs)
			}
			colDef := colDefs[i]
			if colDef.Priority != 0 {
				continue
			}
			cellStr := fmt.Sprint(cell)
			if colDef.Name == "Status" {
				status := ParseStatus(cellStr)
				switch status {
				case StatusError:
					tableRow.Status = table.StatusError
				case StatusWarning:
					tableRow.Status = table.StatusWarning
				}
				if eventType == watch.Deleted {
					cellStr = "Deleted"
				}
			}
			if colDef.Name == "Age" {
				tableRow.Fields = append(tableRow.Fields, creationTime)
			} else {
				tableRow.Fields = append(tableRow.Fields, cellStr)
			}
		}
		if eventType == watch.Error {
			tableRow.Status = table.StatusError
		}
		if eventType == watch.Deleted {
			tableRow.Status = table.StatusDeleted
		}
		// it's fine to only use the latest returned cmd, because of how
		// [table.AddRow] is implemented
		cmd = t.AddRow(tableRow)
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
