package logalize

import (
	"embed"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

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
	builtinLogFormats, err := builtins.ReadDir("builtins/logformats")
	if err != nil {
		return err
	}
	for _, entry := range builtinLogFormats {
		file, _ := builtins.ReadFile("builtins/logformats/" + entry.Name())
		if err = config.Load(rawbytes.Provider(file), yaml.Parser()); err != nil {
			return err
		}
	}

	builtinWords, err := builtins.ReadDir("builtins/words")
	if err != nil {
		return err
	}
	for _, entry := range builtinWords {
		file, _ := builtins.ReadFile("builtins/words/" + entry.Name())
		if err = config.Load(rawbytes.Provider(file), yaml.Parser()); err != nil {
			return err
		}
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
