package logalize

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

func BuildConfig(builtins fs.FS, opts Settings) (*koanf.Koanf, error) {
	config := koanf.New(".")

	// load built-in configuration
	if err := loadBuiltinConfig(config, builtins, opts); err != nil {
		return nil, err
	}

	// read configuration from default paths
	if err := loadConfig(config, defaultConfigPaths, true); err != nil {
		return nil, err
	}

	// read configuration from ./.logalize.yaml
	if err := loadConfig(config, []string{"./.logalize.yaml"}, true); err != nil {
		return nil, err
	}

	// read configuration from user defined path(s)
	if err := loadConfig(config, opts.ConfigPaths, false); err != nil {
		return nil, err
	}

	// check theme availability
	if !config.Exists("themes." + opts.Theme) {
		return nil, fmt.Errorf("Theme \"%s\" is not defined. Use -T/--list-themes flag to see the list of all available themes", opts.Theme)
	}

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

	return config, nil
}

func loadBuiltinConfig(config *koanf.Koanf, builtins fs.FS, opts Settings) error {
	var loadFromDirRecursively func(entries []fs.DirEntry, path string) error
	loadFromDirRecursively = func(entries []fs.DirEntry, path string) error {
		for _, entry := range entries {
			if entry.IsDir() {
				dir, _ := fs.ReadDir(builtins, path+entry.Name())
				if err := loadFromDirRecursively(dir, path+entry.Name()+"/"); err != nil {
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

	// read main configuration file
	builtinsDir, _ := fs.ReadDir(builtins, "builtins")
	if err := loadFromDirRecursively(builtinsDir, "builtins/"); err != nil {
		return err
	}

	// read theme files
	themesDir, _ := fs.ReadDir(builtins, "themes")
	if err := loadFromDirRecursively(themesDir, "themes/"); err != nil {
		return err
	}

	if opts.NoBuiltinLogFormats {
		config.Delete("formats")
	}

	if opts.NoBuiltinPatterns {
		config.Delete("patterns")
	}

	if opts.NoBuiltinWords {
		config.Delete("words")
	}

	if opts.NoBuiltins {
		config.Delete("formats")
		config.Delete("patterns")
		config.Delete("words")
	}

	return nil
}

func loadConfig(config *koanf.Koanf, paths []string, ignoreNonExistent bool) error {
	for _, path := range paths {
		if ok, err := checkFileIsReadable(path); ok {
			if err := config.Load(file.Provider(path), yaml.Parser()); err != nil {
				return err
			}
			// ignore only errors about non-existent files
		} else if !(os.IsNotExist(err) && ignoreNonExistent) {
			return err
		}
	}
	return nil
}

func checkFileIsReadable(filePath string) (bool, error) {
	_, err := os.Open(filePath)
	return err == nil, err
}
