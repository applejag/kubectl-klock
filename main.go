package main

import (
	"github.com/jilleJr/kubectl-klock/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	cmd.InitAndExecute()
}
