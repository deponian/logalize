package core

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deponian/logalize/internal/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

var builtins embed.FS

func TestRunGood(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		// here we care only about multiline input and carriage return
		// other cases are tested in the highlighter package
		{"127.0.0.1 - [test] \"testing\"\nHello true false\n", "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m\nHello \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[48;2;240;108;97mfalse\x1b[0m\n"},
		{"wenzel failed\n127 times", "\x1b[38;2;248;52;178;4mwenzel\x1b[0m \x1b[48;2;240;108;97mfailed\x1b[0m\n\x1b[38;2;80;80;80m127\x1b[0m times"},
		{"true\rfalse", "\x1b[38;2;81;250;138;1mtrue\x1b[0m\r\x1b[48;2;240;108;97mfalse\x1b[0m"},
		{"\rtrue\rfalse", "\r\x1b[38;2;81;250;138;1mtrue\x1b[0m\r\x1b[48;2;240;108;97mfalse\x1b[0m"},
		{"\nfalse\rtrue\n", "\n\x1b[48;2;240;108;97mfalse\x1b[0m\r\x1b[38;2;81;250;138;1mtrue\x1b[0m\n"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/core/Run/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.theme", "test")
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(builtins, cfg, nil, true)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	for _, tt := range tests {
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run("TestRunGood"+tt.plain, func(t *testing.T) {
			err := Run(input, &output, settings)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestRunBad(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/core/Run/02_bad.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.theme", "test")
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(builtins, cfg, nil, true)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}

	input := strings.NewReader("")
	output := bytes.Buffer{}

	t.Run("TestRunBad", func(t *testing.T) {
		if err := Run(input, &output, settings); err == nil {
			t.Error("Run() should have failed")
		}
	})
}

func TestRunBadWriter(t *testing.T) {
	filename := t.TempDir() + "/output.txt"
	file, err := os.Create(filepath.Clean(filename))
	if err != nil {
		t.Errorf("Wasn't able to create test %s: %s", filename, err)
	}
	err = file.Chmod(0o444)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", filename, err)
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Wasn't able to close %s: %s", filename, err)
	}

	settings := config.Settings{}

	input := strings.NewReader("")
	t.Run("TestRunBadWriter", func(t *testing.T) {
		err := Run(input, file, settings)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("Run() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}

func TestRunBadReader(t *testing.T) {
	filename := t.TempDir() + "/input.txt"
	file, err := os.Create(filepath.Clean(filename))
	if err != nil {
		t.Errorf("Wasn't able to create test %s: %s", filename, err)
	}
	err = file.Chmod(0o444)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", filename, err)
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Wasn't able to close %s: %s", filename, err)
	}

	settings := config.Settings{}

	t.Run("TestRunBadReader", func(t *testing.T) {
		err := Run(file, &bytes.Buffer{}, settings)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("Run() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}
