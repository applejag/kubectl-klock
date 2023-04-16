package table

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	}
}

func (m *Model) AddRow(row Row) tea.Cmd {
	var cmd tea.Cmd
	var modified bool
	for i, listItem := range m.list.Items() {
		item, ok := listItem.(Row)
		if ok && item.ID == row.ID {
			cmd = m.list.SetItem(i, row)
			modified = true
			break
		}
	}

	if !modified {
		cmd = m.list.InsertItem(len(m.list.Items()), row)
	}
	m.sortItems()
	m.list.SetHeight(len(m.list.Items()) + extraHeight)
	m.delegate.UpdateFieldMaxLengths(m.list.Items(), m.headers)
	return cmd
}

func (m *Model) sortItems() {
	items := m.list.Items()
	sort.Slice(items, func(i, j int) bool {
		a := items[i].(Row)
		b := items[j].(Row)
		return a.Fields[0] < b.Fields[0]
	})
}

func (m *Model) SetRows(rows []Row) tea.Cmd {
	items := make([]list.Item, len(rows))
	for i, row := range rows {
		items[i] = row
	}
	cmd := m.list.SetItems(items)
	m.sortItems()
	m.list.SetHeight(len(m.list.Items()) + extraHeight)
	m.delegate.UpdateFieldMaxLengths(m.list.Items(), m.headers)
	return cmd
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
	}
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
