// SPDX-FileCopyrightText: 2024 Kalle Fagerberg
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

package util

import (
	"slices"
	"testing"
)

func TestSplitsFromStart(t *testing.T) {
	const sep = '-'
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{
			name: "empty",
			s:    "",
			want: nil,
		},
		{
			name: "1 split",
			s:    "foo",
			want: []string{"foo"},
		},
		{
			name: "2 splits",
			s:    "foo-bar",
			want: []string{"foo", "foo-bar"},
		},
		{
			name: "deployment pod name",
			s:    "thing-operator-675ffd4bbb-jfsfn",
			want: []string{"thing", "thing-operator", "thing-operator-675ffd4bbb", "thing-operator-675ffd4bbb-jfsfn"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitsFromStart(tc.s, sep)
			if !slices.Equal(tc.want, got) {
				t.Errorf("wrong result\ns:    %q\nsep:  %q\nwant: %v\ngot:  %v", tc.s, sep, tc.want, got)
			}
		})
	}
}
