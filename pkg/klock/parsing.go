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
	"fmt"
	"io"
	"regexp"
	"time"
)

type Fraction struct {
	Count int
	Total int
}

func (f Fraction) String() string {
	return fmt.Sprintf("%d/%d", f.Count, f.Total)
}

func ParseFraction(s string) (Fraction, bool) {
	var f Fraction
	if _, err := fmt.Sscanf(s, "%d/%d", &f.Count, &f.Total); err != nil {
		return Fraction{}, false
	}
	return f, true
}

var podRestartsRegex = regexp.MustCompile(`^(\d+)(?: \((\S+) ago\))$`)

func parsePodRestarts(s string) (string, time.Duration, bool) {
	// 0, the most common case
	if s == "0" {
		return s, 0, false
	}
	groups := podRestartsRegex.FindStringSubmatch(s)
	if groups == nil {
		// No match
		return s, 0, false
	}
	groupCount := groups[1]
	groupDur := groups[2]
	dur, ok := parseHumanDuration(groupDur)
	if !ok {
		return s, 0, false
	}
	return groupCount, dur, true
}

func parseHumanDuration(s string) (time.Duration, bool) {
	const (
		DAY = time.Hour * 24
		// This is how [k8s.io/apimachinery/pkg/util/duration] defines a year
		YEAR = DAY * 365
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
			dur += time.Duration(num) * YEAR
		case 'd':
			dur += time.Duration(num) * DAY
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
