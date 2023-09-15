<!--
SPDX-FileCopyrightText: 2023 Kalle Fagerberg

SPDX-License-Identifier: CC-BY-4.0
-->

# kubectl-klock

![demonstration animation](docs/demo.gif)

[![Latest Release](https://img.shields.io/github/release/jilleJr/kubectl-klock.svg)](https://github.com/jilleJr/kubectl-klock/releases)
[![REUSE status](https://api.reuse.software/badge/github.com/jilleJr/kubectl-klock)](https://api.reuse.software/info/github.com/jilleJr/kubectl-klock)

A `kubectl` plugin to render the `kubectl get pods --watch` output in a
much more readable fashion.

Think of it as running `watch kubectl get pods`, but instead of polling,
it uses the regular watch feature to stream updates as soon as they occur.

## Installation

[![Packaging status](https://repology.org/badge/vertical-allrepos/kubectl-klock.svg)](https://repology.org/project/kubectl-klock/versions)

<details open>
<summary><h3>Krew</h3></summary>

Can be installed using the krew kubectl plugin manager: <https://krew.sigs.k8s.io/>

```sh
kubectl krew install klock
kubectl klock pods
```

</details>

<details>
<summary><h3>Nix</h3></summary>

Also packaged as a Nix package: <https://search.nixos.org/packages?channel=unstable&show=kubectl-klock>

Try it out with nix-shell:

```sh
nix-shell -p kubectl-klock
kubectl klock pods
```

> [!IMPORTANT]
> It has not reached the stable channel yet, so requires using the unstable
> Nixpkgs channel.

</details>

<details>
<summary><h3>Prebuilt binaries</h3></summary>

You can download prebuilt binaries from the latest GitHub release: <https://github.com/jilleJr/kubectl-klock/releases/latest>

Download the one that fits your OS and architecture, extract the
tarball/zip file, and move the `klock` binary to somewhere in your PATH.
For example:

```sh
tar -xzf klock_linux_amd64.tar.gz
sudo mv ./klock /usr/local/bin
klock pods
```

For it to work as a subcommand to `kubectl`, it must be called `kubectl-klock`
instead. If you want that, then rename it.

```sh
sudo mv /usr/local/bin/klock /usr/local/bin/kubectl-klock
kubectl klock pods
```

</details>

<details>
<summary><h3>From source</h3></summary>

If you have Go installed, then you can use `go install` to let Go download
and build kubectl-klock for you:

```sh
go install github.com/jilleJr/kubectl-klock@latest
kubectl klock pods
```

Requires Go 1.21 (or later)

</details>

## Usage

Supports a wide range of flags

```sh
kubectl klock <resource> [name(s)] [flags]
```

### Examples

```sh
# Watch all pods
kubectl klock pods

# Watch all pods with more information (such as node name)
kubectl klock pods -o wide

# Watch a specific pod
kubectl klock pods my-pod-7d68885db5-6dfst

# Watch a subset of pods, filtering on labels
kubectl klock pods --selector app=my-app
kubectl klock pods -l app=my-app

# Watch all pods in all namespaces
kubectl klock pods --all-namespaces
kubectl klock pods -A

# Watch other resource types
kubectl klock cronjobs
kubectl klock deployments
kubectl klock statefulsets
kubectl klock nodes

# Watch all pods, but restart the watch when your ~/.kube/config file changes,
# such as when using "kubectl config use-context NAME"
kubectl klock pods --watch-kubeconfig
kubectl klock pods -W
```

There's also some hotkeys available:

```text
  →/l/pgdn next page      d     show/hide deleted               ctrl+c quit
  ←/h/pgup prev page      f     toggle fullscreen               ?/esc  close help
  g/home   go to start    /     filter by text
  G/end    go to end      enter close the filter input field
                          esc   clear the applied filter
```

## Features

- Pagination, for when the terminal window gets too small (height-wise)

- Same output format as `kubectl get`

- Watch arbitrary resources, just like `kubectl get <resource> [name]`

- Filter results

- Auto updating age column.

- Colors on statuses (e.g `Running`) and fractions (e.g `1/1`) to make
  them stand out more.

- Restart watch when kubeconfig file changes (flag: `--watch-kubeconfig`, `-W`),
  such as when changed by [kubectx](https://github.com/ahmetb/kubectx).
