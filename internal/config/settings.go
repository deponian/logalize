package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	goyaml "github.com/goccy/go-yaml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
	"github.com/spf13/pflag"
)

// Settings is the representation of the whole application configuration.
type Settings struct {
	Config       *koanf.Koanf
	Opts         Options
	Builtins     fs.FS
	ColorProfile termenv.Profile
}

// NewSettings creates new Settings instance from built-ins (formats, patterns, words, etc.),
// user configuration (files in /etc/logalize/..., ~/.config/logalize and ./.logalize.yaml)
// and command line flags.
func NewSettings(builtins fs.FS, userConfig *koanf.Koanf, flags *pflag.FlagSet) (Settings, error) {
	// build options step by step
	// first get defaults, then override with values from user configuration
	// then override with everything we get from flags
	opts := NewOptions()
	opts.ReadFromConfig(userConfig)
	opts.ReadFromFlags(flags)

	// load (on not) the built-in configuration based on options
	config := loadBuiltinConfigs(builtins, userConfig, *opts)

	// check theme availability
	if !config.Exists("themes." + opts.Theme) {
		return Settings{},
			fmt.Errorf(
				"theme \"%s\" is not defined. Use -T/--list-themes flag to see the list of all available themes",
				opts.Theme,
			)
	}

	return Settings{
		Config:       config,
		Opts:         *opts,
		Builtins:     builtins,
		ColorProfile: termenv.NewOutput(os.Stdout, termenv.WithUnsafe()).EnvColorProfile(),
	}, nil
}

// CreateUserConfig builds configuration instance from default paths
// (/etc/logalize/..., ~/.config/logalize/... and ./.logalize.yaml) and
// other paths from userPaths variable (most likely these come from --config flag(s)).
func CreateUserConfig(userPaths []string) (*koanf.Koanf, error) {
	loadConfig := func(cfg *koanf.Koanf, paths []string, ignoreNonExistent bool) error {
		for _, path := range paths {
			err := cfg.Load(file.Provider(path), yaml.Parser())
			// ignore only errors about non-existent files
			if err != nil && (!errors.Is(err, os.ErrNotExist) || !ignoreNonExistent) {
				return err
			}
		}

		return nil
	}

	defaultPaths := func() []string {
		homeDir, _ := os.UserHomeDir()

		return []string{
			"/etc/logalize/logalize.yaml",
			homeDir + "/.config/logalize/logalize.yaml",
			"./.logalize.yaml",
		}
	}

	cfg := koanf.New(".")

	// read configuration from default paths
	if err := loadConfig(cfg, defaultPaths(), true); err != nil {
		return nil, err
	}

	// read configuration from user defined path(s)
	if err := loadConfig(cfg, userPaths, false); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadBuiltinConfigs(builtins fs.FS, userConfig *koanf.Koanf, opts Options) *koanf.Koanf {
	var loadRecursively func(config *koanf.Koanf, entries []fs.DirEntry, path string)
	loadRecursively = func(config *koanf.Koanf, entries []fs.DirEntry, path string) {
		for _, entry := range entries {
			if entry.IsDir() {
				dir, _ := fs.ReadDir(builtins, path+entry.Name())
				loadRecursively(config, dir, path+entry.Name()+"/")
			} else {
				file, _ := fs.ReadFile(builtins, path+entry.Name())
				_ = config.Load(rawbytes.Provider(file), yaml.Parser())
			}
		}
	}

	cfg := koanf.New(".")

	// read theme files
	themesDir, _ := fs.ReadDir(builtins, "themes")
	loadRecursively(cfg, themesDir, "themes/")

	// read main configuration file
	builtinsDir, _ := fs.ReadDir(builtins, "builtins")
	loadRecursively(cfg, builtinsDir, "builtins/")

	if opts.NoBuiltinLogFormats {
		cfg.Delete("formats")
	}

	if opts.NoBuiltinPatterns {
		cfg.Delete("patterns")
	}

	if opts.NoBuiltinWords {
		cfg.Delete("words")
	}

	if opts.NoBuiltins {
		cfg.Delete("formats")
		cfg.Delete("patterns")
		cfg.Delete("words")
	}

	// return builtin configuration if user configuration doesn't exit
	if userConfig == nil {
		return cfg
	}

	// apply user configuration on top of builtin configuration,
	// so user configuration has higher priority
	_ = cfg.Merge(userConfig)

	return cfg
}

// ProcessSpecialFlags checks flags like --print-config and --list-themes
// that should produce some text output and then exit the program.
func (s Settings) ProcessSpecialFlags() (data string, exit bool) {
	if s.Opts.PrintConfig {
		return s.printConfig(), true
	}
	if s.Opts.PrintBuiltins {
		return s.printBuiltins(), true
	}
	if s.Opts.ListThemes {
		return s.listThemes(), true
	}

	return "", false
}

func (s Settings) printConfig() string {
	var buf bytes.Buffer
	enc := goyaml.NewEncoder(&buf, goyaml.IndentSequence(true))
	_ = enc.Encode(s.Config.Raw())
	_ = enc.Close()

	return buf.String()
}

func (s Settings) printBuiltins() string {
	var b strings.Builder
	_ = fs.WalkDir(s.Builtins, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, _ := fs.ReadFile(s.Builtins, p)
		b.WriteString("---\n")
		b.Write(data)
		b.WriteByte('\n')

		return nil
	})

	return b.String()
}

func (s Settings) listThemes() string {
	themes := s.Config.MapKeys("themes")

	var result strings.Builder
	fmt.Fprintln(&result, "Available themes:")
	for _, theme := range themes {
		fmt.Fprintf(&result, "  - %s\n", theme)
	}
	fmt.Fprintf(&result, "\nUse one of these with -t/--theme flag\n")

	return result.String()
}
