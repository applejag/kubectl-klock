package table

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/typ.v4"
	"gopkg.in/typ.v4/avl"
)

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

func NewModel() Model {
	var listItems = []list.Item{
		Row{Fields: []string{"pod-1", "1/1", "Running", "0", "0s"}},
		Row{Fields: []string{"pod-2", "1/1", "Running", "10", "0s"}},
		Row{Fields: []string{"pod-2", "1/1", "Error", "0", "0s"}, Deleted: true},
	}
	delegate := &RowDelegate{FieldSpacing: 3}
	headers := []string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"}
	delegate.UpdateFieldMaxLengths(listItems, headers)

	l := list.New(listItems, delegate, 80, len(listItems)+2)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	return Model{
		Styles:   DefaultStyles,
		list:     l,
		delegate: delegate,
		headers:  headers,
		rows: avl.New(func(a, b Row) int {
			return typ.Compare(a.ID, b.ID)
		}),
	}
}

func (m *Model) AddRow(row Row) {
	var listItems = []list.Item{
		Row{Fields: []string{"pod-1", "1/1", "Running", "0", "0s"}},
		Row{Fields: []string{"pod-2", "1/1", "Running", "10", "0s"}},
		Row{Fields: []string{"pod-2", "1/1", "Error", "0", "0s"}, Deleted: true},
	}
	m.delegate.UpdateFieldMaxLengths(listItems, m.headers)
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *Model) SetTitle(headers []string) {
	m.headers = headers
	m.delegate.UpdateFieldMaxLengths(m.list.Items(), headers)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
