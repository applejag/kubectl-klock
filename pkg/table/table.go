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
	"slices"
	"strings"
	"time"

	"github.com/applejag/kubectl-klock/internal/util"
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
	Header lipgloss.Style
	Row    RowStyles

	NoneFound         lipgloss.Style
	Error             lipgloss.Style
	Pagination        lipgloss.Style
	FilterPrompt      lipgloss.Style
	FilterInfo        lipgloss.Style
	FilterNoneVisible lipgloss.Style
	StatusDelim       lipgloss.Style

	Toggles lipgloss.Style
}

var subduedColor = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

var DefaultStyles = Styles{
	Header: lipgloss.NewStyle(),
	Row:    DefaultRowStyle,

	NoneFound: lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(3)).
		SetString("No resources found"),
	Error: lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(9)).
		SetString("ERROR:"),
	Pagination: lipgloss.NewStyle().
		Foreground(subduedColor).
		SetString("PAGE:"),
	FilterPrompt: lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(11)).
		SetString("FILTER:"),
	FilterInfo: lipgloss.NewStyle().
		Foreground(subduedColor).
		SetString("FILTERING:"),
	FilterNoneVisible: lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(3)).
		SetString("No resources visible"),
	StatusDelim: lipgloss.NewStyle().
		Foreground(subduedColor).
		Bold(true).
		SetString(" | "),

	Toggles: lipgloss.NewStyle().
		Foreground(subduedColor),
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

	err                 error
	headers             []string
	maxHeight           int
	rows                []Row
	filteredRows        []Row
	columnWidths        []int
	fullscreenOverride  bool
	quitting            bool
	prevSuggestionCount int

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
	m.updateRows()
	fullscreenCmd := m.updateFullscreenCmd()
	return fullscreenCmd
}

func (m *Model) SetRows(rows []Row) tea.Cmd {
	m.rows = slices.Clone(rows)
	m.sortItems()
	if len(m.rows) > 0 {
		m.StopSpinner()
	}
	m.updateRows()
	return m.updateFullscreenCmd()
}

func (m *Model) updateRows() {
	m.updateFilteredRows()
	m.updateFilterSuggestions()
	m.updatePagination()
	m.updateColumnWidths()
}

func (m *Model) updateFilteredRows() {
	filterText := m.filterText()
	m.filteredRows = make([]Row, 0, len(m.rows))
	for _, row := range m.rows {
		if m.HideDeleted && row.Status == StatusDeleted {
			continue
		}
		if filterText != "" && !rowMatchesText(row, filterText) {
			continue
		}
		m.filteredRows = append(m.filteredRows, row)
	}
}

func rowMatchesText(row Row, needle string) bool {
	for _, field := range row.RenderedFields() {
		if strings.Contains(field, needle) {
			return true
		}
	}
	return false
}

func (m *Model) updateFilterSuggestions() {
	m.filterInput.ShowSuggestions = true
	suggestionsMap := make(map[string]struct{}, m.prevSuggestionCount)
	suggestions := make([]string, 0, m.prevSuggestionCount)

	for _, row := range m.filteredRows {
		for _, split := range util.SplitsFromStart(row.Suggestion, '-') {
			_, isDuplicate := suggestionsMap[split]
			if isDuplicate {
				continue
			}
			suggestionsMap[split] = struct{}{}
			suggestions = append(suggestions, split)
		}
	}

	m.prevSuggestionCount = len(suggestions)
	m.filterInput.SetSuggestions(suggestions)
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
		switch {
		case m.filterInputEnabled && !m.KeyMap.EscapeFilterText(msg):
			m.filterInput.KeyMap.NextSuggestion = m.KeyMap.NextSuggestion
			m.filterInput.KeyMap.PrevSuggestion = m.KeyMap.PrevSuggestion
			m.filterInput.KeyMap.AcceptSuggestion = m.KeyMap.AcceptSuggestion
			i, cmd := m.filterInput.Update(msg)
			m.filterInput = i
			m.updateRows()
			return m, cmd
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
			m.updateRows()
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
		case key.Matches(msg, m.KeyMap.CloseFilter):
			m.filterInputEnabled = false
		case key.Matches(msg, m.KeyMap.ClearFilter):
			m.filterInputEnabled = false
			m.filterInput.SetValue("")
			m.updateRows()
		case key.Matches(msg, m.KeyMap.Filter):
			m.filterInputEnabled = true
			m.updateRows()
			return m, m.filterInput.Focus()
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
	if m.showSpinner {
		return m.spinner.View()
	}
	var buf bytes.Buffer

	currentPage := m.currentPaginatedPage()

	if m.maxHeight > 1 {
		if m.filterInputEnabled {
			m.filterInput.Prompt = ""
			m.filterInput.PromptStyle = m.Styles.FilterPrompt
			buf.WriteString(m.filterInput.View())
			buf.WriteByte('\n')
		} else if len(currentPage) > 0 {
			m.columnsView(&buf, m.headers, m.Styles.Header)
			buf.WriteByte('\n')
		}
	}

	var status []string
	m.viewWriteRows(&buf, currentPage)

	paginatorVisible := m.paginatorVisible()
	if paginatorVisible {
		// Add padding below empty lines
		for i := len(currentPage); i < m.Paginator.PerPage; i++ {
			buf.WriteByte('\n')
		}
		status = append(status, m.Styles.Pagination.Render(m.Paginator.View()))
	}

	if m.filterText() != "" {
		filterInfo := fmt.Sprintf(`"%s" (%d/%d rows)`, m.filterText(), len(m.filteredRows), len(m.rows))
		status = append(status, m.Styles.FilterInfo.Render(filterInfo))
	}

	if len(m.rows) == 0 {
		status = append(status, m.Styles.NoneFound.String())
	} else if len(m.filteredRows) == 0 {
		status = append(status, m.Styles.FilterNoneVisible.String())
	}

	if m.err != nil {
		status = append(status, m.Styles.Error.Render(m.err.Error()))
	}

	if m.fullscreenOverride {
		status = append(status, m.Styles.Toggles.Render("force fullscreen"))
	}

	if m.HideDeleted {
		status = append(status, m.Styles.Toggles.Render("hide deleted"))
	}

	if len(status) > 0 {
		if len(currentPage) > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(strings.Join(status, m.Styles.StatusDelim.String()))
	}

	if m.quitting {
		buf.WriteByte('\n')
	}

	return buf.String()
}

func (m Model) currentPaginatedPage() []Row {
	if len(m.filteredRows) == 0 {
		return nil
	}
	start, end := m.Paginator.GetSliceBounds(len(m.filteredRows))
	return m.filteredRows[start:end]
}

func (m Model) viewWriteRows(buf *bytes.Buffer, currentPage []Row) {
	for i, row := range currentPage {
		if i > 0 {
			buf.WriteByte('\n')
		}
		m.rowView(buf, row)
	}
}

func (m Model) rowView(buf *bytes.Buffer, row Row) {
	style := m.Styles.Row.Cell
	switch row.Status {
	case StatusError:
		style = m.Styles.Row.Error
	case StatusDeleted:
		style = m.Styles.Row.Deleted
	}
	m.columnsView(buf, row.RenderedFields(), style)
}

var lotsOfSpaces = strings.Repeat(" ", 200)

func (m Model) columnsView(buf *bytes.Buffer, columns []string, style lipgloss.Style) {
	for i, col := range columns {
		if i > 0 {
			//TODO: test style.Width()
			spacing := m.CellSpacing + m.columnWidths[i-1] - ansi.PrintableRuneWidth(columns[i-1])
			if spacing > 0 {
				fmt.Fprint(buf, lotsOfSpaces[:spacing])
			}
		}
		fmt.Fprint(buf, style.Render(col))
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
