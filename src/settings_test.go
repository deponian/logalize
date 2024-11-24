package logalize

import (
	"io/fs"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

func TestSettingsInit(t *testing.T) {
	// default config
	configDataDefault := `
settings:
  no-builtin-logformats: true
  #no-builtin-patterns: false
  #no-builtin-words: false
  #no-builtins: false

  only-logformats: true
  only-patterns: true
  only-words: true
`
	defaultConfig := t.TempDir() + "/tempDefaultConfig.yaml"
	configRaw := []byte(configDataDefault)
	err := os.WriteFile(defaultConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", defaultConfig, err)
	}

	t.Cleanup(func() {
		err = os.Remove(defaultConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", defaultConfig, err)
		}
	})

	// user config
	configDataUser := `
settings:
  #no-builtin-logformats: false
  no-builtin-patterns: true
  #no-builtin-words: false
  #no-builtins: false

  only-logformats: false
  #only-patterns: false
  #only-words: false
`
	userConfig := t.TempDir() + "/userConfig.yaml"
	configRaw = []byte(configDataUser)
	err = os.WriteFile(userConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", userConfig, err)
	}

	t.Cleanup(func() {
		err = os.Remove(userConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", userConfig, err)
		}
	})

	// .logalize.yaml
	configDataDot := `
settings:
  #no-builtin-logformats: false
  #no-builtin-patterns: false
  no-builtin-words: true
  #no-builtins: false

  #only-logformats: false
  only-patterns: false
  only-words: false
`
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() failed with this error: %s", err)
	}
	dotConfig := wd + "/.logalize.yaml"

	configRaw = []byte(configDataDot)
	err = os.WriteFile(dotConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", dotConfig, err)
	}

	t.Cleanup(func() {
		err = os.Remove(dotConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", dotConfig, err)
		}
	})

	// flags
	flags := pflag.NewFlagSet("test", pflag.PanicOnError)
	flags.StringP("theme", "a", "", "")
	flags.BoolP("no-builtin-logformats", "b", false, "")
	flags.BoolP("no-builtin-patterns", "c", false, "")
	flags.BoolP("no-builtin-words", "d", false, "")
	flags.BoolP("no-builtins", "e", false, "")
	flags.BoolP("only-logformats", "f", false, "")
	flags.BoolP("only-patterns", "g", false, "")
	flags.BoolP("only-words", "h", false, "")
	flags.StringP("config", "i", "", "")
	args := []string{
		"--no-builtins",
		"--only-words",
		"--config",
		userConfig,
	}
	if err := flags.Parse(args); err != nil {
		t.Errorf("flags.Parse() failed with an error: %s", err)
	}

	correctOpts := Settings{
		ConfigPath: userConfig,

		Theme: "tokyonight-dark",

		NoBuiltinLogFormats: true,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      true,
		NoBuiltins:          true,

		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      true,
	}

	defaultConfigPaths = append(defaultConfigPaths, defaultConfig)

	t.Run("TestSettingsFromInitGood", func(t *testing.T) {
		if err := InitSettings(flags); err != nil {
			t.Errorf("InitSettings() failed with an error: %s", err)
		}

		if !cmp.Equal(Opts, correctOpts) {
			t.Errorf("got %v, want %v", Opts, correctOpts)
		}
	})

	err = os.Chmod(dotConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", dotConfig, err)
	}

	t.Run("TestSettingsFromInitBadDotConfig", func(t *testing.T) {
		err := InitSettings(flags)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitSettings() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	err = os.Chmod(userConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", userConfig, err)
	}

	t.Run("TestSettingsFromInitBadDotConfig", func(t *testing.T) {
		err := InitSettings(flags)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitSettings() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	err = os.Chmod(defaultConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", defaultConfig, err)
	}

	t.Run("TestSettingsFromInitBadDotConfig", func(t *testing.T) {
		err := InitSettings(flags)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitSettings() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	defaultConfigPaths = defaultConfigPaths[:len(defaultConfigPaths)-1]
}

func TestSettingsFromConfig(t *testing.T) {
	configData := `
settings:
  theme: "test"

  no-builtin-logformats: true
  no-builtin-patterns: true
  no-builtin-words: true
  no-builtins: true

  only-logformats: true
  only-patterns: true
  only-words: true
`
	configRaw := []byte(configData)
	config := koanf.New(".")
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}

	correctOpts := Settings{
		Theme: "test",

		NoBuiltinLogFormats: true,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      true,
		NoBuiltins:          true,

		HighlightOnlyLogFormats: true,
		HighlightOnlyPatterns:   true,
		HighlightOnlyWords:      true,
	}

	t.Run("TestSettingsFromConfig", func(t *testing.T) {
		opts := getSettingFromConfig(Settings{}, config)

		if !cmp.Equal(opts, correctOpts) {
			t.Errorf("got %v, want %v", opts, correctOpts)
		}
	})
}

func TestSettingsFromFlags(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.PanicOnError)
	flags.StringP("theme", "a", "", "")
	flags.BoolP("no-builtin-logformats", "b", false, "")
	flags.BoolP("no-builtin-patterns", "c", false, "")
	flags.BoolP("no-builtin-words", "d", false, "")
	flags.BoolP("no-builtins", "e", false, "")
	flags.BoolP("only-logformats", "f", false, "")
	flags.BoolP("only-patterns", "g", false, "")
	flags.BoolP("only-words", "h", false, "")
	flags.BoolP("dry-run", "i", false, "")
	args := []string{
		"--theme",
		"test",
		"--no-builtin-logformats",
		"--no-builtin-patterns",
		"--no-builtin-words",
		"--no-builtins",
		"--only-logformats",
		"--only-patterns",
		"--only-words",
		"--dry-run",
	}
	if err := flags.Parse(args); err != nil {
		t.Errorf("flags.Parse() failed with an error: %s", err)
	}
	correctOpts := Settings{
		Theme: "test",

		NoBuiltinLogFormats: true,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      true,
		NoBuiltins:          true,

		HighlightOnlyLogFormats: true,
		HighlightOnlyPatterns:   true,
		HighlightOnlyWords:      true,
		DryRun:                  true,
	}

	t.Run("TestSettingsFromFlags", func(t *testing.T) {
		opts := getSettingFromFlags(Settings{}, flags)

		if !cmp.Equal(opts, correctOpts) {
			t.Errorf("got %v, want %v", opts, correctOpts)
		}
	})
}
