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
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the menu.
type KeyMap struct {
	// Keybindings used when browsing the list.
	NextPage  key.Binding
	PrevPage  key.Binding
	GoToStart key.Binding
	GoToEnd   key.Binding

	// Keybindings for view settings
	Filter           key.Binding
	ToggleDeleted    key.Binding
	ToggleFullscreen key.Binding

	// Keybindings used while the text-filter is enabled.
	CloseFilter key.Binding
	ClearFilter key.Binding

	// Help toggle keybindings.
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding

	// The quit-no-matter-what keybinding. This will be caught when filtering.
	ForceQuit key.Binding
}

func (k KeyMap) EscapeFilterText(keyMsg tea.KeyMsg) bool {
	return key.Matches(keyMsg, k.ForceQuit, k.ClearFilter, k.CloseFilter, k.NextPage, k.PrevPage)
}

// DefaultKeyMap is a default set of keybindings.
var DefaultKeyMap = KeyMap{
	// Browsing.
	PrevPage: key.NewBinding(
		key.WithKeys("left", "h", "pgup"),
		key.WithHelp("←/h/pgup", "prev page"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right", "l", "pgdown"),
		key.WithHelp("→/l/pgdn", "next page"),
	),
	GoToStart: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
	),
	GoToEnd: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
	),
	ToggleFullscreen: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "toggle fullscreen"),
	),

	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter by text"),
	),
	ToggleDeleted: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "show/hide deleted"),
	),

	// Filtering.
	CloseFilter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "close the filter input field"),
	),
	ClearFilter: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear the applied filter"),
	),

	// Toggle help.
	ShowFullHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "more"),
	),
	CloseFullHelp: key.NewBinding(
		key.WithKeys("?", "esc"),
		key.WithHelp("?/esc", "close help"),
	),

	// Quitting.
	ForceQuit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

// FullHelp returns bindings to show the full help view. It's part of the
// help.KeyMap interface.
func (m Model) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.KeyMap.NextPage,
		m.KeyMap.PrevPage,
		m.KeyMap.GoToStart,
		m.KeyMap.GoToEnd,
	}}

	//filtering := m.filterState == Filtering

	//// If the delegate implements the help.KeyMap interface add full help
	//// keybindings to a special section of the full help.
	//if !filtering {
	//	if b, ok := m.delegate.(help.KeyMap); ok {
	//		kb = append(kb, b.FullHelp()...)
	//	}
	//}

	listLevelBindings := []key.Binding{
		m.KeyMap.ToggleDeleted,
		m.KeyMap.ToggleFullscreen,
		m.KeyMap.Filter,
		m.KeyMap.CloseFilter,
		m.KeyMap.ClearFilter,
	}

	return append(kb,
		listLevelBindings,
		[]key.Binding{
			m.KeyMap.ForceQuit,
			m.KeyMap.CloseFullHelp,
		})
}
