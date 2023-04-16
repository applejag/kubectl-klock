# klock kubectl

A `kubectl` plugin to render the `kubectl get pods --watch` output in a
much more readable fashion.

Think of it as running `watch kubectl get pods`, but instead of polling,
it uses the regular watch feature to stream updates as soon as they occur.

![demo](./docs/kubectl-klock-demo.gif)

## Quick Start

Requires Go 1.20 (or later) installed.

```sh
go install github.com/jilleJr/kubectl-klock@latest

kubectl klock pods
```

