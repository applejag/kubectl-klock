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

	"github.com/applejag/kubectl-klock/pkg/table"
	"github.com/charmbracelet/lipgloss"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestParseCellFractionColoring(t *testing.T) {
	tests := []struct {
		name      string
		cellValue string
		wantStyle lipgloss.Style
	}{
		{
			name:      "ready fraction ok",
			cellValue: "1/1",
			wantStyle: StyleFractionOK,
		},
		{
			name:      "ready fraction warning",
			cellValue: "0/1",
			wantStyle: StyleFractionWarning,
		},
	}

	p := &Printer{}
	colDef := metav1.TableColumnDefinition{Name: "Ready"}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.parseCell(test.cellValue, metav1.TableRow{}, watch.Added, nil, colDef, time.Now())
			styled, ok := got.(table.StyledColumn)
			if !ok {
				t.Fatalf("expected table.StyledColumn, got %T (%v)", got, got)
			}
			if got, want := styled.Style.GetForeground(), test.wantStyle.GetForeground(); got != want {
				t.Errorf("wrong foreground for %q\nwant: %v\ngot:  %v", test.cellValue, want, got)
			}
		})
	}
}

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

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := labelColumnHeader(test.input)
			if got != test.want {
				t.Errorf("value did not match\nwant: %q\ngot:  %q", test.want, got)
			}
		})
	}
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "single arg",
			args:    []string{"pods"},
			wantErr: "",
		},
		{
			name:    "multiple args",
			args:    []string{"pods", "nginx"},
			wantErr: "",
		},
		{
			name:    "resource/name",
			args:    []string{"pods/nginx"},
			wantErr: "",
		},
		{
			name:    "comma separated args",
			args:    []string{"pods,nodes"},
			wantErr: "you may only specify a single resource type",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateArgs(test.args)
			t.Logf("args: %#v", test.args)
			if test.wantErr == "" && err != nil {
				t.Errorf("unexpected error: %q", err)
			}
			if test.wantErr != "" && (err == nil || err.Error() != test.wantErr) {
				t.Errorf("wrong error result\nwant: %q\ngot:  %q", test.wantErr, err)
			}
		})
	}
}
