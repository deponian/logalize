package logalize

import (
	"testing"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

func TestPatternsInit(t *testing.T) {
	configData := `
patterns:
  string:
    priority: 500
    pattern: ("[^"]+"|'[^']+')
    fg: "#00ff00"

  number:
    pattern: (\d+)
    bg: "#00ffff"

  ipv4-address:
    pattern: (\d{1,3}(\.\d{1,3}){3})
    fg: "#ff0000"
    bg: "#ffff00"
    style: bold
`
	correctPatterns := []Pattern{
		{"string", 500, &CapGroup{`("[^"]+"|'[^']+')`, "#00ff00", "", "", nil, nil}},
		{"ipv4-address", 0, &CapGroup{`(\d{1,3}(\.\d{1,3}){3})`, "#ff0000", "#ffff00", "bold", nil, nil}},
		{"number", 0, &CapGroup{`(\d+)`, "", "#00ffff", "", nil, nil}},
	}

	colorProfile = termenv.TrueColor

	config := koanf.New(".")
	configRaw := []byte(configData)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestPatternsInit", func(t *testing.T) {
		if err := initPatterns(config); err != nil {
			t.Errorf("InitPatterns() failed with this error: %s", err)
		}
		for i, pattern := range Patterns {
			pattern.CapGroup.Regexp = nil
			if !cmp.Equal(pattern.Name, correctPatterns[i].Name) {
				t.Errorf("got %v, want %v", pattern.Name, correctPatterns[i].Name)
			}
			if !cmp.Equal(pattern.Priority, correctPatterns[i].Priority) {
				t.Errorf("got %v, want %v", pattern.Priority, correctPatterns[i].Priority)
			}
			if !cmp.Equal(*pattern.CapGroup, *correctPatterns[i].CapGroup) {
				t.Errorf("got %v, want %v", *pattern.CapGroup, *correctPatterns[i].CapGroup)
			}
		}
	})

	configDataBadYAML := `
patterns:
  string:priority: 100
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadYAML)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestPatternsInitBadYAML", func(t *testing.T) {
		if err := initPatterns(config); err == nil {
			t.Errorf("InitPatterns() should have failed")
		}
	})

	configDataBadPattern := `
patterns:
  string:
    priority: 100
    pattern: .*
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadPattern)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestPatternsInitBadPattern", func(t *testing.T) {
		if err := initPatterns(config); err == nil {
			t.Errorf("InitPatterns() should have failed")
		}
	})
}

func TestHighlightPatternsAndWords(t *testing.T) {
	configDataGood := `
patterns:
  string:
    priority: 500
    pattern: ("[^"]+"|'[^']+')
    fg: "#00ff00"

  ipv4-address:
    priority: 400
    pattern: (\d{1,3}(\.\d{1,3}){3})
    fg: "#ff0000"
    bg: "#ffff00"
    style: bold

  number:
    pattern: (\d+)
    bg: "#005050"

  http-status-code:
    priority: 300
    pattern: (\d\d\d)
    fg: "#ffffff"
    alternatives:
      - pattern: (1\d\d)
        fg: "#505050"
      - pattern: (2\d\d)
        fg: "#00ff00"
        style: overline
      - pattern: (3\d\d)
        fg: "#00ffff"
        style: crossout
      - pattern: (4\d\d)
        fg: "#ff0000"
        style: reverse
      - pattern: (5\d\d)
        fg: "#ff00ff"

words:
  good:
    fg: "#52fa8a"
    style: bold
    list:
      - "true"
  bad:
    bg: "#f06c62"
    style: underline
    list:
      - "fail"
      - "fatal"
`
	tests := []struct {
		plain   string
		colored string
	}{
		{"hello", "hello"},
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{"42", "\x1b[48;2;0;80;80m42\x1b[0m"},
		{"127.0.0.1", "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[38;2;80;80;80m101\x1b[0m"},
		{`true bad fail`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[48;2;240;108;97;4mfail\x1b[0m"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status 103`, "status \x1b[38;2;80;80;80m103\x1b[0m"},
		{`status 200`, "status \x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status 302`, "status \x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status 404`, "status \x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status 503`, "status \x1b[38;2;255;0;255m503\x1b[0m"},
		{`status 700`, "status \x1b[38;2;255;255;255m700\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	colorProfile = termenv.TrueColor

	for _, tt := range tests {
		testname := tt.plain
		config := koanf.New(".")
		configRaw := []byte(configDataGood)
		if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
			t.Errorf("Error during config loading: %s", err)
		}
		if err := initPatterns(config); err != nil {
			t.Errorf("InitPatterns() failed with this error: %s", err)
		}
		if err := initWords(config, lemmatizer); err != nil {
			t.Errorf("InitWords() failed with this error: %s", err)
		}
		t.Run(testname, func(t *testing.T) {
			colored := Patterns.highlight(tt.plain, true)
			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}
