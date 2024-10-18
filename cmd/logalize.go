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
	"github.com/goccy/go-yaml"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

var LogalizeCmd *cobra.Command

func Init(builtins embed.FS, version, commit, date string) {
	options := logalize.Settings{}

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
			err = logalize.InitSettings(LogalizeCmd.Flags())
			if err != nil {
				log.Fatal(err)
			}
			err = logalize.InitConfig(builtins)
			if err != nil {
				log.Fatal(err)
			}

			// print config
			if options.PrintConfig {
				printConfig(logalize.Config)
				os.Exit(0)
			}

			// list themes
			if options.ListThemes {
				listThemes(logalize.Config)
				os.Exit(0)
			}

			// run the app
			err = logalize.Run(os.Stdin, os.Stdout, lemmatizer)
			if err != nil {
				log.Fatal(err)
			}
		},
		DisableAutoGenTag: true,
	}

	LogalizeCmd.Flags().StringVarP(&options.ConfigPath, "config", "c", "", "path to user configuration file")
	LogalizeCmd.Flags().BoolVarP(&options.PrintConfig, "print-config", "C", false, "print full configuration file")
	LogalizeCmd.Flags().StringVarP(&options.Theme, "theme", "t", "tokyonight", "set the theme")
	LogalizeCmd.Flags().BoolVarP(&options.ListThemes, "list-themes", "T", false, "display a list of all available themes")

	LogalizeCmd.Flags().BoolVarP(&options.PrintBuiltins, "print-builtins", "b", false, "print built-in log formats, patterns and words as separate YAML files")

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

func listThemes(config *koanf.Koanf) {
	themes := config.MapKeys("themes")
	if len(themes) == 0 {
		fmt.Println("There are no themes available")
	} else {
		fmt.Println("Available themes:")
		for _, theme := range themes {
			fmt.Printf("  - %s\n", theme)
		}
		fmt.Printf("\nUse one of these with -t/--theme flag\n")
	}
}

// custom YAML parser for koanf
// to print the config indented by two spaces instead of four
type YAML struct{}

func (p *YAML) Unmarshal(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *YAML) Marshal(o map[string]interface{}) ([]byte, error) {
	return yaml.MarshalWithOptions(o, yaml.Indent(2), yaml.IndentSequence(true))
}

func printConfig(config *koanf.Koanf) {
	configBytes, _ := config.Marshal(&YAML{})
	fmt.Print(string(configBytes))
}
