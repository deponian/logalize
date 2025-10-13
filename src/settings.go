package logalize

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

type Settings struct {
	Config *koanf.Koanf
	Opts   Options
}

func NewSettings(builtins fs.FS, flags *pflag.FlagSet) (Settings, error) {
	config := koanf.New(".")

	// read configuration from default paths
	if err := loadConfig(config, defaultConfigPaths, true); err != nil {
		return Settings{}, err
	}

	// read configuration from ./.logalize.yaml
	if err := loadConfig(config, []string{"./.logalize.yaml"}, true); err != nil {
		return Settings{}, err
	}

	// read configuration from user defined path(s)
	userConfigs, _ := flags.GetStringArray("config")
	if err := loadConfig(config, userConfigs, false); err != nil {
		return Settings{}, err
	}

	// build options step by step
	// first get defaults, then override with values from configs
	// then override with everything we get from flags
	opts := getBuiltinSettings()
	opts = getSettingFromConfig(opts, config)
	opts = getSettingFromFlags(opts, flags)

	// load (on not) the built-in configuration based on options
	config, err := loadBuiltinConfig(config, builtins, opts)
	if err != nil {
		return Settings{}, err
	}

	// check theme availability
	if !config.Exists("themes." + opts.Theme) {
		return Settings{}, fmt.Errorf("Theme \"%s\" is not defined. Use -T/--list-themes flag to see the list of all available themes", opts.Theme)
	}
	config.MergeAt(config.Cut("themes."+opts.Theme), "theme")
	config.Delete("themes")

	// keep in the config only things we want to colorize
	if opts.HighlightOnlyLogFormats || opts.HighlightOnlyPatterns || opts.HighlightOnlyWords {
		configBackup := config.Copy()

		config.Delete("formats")
		config.Delete("patterns")
		config.Delete("words")

		if opts.HighlightOnlyLogFormats {
			config.MergeAt(configBackup.Cut("formats"), "formats")
		}
		if opts.HighlightOnlyPatterns {
			config.MergeAt(configBackup.Cut("patterns"), "patterns")
		}
		if opts.HighlightOnlyWords {
			config.MergeAt(configBackup.Cut("words"), "words")
		}
	}

	return Settings{Config: config, Opts: opts}, nil
}

func loadConfig(config *koanf.Koanf, paths []string, ignoreNonExistent bool) error {
	for _, path := range paths {
		err := config.Load(file.Provider(path), yaml.Parser())
		// ignore only errors about non-existent files
		if err != nil && !(errors.Is(err, os.ErrNotExist) && ignoreNonExistent) {
			return err
		}
	}
	return nil
}

func loadBuiltinConfig(main *koanf.Koanf, builtins fs.FS, opts Options) (*koanf.Koanf, error) {
	var loadRecursively func(config *koanf.Koanf, entries []fs.DirEntry, path string) error
	loadRecursively = func(config *koanf.Koanf, entries []fs.DirEntry, path string) error {
		for _, entry := range entries {
			if entry.IsDir() {
				dir, _ := fs.ReadDir(builtins, path+entry.Name())
				if err := loadRecursively(config, dir, path+entry.Name()+"/"); err != nil {
					return err
				}
			} else {
				file, _ := fs.ReadFile(builtins, path+entry.Name())
				if err := config.Load(rawbytes.Provider(file), yaml.Parser()); err != nil {
					return err
				}
			}
		}
		return nil
	}

	builtinConfig := koanf.New(".")

	// read theme files
	themesDir, _ := fs.ReadDir(builtins, "themes")
	if err := loadRecursively(builtinConfig, themesDir, "themes/"); err != nil {
		return nil, err
	}

	// read main configuration file
	builtinsDir, _ := fs.ReadDir(builtins, "builtins")
	if err := loadRecursively(builtinConfig, builtinsDir, "builtins/"); err != nil {
		return nil, err
	}

	if opts.NoBuiltinLogFormats {
		builtinConfig.Delete("formats")
	}

	if opts.NoBuiltinPatterns {
		builtinConfig.Delete("patterns")
	}

	if opts.NoBuiltinWords {
		builtinConfig.Delete("words")
	}

	if opts.NoBuiltins {
		builtinConfig.Delete("formats")
		builtinConfig.Delete("patterns")
		builtinConfig.Delete("words")
	}

	// apply main config on top of builtinConfig
	if err := builtinConfig.Merge(main); err != nil {
		return nil, err
	}

	return builtinConfig, nil
}

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
