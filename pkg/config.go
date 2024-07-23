package logalize

import (
	"embed"
	"io/fs"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

// Options stores the values of command-line options
type Options struct {
	ConfigPath    string // path to configuration file
	NoBuiltins    bool   // disable built-in log formats and words
	PrintBuiltins bool   // print built-in log formats and words
}

func InitConfig(opts Options, builtins embed.FS) (*koanf.Koanf, error) {
	config := koanf.New(".")

	// load built-in configuration
	if !opts.NoBuiltins {
		if err := loadBuiltinConfig(config, builtins); err != nil {
			return nil, err
		}
	}

	// read configuration from default paths
	if err := loadDefaultConfig(config); err != nil {
		return nil, err
	}

	// read configuration from user defined path
	if opts.ConfigPath != "" {
		if err := loadUserDefinedConfig(config, opts.ConfigPath); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func loadBuiltinConfig(config *koanf.Koanf, builtins embed.FS) error {
	var loadFromDirRecursively func(entries []fs.DirEntry, path string) error
	loadFromDirRecursively = func(entries []fs.DirEntry, path string) error {
		for _, entry := range entries {
			if entry.IsDir() {
				dir, _ := builtins.ReadDir(path + entry.Name())
				if err := loadFromDirRecursively(dir, path+entry.Name()+"/"); err != nil {
					return err
				}
			} else {
				file, _ := builtins.ReadFile(path + entry.Name())
				if err := config.Load(rawbytes.Provider(file), yaml.Parser()); err != nil {
					return err
				}
			}
		}
		return nil
	}

	builtinsDir, _ := builtins.ReadDir("builtins")
	if err := loadFromDirRecursively(builtinsDir, "builtins/"); err != nil {
		return err
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
