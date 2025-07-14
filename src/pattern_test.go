package logalize

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/muesli/termenv"
)

func TestPatternsInit(t *testing.T) {
	configData := `
patterns:
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')

  number:
    regexp: (\d+)

  ipv4-address:
    regexps:
      - regexp: (\d\d\d(\.\d\d\d){3})
        name: one
      - regexp: ((:\d{1,5})?)
        name: two

themes:
  test:
    patterns:
      string:
        fg: "#00ff00"

      number:
        bg: "#00ffff"
        style: bold

      ipv4-address:
        one:
          fg: "#ffc777"
        two:
          fg: "#ff966c"
`
	correctPatterns := []Pattern{
		{"string", 500, &CapGroupList{
			[]CapGroup{
				{
					"", `("[^"]+"|'[^']+')`, "#00ff00", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:"[^"]+"|'[^']+'))`),
		}},
		{"ipv4-address", 0, &CapGroupList{
			[]CapGroup{
				{
					"", `(\d\d\d(\.\d\d\d){3})`, "#ffc777", "", "", nil, nil,
				},
				{
					"", `((:\d{1,5})?)`, "#ff966c", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d\d\d(\.\d\d\d){3}))(?P<capGroup1>(?:(:\d{1,5})?))`),
		}},
		{"number", 0, &CapGroupList{
			[]CapGroup{
				{
					"", `(\d+)`, "", "#00ffff", "bold", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d+))`),
		}},
	}

	comparePatterns := func(pattern1, pattern2 Pattern) error {
		if pattern1.Name != pattern2.Name || pattern1.Priority != pattern2.Priority {
			return fmt.Errorf("[pattern1: %s, pattern2: %s] names or priorities aren't equal", pattern1.Name, pattern2.Name)
		}
		if err := compareCapGroupLists(*pattern1.CapGroups, *pattern2.CapGroups); err != nil {
			return err
		}
		return nil
	}

	colorProfile = termenv.TrueColor

	var builtins embed.FS

	testConfig := t.TempDir() + "/testConfig.yaml"
	configRaw := []byte(configData)
	err := os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}
	Opts = Settings{
		ConfigPath: testConfig,
		NoBuiltins: true,
		Theme:      "test",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	t.Run("TestPatternsInit", func(t *testing.T) {
		if err := initPatterns(); err != nil {
			t.Errorf("InitPatterns() failed with this error: %s", err)
		}

		for i, pattern := range Patterns {
			if err := comparePatterns(pattern, correctPatterns[i]); err != nil {
				t.Errorf("%s", err)
			}
		}
	})

	configDataBadYAML1 := `
patterns:
  string:priority: 100
themes:
  test:
`

	testConfig = t.TempDir() + "/testConfig.yaml"
	configRaw = []byte(configDataBadYAML1)
	err = os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}
	Opts = Settings{
		ConfigPath: testConfig,
		NoBuiltins: true,
		Theme:      "test",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	t.Run("TestPatternsInitBadYAML1", func(t *testing.T) {
		if err := initPatterns(); err == nil {
			t.Errorf("InitPatterns() should have failed")
		}
	})

	configDataBadYAML2 := `
patterns:
  test:
    regexps: 4
themes:
  test:
`

	testConfig = t.TempDir() + "/testConfig.yaml"
	configRaw = []byte(configDataBadYAML2)
	err = os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}
	Opts = Settings{
		ConfigPath: testConfig,
		NoBuiltins: true,
		Theme:      "test",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	t.Run("TestPatternsInitBadYAML2", func(t *testing.T) {
		if err := initPatterns(); err == nil {
			t.Errorf("InitPatterns() should have failed")
		}
	})

	configDataBadRegExp := `
patterns:
  string:
    priority: 100
    regexp: .*
themes:
  test:
`

	testConfig = t.TempDir() + "/testConfig.yaml"
	configRaw = []byte(configDataBadRegExp)
	err = os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}
	Opts = Settings{
		ConfigPath: testConfig,
		NoBuiltins: true,
		Theme:      "test",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	t.Run("TestPatternsInitBadRegExp", func(t *testing.T) {
		if err := initPatterns(); err == nil {
			t.Errorf("InitPatterns() should have failed")
		}
	})

	configDataBadStyle := `
patterns:
  string:
    regexp: (.*)

themes:
  test:
    patterns:
      string:
        style: hello
`

	testConfig = t.TempDir() + "/testConfig.yaml"
	configRaw = []byte(configDataBadStyle)
	err = os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}
	Opts = Settings{
		ConfigPath: testConfig,
		NoBuiltins: true,
		Theme:      "test",
	}

	err = InitConfig(builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	t.Run("TestPatternsInitBadRegExp", func(t *testing.T) {
		if err := initPatterns(); err == nil {
			t.Errorf("InitPatterns() should have failed")
		}
	})
}

func TestHighlightPatternsAndWords(t *testing.T) {
	configDataGood := `
patterns:
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')

  ipv4-address:
    priority: 400
    regexp: (\d{1,3}(\.\d{1,3}){3})

  number:
    regexp: (\d+)

  http-status:
    priority: 300
    regexps:
    - regexp: (status:)
      name: status
    - regexp: (\d\d\d)
      name: code
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

themes:
  test:
    patterns:
      string:
        fg: "#00ff00"

      ipv4-address:
        fg: "#ff0000"
        bg: "#ffff00"
        style: bold

      number:
        bg: "#005050"

      http-status:
        code:
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
`
	tests := []struct {
		plain   string
		colored string
	}{
		{"hello", "hello"},
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{"42", "\x1b[48;2;0;80;80m42\x1b[0m"},
		{"127.0.0.1", "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[48;2;0;80;80m101\x1b[0m"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status:103`, "status:\x1b[38;2;80;80;80m103\x1b[0m"},
		{`status:200`, "status:\x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status:302`, "status:\x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status:404`, "status:\x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status:503`, "status:\x1b[38;2;255;0;255m503\x1b[0m"},
		{`status:700`, "status:\x1b[38;2;255;255;255m700\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	colorProfile = termenv.TrueColor

	var builtins embed.FS

	testConfig := t.TempDir() + "/testConfig.yaml"
	configRaw := []byte(configDataGood)
	err = os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}
	Opts = Settings{
		ConfigPath: testConfig,
		NoBuiltins: true,
		Theme:      "test",
	}

	for _, tt := range tests {
		testname := tt.plain

		err := InitConfig(builtins)
		if err != nil {
			t.Errorf("InitConfig() failed with this error: %s", err)
		}

		if err := initPatterns(); err != nil {
			t.Errorf("InitPatterns() failed with this error: %s", err)
		}
		if err := initWords(lemmatizer); err != nil {
			t.Errorf("InitWords() failed with this error: %s", err)
		}
		t.Run(testname, func(t *testing.T) {
			colored := Patterns.highlight(tt.plain)
			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}

// Below are the tests for all built-in patterns

// rfc3339
func TestPatternsRFC3339(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{
			`2024-02-17T06:56:10Z`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mT\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255mZ\x1b[0m",
		},
		{
			`2024-02-17T06:56:10+05:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mT\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255m+05:00\x1b[0m",
		},
		{
			`2024-02-17T06:56:10.636960544-01:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mT\x1b[0m\x1b[38;2;252;167;234m06:56:10.636960544\x1b[0m\x1b[38;2;130;170;255m-01:00\x1b[0m",
		},
		{
			`2024-02-17t06:56:10z`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mt\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255mz\x1b[0m",
		},
		{
			`2024-02-17t06:56:10+05:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mt\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255m+05:00\x1b[0m",
		},
		{
			`2024-02-17t06:56:10.636960544-01:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mt\x1b[0m\x1b[38;2;252;167;234m06:56:10.636960544\x1b[0m\x1b[38;2;130;170;255m-01:00\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// time
func TestPatternsTime(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`23:42:12`, "\x1b[38;2;252;167;234m23:42:12\x1b[0m"},
		{`01:37:59.743`, "\x1b[38;2;252;167;234m01:37:59.743\x1b[0m"},
		{`17:49:37.034123`, "\x1b[38;2;252;167;234m17:49:37.034123\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// date-1
// date-2
// date-3
// date-4
func TestPatternsDates(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`1999-07-10`, "\x1b[38;2;192;153;255m1999-07-10\x1b[0m"},
		{`1999/07/10`, "\x1b[38;2;192;153;255m1999/07/10\x1b[0m"},
		{`07-10-1999`, "\x1b[38;2;192;153;255m07-10-1999\x1b[0m"},
		{`07/10/1999`, "\x1b[38;2;192;153;255m07/10/1999\x1b[0m"},
		{`27 Jan`, "\x1b[38;2;192;153;255m27 Jan\x1b[0m"},
		{`27 January`, "\x1b[38;2;192;153;255m27 January\x1b[0m"},
		{`27 Jan 2023`, "\x1b[38;2;192;153;255m27 Jan 2023\x1b[0m"},
		{`27 August 2023`, "\x1b[38;2;192;153;255m27 August 2023\x1b[0m"},
		{`27-Jan-2023`, "\x1b[38;2;192;153;255m27-Jan-2023\x1b[0m"},
		{`27-August-2023`, "\x1b[38;2;192;153;255m27-August-2023\x1b[0m"},
		{`27/Jan/2023`, "\x1b[38;2;192;153;255m27/Jan/2023\x1b[0m"},
		{`27/August/2023`, "\x1b[38;2;192;153;255m27/August/2023\x1b[0m"},
		{`Jan 27`, "\x1b[38;2;192;153;255mJan 27\x1b[0m"},
		{`January 27`, "\x1b[38;2;192;153;255mJanuary 27\x1b[0m"},
		{`Jan 27 2023`, "\x1b[38;2;192;153;255mJan 27 2023\x1b[0m"},
		{`August 27 2023`, "\x1b[38;2;192;153;255mAugust 27 2023\x1b[0m"},
		{`Jan-27-2023`, "\x1b[38;2;192;153;255mJan-27-2023\x1b[0m"},
		{`August-27-2023`, "\x1b[38;2;192;153;255mAugust-27-2023\x1b[0m"},
		{`Jan/27/2023`, "\x1b[38;2;192;153;255mJan/27/2023\x1b[0m"},
		{`August/27/2023`, "\x1b[38;2;192;153;255mAugust/27/2023\x1b[0m"},
		{`Mon 17`, "\x1b[38;2;192;153;255mMon 17\x1b[0m"},
		{`Sunday 3`, "\x1b[38;2;192;153;255mSunday 3\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// duration
func TestPatternsDuration(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`75.984854ms`, "\x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m"},
		{`5s`, "\x1b[38;2;79;214;190m5\x1b[0m\x1b[38;2;65;166;181ms\x1b[0m"},
		{`784m`, "\x1b[38;2;79;214;190m784\x1b[0m\x1b[38;2;65;166;181mm\x1b[0m"},
		{`7.5h`, "\x1b[38;2;79;214;190m7.5\x1b[0m\x1b[38;2;65;166;181mh\x1b[0m"},
		{`25d`, "\x1b[38;2;79;214;190m25\x1b[0m\x1b[38;2;65;166;181md\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// logfmt-general
func TestPatternsLogfmtGeneral(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`key=value`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0mvalue"},
		{`key=5s`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;79;214;190m5\x1b[0m\x1b[38;2;65;166;181ms\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// logfmt-string
func TestPatternsLogfmtString(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`key="value"`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0mvalue\x1b[38;2;154;173;236m\"\x1b[0m"},
		{`key="5s"`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m\x1b[38;2;79;214;190m5\x1b[0m\x1b[38;2;65;166;181ms\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// ipv4-address
func TestPatternsIPv4(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`127.0.0.1`, "\x1b[38;2;118;211;255m127.0.0.1\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`255.255.255.255`, "\x1b[38;2;118;211;255m255.255.255.255\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`0.0.0.0`, "\x1b[38;2;118;211;255m0.0.0.0\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`10.0.0.200/16`, "\x1b[38;2;118;211;255m10.0.0.200\x1b[0m\x1b[38;2;13;185;215m/16\x1b[0m"},
		{`10.0.0.0/8`, "\x1b[38;2;118;211;255m10.0.0.0\x1b[0m\x1b[38;2;13;185;215m/8\x1b[0m"},
		{`10.0.7.107:80`, "\x1b[38;2;118;211;255m10.0.7.107\x1b[0m\x1b[38;2;13;185;215m:80\x1b[0m"},
		{`8.9.10.237:8080`, "\x1b[38;2;118;211;255m8.9.10.237\x1b[0m\x1b[38;2;13;185;215m:8080\x1b[0m"},
		{`1.2.3.4:17846`, "\x1b[38;2;118;211;255m1.2.3.4\x1b[0m\x1b[38;2;13;185;215m:17846\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// ipv6-address
func TestPatternsIPv6(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{
			`2001:db8:4006:812::200e`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:4006:812::200e\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0db8:0000:cd30:0000:0000:0000:0000`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0db8:0000:cd30:0000:0000:0000:0000\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0db8::cd30:0:0:0:0`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0db8::cd30:0:0:0:0\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0db8:0:cd30::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0db8:0:cd30::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff02:0:0:0:0:1:ff00:0000`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:1:ff00:0000\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff02:0:0:0:0:1:ffff:ffff`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:1:ffff:ffff\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8::1234:5678`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8::1234:5678\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff02:0:0:0:0:0:0:2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:0:0:2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`fdf8:f53b:82e4::53`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mfdf8:f53b:82e4::53\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`fe80::200:5aee:feaa:20a2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mfe80::200:5aee:feaa:20a2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0000:4136:e378:`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0000:4136:e378:\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`8000:63bf:3fff:fdd2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m8000:63bf:3fff:fdd2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::1234:5678`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::1234:5678\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2000::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2000::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8:a0b:12f0::1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:a0b:12f0::1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:4:112:cd:65a:753:0:a1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:4:112:cd:65a:753:0:a1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0002:6c::430`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0002:6c::430\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:5::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:5::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`fe08::7:8`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mfe08::7:8\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2002:cb0a:3cdd:1::1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2002:cb0a:3cdd:1::1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8:8:4::2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:8:4::2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff01:0:0:0:0:0:0:2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff01:0:0:0:0:0:0:2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:0:0`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:0:0\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0000::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0000::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:192.0.2.47`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:192.0.2.47\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:0.0.0.0`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:0.0.0.0\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:255.255.255.255`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:255.255.255.255\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:10.0.0.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:10.0.0.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::192.168.0.1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::192.168.0.1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::255.255.255.255`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::255.255.255.255\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8:122:344::192.0.2.33`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:122:344::192.0.2.33\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`0:0:0:0:0:0:13.1.68.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m0:0:0:0:0:0:13.1.68.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`0:0:0:0:0:ffff:129.144.52.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m0:0:0:0:0:ffff:129.144.52.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::13.1.68.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::13.1.68.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:129.144.52.38`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:129.144.52.38\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`59fb:0:0:0:0:1005:cc57:6571`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m59fb:0:0:0:0:1005:cc57:6571\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`[2001:5::]:22`,
			"\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255m2001:5::\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:22\x1b[0m",
		},
		{
			`[2001:db8:4006:812::200e]:8080`,
			"\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255m2001:db8:4006:812::200e\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:8080\x1b[0m",
		},
		{
			`[ff02:0:0:0:0:1:ffff:ffff]:23456`,
			"\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:1:ffff:ffff\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:23456\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// mac-address
func TestPatternsMACAddress(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`3D:F2:C9:A6:B3:4F`, "\x1b[38;2;79;214;190m3D:F2:C9:A6:B3:4F\x1b[0m"},
		{`3D-F2-C9-A6-B3-4F`, "\x1b[38;2;79;214;190m3D-F2-C9-A6-B3-4F\x1b[0m"},
		{`3d:f2:c9:a6:b3:4f`, "\x1b[38;2;79;214;190m3d:f2:c9:a6:b3:4f\x1b[0m"},
		{`3d-f2-c9-a6-b3-4f`, "\x1b[38;2;79;214;190m3d-f2-c9-a6-b3-4f\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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

// uuid
func TestPatternsUUID(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{`0a99af43-0ad4-4237-b9cd-064966eb2803`, "\x1b[38;2;134;225;252m0a99af43-0ad4-4237-b9cd-064966eb2803\x1b[0m"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight-dark",
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
