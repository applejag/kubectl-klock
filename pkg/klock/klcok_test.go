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

package klock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_NormalizedLabelColumns(t *testing.T) {
	labelColumnsTests := []struct {
		labelColumns           []string
		normalizedLabelColumns []string
	}{
		{
			labelColumns:           []string{"app"},
			normalizedLabelColumns: []string{"app"},
		},
		{
			labelColumns:           []string{"app,version"},
			normalizedLabelColumns: []string{"app", "version"},
		},
		{
			labelColumns:           []string{"app", "version"},
			normalizedLabelColumns: []string{"app", "version"},
		},
		{
			labelColumns:           []string{"app", "version,role"},
			normalizedLabelColumns: []string{"app", "version", "role"},
		},
		{
			labelColumns:           []string{" app , version ", " role   "},
			normalizedLabelColumns: []string{"app", "version", "role"},
		},
		{
			labelColumns:           []string{" , app, , version,, ,", ",role, "},
			normalizedLabelColumns: []string{"app", "version", "role"},
		},
		{
			labelColumns:           []string{",", " , ", " ", ""},
			normalizedLabelColumns: []string{},
		},
	}

	for _, labelColumnsTest := range labelColumnsTests {
		o := Options{LabelColumns: labelColumnsTest.labelColumns}
		assert.Exactly(t, labelColumnsTest.normalizedLabelColumns, o.NormalizedLabelColumns())
	}
}
