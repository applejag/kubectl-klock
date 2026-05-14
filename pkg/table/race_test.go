// SPDX-FileCopyrightText: 2025 Kalle Fagerberg
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
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestRace_AddRowWhileUpdate reproduces the data race between the Kubernetes
// watch goroutine calling AddRow and the bubbletea event loop calling Update
// (on TickMsg) and View concurrently.
//
// This is the root cause of https://github.com/applejag/kubectl-klock/issues/161
//
// Run with: go test -race -run TestRace_AddRowWhileUpdate -count=1 ./pkg/table/
func TestRace_AddRowWhileUpdate(t *testing.T) {
	m := New()
	m.SetHeaders([]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"})

	// Pre-populate some rows so the TickMsg handler has something to iterate
	for i := 0; i < 10; i++ {
		m.AddRow(Row{
			ID:     fmt.Sprintf("uid-%d", i),
			Fields: []any{fmt.Sprintf("pod-%d", i), "1/1", "Running", "0", time.Now()},
		})
	}

	const (
		numWriters  = 3
		numUpdates  = 200
		numAddRows  = 200
	)

	var wg sync.WaitGroup

	// Simulate the Kubernetes watch goroutine: call AddRow, SetHeaders,
	// SetError, SetRows, StartSpinner, StopSpinner concurrently.
	// This is what happens in klock.go watch() and pipeEvents().
	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < numAddRows; i++ {
				uid := fmt.Sprintf("uid-%d-%d", writerID, i)
				row := Row{
					ID:     uid,
					Fields: []any{fmt.Sprintf("pod-%d-%d", writerID, i), "1/1", "Running", "0", time.Now()},
				}
				m.AddRow(row)

				// Also exercise the other methods called from background goroutines
				if i%50 == 0 {
					m.SetHeaders([]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"})
				}
				if i%70 == 0 {
					m.SetError(fmt.Errorf("transient error %d", i))
				}
				if i%71 == 0 {
					m.SetError(nil)
				}
				if i%100 == 0 {
					m.StopSpinner()
				}
			}
		}(w)
	}

	// Simulate the bubbletea event loop: call Update(TickMsg) and View()
	// concurrently with the AddRow goroutines above.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numUpdates; i++ {
			m.Update(TickMsg(time.Now()))
			_ = m.View()
		}
	}()

	wg.Wait()
}

// TestRace_SetRowsWhileView reproduces the race between Clear/SetRows
// (called from klock.go watch() on restart) and View (bubbletea event loop).
func TestRace_SetRowsWhileView(t *testing.T) {
	m := New()
	m.SetHeaders([]string{"NAME", "STATUS"})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			rows := make([]Row, 5)
			for j := range rows {
				rows[j] = Row{
					ID:     fmt.Sprintf("uid-%d-%d", i, j),
					Fields: []any{fmt.Sprintf("pod-%d-%d", i, j), "Running"},
				}
			}
			m.SetRows(rows)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			m.Update(TickMsg(time.Now()))
			_ = m.View()
		}
	}()

	wg.Wait()
}
