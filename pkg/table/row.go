package table

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"k8s.io/apimachinery/pkg/util/duration"
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
	Fields []any
	Status Status

	renderedFields []string
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
	if len(r.Fields) == 0 {
		return ""
	}
	str, ok := r.Fields[0].(string)
	if !ok {
		return ""
	}
	return str
}

func (r *Row) RenderedFields() []string {
	if len(r.renderedFields) != len(r.Fields) {
		r.ReRenderFields()
	}
	return r.renderedFields
}

func (r *Row) ReRenderFields() {
	r.renderedFields = resizeSlice(r.renderedFields, len(r.Fields))
	for i, col := range r.Fields {
		r.renderedFields[i] = renderColumn(col)
	}
}

func renderColumn(value any) string {

	switch value := value.(type) {
	case string:
		return value
	case time.Time:
		dur := time.Since(value)
		return duration.HumanDuration(dur)
	default:
		return fmt.Sprint(value)
	}
}
