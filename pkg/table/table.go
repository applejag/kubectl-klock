// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package table

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/ansi"
)

type Styles struct {
	TitleBar lipgloss.Style
	Title    lipgloss.Style
	Row      RowStyles

	Error      lipgloss.Style
	Pagination lipgloss.Style
}

var subduedColor = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

var DefaultStyles = Styles{
	TitleBar: lipgloss.NewStyle(),
	Title:    lipgloss.NewStyle(),
	Row:      DefaultRowStyle,

	Error: lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(9)).
		SetString("ERROR:"),
	Pagination: lipgloss.NewStyle().
		Foreground(subduedColor).
		SetString("PAGE:"),
}

type Model struct {
	Styles      Styles
	CellSpacing int
	HideDeleted bool
	ShowHelp    bool

	// Key mappings for navigating the list.
	KeyMap KeyMap

	help        help.Model
	Paginator   paginator.Model
	spinner     spinner.Model
	filterInput textinput.Model
	showSpinner bool

	err                error
	headers            []string
	maxHeight          int
	rows               []Row
	filteredRows       []Row
	columnWidths       []int
	fullscreenOverride bool
	quitting           bool

	filterInputEnabled bool
}

func New() *Model {
	return &Model{
		Styles:      DefaultStyles,
		KeyMap:      DefaultKeyMap,
		Paginator:   paginator.New(),
		CellSpacing: 3,

		help:    help.New(),
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),

		filterInput:        textinput.New(),
		filterInputEnabled: false,

		headers:   nil,
		maxHeight: 30,
		rows:      nil,
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
	m.StopSpinner()
	m.updateFilteredRows()
	m.updateColumnWidths()
	m.updatePagination()
	fullscreenCmd := m.updateFullscreenCmd()
	return fullscreenCmd
}

func (m *Model) SetRows(rows []Row) tea.Cmd {
	m.rows = slices.Clone(rows)
	m.sortItems()
	if len(m.rows) > 0 {
		m.StopSpinner()
	}
	m.updateFilteredRows()
	m.updateColumnWidths()
	m.updatePagination()
	return m.updateFullscreenCmd()
}

func (m *Model) updateFilteredRows() {
	rows := m.rows
	if filterText := m.filterText(); filterText != "" {
		rows = make([]Row, 0)
		for _, row := range m.rows {
			for _, field := range row.RenderedFields() {
				if strings.Contains(field, filterText) {
					rows = append(rows, row)
					break
				}
			}
		}
	}

	if !m.HideDeleted {
		m.filteredRows = rows
		return
	}
	m.filteredRows = make([]Row, 0, len(rows))
	for _, row := range rows {
		if m.HideDeleted && row.Status == StatusDeleted {
			continue
		}
		m.filteredRows = append(m.filteredRows, row)
	}
}

func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) updateFullscreenCmd() tea.Cmd {
	if m.fullscreenOverride || m.windowTooShort() {
		return tea.EnterAltScreen
	}
	return tea.ExitAltScreen
}

func (m *Model) paginatorVisible() bool {
	if m.maxHeight <= 2 {
		return false
	}
	return m.windowTooShort()
}

func (m *Model) windowTooShort() bool {
	height := len(m.filteredRows) + 1 // +1 for header
	if m.err != nil {
		height++
	}
	return height > m.maxHeight
}

func (m *Model) sortItems() {
	slices.SortStableFunc(m.rows, func(a, b Row) int {
		return cmp.Compare(a.SortValue(), b.SortValue())
	})
}

func (m *Model) updatePagination() {
	perPage := m.maxHeight - 2 // 1 for header & 1 for paginator
	if perPage < 1 {
		perPage = 1
	}
	m.Paginator.PerPage = perPage
	m.Paginator.SetTotalPages(len(m.filteredRows))

	// Make sure the page stays in bounds
	if m.Paginator.Page >= m.Paginator.TotalPages-1 {
		m.Paginator.Page = m.Paginator.TotalPages - 1
	}
}

func (m Model) Init() tea.Cmd {
	cmd := doTick()
	if m.showSpinner {
		cmd = tea.Batch(cmd, m.spinner.Tick)
	}
	return cmd
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

func (m *Model) StartSpinner() tea.Cmd {
	if m.showSpinner {
		return nil
	}
	m.showSpinner = true
	return m.spinner.Tick
}

func (m *Model) StopSpinner() {
	m.showSpinner = false
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterInputEnabled && !key.Matches(msg, m.KeyMap.ForceQuit, m.KeyMap.Filter) {
			i, cmd := m.filterInput.Update(msg)
			m.filterInput = i
			m.updateFilteredRows()
			m.updatePagination()
			m.updateColumnWidths()
			return m, cmd
		}
		switch {
		case key.Matches(msg, m.KeyMap.ForceQuit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.PrevPage):
			m.Paginator.PrevPage()
			m.updateColumnWidths()
			return m, nil
		case key.Matches(msg, m.KeyMap.NextPage):
			m.Paginator.NextPage()
			m.updateColumnWidths()
			return m, nil
		case key.Matches(msg, m.KeyMap.ToggleDeleted):
			m.HideDeleted = !m.HideDeleted
			m.updateFilteredRows()
			m.updatePagination()
			m.updateColumnWidths()
			return m, nil
		case key.Matches(msg, m.KeyMap.ToggleFullscreen):
			m.fullscreenOverride = !m.fullscreenOverride
			return m, m.updateFullscreenCmd()
		case !m.ShowHelp && key.Matches(msg, m.KeyMap.ShowFullHelp):
			m.ShowHelp = true
			return m, nil
		case m.ShowHelp && key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.ShowHelp = false
			return m, nil
		case key.Matches(msg, m.KeyMap.Filter):
			m.filterInputEnabled = !m.filterInputEnabled
			m.updateFilteredRows()
			m.updatePagination()
			m.updateColumnWidths()
			var cmd tea.Cmd
			if m.filterInputEnabled {
				cmd = m.filterInput.Focus()
			}
			return m, cmd
		}

	case spinner.TickMsg:
		s, cmd := m.spinner.Update(msg)
		m.spinner = s
		if m.showSpinner {
			return m, cmd
		}
	case TickMsg:
		for i := range m.rows {
			m.rows[i].ReRenderFields()
		}
		m.updateColumnWidths()
		return m, doTick()

	case tea.WindowSizeMsg:
		m.maxHeight = msg.Height
		m.help.Width = msg.Width
		m.updatePagination()
		return m, m.updateFullscreenCmd()
	}
	return m, nil
}

func (m Model) View() string {
	if m.ShowHelp {
		return m.help.FullHelpView(m.FullHelp())
	}
	if len(m.rows) == 0 {
		if m.showSpinner {
			return m.spinner.View()
		}
		return "No resources found"
	}
	if len(m.filteredRows) == 0 {
		if m.filterInputEnabled {
			return m.filterInput.View()
		} else {
			return "No resources visible"
		}
	}
	var buf bytes.Buffer
	if m.maxHeight > 1 {
		if m.filterInputEnabled {
			buf.WriteString(m.filterInput.View())
		} else {
			m.columnsView(&buf, m.headers, lipgloss.Style{})
		}
		buf.WriteByte('\n')
	}

	currentPage := m.currentPaginatedPage()
	for i, row := range currentPage {
		if i > 0 {
			buf.WriteByte('\n')
		}
		m.rowView(&buf, row)
	}

	paginatorVisible := m.paginatorVisible()
	if paginatorVisible {
		for i := len(currentPage); i < m.Paginator.PerPage; i++ {
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
		buf.WriteString(m.Styles.Pagination.Render(m.Paginator.View()))
	}

	if m.err != nil {
		if paginatorVisible {
			buf.WriteString("  ")
		} else {
			buf.WriteByte('\n')
		}
		buf.WriteString(m.Styles.Error.Render(m.err.Error()))
	}

	buf.WriteByte('\n')

	return buf.String()
}

func (m Model) currentPaginatedPage() []Row {
	if len(m.filteredRows) == 0 {
		return nil
	}
	start, end := m.Paginator.GetSliceBounds(len(m.filteredRows))
	return m.filteredRows[start:end]
}

func (m Model) rowView(w io.Writer, row Row) {
	style := m.Styles.Row.Cell
	switch row.Status {
	case StatusError:
		style = m.Styles.Row.Error
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
			spacing := m.CellSpacing + m.columnWidths[i-1] - ansi.PrintableRuneWidth(columns[i-1])
			if spacing > 0 {
				fmt.Fprint(w, lotsOfSpaces[:spacing])
			}
		}
		fmt.Fprint(w, style.Render(col))
	}
}

func (m *Model) updateColumnWidths() {
	lengths := expandToMaxLengths(nil, m.headers)
	for _, row := range m.currentPaginatedPage() {
		lengths = expandToMaxLengths(lengths, row.RenderedFields())
	}
	m.columnWidths = lengths
}

func (m *Model) filterText() string {
	if !m.filterInputEnabled {
		return ""
	}
	return m.filterInput.Value()
}

func expandToMaxLengths(lengths []int, row []string) []int {
	lengths = expandSlice(lengths, len(row))
	for i, f := range row {
		lengths[i] = max(ansi.PrintableRuneWidth(f), lengths[i])
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
	if cap(slice) >= minLen {
		return slice[:minLen]
	}
	return append(slice, make(S, delta)...)
}
