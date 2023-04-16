package table

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const extraHeight = 4

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

	list      list.Model
	delegate  *RowDelegate
	headers   []string
	maxHeight int
}

func NewModel() *Model {
	delegate := &RowDelegate{FieldSpacing: 3}

	l := list.New(nil, delegate, 80, 30)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	return &Model{
		Styles:    DefaultStyles,
		list:      l,
		delegate:  delegate,
		headers:   nil,
		maxHeight: 30,
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
	return tea.Batch(cmd, m.updateHeight())
}

func (m *Model) SetRows(rows []Row) tea.Cmd {
	items := make([]list.Item, len(rows))
	for i, row := range rows {
		items[i] = row
	}
	cmd := m.list.SetItems(items)
	m.sortItems()
	return tea.Batch(cmd, m.updateHeight())
}

func (m *Model) updateHeight() tea.Cmd {
	height := len(m.list.Items()) + extraHeight
	if height > m.maxHeight {
		height = m.maxHeight
	}
	m.list.SetHeight(height)
	shouldShowPagination := height > m.maxHeight
	if m.list.ShowPagination() != shouldShowPagination {
		m.list.SetShowPagination(shouldShowPagination)
		if shouldShowPagination {
			return tea.EnterAltScreen
		}
		return tea.ExitAltScreen
	}
	return nil
}

func (m *Model) sortItems() {
	items := m.list.Items()
	sort.Slice(items, func(i, j int) bool {
		return items[i].FilterValue() < items[j].FilterValue()
	})
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *Model) SetHeaders(headers []string) {
	m.headers = headers
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.maxHeight = msg.Height
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	m.delegate.UpdateFieldMaxLengths(m.populatedViewItems(), m.headers)
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

func (m Model) populatedViewItems() []list.Item {
	items := m.list.VisibleItems() // more like, "filtered items"
	if len(items) > 0 {
		start, end := m.list.Paginator.GetSliceBounds(len(items))
		return items[start:end]
	}
	return items
}
