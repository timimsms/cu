package main

import (
	"os"

	"github.com/timimsms/cu/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
