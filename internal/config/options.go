package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// Options stores the values of command-line and config options
type Options struct {
	ConfigPaths []string // path(s) to configuration file(s)

	Theme string // the name of the theme to be used

	Debug bool // add debug info to the output

	NoBuiltinLogFormats bool // disable built-in log formats
	NoBuiltinPatterns   bool // disable built-in patterns
	NoBuiltinWords      bool // disable built-in words
	NoBuiltins          bool // disable built-in log formats, patterns and words

	HighlightOnlyLogFormats bool // highlight only log formats
	HighlightOnlyPatterns   bool // highlight only patterns
	HighlightOnlyWords      bool // highlight only words

	DryRun bool // don't alter the input

	NoANSIEscapeSequencesStripping bool // disable removing of ANSI escape sequences from the input
}

func getBuiltinSettings() Options {
	return Options{
		ConfigPaths: []string{},

		Theme: "tokyonight-dark",

		Debug: false,

		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      false,
		NoBuiltins:          false,

		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      false,

		DryRun: false,

		NoANSIEscapeSequencesStripping: false,
	}
}

func getSettingFromConfig(opts Options, config *koanf.Koanf) Options {
	if config.Exists("settings.theme") {
		opts.Theme = config.String("settings.theme")
	}

	if config.Exists("settings.no-builtin-logformats") {
		opts.NoBuiltinLogFormats = config.Bool("settings.no-builtin-logformats")
	}
	if config.Exists("settings.no-builtin-patterns") {
		opts.NoBuiltinPatterns = config.Bool("settings.no-builtin-patterns")
	}
	if config.Exists("settings.no-builtin-words") {
		opts.NoBuiltinWords = config.Bool("settings.no-builtin-words")
	}
	if config.Exists("settings.no-builtins") {
		opts.NoBuiltins = config.Bool("settings.no-builtins")
	}

	if config.Exists("settings.only-logformats") {
		opts.HighlightOnlyLogFormats = config.Bool("settings.only-logformats")
	}
	if config.Exists("settings.only-patterns") {
		opts.HighlightOnlyPatterns = config.Bool("settings.only-patterns")
	}
	if config.Exists("settings.only-words") {
		opts.HighlightOnlyWords = config.Bool("settings.only-words")
	}

	if config.Exists("settings.no-ansi-escape-sequences-stripping") {
		opts.NoANSIEscapeSequencesStripping = config.Bool("settings.no-ansi-escape-sequences-stripping")
	}

	return opts
}

func getSettingFromFlags(opts Options, flags *pflag.FlagSet) Options {
	if flags.Changed("config") {
		opts.ConfigPaths, _ = flags.GetStringArray("config")
	}

	if flags.Changed("theme") {
		opts.Theme, _ = flags.GetString("theme")
	}

	if flags.Changed("debug") {
		opts.Debug, _ = flags.GetBool("debug")
	}

	if flags.Changed("no-builtin-logformats") {
		opts.NoBuiltinLogFormats, _ = flags.GetBool("no-builtin-logformats")
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

	if flags.Changed("only-logformats") {
		opts.HighlightOnlyLogFormats, _ = flags.GetBool("only-logformats")
	}
	if flags.Changed("only-patterns") {
		opts.HighlightOnlyPatterns, _ = flags.GetBool("only-patterns")
	}
	if flags.Changed("only-words") {
		opts.HighlightOnlyWords, _ = flags.GetBool("only-words")
	}
	if flags.Changed("dry-run") {
		opts.DryRun, _ = flags.GetBool("dry-run")
	}

	if flags.Changed("no-ansi-escape-sequences-stripping") {
		opts.NoANSIEscapeSequencesStripping, _ = flags.GetBool("no-ansi-escape-sequences-stripping")
	}

	return opts
}
