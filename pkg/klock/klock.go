package klock

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jilleJr/kubectl-klock/pkg/table"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Execute(configFlags *genericclioptions.ConfigFlags) error {
	p := tea.NewProgram(table.NewModel())
	return p.Start()
}
