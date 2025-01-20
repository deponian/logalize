package logalize

import (
	"embed"
	"os"
	"regexp"
	"testing"

	"github.com/muesli/termenv"
)

func TestLogFormatsInit(t *testing.T) {
	configData := `
formats:
  test:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      name: one
    - regexp: ([^ ]+ )
      name: two
    - regexp: (\[.+\] )
      name: three
    - regexp: ("[^"]+")
      name: four
    - regexp: (\d\d\d)
      name: five
      alternatives:
        - regexp: (1\d\d)
          name: 1
        - regexp: (2\d\d)
          name: 2
        - regexp: (3\d\d)
          name: 3
        - regexp: (4\d\d)
          name: 4
        - regexp: (5\d\d)
          name: 5

themes:
  test:
    formats:
      test:
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
        five:
          1:
            fg: "#505050"
          2:
            fg: "#00ff00"
            style: overline
          3:
            fg: "#00ffff"
            style: crossout
          4:
            fg: "#ff0000"
            style: reverse
          5:
            fg: "#ff00ff"
`

	correctCapGroupList := CapGroupList{
		[]CapGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "#f5ce42", "", "", nil, nil},
			{"two", `([^ ]+ )`, "", "#764a9e", "", nil, nil},
			{"three", `(\[.+\] )`, "", "", "bold", nil, nil},
			{"four", `("[^"]+")`, "#9daf99", "#76fb99", "underline", nil, nil},
			{
				"five",
				`(\d\d\d)`, "", "", "",
				[]CapGroup{
					{"1", `(1\d\d)`, "#505050", "", "", nil, regexp.MustCompile(`(1\d\d)`)},
					{"2", `(2\d\d)`, "#00ff00", "", "overline", nil, regexp.MustCompile(`(2\d\d)`)},
					{"3", `(3\d\d)`, "#00ffff", "", "crossout", nil, regexp.MustCompile(`(3\d\d)`)},
					{"4", `(4\d\d)`, "#ff0000", "", "reverse", nil, regexp.MustCompile(`(4\d\d)`)},
					{"5", `(5\d\d)`, "#ff00ff", "", "", nil, regexp.MustCompile(`(5\d\d)`)},
				},
				nil,
			},
		},
		regexp.MustCompile(`^(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3} ))(?P<capGroup1>(?:[^ ]+ ))(?P<capGroup2>(?:\[.+\] ))(?P<capGroup3>(?:"[^"]+"))(?P<capGroup4>(?:\d\d\d))$`),
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

	t.Run("TestFormatsInit", func(t *testing.T) {
		if err := initLogFormats(); err != nil {
			t.Errorf("InitLogFormats() failed with this error: %s", err)
		}

		if err := compareCapGroupLists(*LogFormats[0].CapGroups, correctCapGroupList); err != nil {
			t.Errorf("%s", err)
		}
	})

	configDataBadYAML := `
formats:
  test:
  regexp: bad
themes:
  test:
`
	testConfig = t.TempDir() + "/testConfig.yaml"
	configRaw = []byte(configDataBadYAML)
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

	t.Run("TestFormatsInitBadYAML", func(t *testing.T) {
		if err := initLogFormats(); err == nil {
			t.Errorf("InitLogFormats() should have failed")
		}
	})

	configDataBadRegExp := `
formats:
  test:
    - regexp: 'd{1,3}(\.\d{1,3}){3}'
      name: one
    - regexp: '[^ ]+'
      name: two

themes:
  test:
    formats:
      test:
        one:
          fg: "#f5ce42"
        two:
          bg: "#764a9e"
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

	t.Run("TestFormatsInitBadRegExp", func(t *testing.T) {
		if err := initLogFormats(); err == nil {
			t.Errorf("InitLogFormats() should have failed")
		}
	})
}

func TestLogFormatHighlight(t *testing.T) {
	configData := `
formats:
  test:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      name: one
    - regexp: ([^ ]+ )
      name: two
    - regexp: (\[.+\] )
      name: three
    - regexp: ("[^"]+")
      name: four

themes:
  test:
    formats:
      test:
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
`
	tests := []struct {
		plain   string
		colored string
	}{
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m"},
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

	for _, tt := range tests {
		testname := tt.plain

		err := InitConfig(builtins)
		if err != nil {
			t.Errorf("InitConfig() failed with this error: %s", err)
		}

		if err := initLogFormats(); err != nil {
			t.Errorf("InitLogFormats() failed with this error: %s", err)
		}
		t.Run(testname, func(t *testing.T) {
			colored := LogFormats[0].highlight(tt.plain)
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}
