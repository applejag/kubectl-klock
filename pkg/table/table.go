package table

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/typ.v4"
	"gopkg.in/typ.v4/avl"
)

const extraHeight = 2

type Styles struct {
	TitleBar lipgloss.Style
	Title    lipgloss.Style
	Row      RowStyles
}

var DefaultStyles = Styles{
	TitleBar: lipgloss.NewStyle(),
	Title:    lipgloss.NewStyle(),
	Row:      DefaultRowStyle,
}

type Model struct {
	Styles Styles

	list     list.Model
	delegate *RowDelegate
	headers  []string
	rows     avl.Tree[Row]
}

func NewModel() *Model {
	delegate := &RowDelegate{FieldSpacing: 3}

	l := list.New(nil, delegate, 80, extraHeight)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	return &Model{
		Styles:   DefaultStyles,
		list:     l,
		delegate: delegate,
		headers:  nil,
		rows: avl.New(func(a, b Row) int {
			if len(a.Fields) > 0 && len(b.Fields) > 0 {
				if cmp := typ.Compare(a.Fields[0], b.Fields[0]); cmp != 0 {
					return cmp
				}
			}
			if cmp := typ.Compare(a.Status, b.Status); cmp != 0 {
				return cmp
			}
			return typ.Compare(a.ID, b.ID)
		}),
	}
}

func (m *Model) AddRow(row Row) tea.Cmd {
	m.rows.Add(row)
	cmd := m.list.SetItems(m.itemsFromTree())
	m.list.SetHeight(len(m.list.Items()) + extraHeight)
	m.delegate.UpdateFieldMaxLengths(m.list.Items(), m.headers)
	return cmd
}

func (m *Model) SetRows(rows []Row) tea.Cmd {
	for _, row := range rows {
		m.rows.Add(row)
	}
	cmd := m.list.SetItems(m.itemsFromTree())
	m.list.SetHeight(len(m.list.Items()) + extraHeight)
	m.delegate.UpdateFieldMaxLengths(m.list.Items(), m.headers)
	return cmd
}

func (m *Model) itemsFromTree() []list.Item {
	slice := make([]list.Item, 0, m.rows.Len())
	m.rows.WalkInOrder(func(value Row) {
		slice = append(slice, value)
	})
	return slice
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *Model) SetHeaders(headers []string) {
	m.headers = headers
	m.delegate.UpdateFieldMaxLengths(m.list.Items(), headers)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var title strings.Builder
	m.delegate.Styles = m.Styles.Row
	m.delegate.RenderColumns(&title, m.headers, lipgloss.Style{})
	m.list.Title = title.String()
	m.list.Styles.Title = m.Styles.Title
	m.list.Styles.TitleBar = m.Styles.TitleBar
	return m.list.View()
}
