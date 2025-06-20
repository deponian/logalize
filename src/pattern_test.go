package logalize

import (
	"embed"
	"fmt"
	"os"
	"regexp"
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
