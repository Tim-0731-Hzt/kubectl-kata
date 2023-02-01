package main

import (
	"github.com/Tim-0731-Hzt/kubectl-kata/cmd/plugin/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	cli.InitAndExecute()
}
