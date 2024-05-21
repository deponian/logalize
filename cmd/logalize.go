package cmd

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	logalize "github.com/deponian/logalize/pkg"
	"github.com/spf13/cobra"
)

func Execute(builtins embed.FS, version, commit, date string) {
	options := logalize.Options{}

	logalizeCmd := &cobra.Command{
		Use:   "logalize",
		Short: "Fast and extensible log colorizer. Alternative to ccze",
		Long: `Logalize is fast and extensible log colorizer
Alternative to ccze and colorize`,
		Version: fmt.Sprintf("%s (%s) %s", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			// build config
			lemmatizer, err := golem.New(en.New())
			if err != nil {
				log.Fatal(err)
			}
			config, err := logalize.InitConfig(options, builtins)
			if err != nil {
				log.Fatal(err)
			}

			// run the app
			err = logalize.Run(os.Stdin, os.Stdout, config, builtins, lemmatizer)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	logalizeCmd.Flags().StringVarP(&options.ConfigPath, "config", "c", "", "path to configuration file")
	logalizeCmd.Flags().BoolVarP(&options.NoBuiltins, "no-builtins", "n", false, "disable built-in log formats and words")

	err := logalizeCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
