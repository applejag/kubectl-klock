package table

import (
	"github.com/charmbracelet/lipgloss"
)

type RowStyles struct {
	Cell    lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Deleted lipgloss.Style
}

var DefaultRowStyle = RowStyles{
	Cell:    lipgloss.NewStyle(),
	Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
	Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
	Deleted: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
}

type Row struct {
	ID     string
	Fields []string
	Status Status
}

type Status int

const (
	StatusDefault Status = iota
	StatusError
	StatusWarning
	StatusDeleted
)

// SortValue value is the value we use when sorting the list.
func (r Row) SortValue() string {
	return r.Fields[0]
}
