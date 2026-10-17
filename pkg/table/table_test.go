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
	"testing"
	"time"

	"github.com/kubecolor/kubecolor/config"
	"github.com/kubecolor/kubecolor/config/color"
)

func TestFreshColor(t *testing.T) {
	green := color.MustParse("green")
	yellow := color.MustParse("yellow")
	cfg := &config.Config{
		ObjFreshThreshold: config.MustParseDurationSlice("1m/1h/1d"),
	}
	// Fewer colors than thresholds: the last threshold has no color.
	cfg.Theme.Data.DurationFresh = color.Slice{green, yellow}

	tests := []struct {
		name      string
		age       time.Duration
		wantColor color.Color
		wantOK    bool
	}{
		{name: "under first threshold", age: 30 * time.Second, wantColor: green, wantOK: true},
		{name: "under second threshold", age: 30 * time.Minute, wantColor: yellow, wantOK: true},
		{name: "threshold without color", age: 2 * time.Hour, wantOK: false},
		{name: "older than all thresholds", age: 48 * time.Hour, wantOK: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := freshColor(tc.age, cfg)
			if ok != tc.wantOK {
				t.Fatalf("want ok=%t, got ok=%t", tc.wantOK, ok)
			}
			if ok && got.String() != tc.wantColor.String() {
				t.Errorf("want color %q, got %q", tc.wantColor.String(), got.String())
			}
		})
	}

	if _, ok := freshColor(time.Second, &config.Config{}); ok {
		t.Error("want no fresh color when no thresholds are configured")
	}
	if _, ok := freshColor(time.Second, nil); ok {
		t.Error("want no fresh color when config is nil")
	}
}
