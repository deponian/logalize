package main

import (
	"embed"
	"os"

	"github.com/deponian/logalize/cmd"
)

func main() {
	var builtins embed.FS
	cmd.Init(builtins, "", "", "")

	shell := os.Args[1]

	switch shell {
	case "bash":
		cmd.LogalizeCmd.Root().GenBashCompletionV2(os.Stdout, true)
	case "zsh":
		cmd.LogalizeCmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		cmd.LogalizeCmd.Root().GenFishCompletion(os.Stdout, true)
	}
}
