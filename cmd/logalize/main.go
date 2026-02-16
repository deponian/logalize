package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/deponian/logalize/internal/config"
	"github.com/deponian/logalize/internal/core"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var (
	version = "0.0.0"
	commit  = "latest"
	date    = "2024-05-12"
)

func NewCommand(builtins embed.FS) *cobra.Command {
	root := &cobra.Command{
		Use:   "logalize",
		Short: "fast and extensible log colorizer",
		Long: `Logalize is a log colorizer.
It's fast and extensible alternative to ccze and colorize.`,
		Version:      fmt.Sprintf("%s (%s) %s", version, commit, date),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// build user configuration from default paths and from --config flag
			paths, _ := cmd.Flags().GetStringArray("config")
			cfg, err := config.CreateUserConfig(paths)
			if err != nil {
				return err
			}

			// build application settings
			settings, err := config.NewSettings(builtins, cfg, cmd.Flags(), hasDarkBackground())
			if err != nil {
				return err
			}

			// check special flags like --print-config and --list-themes
			str, exit := settings.ProcessSpecialFlags()
			if exit {
				fmt.Print(str)

				return nil
			}

			// run the main loop
			err = core.Run(os.Stdin, os.Stdout, settings)
			if err != nil {
				return err
			}

			return nil
		},
		DisableAutoGenTag: true,
	}

	// these flags won't stop the program from running
	root.Flags().StringArrayP("config", "c", []string{}, "path to user configuration file (can be repeated)")
	root.Flags().StringP("theme", "t", "tokyonight-dark", "set the theme")

	root.Flags().BoolP("debug", "d", false, "add debug info to the output")

	root.Flags().BoolP("no-builtin-formats", "L", false, "disable built-in formats highlighting")
	root.Flags().BoolP("no-builtin-patterns", "P", false, "disable built-in patterns highlighting")
	root.Flags().BoolP("no-builtin-words", "W", false, "disable built-in words highlighting")
	root.Flags().BoolP("no-builtins", "N", false, "disable built-in formats, patterns and words highlighting")

	root.Flags().BoolP("only-formats", "f", false, "highlight only formats (can be combined with -p and -w)")
	root.Flags().BoolP("only-patterns", "p", false, "highlight only patterns (can be combined with -f and -w)")
	root.Flags().BoolP("only-words", "w", false, "highlight only words (can be combined with -f and -p)")
	root.Flags().BoolP("dry-run", "n", false, "don't alter the input in any way")

	root.Flags().BoolP("no-ansi-escape-sequences-stripping", "s", false, "disable removing of ANSI escape sequences (save input colors)")

	// these flags will print something and stop the program
	root.Flags().BoolP("print-config", "C", false, "print full configuration file")
	root.Flags().BoolP("list-themes", "T", false, "display a list of all available themes")
	root.Flags().BoolP("print-builtins", "B", false, "print built-in formats, patterns and words as separate YAML files")

	return root
}

// We need to query the terminal outside the main application package
// because if we include this code inside, it will be impossible to test
// it completely and achieve 100% coverage.
// Don't look at me like that.
func hasDarkBackground() bool {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return true
	}
	defer tty.Close()

	detector := termenv.NewOutput(tty)

	if detector.HasDarkBackground() {
		return true
	} else {
		return false
	}
}

func Run(builtins embed.FS) int {
	command := NewCommand(builtins)

	if err := command.Execute(); err != nil {
		return 1
	}

	return 0
}
