package logalize

import (
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// Settings stores the values of command-line and config options
type Settings struct {
	ConfigPath string // path to configuration file

	Theme string // the name of the theme to be used

	NoBuiltinLogFormats bool // disable built-in log formats
	NoBuiltinPatterns   bool // disable built-in patterns
	NoBuiltinWords      bool // disable built-in words
	NoBuiltins          bool // disable built-in log formats, patterns and words

	HighlightOnlyLogFormats bool // highlight only log formats
	HighlightOnlyPatterns   bool // highlight only patterns
	HighlightOnlyWords      bool // highlight only words
	DryRun                  bool // highlight nothing
}

var Opts Settings

func InitSettings(flags *pflag.FlagSet) error {
	config := koanf.New(".")

	// read settings from default paths
	if err := loadConfig(config, defaultConfigPaths, true); err != nil {
		return err
	}

	// read settings from user defined path
	userConfig, _ := flags.GetString("config")
	if err := loadConfig(config, []string{userConfig}, false); err != nil {
		return err
	}

	// read settings from ./.logalize.yaml
	if err := loadConfig(config, []string{"./.logalize.yaml"}, true); err != nil {
		return err
	}

	// build settings step by step
	// first get defaults, then override with values from configs
	// then override with everything we get from flags
	opts := getBuiltinSettings()
	opts = getSettingFromConfig(opts, config)
	opts = getSettingFromFlags(opts, flags)

	Opts = opts

	return nil
}

func getBuiltinSettings() Settings {
	return Settings{
		ConfigPath: "",

		Theme: "tokyonight-dark",

		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      false,
		NoBuiltins:          false,

		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      false,
		DryRun:                  false,
	}
}

func getSettingFromConfig(opts Settings, config *koanf.Koanf) Settings {
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

	return opts
}

func getSettingFromFlags(opts Settings, flags *pflag.FlagSet) Settings {
	if flags.Changed("config") {
		opts.ConfigPath, _ = flags.GetString("config")
	}

	if flags.Changed("theme") {
		opts.Theme, _ = flags.GetString("theme")
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

	return opts
}
