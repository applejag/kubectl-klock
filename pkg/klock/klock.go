package klock

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jilleJr/kubectl-klock/pkg/table"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/scheme"
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
	//infos, err := r.Infos()
	//if err != nil {
	//	return err
	//}
	//info := infos[0]
	//mapping := info.ResourceMapping()
	//printer, err := newPrinter(mapping, o.PrintFlags.Copy(), o.AllNamespaces)
	//if err != nil {
	//	return err
	//}

	watch, err := r.Watch("0")
	if err != nil {
		return err
	}

	//w := printers.GetNewTabWriter(os.Stdout)

	t := table.NewModel()
	p := tea.NewProgram(t)

	//headers := []string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"}
	//t.SetHeaders(headers)

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
				//if err := printObj(printer, event, o.OutputWatchEvents, w); err != nil {
				//	//p.Quit()
				//	fmt.Fprintf(os.Stderr, "err: %s\n", err)
				//	return
				//}
				//w.Flush()
			}
		}
		//for podEvent := range watch.ResultChan() {
		//pod, ok := podEvent.Object.(*v1.Pod)
		//if !ok {
		//	continue
		//}

		//var ready int
		//var count int
		//var restarts int
		//status := table.StatusDefault
		//statusText := ""
		//for _, container := range pod.Status.ContainerStatuses {
		//	count++
		//	restarts += int(container.RestartCount)
		//	if container.Ready {
		//		ready++
		//	}
		//}
		//var buf bytes.Buffer
		////if err := printers.NewTypeSetter(scheme.Scheme).WrapToPrinter(delegate printers.ResourcePrinter, err error).PrintObj(pod, &buf); err != nil {
		////	fmt.Fprintf(&buf, "\nerror printing: %s\n", err)
		////}
		//fmt.Println(buf.String())
		//time.Sleep(1)
		//if pod.DeletionTimestamp != nil {
		//	status = table.StatusError
		//	statusText = "Terminating"
		//}
		//if podEvent.Type == watch.Deleted {
		//	status = table.StatusDeleted
		//	statusText = "Deleted"
		//}
		//p.Send(t.AddRow(table.Row{
		//	ID: string(pod.UID),
		//	Fields: []string{
		//		pod.Name,
		//		fmt.Sprintf("%d/%d", ready, count),
		//		statusText,
		//		strconv.Itoa(restarts),
		//		time.Since(pod.CreationTimestamp.Time).Truncate(time.Second).String(),
		//	},
		//	Status: status,
		//}))
		//}
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
		tableRow := table.Row{
			ID:     uid,
			Fields: make([]string, 0, len(colDefs)),
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
			}
			tableRow.Fields = append(tableRow.Fields, cellStr)
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

func newPrinter(mapping *meta.RESTMapping, printFlags get.PrintFlags, withNamespace bool) (printers.ResourcePrinter, error) {
	fmt.Println("mapping:", mapping)
	if mapping != nil {
		printFlags.SetKind(mapping.GroupVersionKind.GroupKind())
	}
	printer, err := printFlags.ToPrinter()
	if err != nil {
		return nil, err
	}
	if withNamespace {
		printFlags.EnsureWithNamespace()
	}
	printer, err = printers.NewTypeSetter(scheme.Scheme).WrapToPrinter(printer, nil)
	if err != nil {
		return nil, err
	}
	printer = &get.TablePrinter{Delegate: printer}
	return printer, nil
}

func printObj(printer printers.ResourcePrinter, event watch.Event, outputWatchEvents bool, w io.Writer) error {
	objToPrint := event.Object
	if outputWatchEvents {
		objToPrint = &metav1.WatchEvent{Type: string(event.Type), Object: runtime.RawExtension{Object: event.Object}}
	}
	if err := printer.PrintObj(objToPrint, w); err != nil {
		return err
	}
	return nil
}

func transformRequests(req *rest.Request) {
	//if !o.ServerPrint || !o.IsHumanReadablePrinter {
	//	return
	//}

	req.SetHeader("Accept", strings.Join([]string{
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName),
		"application/json",
	}, ","))
}
