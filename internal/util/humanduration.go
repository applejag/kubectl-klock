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
	"fmt"
	"io"
	"time"
)

func ParseHumanDuration(s string) (time.Duration, bool) {
	const (
		DAY = time.Hour * 24
	)

	rest := s
	var dur time.Duration

	for rest != "" {
		num, char, newRest, ok := parseHumanDurationSegment(rest)
		if !ok {
			return dur, false
		}
		rest = newRest
		switch char {
		case 'y':
			now := time.Now()
			dur += now.AddDate(num, 0, 0).Sub(now)
		case 'M':
			now := time.Now()
			dur += now.AddDate(0, num, 0).Sub(now)
		case 'w':
			now := time.Now()
			dur += now.AddDate(0, 0, 7*num).Sub(now)
		case 'd':
			now := time.Now()
			dur += now.AddDate(0, 0, num).Sub(now)
		case 'h':
			dur += time.Duration(num) * time.Hour
		case 'm':
			dur += time.Duration(num) * time.Minute
		case 's':
			dur += time.Duration(num) * time.Second
		default:
			return dur, false
		}
	}
	return dur, true
}

func parseHumanDurationSegment(s string) (num int, char rune, rest string, ok bool) {
	n, err := fmt.Sscanf(s, "%d%c%s", &num, &char, &rest)
	ok = (err == io.EOF && n == 2) || err == nil
	return
}
