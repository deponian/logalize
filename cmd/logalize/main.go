package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/deponian/logalize/internal/config"
	"github.com/deponian/logalize/internal/core"
	"github.com/spf13/cobra"
)

var LogalizeCmd *cobra.Command

func Init(builtins embed.FS, version, commit, date string) {
	var printBuiltinsFlag bool
	var printConfigFlag bool
	var listThemesFlag bool

	LogalizeCmd = &cobra.Command{
		Use:   "logalize",
		Short: "fast and extensible log colorizer",
		Long: `Logalize is a log colorizer.
It's fast and extensible alternative to ccze and colorize.`,
		Version: fmt.Sprintf("%s (%s) %s", version, commit, date),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// print built-in log formats and words and exit
			if printBuiltinsFlag {
				err := printBuiltins(builtins)
				if err != nil {
					log.Fatal(err)
				} else {
					os.Exit(0)
				}
			}

			// build config
			settings, err := config.NewSettings(builtins, LogalizeCmd.Flags())
			if err != nil {
				log.Fatal(err)
			}

			// print config
			if printConfigFlag {
				fmt.Print(settings.PrintConfig())
				os.Exit(0)
			}

			// list themes
			if listThemesFlag {
				fmt.Printf(settings.PrintThemes())
				os.Exit(0)
			}

			// run the app
			err = core.Run(os.Stdin, os.Stdout, settings)
			if err != nil {
				log.Fatal(err)
			}
		},
		DisableAutoGenTag: true,
	}

	// these flags are used outside of the logalize package
	LogalizeCmd.Flags().BoolVarP(&printConfigFlag, "print-config", "C", false, "print full configuration file")
	LogalizeCmd.Flags().BoolVarP(&listThemesFlag, "list-themes", "T", false, "display a list of all available themes")
	LogalizeCmd.Flags().BoolVarP(&printBuiltinsFlag, "print-builtins", "b", false, "print built-in log formats, patterns and words as separate YAML files")

	// these flags are used inside the logalize package
	// they will be processed by InitSettings()
	LogalizeCmd.Flags().StringArrayP("config", "c", []string{}, "path to user configuration file (can be repeated)")
	LogalizeCmd.Flags().StringP("theme", "t", "tokyonight-dark", "set the theme")

	LogalizeCmd.Flags().BoolP("debug", "d", false, "add debug info to the output")

	LogalizeCmd.Flags().BoolP("no-builtin-logformats", "L", false, "disable built-in log formats highlighting")
	LogalizeCmd.Flags().BoolP("no-builtin-patterns", "P", false, "disable built-in patterns highlighting")
	LogalizeCmd.Flags().BoolP("no-builtin-words", "W", false, "disable built-in words highlighting")
	LogalizeCmd.Flags().BoolP("no-builtins", "N", false, "disable built-in log formats, patterns and words highlighting")

	LogalizeCmd.Flags().BoolP("only-logformats", "l", false, "highlight only log formats (can be combined with -p and -w)")
	LogalizeCmd.Flags().BoolP("only-patterns", "p", false, "highlight only patterns (can be combined with -l and -w)")
	LogalizeCmd.Flags().BoolP("only-words", "w", false, "highlight only words (can be combined with -l and -p)")
	LogalizeCmd.Flags().BoolP("dry-run", "n", false, "don't alter the input in any way")

	LogalizeCmd.Flags().BoolP("no-ansi-escape-sequences-stripping", "s", false, "disable removing of ANSI escape sequences (save input colors)")
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
