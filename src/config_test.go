package logalize

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/muesli/termenv"
)

//go:embed builtins/*
//go:embed themes/*
var builtins embed.FS

func TestConfigLoadBuiltinFlagNoBuiltins(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      false,
		NoBuiltins:          true,
		Theme:               "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagNoBuiltinLogFormats(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C \x1b[38;2;192;153;255m17 Feb 2024\x1b[0m \x1b[38;2;252;167;234m00:39:12.557\x1b[0m * Parent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},

		// words
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m \x1b[38;2;118;211;255m7.7.7.7\x1b[0m\x1b[38;2;13;185;215m\x1b[0m \x1b[38;2;252;167;234m01:37:59.743\x1b[0m \x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: true,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      false,
		NoBuiltins:          false,
		Theme:               "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagNoBuiltinPatterns(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      false,
		NoBuiltins:          false,
		Theme:               "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagNoBuiltinWords(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail \x1b[38;2;118;211;255m7.7.7.7\x1b[0m\x1b[38;2;13;185;215m\x1b[0m \x1b[38;2;252;167;234m01:37:59.743\x1b[0m \x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      true,
		NoBuiltins:          false,
		Theme:               "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagNoBuiltinPatternsAndWords(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      true,
		NoBuiltins:          false,
		Theme:               "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagDryRun(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      false,
		DryRun:                  true,
		Theme:                   "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagHighlightOnlyLogFormats(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: true,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      false,
		DryRun:                  false,
		Theme:                   "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagHighlightOnlyPatterns(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C \x1b[38;2;192;153;255m17 Feb 2024\x1b[0m \x1b[38;2;252;167;234m00:39:12.557\x1b[0m * Parent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail \x1b[38;2;118;211;255m7.7.7.7\x1b[0m\x1b[38;2;13;185;215m\x1b[0m \x1b[38;2;252;167;234m01:37:59.743\x1b[0m \x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   true,
		HighlightOnlyWords:      false,
		DryRun:                  false,
		Theme:                   "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinFlagHighlightOnlyWords(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      true,
		DryRun:                  false,
		Theme:                   "tokyonight-dark",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadBuiltinBad(t *testing.T) {
	colorProfile = termenv.TrueColor

	Opts = Settings{
		Theme: "tokyonight-dark",
	}

	builtinsLogformatsBad := fstest.MapFS{
		"builtins/logformats/bad.yaml": {
			Data: []byte("formats:\n  test:\n  regexp: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinLogformatsBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsLogformatsBad)
		if err == nil || err.Error() != "yaml: line 3: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	builtinsPatternsBad := fstest.MapFS{
		"builtins/patterns/bad.yaml": {
			Data: []byte("patterns:\n  bad: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinPatternsBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsPatternsBad)
		if err == nil || err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	builtinsWordsBad := fstest.MapFS{
		"builtins/words/bad.yaml": {
			Data: []byte("words:\n  bad: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinWordsBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsWordsBad)
		if err == nil || err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	builtinsThemesBad := fstest.MapFS{
		"themes/bad.yaml": {
			Data: []byte("themes:\n  bad: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinThemesBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsThemesBad)
		if err == nil || err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})
}

func TestConfigLoadUserDefinedGood(t *testing.T) {
	colorProfile = termenv.TrueColor
	configData1 := `
formats:
  menetekel:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      name: one
    - regexp: ([^ ]+ )
      name: two
    - regexp: (\[.+\] )
      name: three
    - regexp: ("[^"]+")
      name: four

patterns:
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')

  ipv4-address:
    priority: 400
    regexp: (\d{1,3}(\.\d{1,3}){3})

  number:
    regexp: (\d+)

  http-status-code:
    priority: 300
    regexp: (\d\d\d)
    alternatives:
      - regexp: (1\d\d)
        name: 1xx
      - regexp: (2\d\d)
        name: 2xx
      - regexp: (3\d\d)
        name: 3xx
      - regexp: (4\d\d)
        name: 4xx
      - regexp: (5\d\d)
        name: 5xx

words:
  friends:
    - "toni"
    - "wenzel"
  foes:
    - "argus"
    - "cletus"

themes:
  # it will be overridden in the next config
  test:
    formats:
      menetekel:
        one:
          fg: "#ff0000"
        two:
          bg: "#ff0000"
        three:
          style: bold
        four:
          fg: "#ff0000"
          bg: "#ff0000"
          style: underline
`
	configData2 := `
themes:
  test:
    formats:
      menetekel:
        one:
          fg: "#f5ce42"
        two:
          bg: "#764a9e"
        three:
          style: bold
        four:
          fg: "#9daf99"
          bg: "#76fb99"
          style: underline

    patterns:
      string:
        fg: "#00ff00"

      ipv4-address:
        fg: "#ff0000"
        bg: "#ffff00"
        style: bold

      number:
        bg: "#005050"

      http-status-code:
        default:
          fg: "#ffffff"
        1xx:
          fg: "#505050"
        2xx:
          fg: "#00ff00"
          style: overline
        3xx:
          fg: "#00ffff"
          style: crossout
        4xx:
          fg: "#ff0000"
          style: reverse
        5xx:
          fg: "#ff00ff"

    words:
      friends:
        fg: "#f834b2"
        style: underline
      foes:
        fg: "#120fbb"
        style: underline
`
	tests := []struct {
		plain   string
		colored string
	}{
		// log format
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m"},

		// pattern
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{`42`, "\x1b[48;2;0;80;80m42\x1b[0m"},
		{`127.0.0.1`, "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[38;2;80;80;80m101\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status 103`, "status \x1b[38;2;80;80;80m103\x1b[0m"},
		{`status 200`, "status \x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status 302`, "status \x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status 404`, "status \x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status 503`, "status \x1b[38;2;255;0;255m503\x1b[0m"},
		{`status 700`, "status \x1b[38;2;255;255;255m700\x1b[0m"},

		// words
		{"wenzel", "\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"argus", "\x1b[38;2;18;15;187;4margus\x1b[0m"},

		{"not toni", "not \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"Not wenzel", "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"wasn't argus", "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"won't cletus", "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
		{"cannot toni", "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"won't be wenzel", "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"cannot be argus", "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"should not be cletus", "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m \x1b[48;2;0;80;80m01\x1b[0m:\x1b[48;2;0;80;80m37\x1b[0m:\x1b[48;2;0;80;80m59\x1b[0m.\x1b[38;2;255;255;255m743\x1b[0m \x1b[48;2;0;80;80m75\x1b[0m.\x1b[38;2;255;255;255m984\x1b[0m\x1b[38;2;255;255;255m854\x1b[0mms",
		},
		{
			`"wenzel" and wenzel`,
			"\x1b[38;2;0;255;0m\"wenzel\"\x1b[0m and \x1b[38;2;248;52;178;4mwenzel\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	userConfig1 := t.TempDir() + "/userConfig1.yaml"
	configRaw := []byte(configData1)
	err = os.WriteFile(userConfig1, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", userConfig1, err)
	}

	userConfig2 := t.TempDir() + "/userConfig2.yaml"
	configRaw = []byte(configData2)
	err = os.WriteFile(userConfig2, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", userConfig2, err)
	}

	Opts = Settings{
		ConfigPaths: []string{userConfig1, userConfig2},
		NoBuiltins:  true,
		Theme:       "test",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			err := Run(input, &output, lemmatizer)
			if err != nil {
				t.Errorf("Run() failed with this error: %s", err)
			}

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestConfigLoadUserDefinedBad(t *testing.T) {
	colorProfile = termenv.TrueColor

	configDataBadYAML := `
formats:
  test:
  regexp: bad:
`

	userConfig := t.TempDir() + "/userConfig.yaml"
	configRaw := []byte(configDataBadYAML)
	err := os.WriteFile(userConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", userConfig, err)
	}

	Opts = Settings{
		ConfigPaths: []string{userConfig},
		NoBuiltins:  true,
		Theme:       "tokyonight-dark",
	}

	t.Run("TestConfigLoadUserDefinedBadYAML", func(t *testing.T) {
		err := InitConfig(builtins)
		if err == nil || err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	Opts = Settings{
		ConfigPaths: []string{userConfig + "error"},
		NoBuiltins:  true,
		Theme:       "tokyonight-dark",
	}

	t.Run("TestConfigLoadUserDefinedFileDoesntExist", func(t *testing.T) {
		err := InitConfig(builtins)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitConfig() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	userConfigReadOnly := t.TempDir() + "/userConfigReadOnly.yaml"
	configRaw = []byte(configDataBadYAML)
	err = os.WriteFile(userConfigReadOnly, configRaw, 0200)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", userConfig, err)
	}

	Opts = Settings{
		ConfigPaths: []string{userConfigReadOnly},
		NoBuiltins:  false,
		Theme:       "tokyonight-dark",
	}

	t.Run("TestConfigLoadUserDefinedReadOnly", func(t *testing.T) {
		err := InitConfig(builtins)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitConfig() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}

func TestConfigLoadDefault(t *testing.T) {
	colorProfile = termenv.TrueColor
	configDataBadYAML := `
formats:
  test:
  regexp: bad:
`

	tempDefaultConfig := t.TempDir() + "/tempDefaultConfig.yaml"
	configRaw := []byte(configDataBadYAML)
	err := os.WriteFile(tempDefaultConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", tempDefaultConfig, err)
	}

	t.Cleanup(func() {
		err = os.Remove(tempDefaultConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", tempDefaultConfig, err)
		}
	})

	Opts = Settings{
		ConfigPaths: []string{},
		NoBuiltins:  true,
		Theme:       "tokyonight-dark",
	}

	defaultConfigPaths = append(defaultConfigPaths, tempDefaultConfig)

	t.Run("TestConfigLoadDefaultBadYAML", func(t *testing.T) {
		err := InitConfig(builtins)
		if err == nil || err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	err = os.Chmod(tempDefaultConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", tempDefaultConfig, err)
	}

	t.Run("TestConfigLoadDefaultReadOnly", func(t *testing.T) {
		err := InitConfig(builtins)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitConfig() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	defaultConfigPaths = defaultConfigPaths[:len(defaultConfigPaths)-1]
}

func TestConfigLoadDotLogalize(t *testing.T) {
	colorProfile = termenv.TrueColor
	configDataBadYAML := `
formats:
  test:
  regexp: bad:
`

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() failed with this error: %s", err)
	}
	dotConfig := wd + "/.logalize.yaml"
	configRaw := []byte(configDataBadYAML)
	if ok, err := checkFileIsReadable(dotConfig); ok {
		if err != nil {
			t.Errorf("checkFileIsReadable() failed with this error: %s", err)
		}
		err = os.Remove(dotConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", dotConfig, err)
		}
	}

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

	Opts = Settings{
		ConfigPaths: []string{},
		NoBuiltins:  true,
		Theme:       "tokyonight-dark",
	}

	t.Run("TestConfigLoadDotLogalizeBadYAML", func(t *testing.T) {
		err := InitConfig(builtins)
		if err == nil || err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	err = os.Chmod(dotConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", dotConfig, err)
	}

	t.Run("TestConfigLoadDotLogalizeReadOnly", func(t *testing.T) {
		err := InitConfig(builtins)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitConfig() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}

func TestConfigTheme(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS

	configData := `
themes:
  test: {}
`

	testConfig := t.TempDir() + "/testConfig.yaml"
	configRaw := []byte(configData)
	err := os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}

	// good
	Opts = Settings{
		ConfigPaths: []string{testConfig},
		Theme:       "test",
	}

	t.Run("TestConfigThemeIsDefined", func(t *testing.T) {
		err := InitConfig(builtins)
		if err != nil {
			t.Errorf("InitConfig() failed with this error: %s", err)
		}
	})

	// bad
	Opts = Settings{
		ConfigPaths: []string{testConfig},
		Theme:       "idontexist",
	}

	t.Run("TestConfigThemeIsNotDefined", func(t *testing.T) {
		err := InitConfig(builtins)
		if err == nil || err.Error() != "Theme \"idontexist\" is not defined. Use -T/--list-themes flag to see the list of all available themes" {
			t.Errorf("InitConfig() should have failed with \"Theme \"idontexist\" is not defined. Use -T/--list-themes flag to see the list of all available themes\", got: [%T] %s", err, err)
		}
	})
}
