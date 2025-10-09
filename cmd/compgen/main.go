package main

import (
	"embed"
	"log"
	"os"

	cmd "github.com/deponian/logalize/cmd/logalize"
)

func main() {
	var builtins embed.FS
	logalizeCmd := cmd.NewCommand(builtins)

	shell := os.Args[1]

	switch shell {
	case "bash":
		err := logalizeCmd.Root().GenBashCompletionV2(os.Stdout, true)
		if err != nil {
			log.Fatal(err)
		}
	case "zsh":
		err := logalizeCmd.Root().GenZshCompletion(os.Stdout)
		if err != nil {
			log.Fatal(err)
		}
	case "fish":
		err := logalizeCmd.Root().GenFishCompletion(os.Stdout, true)
		if err != nil {
			log.Fatal(err)
		}
	}
}
