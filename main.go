package main

import (
	"embed"
	"os"

	cmd "github.com/deponian/logalize/cmd/logalize"
)

//go:embed builtins/*
//go:embed themes/*
var builtins embed.FS

func main() {
	os.Exit(cmd.Run(builtins))
}
