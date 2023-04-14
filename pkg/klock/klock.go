package klock

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func Execute(o Options, resourceType string) error {
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
		ResourceTypeOrNameArgs(true, resourceType).
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
	info := infos[0]
	mapping := info.ResourceMapping()
	printer, err := newPrinter(mapping, o.PrintFlags.Copy(), o.AllNamespaces)
	if err != nil {
		return err
	}

	watch, err := r.Watch("0")
	if err != nil {
		return err
	}

	w := printers.GetNewTabWriter(os.Stdout)

	//t := table.NewModel()
	//p := tea.NewProgram(t)

	//headers := []string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"}
	//t.SetHeaders(headers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				watch.Stop()
				return
			case event := <-watch.ResultChan():
				if err := printObj(printer, event, o.OutputWatchEvents, w); err != nil {
					//p.Quit()
					fmt.Fprintf(os.Stderr, "err: %s\n", err)
					return
				}
				w.Flush()
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
	//return p.Start()
	wg.Wait()
	return nil
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
