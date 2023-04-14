package klock

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jilleJr/kubectl-klock/pkg/table"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Execute(configFlags *genericclioptions.ConfigFlags) error {
	t := table.NewModel()
	p := tea.NewProgram(t)

	var rows = []table.Row{
		{ID: "pod-1", Fields: []string{"pod-1", "1/1", "Running", "0", "0s"}},
		{ID: "pod-2", Fields: []string{"pod-2", "1/1", "Running", "10", "0s"}},
		{ID: "pod-3", Fields: []string{"pod-3", "0/1", "Error", "5", "15m30s"}, Status: table.StatusError},
	}
	t.SetRows(rows)

	headers := []string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"}
	t.SetHeaders(headers)

	go func() {
		time.Sleep(1 * time.Second)
		p.Send(t.AddRow(table.Row{
			ID:     "pod-4",
			Fields: []string{"pod-4", "0/0", "Deleted", "0", "5h"},
			Status: table.StatusDeleted,
		}))

		time.Sleep(1 * time.Second)
		p.Send(t.AddRow(table.Row{
			ID:     "foobar",
			Fields: []string{"pod-5", "0/1", "ContainerCreating", "0", "0s"},
		}))
	}()
	return p.Start()
}
