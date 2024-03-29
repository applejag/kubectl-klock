<!--
SPDX-FileCopyrightText: 2023 Kalle Fagerberg

SPDX-License-Identifier: CC-BY-4.0
-->

# kubectl-klock's changelog

This project tries to follow [SemVer 2.0.0](https://semver.org/).

<!--
	When composing new changes to this list, try to follow convention.
	The WIP release shall be updated just before adding the Git tag.
	Replace (WIP) by (YYYY-MM-DD), e.g. (2021-02-09) for 9th of February, 2021
	A good source on conventions can be found here:
	https://changelog.md/
-->

## v0.6.0 (2024-03-11)

- Added status coloring on multiple statuses,
  where `NotReady,SchedulingDisabled` would have red `NotReady`
  and orange `SchedulingDisabled`. (#69)

- Added coloring on PV/StorageClass reclaim policy and PV/PVC status. (#73)

- Changed to Go 1.21.5 to resolve vulnerability [GO-2023-2182](https://pkg.go.dev/vuln/GO-2023-2182),
  for denial of service in net/http, and [GO-2023-2185](https://pkg.go.dev/vuln/GO-2023-2185)
  for insecure parsing of Windows paths with a `\??\` prefix. (#72)

- Changed to Go 1.22.1 to stay up to date, but also to resolve
  vulnerabilities [GO-2024-2600](https://pkg.go.dev/vuln/GO-2024-2600),
  [GO-2024-2599](https://pkg.go.dev/vuln/GO-2024-2599),
  and [GO-2024-2598](https://pkg.go.dev/vuln/GO-2024-2598). (#85)

## v0.5.1 (2023-11-11)

- Fixes `--watch-kubeconfig` to reading a kubeconfig that has not yet been
  fully flushed to disk. Only relevant on bigger kubeconfig files.

  The fix is just adding a small 150ms sleep, which is hopefully enough time
  for tools like `kubectx` to finish writing the kubeconfig file. (#63)

- Added spinner when the watch restarts (either from `--watch-kubeconfig` or
  from an error), to indicate that it's loading. (#64)

- Changed to Go 1.21.4 to resolve vulnerability [GO-2023-2186](https://pkg.go.dev/vuln/GO-2023-2186),
  where `filepath.IsLocal` incorrectly treated reserved names as local. (#66)

## v0.5.0 (2023-11-04)

- BREAKING: Changed binary name from `klock` to `kubectl-klock`.
  Any automated tooling downloading from GitHub release assets may break. (#41)

- BREAKING: Changed my username from `jilleJr` to `applejag`.
  Any automated tooling may break from this. (#42)

- Added `completion` subcommand. (#34)

- Added completion to resource type and name. (#34)

- Added completion on flags: (#39)

  - `--output`, `-o`
  - `--namespace`, `-n`
  - `--cluster`
  - `--context`
  - `--user`

- Added `--version` flag to print the command's version.
  (#40, thanks @semihbkgr!)

- Added Snap: `sudo snap install klock --edge` (#43)

- Added `--label-columns` / `-L` flag to present labels as columns.
  (#55, thanks @semihbkgr!)

- Added parsing of a pod's "RESTART" column (e.g `5 (3m ago)`)
  so it auto updates, similarly to the "AGE" column. (#56)

- Added timer on pod's "STATUS" column when a pod is deleted
  (e.g `Deleted (3m ago)`). (#56)

- Added auto updating on event's "LAST SEEN" column. (#58)

- Added auto updating on job's "DURATION" column. (#60)

- Added auto updating on cronjob's "LAST SCHEDULE" column. (#60)

- Added coloring on event's "REASON" column. (#58)

- Fixed glitches when using flag `--watch-kubeconfig` / `-W`.
  The watch was not properly restarting, but works great now. (#57)

## v0.4.0 (2023-09-03)

- Added text filtering. (#32, thanks @semihbkgr!)

- Added toggles info to status bar. (7298803)

- Updated Go from v1.20 to v1.21. (c71519b)

- Updated k8s.io dependencies from v0.27.4 to v0.28.0. (514a851)

## v0.3.2 (2023-09-19)

- Fixed some cells being printed as `<nil>`. Now they are printed as empty
  cells instead. (#23, thanks @semihbkgr!)

## v0.3.1 (2023-08-05)

- Fixed namespace column when using `--all-namespaces` not rendering or
  sorting properly (fd3b165)

## v0.3.0 (2023-08-05)

- Added support for `exec` auth plugin. (5f122fa, #13)

- Fixed spinner still showing when there's no results. (edec13c, #18)

- Fixed not redrawing on update, but instead only on age timer tick.
  Now it redraws immediately. (f65ef48, #20)

- Fixed namespace column not showing when using `--all-namespaces`.
  (8b5a172, #21)

- Fixed non-namespaced resources' `CREATED AT` column not showing as duration.
  (d1323de)

## v0.2.0 (2023-04-28)

- Now available on krew index!

  ```bash
  kubectl krew install klock
  kubectl klock pods
  ```

- Added warning color on pod restarts when >0. (7f7a1b9)

- Added flag `--watch-kubeconfig` / `-W` to restart the watch when your
  kubeconfig (e.g `~/.kube/config`) file was changed. Such as when changing
  context via `kubectl config use-context NAME` or
  [kubectx](https://github.com/ahmetb/kubectx). (25a1f97)

- Added keybinding `f` to toggle fullscreen manually. (30639c5)

- Fixed bug where kubectl-klock would panic after ~50min of inactivity. (#12)

- Added better help text in README and `--help`. (9aa6d22)

- Added support for ARM architecture. All supported architectures: (246223c)

  - darwin/386
  - darwin/amd64
  - darwin/arm64 *(new!)*
  - linux/386
  - linux/amd64
  - linux/arm64 *(new!)*
  - windows/386
  - windows/amd64
  - windows/arm64 *(new!)*

## v0.1.1 (2023-04-17)

- Added `--output wide` / `-o wide` support. (18beed4)

## v0.1.0 (2023-04-16)

- Initial release. Features:

  - Watch arbitrary resources, just like `kubectl get <resource> [name]`

  - Auto updating age

  - Colors on statuses (e.g `Running`) and fractions (e.g `1/1`) to make them
    stand out more

- Supported architectures:

  - darwin/386
  - darwin/amd64
  - linux/386
  - linux/amd64
  - windows/386
  - windows/amd64
