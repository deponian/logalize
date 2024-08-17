package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

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
		Long: `Logalize is a log colorizer.
It's fast and extensible alternative to ccze and colorize.`,
		Version: fmt.Sprintf("%s (%s) %s", version, commit, date),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// print built-in log formats and words and exit
			if options.PrintBuiltins {
				err := printBuiltins(builtins)
				if err != nil {
					log.Fatal(err)
				} else {
					os.Exit(0)
				}
			}

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
			err = logalize.Run(os.Stdin, os.Stdout, config, lemmatizer)
			if err != nil {
				log.Fatal(err)
			}
		},
		DisableAutoGenTag: true,
	}

	LogalizeCmd.Flags().StringVarP(&options.ConfigPath, "config", "c", "", "path to configuration file")

	LogalizeCmd.Flags().BoolVarP(&options.PrintBuiltins, "print-builtins", "b", false, "print built-in log formats, patterns and words")

	LogalizeCmd.Flags().BoolVarP(&options.NoBuiltins, "no-builtins", "N", false, "disable built-in log formats, patterns and words highlighting")
	LogalizeCmd.Flags().BoolVarP(&options.NoBuiltinLogFormats, "no-builtin-logformats", "L", false, "disable built-in log formats highlighting")
	LogalizeCmd.Flags().BoolVarP(&options.NoBuiltinPatterns, "no-builtin-patterns", "P", false, "disable built-in patterns highlighting")
	LogalizeCmd.Flags().BoolVarP(&options.NoBuiltinWords, "no-builtin-words", "W", false, "disable built-in words highlighting")

	LogalizeCmd.Flags().BoolVarP(&options.DryRun, "dry-run", "n", false, "disable any colorization")
	LogalizeCmd.Flags().BoolVarP(&options.HighlightOnlyLogFormats, "only-logformats", "l", false, "highlight only log formats (can be combined with -p and -w)")
	LogalizeCmd.Flags().BoolVarP(&options.HighlightOnlyPatterns, "only-patterns", "p", false, "highlight only patterns (can be combined with -l and -w)")
	LogalizeCmd.Flags().BoolVarP(&options.HighlightOnlyWords, "only-words", "w", false, "highlight only words (can be combined with -l and -p)")
}

func Execute() {
	err := LogalizeCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func printBuiltins(builtins embed.FS) error {
	var printDirRecursively func(entries []fs.DirEntry, path string) error
	printDirRecursively = func(entries []fs.DirEntry, path string) error {
		for _, entry := range entries {
			if entry.IsDir() {
				dir, _ := builtins.ReadDir(path + entry.Name())
				if err := printDirRecursively(dir, path+entry.Name()+"/"); err != nil {
					return err
				}
			} else {
				filename := entry.Name()
				file, _ := builtins.ReadFile(path + filename)
				group := strings.Split(path, "/")[1]
				fmt.Printf("---\n# [%s] %s\n%v", group, filename, string(file))
			}
		}
		return nil
	}

	builtinsDir, _ := builtins.ReadDir("builtins")
	if err := printDirRecursively(builtinsDir, "builtins/"); err != nil {
		return err
	}

	return nil
}
