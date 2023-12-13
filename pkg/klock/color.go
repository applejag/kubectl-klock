// SPDX-FileCopyrightText: 2020 Hidetatsu Yaginuma
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
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
//
// This file contains modified version taken from kubecolor's source:
// https://github.com/kubecolor/kubecolor/blob/aace98d870cc1a7da5f6f99ab61f99eea61b6418/printer/kubectl_output_colored_printer.go

package klock

import (
	"strings"

	"github.com/applejag/kubectl-klock/pkg/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	StyleRestartsWarning = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(3))

	StyleFractionOK      = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(2))
	StyleFractionWarning = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(3))

	StyleStatusDefault = lipgloss.NewStyle()
	StyleStatusOK      = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(2))
	StyleStatusError   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(1))
	StyleStatusWarning = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(3))
)

func FractionStyle(str string) (lipgloss.Style, bool) {
	f, ok := ParseFraction(str)
	if !ok {
		return lipgloss.Style{}, false
	}
	if f.Count == f.Total {
		return StyleFractionOK, true
	}
	return StyleFractionWarning, true
}

func StatusColumn(status string) any {
	if !strings.Contains(status, ",") {
		return table.StyledColumn{
			Value: status,
			Style: StatusStyle(status),
		}
	}
	column := table.JoinedColumn{
		Delimiter: ",",
	}
	for _, s := range strings.Split(status, ",") {
		column.Values = append(column.Values, table.StyledColumn{
			Value: s,
			Style: StatusStyle(s),
		})
	}
	return column
}

func StatusStyle(status string) lipgloss.Style {
	switch status {
	case
		// from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/events/event.go
		// Container event reason list
		"Failed",
		"BackOff",
		"ExceededGracePeriod",
		// Pod event reason list
		"FailedKillPod",
		"FailedCreatePodContainer",
		// "Failed",
		"NetworkNotReady",
		// Image event reason list
		// "Failed",
		"InspectFailed",
		"ErrImageNeverPull",
		// "BackOff",
		// kubelet event reason list
		"NodeNotSchedulable",
		"KubeletSetupFailed",
		"FailedAttachVolume",
		"FailedMount",
		"VolumeResizeFailed",
		"FileSystemResizeFailed",
		"FailedMapVolume",
		"ContainerGCFailed",
		"ImageGCFailed",
		"FailedNodeAllocatableEnforcement",
		"FailedCreatePodSandBox",
		"FailedPodSandBoxStatus",
		"FailedMountOnFilesystemMismatch",
		// Image manager event reason list
		"InvalidDiskCapacity",
		"FreeDiskSpaceFailed",
		// Probe event reason list
		"Unhealthy",
		// Pod worker event reason list
		"FailedSync",
		// Config event reason list
		"FailedValidation",
		// Lifecycle hooks
		"FailedPostStartHook",
		"FailedPreStopHook",
		// Node status list
		"NotReady",
		"NetworkUnavailable",

		// some other status
		"CreateContainerConfigError",
		"ContainerStatusUnknown",
		"CrashLoopBackOff",
		"ImagePullBackOff",
		"Evicted",
		"FailedScheduling",
		"Error",
		"ErrImagePull":
		return StyleStatusError
	case
		// from https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/events/event.go
		// Container event reason list
		"Killing",
		"Preempting",
		// Pod event reason list
		// Image event reason list
		// kubelet event reason list
		"NodeNotReady",
		"NodeSchedulable",
		"Starting",
		"AlreadyMountedVolume",
		"SuccessfulAttachVolume",
		"SuccessfulMountVolume",
		"NodeAllocatableEnforced",
		// Image manager event reason list
		// Probe event reason list
		"ProbeWarning",
		// Pod worker event reason list
		// Config event reason list
		// Lifecycle hooks
		// Node event reason list
		"SchedulingDisabled",
		"DiskPressure",
		"MemoryPressure",
		"PIDPressure",

		// some other status
		"Pending",
		"ContainerCreating",
		"PodInitializing",
		"Terminating",
		"Warning",

		// PV reclaim policy
		"Delete":
		return StyleStatusWarning
	case
		"Running",
		"Completed",
		"Pulled",
		"Created",
		"Rebooted",
		"NodeReady",
		"Started",
		"Normal",
		"VolumeResizeSuccessful",
		"FileSystemResizeSuccessful",
		"Ready",

		// PV reclaim policy
		"Retain":
		return StyleStatusOK
	}
	// some ok status, not colored:
	// "SandboxChanged",
	// "Pulling",
	return StyleStatusDefault
}
