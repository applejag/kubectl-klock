package table

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/typ.v4/slices"
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
	Styles      Styles
	CellSpacing int

	// Key mappings for navigating the list.
	KeyMap KeyMap

	headers      []string
	maxHeight    int
	rows         []Row
	columnWidths []int
	quitting     bool
}

func NewModel() *Model {
	return &Model{
		Styles:      DefaultStyles,
		KeyMap:      DefaultKeyMap,
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

type rowUpdateMsg struct{}

func rowUpdate() tea.Msg {
	return rowUpdateMsg{}
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
	return tea.Batch(m.updateHeight(), rowUpdate)
}

func (m *Model) SetRows(rows []Row) tea.Cmd {
	m.rows = slices.Clone(rows)
	m.sortItems()
	m.updateColumnWidths()
	return tea.Batch(m.updateHeight(), rowUpdate)
}

func (m *Model) updateHeight() tea.Cmd {
	height := len(m.rows) + extraHeight
	shouldShowPagination := height > m.maxHeight
	if shouldShowPagination {
		return tea.EnterAltScreen
	}
	return tea.ExitAltScreen
}

func (m *Model) sortItems() {
	slices.SortStableFunc(m.rows, func(a, b Row) bool {
		return a.SortValue() < b.SortValue()
	})
}

func (m Model) Init() tea.Cmd {
	return doTick()
}

func (m *Model) SetHeaders(headers []string) {
	m.headers = headers
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
		if key.Matches(msg, m.KeyMap.ForceQuit) {
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.maxHeight = msg.Height
	case rowUpdateMsg:
		// TODO: update filter
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
	m.columnsView(&buf, m.headers, lipgloss.Style{})
	for _, row := range m.rows {
		buf.WriteByte('\n')
		m.rowView(&buf, row)
	}

	if m.quitting {
		buf.WriteByte('\n')
	}

	return buf.String()
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
