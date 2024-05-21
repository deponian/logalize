package main

import (
	"embed"

	"github.com/deponian/logalize/cmd"
)

var (
	version string = "0.1.0"
	commit  string = "latest"
	date    string = "2024-05-12"
)

//go:embed builtins/*
var builtins embed.FS

func main() {
	cmd.Execute(builtins, version, commit, date)
}
