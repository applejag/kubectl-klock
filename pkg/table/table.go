package table

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/typ.v4/slices"
)

type Styles struct {
	TitleBar lipgloss.Style
	Title    lipgloss.Style
	Row      RowStyles

	Pagination lipgloss.Style
}

var verySubduedColor = lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
var subduedColor = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

var DefaultStyles = Styles{
	TitleBar: lipgloss.NewStyle(),
	Title:    lipgloss.NewStyle(),
	Row:      DefaultRowStyle,

	Pagination: lipgloss.NewStyle().
		Foreground(subduedColor).
		SetString("PAGE:"),
}

type Model struct {
	Styles      Styles
	CellSpacing int

	// Key mappings for navigating the list.
	KeyMap KeyMap

	Paginator paginator.Model

	headers      []string
	maxHeight    int
	rows         []Row
	columnWidths []int
	quitting     bool
}

const (
	ellipsis = "…"
)

func NewModel() *Model {

	p := paginator.NewModel()

	return &Model{
		Styles:      DefaultStyles,
		KeyMap:      DefaultKeyMap,
		Paginator:   p,
		CellSpacing: 3,
		headers:     nil,
		maxHeight:   30,
		rows:        nil,
	}
}

func (m *Model) RowIndex(id string) int {
	for i, row := range m.rows {
		if row.ID == id {
			return i
		}
	}
	return -1
}

func (m *Model) AddRow(row Row) tea.Cmd {
	index := m.RowIndex(row.ID)
	if index == -1 {
		m.rows = append(m.rows, row)
	} else {
		m.rows[index] = row
	}

	m.sortItems()
	m.updateColumnWidths()
	m.updatePagination()
	return m.updateHeight()
}

func (m *Model) SetRows(rows []Row) tea.Cmd {
	m.rows = slices.Clone(rows)
	m.sortItems()
	m.updateColumnWidths()
	m.updatePagination()
	return m.updateHeight()
}

func (m *Model) updateHeight() tea.Cmd {
	if m.paginatorVisible() {
		return tea.EnterAltScreen
	}
	return tea.ExitAltScreen
}

func (m *Model) paginatorVisible() bool {
	if m.maxHeight <= 2 {
		return false
	}
	height := len(m.rows) + 1 // +1 for header
	return height > m.maxHeight
}

func (m *Model) sortItems() {
	slices.SortStableFunc(m.rows, func(a, b Row) bool {
		return a.SortValue() < b.SortValue()
	})
}

func (m *Model) updatePagination() {
	perPage := m.maxHeight - 2 // 1 for header & 1 for paginator
	if perPage < 1 {
		perPage = 1
	}
	m.Paginator.PerPage = perPage
	m.Paginator.SetTotalPages(len(m.rows))
}

func (m Model) Init() tea.Cmd {
	return doTick()
}

func (m *Model) SetHeaders(headers []string) {
	m.headers = headers
	m.updateColumnWidths()
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.ForceQuit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.PrevPage):
			m.Paginator.PrevPage()
			return m, nil
		case key.Matches(msg, m.KeyMap.NextPage):
			m.Paginator.NextPage()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.maxHeight = msg.Height
		m.updatePagination()
	case TickMsg:
		for i := range m.rows {
			m.rows[i].ReRenderFields()
		}
		m.updateColumnWidths()
		return m, doTick()
	}
	return m, nil
}

func (m Model) View() string {
	if len(m.rows) == 0 {
		return "No resources found"
	}
	var buf bytes.Buffer
	if m.maxHeight > 1 {
		m.columnsView(&buf, m.headers, lipgloss.Style{})
		buf.WriteByte('\n')
	}

	currentPage := m.currentPaginatedPage()
	for i, row := range currentPage {
		if i > 0 {
			buf.WriteByte('\n')
		}
		m.rowView(&buf, row)
	}

	if m.paginatorVisible() {
		for i := len(currentPage); i < m.Paginator.PerPage; i++ {
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
		buf.WriteString(m.Styles.Pagination.Render(m.Paginator.View()))
	}

	if m.quitting {
		buf.WriteByte('\n')
	}

	return buf.String()
}

func (m Model) currentPaginatedPage() []Row {
	if len(m.rows) == 0 {
		return nil
	}
	start, end := m.Paginator.GetSliceBounds(len(m.rows))
	return m.rows[start:end]
}

func (m Model) rowView(w io.Writer, row Row) {
	style := m.Styles.Row.Cell
	switch row.Status {
	case StatusError:
		style = m.Styles.Row.Error
	case StatusWarning:
		style = m.Styles.Row.Warning
	case StatusDeleted:
		style = m.Styles.Row.Deleted
	}
	m.columnsView(w, row.RenderedFields(), style)
}

var lotsOfSpaces = strings.Repeat(" ", 200)

func (m Model) columnsView(w io.Writer, columns []string, style lipgloss.Style) {
	for i, col := range columns {
		if i > 0 {
			//TODO: test style.Width()
			spacing := m.CellSpacing + m.columnWidths[i-1] - len(columns[i-1])
			if spacing > 0 {
				fmt.Fprint(w, lotsOfSpaces[:spacing])
			}
		}
		fmt.Fprintf(w, style.Render(col))
	}
}

func (m *Model) updateColumnWidths() {
	lengths := expandToMaxLengths(nil, m.headers)
	for _, row := range m.rows {
		lengths = expandToMaxLengths(lengths, row.RenderedFields())
	}
	m.columnWidths = lengths
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

func resizeSlice[S ~[]E, E any](slice S, targetLen int) S {
	slice = expandSlice(slice, targetLen)
	return slice[:targetLen]
}

func expandSlice[S ~[]E, E any](slice S, minLen int) S {
	delta := minLen - len(slice)
	if delta <= 0 {
		return slice
	}
	return append(slice, make(S, delta)...)
}
