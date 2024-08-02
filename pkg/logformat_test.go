package logalize

import (
	"regexp"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

func TestLogFormatsInit(t *testing.T) {
	configData := `
formats:
  test:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      fg: "#f5ce42"
    - regexp: ([^ ]+ )
      bg: "#764a9e"
    - regexp: (\[.+\] )
      style: bold
    - regexp: ("[^"]+")
      fg: "#9daf99"
      bg: "#76fb99"
      style: underline
    - regexp: (\d\d\d)
      alternatives:
        - regexp: (1\d\d)
          fg: "#505050"
        - regexp: (2\d\d)
          fg: "#00ff00"
          style: overline
        - regexp: (3\d\d)
          fg: "#00ffff"
          style: crossout
        - regexp: (4\d\d)
          fg: "#ff0000"
          style: reverse
        - regexp: (5\d\d)
          fg: "#ff00ff"
`

	correctCapGroupList := CapGroupList{
		[]CapGroup{
			{`(\d{1,3}(\.\d{1,3}){3} )`, "#f5ce42", "", "", nil, regexp.MustCompile(`(\d{1,3}(\.\d{1,3}){3} )`)},
			{`([^ ]+ )`, "", "#764a9e", "", nil, regexp.MustCompile(`([^ ]+ )`)},
			{`(\[.+\] )`, "", "", "bold", nil, regexp.MustCompile(`(\[.+\] )`)},
			{`("[^"]+")`, "#9daf99", "#76fb99", "underline", nil, regexp.MustCompile(`("[^"]+")`)},
			{
				`(\d\d\d)`, "", "", "",
				[]CapGroup{
					{`(1\d\d)`, "#505050", "", "", nil, regexp.MustCompile(`(1\d\d)`)},
					{`(2\d\d)`, "#00ff00", "", "overline", nil, regexp.MustCompile(`(2\d\d)`)},
					{`(3\d\d)`, "#00ffff", "", "crossout", nil, regexp.MustCompile(`(3\d\d)`)},
					{`(4\d\d)`, "#ff0000", "", "reverse", nil, regexp.MustCompile(`(4\d\d)`)},
					{`(5\d\d)`, "#ff00ff", "", "", nil, regexp.MustCompile(`(5\d\d)`)},
				},
				regexp.MustCompile(`(\d\d\d)`),
			},
		},
		regexp.MustCompile(`^(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3} ))(?P<capGroup1>(?:[^ ]+ ))(?P<capGroup2>(?:\[.+\] ))(?P<capGroup3>(?:"[^"]+"))(?P<capGroup4>(?:\d\d\d))$`),
	}

	colorProfile = termenv.TrueColor

	config := koanf.New(".")
	configRaw := []byte(configData)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestFormatsInit", func(t *testing.T) {
		if err := initLogFormats(config); err != nil {
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
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadYAML)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestFormatsInitBadYAML", func(t *testing.T) {
		if err := initLogFormats(config); err == nil {
			t.Errorf("InitLogFormats() should have failed")
		}
	})

	configDataBadRegExp := `
formats:
  test:
    - regexp: 'd{1,3}(\.\d{1,3}){3}'
      fg: "#f5ce42"
    - regexp: '[^ ]+'
      bg: "#764a9e"
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadRegExp)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestFormatsInitBadRegExp", func(t *testing.T) {
		if err := initLogFormats(config); err == nil {
			t.Errorf("InitLogFormats() should have failed")
		}
	})
}

func TestLogFormatHighlight(t *testing.T) {
	configData := `
formats:
  test:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      fg: "#f5ce42"
    - regexp: ([^ ]+ )
      bg: "#764a9e"
    - regexp: (\[.+\] )
      style: bold
    - regexp: ("[^"]+")
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

	for _, tt := range tests {
		testname := tt.plain
		config := koanf.New(".")
		configRaw := []byte(configData)
		if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
			t.Errorf("Error during config loading: %s", err)
		}

		if err := initLogFormats(config); err != nil {
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
