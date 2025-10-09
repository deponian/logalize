package config

import (
	"bytes"
	"embed"
	"os"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
	"github.com/spf13/pflag"
)

func compareConfigs(c1, c2 *koanf.Koanf) (bool, error) {
	c1j, err := c1.Marshal(json.Parser())
	if err != nil {
		return false, err
	}
	c2j, err := c2.Marshal(json.Parser())
	if err != nil {
		return false, err
	}

	return bytes.Equal(c1j, c2j), nil
}

func TestSettingsNewGood(t *testing.T) {
	correctOpts := Options{
		ConfigPaths: []string{"test1", "test2", "test3"},

		Theme: "test",

		NoBuiltinLogFormats: true,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      true,
		NoBuiltins:          true,

		HighlightOnlyLogFormats: true,
		HighlightOnlyPatterns:   true,
		HighlightOnlyWords:      true,

		NoANSIEscapeSequencesStripping: true,

		Debug:  true,
		DryRun: true,

		PrintConfig:   true,
		PrintBuiltins: true,
		ListThemes:    true,
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/settings/NewSettings/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	flags := pflag.NewFlagSet("test", pflag.PanicOnError)

	flags.StringArrayP("config", "c", []string{}, "")

	flags.StringP("theme", "t", "tokyonight-dark", "")

	flags.BoolP("no-builtin-logformats", "L", false, "")
	flags.BoolP("no-builtin-patterns", "P", false, "")
	flags.BoolP("no-builtin-words", "W", false, "")
	flags.BoolP("no-builtins", "N", false, "")

	flags.BoolP("only-logformats", "l", false, "")
	flags.BoolP("only-patterns", "p", false, "")
	flags.BoolP("only-words", "w", false, "")

	flags.BoolP("no-ansi-escape-sequences-stripping", "s", false, "")

	flags.BoolP("debug", "d", false, "")
	flags.BoolP("dry-run", "n", false, "")

	flags.BoolP("print-config", "C", false, "")
	flags.BoolP("list-themes", "T", false, "")
	flags.BoolP("print-builtins", "b", false, "")

	args := []string{
		"--config", "test1",
		"--config", "test2",
		"--config", "test3",
		"--no-builtin-logformats",
		"--no-builtin-patterns",
		"--no-builtin-words",
		"--no-builtins",
		"--no-ansi-escape-sequences-stripping",
		"--debug",
		"--dry-run",
		"--print-config",
		"--list-themes",
		"--print-builtins",
	}

	if err := flags.Parse(args); err != nil {
		t.Errorf("flags.Parse() failed with an error: %s", err)
	}

	settings, err := NewSettings(embed.FS{}, cfg, flags)
	if err != nil {
		t.Fatalf("NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	t.Run("TestSettingsNewGood", func(t *testing.T) {
		if !cmp.Equal(settings.Opts, correctOpts) {
			t.Errorf("got %v, want %v", settings.Opts, correctOpts)
		}
	})
}

func TestSettingsNewBad(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/settings/NewSettings/02_bad.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	_, err = NewSettings(embed.FS{}, cfg, nil)
	if err == nil {
		t.Error("NewSettings(...) should have failed")
	}
}

func TestSettingsCreateUserConfigGood(t *testing.T) {
	correctConfig := koanf.New(".")
	err := correctConfig.Load(file.Provider("./testdata/settings/CreateUserConfig/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("correctConfig.Load(...) failed with this error: %s", err)
	}

	cfg, err := CreateUserConfig([]string{"./testdata/settings/CreateUserConfig/01_good.yaml"})
	if err != nil {
		t.Fatalf("CreateUserConfig(...) failed with this error: %s", err)
	}

	t.Run("TestSettingsCreateUserConfigGood", func(t *testing.T) {
		if ok, err := compareConfigs(cfg, correctConfig); err != nil || !ok {
			t.Errorf("got %v, want %v", cfg, correctConfig)
		}
	})
}

func TestSettingsCreateUserConfigBadUserConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() failed with this error: %s", err)
	}
	badConfig := wd + "/bad.yaml"
	err = os.WriteFile(badConfig, []byte{}, 0o200)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", badConfig, err)
	}
	t.Cleanup(func() {
		err = os.Remove(badConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", badConfig, err)
		}
	})

	t.Run("TestSettingsCreateUserConfigBadUserConfig", func(t *testing.T) {
		_, err := CreateUserConfig([]string{badConfig})
		if err == nil {
			t.Error("CreateUserConfig(...) should have failed")
		}
	})
}

func TestSettingsCreateUserConfigBadDefaultConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() failed with this error: %s", err)
	}
	dotConfig := wd + "/.logalize.yaml"
	err = os.WriteFile(dotConfig, []byte{}, 0o600)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", dotConfig, err)
	}
	t.Cleanup(func() {
		err = os.Remove(dotConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", dotConfig, err)
		}
	})
	err = os.Chmod(dotConfig, 0o200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", dotConfig, err)
	}

	t.Run("TestSettingsCreateUserConfigBadDefaultConfig", func(t *testing.T) {
		_, err := CreateUserConfig([]string{""})
		if err == nil {
			t.Error("CreateUserConfig(...) should have failed")
		}
	})
}

func TestLoadBuiltinConfigsGood(t *testing.T) {
	builtinConfig, err := os.ReadFile("./testdata/settings/loadBuiltinConfigs/01_good_main.yaml")
	if err != nil {
		t.Fatalf("os.ReadFile(...) failed with this error: %s", err)
	}
	builtins := fstest.MapFS{
		"builtins/patterns/test.yaml": {
			Data: builtinConfig,
		},
	}

	correctConfig := koanf.New(".")
	err = correctConfig.Load(file.Provider("./testdata/settings/loadBuiltinConfigs/01_good_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("correctConfig.Load(...) failed with this error: %s", err)
	}

	t.Run("TestLoadBuiltinConfigsGood", func(t *testing.T) {
		cfg := loadBuiltinConfigs(builtins, nil, Options{})
		if ok, err := compareConfigs(cfg, correctConfig); err != nil || !ok {
			t.Errorf("got %v, want %v", cfg, correctConfig)
		}
	})
}

func TestLoadBuiltinConfigsGoodWithUserConfig(t *testing.T) {
	builtinConfig, err := os.ReadFile("./testdata/settings/loadBuiltinConfigs/02_builtin_config.yaml")
	if err != nil {
		t.Fatalf("os.ReadFile(...) failed with this error: %s", err)
	}
	builtins := fstest.MapFS{
		"builtins/patterns/test.yaml": {
			Data: builtinConfig,
		},
	}

	userConfig := koanf.New(".")
	err = userConfig.Load(file.Provider("./testdata/settings/loadBuiltinConfigs/02_user_config.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("userConfig.Load(...) failed with this error: %s", err)
	}

	correctConfig := koanf.New(".")
	err = correctConfig.Load(file.Provider("./testdata/settings/loadBuiltinConfigs/02_correct_config.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("correctConfig.Load(...) failed with this error: %s", err)
	}

	t.Run("TestLoadBuiltinConfigsGoodWithUserConfig", func(t *testing.T) {
		cfg := loadBuiltinConfigs(builtins, userConfig, Options{})
		if ok, err := compareConfigs(cfg, correctConfig); err != nil || !ok {
			t.Errorf("got %v, want %v", cfg, correctConfig)
		}
	})
}

func TestLoadBuiltinConfigsGoodNoBuiltns(t *testing.T) {
	builtinConfig, err := os.ReadFile("./testdata/settings/loadBuiltinConfigs/03_no_builtins.yaml")
	if err != nil {
		t.Fatalf("os.ReadFile(...) failed with this error: %s", err)
	}
	builtins := fstest.MapFS{
		"builtins/test.yaml": {
			Data: builtinConfig,
		},
	}

	t.Run("TestLoadBuiltinConfigsGoodNoBuiltnsThreeFlags", func(t *testing.T) {
		opts := Options{
			NoBuiltinLogFormats: true,
			NoBuiltinPatterns:   true,
			NoBuiltinWords:      true,
		}
		cfg := loadBuiltinConfigs(builtins, nil, opts)
		if cfg.Exists("formats") || cfg.Exists("patterns") || cfg.Exists("words") {
			t.Errorf("config shouldn't have any formats, patterns or words")
		}
	})

	t.Run("TestLoadBuiltinConfigsGoodNoBuiltnsOneFlag", func(t *testing.T) {
		opts := Options{
			NoBuiltins: true,
		}
		cfg := loadBuiltinConfigs(builtins, nil, opts)
		if cfg.Exists("formats") || cfg.Exists("patterns") || cfg.Exists("words") {
			t.Errorf("config shouldn't have any formats, patterns or words")
		}
	})
}

func TestSettingsProcessSpecialFlagsPrintConfig(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/settings/ProcessSpecialFlags/01_print_config.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings, err := NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("NewSettings(...) failed with this error: %s", err)
	}
	settings.Opts.PrintConfig = true

	correctConfig, err := os.ReadFile("./testdata/settings/ProcessSpecialFlags/01_print_config.yaml")
	if err != nil {
		t.Fatalf("os.ReadFile(...) failed with this error: %s", err)
	}

	t.Run("TestSettingsProcessSpecialFlagsPrintConfig", func(t *testing.T) {
		output, _ := settings.ProcessSpecialFlags()
		if output != string(correctConfig) {
			t.Errorf("got %v, want %v", output, string(correctConfig))
		}
	})
}

func TestSettingsProcessSpecialFlagsPrintBuiltins(t *testing.T) {
	builtinConfig, err := os.ReadFile("./testdata/settings/ProcessSpecialFlags/02_builtin_config.yaml")
	if err != nil {
		t.Fatalf("os.ReadFile(...) failed with this error: %s", err)
	}
	builtins := fstest.MapFS{
		"builtins/patterns/test.yaml": {
			Data: builtinConfig,
		},
	}

	cfg := koanf.New(".")
	err = cfg.Load(file.Provider("./testdata/settings/ProcessSpecialFlags/02_print_builtins.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings, err := NewSettings(builtins, cfg, nil)
	if err != nil {
		t.Fatalf("NewSettings(...) failed with this error: %s", err)
	}
	settings.Opts.PrintBuiltins = true

	correctBuiltins := "---\npatterns:\n  test:\n    regexp: (.*)\n\n"

	t.Run("TestSettingsProcessSpecialFlagsPrintBuiltins", func(t *testing.T) {
		output, _ := settings.ProcessSpecialFlags()
		if output != correctBuiltins {
			t.Errorf("got %v, want %v", output, correctBuiltins)
		}
	})
}

func TestSettingsProcessSpecialFlagsListThemes(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/settings/ProcessSpecialFlags/03_list_themes.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings, err := NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("NewSettings(...) failed with this error: %s", err)
	}
	settings.Opts.ListThemes = true

	correctThemes := "Available themes:\n  - test\n  - test2\n  - test3\n\nUse one of these with -t/--theme flag\n"

	t.Run("TestSettingsProcessSpecialFlagsListThemes", func(t *testing.T) {
		output, _ := settings.ProcessSpecialFlags()
		if output != correctThemes {
			t.Errorf("got %v, want %v", output, correctThemes)
		}
	})
}

func TestSettingsProcessSpecialFlagsNoFlags(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/settings/ProcessSpecialFlags/04_no_flags.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings, err := NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("NewSettings(...) failed with this error: %s", err)
	}

	t.Run("TestSettingsProcessSpecialFlagsNoFlags", func(t *testing.T) {
		_, result := settings.ProcessSpecialFlags()
		if result != false {
			t.Errorf("got %v, want %v", result, false)
		}
	})
}
