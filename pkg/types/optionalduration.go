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

package types

import (
	"encoding"
	"fmt"
	"time"

	"github.com/applejag/kubectl-klock/internal/util"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/duration"
)

// OptionalDuration is a config option that can be either set to a [time.Duration],
// or a false value via the string "false" or an empty value.
//
// Compared to the standard library duration, the parsing of this duration
// also supports more units:
//
// - "d" for "24h"
// - "w" for "7d"
// - "M" for "30d"
// - "y" for "365d"
type OptionalDuration struct {
	Value    time.Duration
	HasValue bool
}

func NewOptionalDuration(dur time.Duration) OptionalDuration {
	return OptionalDuration{Value: dur, HasValue: true}
}

var _ encoding.TextUnmarshaler = &OptionalDuration{}
var _ pflag.Value = &OptionalDuration{}

func (o OptionalDuration) Duration() (time.Duration, bool) {
	return o.Value, o.HasValue
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (o *OptionalDuration) UnmarshalText(text []byte) error {
	return o.Set(string(text))
}

// Set implements [pflag.Value].
func (o *OptionalDuration) Set(v string) error {
	switch v {
	case "", "false", "False", "FALSE":
		o.HasValue = false
		o.Value = 0
		return nil
	case "0", "true":
		o.HasValue = true
		o.Value = 0
		return nil
	default:
		dur, ok := util.ParseHumanDuration(v)
		if !ok {
			return fmt.Errorf(`invalid duration %q, must be "false", empty string, or a valid time duration using units: s, m, h, d, w, M, y`, v)
		}
		o.HasValue = true
		o.Value = dur
		return nil
	}
}

// String implements [pflag.Value].
func (o OptionalDuration) String() string {
	if !o.HasValue {
		return "false"
	}
	return duration.HumanDuration(o.Value)
}

// Type implements [pflag.Value].
func (o OptionalDuration) Type() string {
	return "duration?"
}
