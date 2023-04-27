<!--
SPDX-FileCopyrightText: 2023 Kalle Fagerberg

SPDX-License-Identifier: CC-BY-4.0
-->

# klock kubectl

[![REUSE status](https://api.reuse.software/badge/github.com/jilleJr/kubectl-klock)](https://api.reuse.software/info/github.com/jilleJr/kubectl-klock)

A `kubectl` plugin to render the `kubectl get pods --watch` output in a
much more readable fashion.

Think of it as running `watch kubectl get pods`, but instead of polling,
it uses the regular watch feature to stream updates as soon as they occur.

## Quick Start

Requires Go 1.20 (or later) installed.

```sh
go install github.com/jilleJr/kubectl-klock@latest

kubectl klock pods
```

## Usage

Supports a wide range of flags

```sh
kubectl klock <resource> [name(s)] [flags]

# Examples:

kubectl klock pods

kubectl klock pods my-pod-7d68885db5-6dfst

kubectl klock pods --selector app=my-app
kubectl klock pods -l app=my-app

kubectl klock pods --all-namespaces
kubectl klock pods -A

kubectl klock cronjobs
kubectl klock deployments
kubectl klock statefulsets
kubectl klock nodes

kubectl klock pods --watch-kubeconfig
```

There's also some hotkeys available:

```text
  →/l/pgdn  next page      d  show/hide deleted    ctrl+c  quit
  ←/h/pgup  prev page                              ?       show help
  g/home    go to start
  G/end     go to end
```

## Features

- Pagination, for when the terminal window gets too small (height-wise)

- Same output format as `kubectl get`

- Watch arbitrary resources, just like `kubectl get <resource> [name]`

- Auto updating age column.

- Colors on statuses (e.g `Running`) and fractions (e.g `1/1`) to make
  them stand out more.

- Restart watch when kubeconfig file changes (flag: `--watch-kubeconfig`, `-W`),
  such as when changed by [kubectx](https://github.com/ahmetb/kubectx).