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

var LogalizeCmd *cobra.Command

func Init(builtins embed.FS, version, commit, date string) {
	options := logalize.Options{}

	LogalizeCmd = &cobra.Command{
		Use:   "logalize",
		Short: "fast and extensible log colorizer",
		Long: `Logalize is log colorizer.
It's fast and extensible alternative to ccze and colorize.`,
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
		DisableAutoGenTag: true,
	}

	LogalizeCmd.Flags().StringVarP(&options.ConfigPath, "config", "c", "", "path to configuration file")
	LogalizeCmd.Flags().BoolVarP(&options.NoBuiltins, "no-builtins", "n", false, "disable built-in log formats and words")
}

func Execute() {
	err := LogalizeCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
