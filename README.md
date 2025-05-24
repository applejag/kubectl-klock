<!--
SPDX-FileCopyrightText: 2023 Kalle Fagerberg

SPDX-License-Identifier: CC-BY-4.0
-->

# kubectl-klock

![demonstration animation](docs/demo.gif)

[![Latest Release](https://img.shields.io/github/release/applejag/kubectl-klock.svg)](https://github.com/applejag/kubectl-klock/releases)
[![REUSE status](https://api.reuse.software/badge/github.com/applejag/kubectl-klock)](https://api.reuse.software/info/github.com/applejag/kubectl-klock)

A `kubectl` plugin to render the `kubectl get pods --watch` output in a
much more readable fashion.

Think of it as running `watch kubectl get pods`, but instead of polling,
it uses the regular watch feature to stream updates as soon as they occur.

## Installation

### Krew

[![krew status](https://img.shields.io/badge/dynamic/yaml?url=https%3A%2F%2Fgithub.com%2Fkubernetes-sigs%2Fkrew-index%2Fraw%2Fmaster%2Fplugins%2Fklock.yaml&query=spec.version&logo=data%3Aimage%2Fpng%3Bbase64%2CiVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAACPVBMVEVHcEwRR3sHR4B6ruEEP3UkW48QRnssZp1Vir4TToQlW5AYV5A1cqoEP3ZXjMD%2F%2F%2F8TToUtZ50HSIEKT4uCncErdbIYaa0hY50tWHwGUY93te0FLk0onfcFL08Rab8SkO8tV3oMl%2F4fnv4hXaIqov8Qa8IjW5wDK0wDL1MPju45qP7%2F9%2F9Pr%2F0SXpwXY6Fck8dTl9Ignv0Qmf8ETosJQnk0fLkkb60RWJIPk%2FYsWX4TQmYRRG1Wo%2BRctv8lVXwHZKkEM1dWh7T4%2F%2F8KVJIxZZgFN2sgaqdYmdIIUY4Pcc4xcqooh9QdYLAFOWwEe9YUlfcqXZErX5IAQXoAK0sASogARoEATo4AKUkARoIAOnEAM2YAUpMAKEcAKEYAJkQAQHkALE0ANGcAU5YAS4gBRoIBTIkAXaYAVJgBPGwAarwAWaABLE4BMVYAUpUAgOABRHkBR4MBK04BSocBOG0BQ30BN2sBSYQANmkAetcAcccAKUgAK00AO3EBNWABiOwAPXQALlUAVpkAWZ0AKEoATZAAJ0UAZbMEkfoAK0wAYqwASYgAKkoARH8EkvoAOnAAPGsBP3cBN2wAMmQAU5gAOmYAM2cAQXsAN2wAN2AAcMMANmoAhegANFsANmAARHcANWkAg%2BQAO3AAMmUAU5oAfdkANmsAabwAOW8BSIQAQnoAZa8AMVcAOWQAgeAAVZUAMFUALU8Ai%2FIATpAAP3cAQXwAft4AcssASIUAg%2BYAb8UANGgCN20AOnMAZLQAd9LE9of%2BAAAAUnRSTlMA6fcJ%2FH7qUxbDe6w8%2BxUCwFH27QhUqYRZ9wn5Xvnt41nlm5Zm8o%2F%2B%2FvQvARTErQwXgc389T2A7NVSwesVEH31%2BxUC7Uz7hg32%2BkxzrPv92GBgoriJcQAAAQFJREFUGNNjYAABPi5GNlYOBhgQERTgjYxk4WRnBvG0tWTEhOKDQICJm4efgUFFsb2uMjc0NCIiIjRFWFyfQb6zsa0iLzM9JCQkOaTHxIVBrrm1tiwrLSksLKxownzXAAY19bi47IzU6OiwkolRUb6BDArSsTGxMTn5BcWTFiUm7vZhkJANB4KqwtKWxVt2Re3xZFDWCAaC7uryrmVbd%2B7d58%2Bg2tCREJwwp79m6tJN23b4eTNoNk2ZMW%2FVkpmTI6cv37zdw4tBqXf2go1r182d1Vc%2FbYWzmzuDjqGRhf2aDesXxodK6pk6gXxja2Np5bB6pbGuqBTcv4521uZmBmAmAERnUiB8Vh3oAAAAAElFTkSuQmCC&label=krew)](https://krew.sigs.k8s.io/plugins/)

```sh
kubectl krew install klock
```

### Snap

[![klock](https://snapcraft.io/klock/badge.svg)](https://snapcraft.io/klock)

```sh
sudo snap install klock
```

### Scoop

![Scoop](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fgithub.com%2Fapplejag%2Fapplejag-bucket%2Fraw%2Fmaster%2Fbucket%2Fkubectl-klock.json&query=%24.version&logo=scoop&label=applejag-bucket%2Fkubectl-klock)

```pwsh
scoop bucket add applejag https://github.com/applejag/applejag-bucket
scoop install applejag/kubectl-klock
```

### Nix

[![Packaging status](https://repology.org/badge/vertical-allrepos/kubectl-klock.svg?header=)](https://repology.org/project/kubectl-klock/versions)

```sh
nix-shell -p kubectl-klock
```

### Pre-built binaries

You can download pre-built binaries from the latest GitHub release: <https://github.com/applejag/kubectl-klock/releases/latest>

Download the one that fits your OS and architecture, extract the
tarball/zip file, and move the `kubectl-klock` binary to somewhere in your PATH.
For example:

```sh
tar -xzf kubectl-klock_linux_amd64.tar.gz
sudo mv ./kubectl-klock /usr/local/bin
```

### From source

Requires Go 1.21 (or later).

```sh
go install github.com/applejag/kubectl-klock@latest
```

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
  →/l/pgdn next page      /        filter by text                  ctrl+c quit
  ←/h/pgup prev page      enter    close the filter input field    ?/esc  close help
  g/home   go to start    esc      clear the applied filter        d      show/hide deleted
  G/end    go to end      ↓/ctrl+n show next suggestion            f      toggle fullscreen
                          ↑/ctrl+p show previous suggestion
                          tab      accept a suggestion
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

- Color themes powered by [kubecolor](https://kubecolor.github.io/)

### Color themes

Klock uses kubecolor's coloring logic and behavior when coloring its output.
See: <https://kubecolor.github.io/customizing/themes/>

Color settings that klock uses:

- `KUBECOLOR_THEME_BASE_DANGER` for rows with errors
- `KUBECOLOR_THEME_BASE_MUTED` for "No resources found"
- `KUBECOLOR_THEME_BASE_MUTED` for deleted rows
- `KUBECOLOR_THEME_BASE_MUTED` for status line
- `KUBECOLOR_THEME_BASE_SECONDARY` for "FILTER:" prompt
- `KUBECOLOR_THEME_BASE_WARNING` for "No resources visible" when filtering
- `KUBECOLOR_THEME_DATA_DURATIONFRESH` for `AGE: 12h` when below threshold
- `KUBECOLOR_THEME_DATA_RATIO_EQUAL` for `READY: 1/1`
- `KUBECOLOR_THEME_DATA_RATIO_UNEQUAL` for `READY: 0/1`
- `KUBECOLOR_THEME_STATUS_ERROR` for `STATUS: CrashLoopBackOff`
- `KUBECOLOR_THEME_STATUS_SUCCESS` for `STATUS: Running`
- `KUBECOLOR_THEME_STATUS_WARNING` for `STATUS: Terminating`
- `KUBECOLOR_THEME_TABLE_COLUMNS` for table columns
- `KUBECOLOR_THEME_TABLE_HEADER` for table header

You can configure these colors either via
[environment variables](https://kubecolor.github.io/reference/environment-variables/)
or via the [`~/.kube/color.yaml` config file](https://kubecolor.github.io/reference/config/)

### Completion

To get completion when writing `kubectl klock`, you need to add
[`./bin/kubectl_complete-klock`](./bin/kubectl_complete-klock)
to your `PATH`.

For example:

```sh
sudo curl https://github.com/applejag/kubectl-klock/raw/main/bin/kubectl_complete-klock -o /usr/local/bin/kubectl_complete-klock
sudo chmod +x /usr/local/bin/kubectl_complete-klock
```
