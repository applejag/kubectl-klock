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

import "strings"

func SplitsFromStart(s string, sep byte) []string {
	if s == "" {
		return nil
	}

	var indexFromStart int
	var result []string

	for {
		index := strings.IndexByte(s[indexFromStart:], sep)
		if index == -1 {
			break
		}

		indexFromStart += index
		result = append(result, s[:indexFromStart])
		indexFromStart++ // skip over the separator
	}

	result = append(result, s)
	return result
}
