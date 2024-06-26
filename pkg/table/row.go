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
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"k8s.io/apimachinery/pkg/util/duration"
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

type StyledColumn struct {
	Value any
	Style lipgloss.Style
}

type JoinedColumn struct {
	Delimiter string
	Values    []any
}

type AgoColumn struct {
	Value string
	Time  time.Time
}

func (c AgoColumn) String() string {
	dur := time.Since(c.Time)
	return fmt.Sprintf("%s (%s ago)", c.Value, duration.HumanDuration(dur))
}

type Row struct {
	ID         string
	Fields     []any
	Status     Status
	SortKey    string
	Suggestion string

	renderedFields []string
}

type Status int

const (
	StatusDefault Status = iota
	StatusError
	StatusWarning
	StatusDeleted
)

// SortValue value is the value we use when sorting the list.
func (r Row) SortValue() string {
	if r.SortKey != "" {
		return r.SortKey
	}
	if len(r.Fields) == 0 {
		return ""
	}
	str, ok := r.Fields[0].(string)
	if !ok {
		return ""
	}
	return str
}

func (r *Row) RenderedFields() []string {
	if len(r.renderedFields) != len(r.Fields) {
		r.ReRenderFields()
	}
	return r.renderedFields
}

func (r *Row) ReRenderFields() {
	r.renderedFields = resizeSlice(r.renderedFields, len(r.Fields))
	for i, col := range r.Fields {
		r.renderedFields[i] = renderColumn(col)
	}
}

func renderColumn(value any) string {
	switch value := value.(type) {
	case JoinedColumn:
		var sb strings.Builder
		for i, v := range value.Values {
			if i > 0 {
				sb.WriteString(value.Delimiter)
			}
			sb.WriteString(renderColumn(v))
		}
		return sb.String()
	case StyledColumn:
		return value.Style.Render(renderColumn(value.Value))
	case string:
		return value
	case time.Time:
		dur := time.Since(value)
		return duration.HumanDuration(dur)
	case fmt.Stringer:
		return value.String()
	default:
		if value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}
}
