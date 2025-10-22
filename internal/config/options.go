// Package config manages options and settings.
package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// Options stores the values of command-line and config options.
type Options struct {
	ConfigPaths []string // path(s) to configuration file(s)

	Theme string // the name of the theme to be used

	Debug bool // add debug info to the output

	NoBuiltinFormats  bool // disable built-in formats
	NoBuiltinPatterns bool // disable built-in patterns
	NoBuiltinWords    bool // disable built-in words
	NoBuiltins        bool // disable built-in formats, patterns and words

	HighlightOnlyFormats  bool // highlight only formats
	HighlightOnlyPatterns bool // highlight only patterns
	HighlightOnlyWords    bool // highlight only words

	DryRun bool // don't alter the input

	NoANSIEscapeSequencesStripping bool // disable removing of ANSI escape sequences from the input

	PrintConfig   bool // print fully merged configuration file and exit the program
	PrintBuiltins bool // print built-in configuration and exit the program
	ListThemes    bool // print all available themes and exit the program
}

// NewOptions create new instance of Options with default values.
func NewOptions() *Options {
	return &Options{
		ConfigPaths: []string{},

		Theme: "tokyonight-dark",

		NoBuiltinFormats:  false,
		NoBuiltinPatterns: false,
		NoBuiltinWords:    false,
		NoBuiltins:        false,

		HighlightOnlyFormats:  false,
		HighlightOnlyPatterns: false,
		HighlightOnlyWords:    false,

		NoANSIEscapeSequencesStripping: false,

		Debug:  false,
		DryRun: false,

		PrintConfig:   false,
		PrintBuiltins: false,
		ListThemes:    false,
	}
}

// ReadFromConfig reads new options from configuration object
// and override corresponding fields in the *Option instance.
func (opts *Options) ReadFromConfig(cfg *koanf.Koanf) {
	if cfg == nil {
		return
	}

	if cfg.Exists("settings.theme") {
		opts.Theme = cfg.String("settings.theme")
	}

	if cfg.Exists("settings.no-builtin-formats") {
		opts.NoBuiltinFormats = cfg.Bool("settings.no-builtin-formats")
	}
	if cfg.Exists("settings.no-builtin-patterns") {
		opts.NoBuiltinPatterns = cfg.Bool("settings.no-builtin-patterns")
	}
	if cfg.Exists("settings.no-builtin-words") {
		opts.NoBuiltinWords = cfg.Bool("settings.no-builtin-words")
	}
	if cfg.Exists("settings.no-builtins") {
		opts.NoBuiltins = cfg.Bool("settings.no-builtins")
	}

	if cfg.Exists("settings.only-formats") {
		opts.HighlightOnlyFormats = cfg.Bool("settings.only-formats")
	}
	if cfg.Exists("settings.only-patterns") {
		opts.HighlightOnlyPatterns = cfg.Bool("settings.only-patterns")
	}
	if cfg.Exists("settings.only-words") {
		opts.HighlightOnlyWords = cfg.Bool("settings.only-words")
	}

	if cfg.Exists("settings.no-ansi-escape-sequences-stripping") {
		opts.NoANSIEscapeSequencesStripping = cfg.Bool("settings.no-ansi-escape-sequences-stripping")
	}

	if cfg.Exists("settings.debug") {
		opts.Debug = cfg.Bool("settings.debug")
	}
	if cfg.Exists("settings.dry-run") {
		opts.DryRun = cfg.Bool("settings.dry-run")
	}
}

// ReadFromFlags reads new options from command line flags
// and override corresponding fields in the *Option instance.
func (opts *Options) ReadFromFlags(flags *pflag.FlagSet) {
	if flags == nil {
		return
	}

	if flags.Changed("config") {
		opts.ConfigPaths, _ = flags.GetStringArray("config")
	}

	if flags.Changed("theme") {
		opts.Theme, _ = flags.GetString("theme")
	}

	if flags.Changed("no-builtin-formats") {
		opts.NoBuiltinFormats, _ = flags.GetBool("no-builtin-formats")
	}
	if flags.Changed("no-builtin-patterns") {
		opts.NoBuiltinPatterns, _ = flags.GetBool("no-builtin-patterns")
	}
	if flags.Changed("no-builtin-words") {
		opts.NoBuiltinWords, _ = flags.GetBool("no-builtin-words")
	}
	if flags.Changed("no-builtins") {
		opts.NoBuiltins, _ = flags.GetBool("no-builtins")
	}

	if flags.Changed("only-formats") {
		opts.HighlightOnlyFormats, _ = flags.GetBool("only-formats")
	}
	if flags.Changed("only-patterns") {
		opts.HighlightOnlyPatterns, _ = flags.GetBool("only-patterns")
	}
	if flags.Changed("only-words") {
		opts.HighlightOnlyWords, _ = flags.GetBool("only-words")
	}

	if flags.Changed("no-ansi-escape-sequences-stripping") {
		opts.NoANSIEscapeSequencesStripping, _ = flags.GetBool("no-ansi-escape-sequences-stripping")
	}

	if flags.Changed("debug") {
		opts.Debug, _ = flags.GetBool("debug")
	}
	if flags.Changed("dry-run") {
		opts.DryRun, _ = flags.GetBool("dry-run")
	}

	if flags.Changed("print-config") {
		opts.PrintConfig, _ = flags.GetBool("print-config")
	}
	if flags.Changed("print-builtins") {
		opts.PrintBuiltins, _ = flags.GetBool("print-builtins")
	}
	if flags.Changed("list-themes") {
		opts.ListThemes, _ = flags.GetBool("list-themes")
	}
}
