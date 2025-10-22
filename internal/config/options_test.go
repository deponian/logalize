package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

func TestOptionsReadFromConfig(t *testing.T) {
	correctOpts := Options{
		ConfigPaths: []string{},

		Theme: "test",

		NoBuiltinFormats:  true,
		NoBuiltinPatterns: true,
		NoBuiltinWords:    true,
		NoBuiltins:        true,

		HighlightOnlyFormats:  true,
		HighlightOnlyPatterns: true,
		HighlightOnlyWords:    true,

		NoANSIEscapeSequencesStripping: true,

		Debug:  true,
		DryRun: true,

		PrintConfig:   false,
		PrintBuiltins: false,
		ListThemes:    false,
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/options/ReadFromConfig/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestOptionsReadFromConfig", func(t *testing.T) {
		opts := NewOptions()
		opts.ReadFromConfig(cfg)
		if !cmp.Equal(*opts, correctOpts) {
			t.Errorf("got: %v, want: %v", *opts, correctOpts)
		}
	})

	t.Run("TestOptionsReadFromConfigNil", func(t *testing.T) {
		opts := NewOptions()
		opts.ReadFromConfig(nil)
		if !cmp.Equal(*opts, *NewOptions()) {
			t.Errorf("got: %v, want: %v", *opts, *NewOptions())
		}
	})
}

func TestOptionsReadFromFlags(t *testing.T) {
	correctOpts := Options{
		ConfigPaths: []string{"test1", "test2", "test3"},

		Theme: "test",

		NoBuiltinFormats:  true,
		NoBuiltinPatterns: true,
		NoBuiltinWords:    true,
		NoBuiltins:        true,

		HighlightOnlyFormats:  true,
		HighlightOnlyPatterns: true,
		HighlightOnlyWords:    true,

		NoANSIEscapeSequencesStripping: true,

		Debug:  true,
		DryRun: true,

		PrintConfig:   true,
		PrintBuiltins: true,
		ListThemes:    true,
	}

	flags := pflag.NewFlagSet("test", pflag.PanicOnError)

	flags.StringArrayP("config", "c", []string{}, "")

	flags.StringP("theme", "t", "tokyonight-dark", "")

	flags.BoolP("no-builtin-formats", "L", false, "")
	flags.BoolP("no-builtin-patterns", "P", false, "")
	flags.BoolP("no-builtin-words", "W", false, "")
	flags.BoolP("no-builtins", "N", false, "")

	flags.BoolP("only-formats", "f", false, "")
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
		"--theme", "test",
		"--no-builtin-formats",
		"--no-builtin-patterns",
		"--no-builtin-words",
		"--no-builtins",
		"--only-formats",
		"--only-patterns",
		"--only-words",
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

	t.Run("TestOptionsReadFromFlags", func(t *testing.T) {
		opts := NewOptions()
		opts.ReadFromFlags(flags)
		if !cmp.Equal(*opts, correctOpts) {
			t.Errorf("got: %v, want: %v", *opts, correctOpts)
		}
	})

	t.Run("TestOptionsReadFromFlagsNil", func(t *testing.T) {
		opts := NewOptions()
		opts.ReadFromFlags(nil)
		if !cmp.Equal(*opts, *NewOptions()) {
			t.Errorf("got: %v, want: %v", *opts, *NewOptions())
		}
	})
}
