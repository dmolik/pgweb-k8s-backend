package main

import (
	"github.com/dmolik/pgweb-k8s-backend/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
