package table

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RowStyles struct {
	Cell    lipgloss.Style
	Error   lipgloss.Style
	Deleted lipgloss.Style
}

var DefaultRowStyle = RowStyles{
	Cell:    lipgloss.NewStyle(),
	Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
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
	StatusDeleted
)

// ensure [Row] implements the interface.
var _ list.Item = Row{}

// Filter value is the value we use when filtering against this item when
// we're filtering the list.
func (i Row) FilterValue() string {
	return i.Fields[0]
}

type RowDelegate struct {
	FieldSpacing    int
	FieldMaxLengths []int
	Styles          RowStyles
}

// ensure [RowDelegate] implements the interface.
var _ list.ItemDelegate = RowDelegate{}

func (d *RowDelegate) UpdateFieldMaxLengths(listItems []list.Item, headers []string) {
	lengths := expandToMaxLengths(nil, headers)
	for _, listItem := range listItems {
		item, ok := listItem.(Row)
		if !ok {
			continue
		}
		lengths = expandToMaxLengths(lengths, item.Fields)
	}
	d.FieldMaxLengths = lengths
}

func expandToMaxLengths(lengths []int, strs []string) []int {
	lengths = expandSlice(lengths, len(strs))
	for i, f := range strs {
		if len(f) > lengths[i] {
			lengths[i] = len(f)
		}
	}
	return lengths
}

func expandSlice[S ~[]E, E any](slice S, minLen int) S {
	delta := minLen - len(slice)
	if delta <= 0 {
		return slice
	}
	return append(slice, make(S, delta)...)
}

// Render renders the item's view.
func (d RowDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Row)
	if !ok {
		return
	}
	switch item.Status {
	case StatusError:
		d.RenderColumns(w, item.Fields, d.Styles.Error)
	case StatusDeleted:
		d.RenderColumns(w, item.Fields, d.Styles.Deleted)
	default:
		d.RenderColumns(w, item.Fields, d.Styles.Cell)
	}
}

var lotsOfSpaces = strings.Repeat(" ", 200)

func (d RowDelegate) RenderColumns(w io.Writer, columns []string, style lipgloss.Style) {
	for i, f := range columns {
		if i > 0 {
			//TODO: test style.Width()
			spacing := d.FieldSpacing + d.FieldMaxLengths[i-1] - len(columns[i-1])
			if spacing > 0 {
				fmt.Fprint(w, lotsOfSpaces[:spacing])
			}
		}
		fmt.Fprintf(w, style.Render(f))
	}
}

// Height is the height of the list item.
func (RowDelegate) Height() int {
	return 1
}

// Spacing is the size of the horizontal gap between list items in cells.
func (RowDelegate) Spacing() int {
	return 0
}

// Update is the update loop for items. All messages in the list's update
// loop will pass through here except when the user is setting a filter.
// Use this method to perform item-level updates appropriate to this
// delegate.
func (RowDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
