package logalize

import (
	"io/fs"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

// Options stores the values of command-line options
type Options struct {
	ConfigPath string // path to configuration file

	PrintBuiltins bool // print built-in log formats and words

	NoBuiltinLogFormats bool // disable built-in log formats
	NoBuiltinPatterns   bool // disable built-in patterns
	NoBuiltinWords      bool // disable built-in words
	NoBuiltins          bool // disable built-in log formats, patterns and words

	HighlightOnlyLogFormats bool // highlight only log formats
	HighlightOnlyPatterns   bool // highlight only patterns
	HighlightOnlyWords      bool // highlight only words
	DryRun                  bool // highlight nothing
}

var Opts Options

func InitConfig(opts Options, builtins fs.FS) (*koanf.Koanf, error) {
	config := koanf.New(".")

	// set options
	Opts = opts

	// load built-in configuration
	if err := loadBuiltinConfig(config, builtins); err != nil {
		return nil, err
	}

	// read configuration from default paths
	if err := loadDefaultConfig(config); err != nil {
		return nil, err
	}

	// read configuration from user defined path
	if err := loadUserDefinedConfig(config, opts.ConfigPath); err != nil {
		return nil, err
	}

	// keep in the config only things we want to colorize
	if Opts.HighlightOnlyLogFormats || Opts.HighlightOnlyPatterns || Opts.HighlightOnlyWords {
		configBackup := config.Copy()

		config.Delete("")

		if Opts.HighlightOnlyLogFormats {
			config.MergeAt(configBackup.Cut("formats"), "formats")
		}
		if Opts.HighlightOnlyPatterns {
			config.MergeAt(configBackup.Cut("patterns"), "patterns")
		}
		if Opts.HighlightOnlyWords {
			config.MergeAt(configBackup.Cut("words"), "words")
		}
	}

	// clear config if we don't want to colorize anything
	if Opts.DryRun {
		config.Delete("")
	}

	return config, nil
}

func loadBuiltinConfig(config *koanf.Koanf, builtins fs.FS) error {
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

	builtinsDir, _ := fs.ReadDir(builtins, "builtins")
	if err := loadFromDirRecursively(builtinsDir, "builtins/"); err != nil {
		return err
	}

	if Opts.NoBuiltinLogFormats {
		config.Delete("formats")
	}

	if Opts.NoBuiltinPatterns {
		config.Delete("patterns")
	}

	if Opts.NoBuiltinWords {
		config.Delete("words")
	}

	if Opts.NoBuiltins {
		config.Delete("")
	}

	return nil
}

func loadDefaultConfig(config *koanf.Koanf) error {
	defaultConfigPaths := [...]string{
		"/etc/logalize/logalize.yaml",
		"~/.config/logalize/logalize.yaml",
		".logalize.yaml",
	}
	for _, path := range defaultConfigPaths {
		if ok, err := checkFileIsReadable(path); ok {
			if err := config.Load(file.Provider(path), yaml.Parser()); err != nil {
				return err
			}
			// ignore only errors about non-existent files
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func loadUserDefinedConfig(config *koanf.Koanf, path string) error {
	if Opts.ConfigPath == "" {
		return nil
	}

	if ok, err := checkFileIsReadable(path); ok {
		if err := config.Load(file.Provider(path), yaml.Parser()); err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func checkFileIsReadable(filePath string) (bool, error) {
	_, err := os.Open(filePath)
	return err == nil, err
}
