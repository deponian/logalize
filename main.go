package main

import (
	"embed"

	"github.com/deponian/logalize/cmd"
)

var (
	version string = "0.0.0"
	commit  string = "latest"
	date    string = "2024-05-12"
)

//go:embed builtins/*
//go:embed themes/*
var builtins embed.FS

func main() {
	cmd.Init(builtins, version, commit, date)
	cmd.Execute()
}
