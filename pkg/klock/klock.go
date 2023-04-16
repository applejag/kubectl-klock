package klock

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jilleJr/kubectl-klock/pkg/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/get"
)

type Options struct {
	ConfigFlags *genericclioptions.ConfigFlags
	PrintFlags  *get.PrintFlags

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

	watch, err := r.Watch("0")
	if err != nil {
		return err
	}

	t := table.NewModel()
	p := tea.NewProgram(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		var colDefs []metav1.TableColumnDefinition
		for {
			select {
			case <-ctx.Done():
				watch.Stop()
				return
			case event := <-watch.ResultChan():
				objTable, err := decodeIntoTable(event.Object)
				if err != nil {
					p.Quit()
					fmt.Fprintf(os.Stderr, "err: %s\n", err)
					return
				}
				colDefs = updateColDefHeaders(t, colDefs, objTable)
				cmd, err := addObjectToTable(t, colDefs, objTable, event.Type)
				if err != nil {
					p.Quit()
					fmt.Fprintf(os.Stderr, "err: %s\n", err)
					return
				}
				p.Send(cmd)
			}
		}
	}()
	return p.Start()
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
	var cmds []tea.Cmd
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
		cmds = append(cmds, t.AddRow(tableRow))
	}
	return tea.Batch(cmds...), nil
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
