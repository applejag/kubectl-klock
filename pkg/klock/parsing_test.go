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
	"time"
)

func TestParseHumanDurationSegment(t *testing.T) {
	tests := []struct {
		input string
		num   int
		char  rune
		rest  string
		ok    bool
	}{
		{
			input: "12y",
			num:   12,
			char:  'y',
			rest:  "",
			ok:    true,
		},
		{
			input: "12y15d",
			num:   12,
			char:  'y',
			rest:  "15d",
			ok:    true,
		},
		{
			input: "invalid",
			ok:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			num, char, rest, ok := parseHumanDurationSegment(tc.input)
			if tc.ok != ok {
				t.Fatalf("want ok=%t, got ok=%t & num=%d & char=%c & rest=%s", tc.ok, ok, num, char, rest)
			}
			if !tc.ok {
				return
			}
			if tc.num != num {
				t.Errorf("want num=%d, got num=%d", tc.num, num)
			}
			if tc.char != char {
				t.Errorf("want char=%c, got char=%c", tc.char, char)
			}
			if tc.rest != rest {
				t.Errorf("want rest=%s, got rest=%s", tc.rest, rest)
			}
		})
	}
}

func TestParseHumanDuration(t *testing.T) {
	const (
		DAY = time.Hour * 24
		// This is how [k8s.io/apimachinery/pkg/util/duration] defines a year
		YEAR = DAY * 365
	)

	tests := []struct {
		input string
		dur   time.Duration
		ok    bool
	}{
		{
			input: "1d",
			dur:   DAY,
			ok:    true,
		},
		{
			input: "1d15m",
			dur:   DAY + 15*time.Minute,
			ok:    true,
		},
		{
			input: "1d15m30s",
			dur:   DAY + 15*time.Minute + 30*time.Second,
			ok:    true,
		},
		{
			input: "1y30d",
			dur:   YEAR + 30*DAY,
			ok:    true,
		},
		{
			input: "1f",
			ok:    false,
		},
		{
			input: "1h30p",
			ok:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			dur, ok := parseHumanDuration(tc.input)
			if tc.ok != ok {
				t.Fatalf("want ok=%t, got ok=%t & dur=%s", tc.ok, ok, dur)
			}
			if !tc.ok {
				return
			}
			if tc.dur != dur {
				t.Errorf("want dur=%s, got dur=%s", tc.dur, dur)
			}
		})
	}
}
