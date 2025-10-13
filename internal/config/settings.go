package config

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
	if err := loadUserConfigs(config, flags); err != nil {
		return Settings{}, err
	}

	// build options step by step
	// first get defaults, then override with values from configs
	// then override with everything we get from flags
	opts := getBuiltinSettings()
	opts = getSettingFromConfig(opts, config)
	opts = getSettingFromFlags(opts, flags)

	// load (on not) the built-in configuration based on options
	config, err := loadBuiltinConfigs(config, builtins, opts)
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

func loadUserConfigs(config *koanf.Koanf, flags *pflag.FlagSet) error {
	loadConfig := func(paths []string, ignoreNonExistent bool) error {
		for _, path := range paths {
			err := config.Load(file.Provider(path), yaml.Parser())
			// ignore only errors about non-existent files
			if err != nil && !(errors.Is(err, os.ErrNotExist) && ignoreNonExistent) {
				return err
			}
		}
		return nil
	}

	// read configuration from default paths
	if err := loadConfig(getDefaultConfigPaths(), true); err != nil {
		return err
	}

	// read configuration from ./.logalize.yaml
	if err := loadConfig([]string{"./.logalize.yaml"}, true); err != nil {
		return err
	}

	// read configuration from user defined path(s)
	paths, _ := flags.GetStringArray("config")
	if err := loadConfig(paths, false); err != nil {
		return err
	}

	return nil
}

func getDefaultConfigPaths() []string {
	homeDir, _ := os.UserHomeDir()
	return []string{
		"/etc/logalize/logalize.yaml",
		homeDir + "/.config/logalize/logalize.yaml",
	}
}

func loadBuiltinConfigs(main *koanf.Koanf, builtins fs.FS, opts Options) (*koanf.Koanf, error) {
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
