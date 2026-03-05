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
)

func TestLabelColumnHeader(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "my-label",
			want:  "MY-LABEL",
		},
		{
			input: "foo/bar",
			want:  "BAR",
		},
		{
			input: "foo/bar/moo",
			want:  "MOO",
		},
		{
			input: "foo/",
			want:  "",
		},
		{
			input: "/",
			want:  "",
		},
		{
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := labelColumnHeader(tc.input)
			if got != tc.want {
				t.Errorf("value did not match\nwant: %q\ngot:  %q", tc.want, got)
			}
		})
	}
}

func Test_parseArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "no args",
			args: args{
				args: []string{},
			},
			wantErr: true,
		}, {
			name: "single arg",
			args: args{
				args: []string{
					"pods",
				},
			},
			wantErr: false,
		}, {
			name: "multiple args",
			args: args{
				args: []string{
					"pods",
					"nginx",
				},
			},
			wantErr: false,
		}, {
			name: "resource/name",
			args: args{
				args: []string{
					"pods/nginx",
				},
			},
			wantErr: false,
		}, {
			name: "comma separated args",
			args: args{
				args: []string{
					"pods,nodes",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseArgs(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
