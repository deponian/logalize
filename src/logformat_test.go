package logalize

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

func TestLogFormatsInit(t *testing.T) {
	configData := `
formats:
  test:
    - pattern: (\d{1,3}(\.\d{1,3}){3} )
      fg: "#f5ce42"
    - pattern: ([^ ]+ )
      bg: "#764a9e"
    - pattern: (\[.+\] )
      style: bold
    - pattern: ("[^"]+")
      fg: "#9daf99"
      bg: "#76fb99"
      style: underline
    - pattern: (\d\d\d)
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
`

	correctCapGroups := CapGroupList{
		{`(\d{1,3}(\.\d{1,3}){3} )`, "#f5ce42", "", "", nil, nil},
		{`([^ ]+ )`, "", "#764a9e", "", nil, nil},
		{`(\[.+\] )`, "", "", "bold", nil, nil},
		{`("[^"]+")`, "#9daf99", "#76fb99", "underline", nil, nil},
		{
			`(\d\d\d)`, "#ffffff", "", "",
			CapGroupList{
				{`(1\d\d)`, "#505050", "", "", nil, nil},
				{`(2\d\d)`, "#00ff00", "", "overline", nil, nil},
				{`(3\d\d)`, "#00ffff", "", "crossout", nil, nil},
				{`(4\d\d)`, "#ff0000", "", "reverse", nil, nil},
				{`(5\d\d)`, "#ff00ff", "", "", nil, nil},
			},
			nil,
		},
	}

	colorProfile = termenv.TrueColor

	config := koanf.New(".")
	configRaw := []byte(configData)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestFormatsInit", func(t *testing.T) {
		formats, err := initLogFormats(config)
		if err != nil {
			t.Errorf("InitLogFormats() failed with this error: %s", err)
		}

		// check alternatives' regexps
		// yeah, I know, it's disgusting, don't look at it, look away
		checkRegexp := func(format *LogFormat, alt int, correctRegexp string) {
			if format.CapGroups[4].Alternatives[alt].Regexp.String() != correctRegexp {
				t.Errorf("got %v, want %v", formats[0].CapGroups[4].Alternatives[alt].Regexp.String(), correctRegexp)
			}
			format.CapGroups[4].Alternatives[alt].Regexp = nil
		}
		checkRegexp(&formats[0], 0, `(1\d\d)`)
		checkRegexp(&formats[0], 1, `(2\d\d)`)
		checkRegexp(&formats[0], 2, `(3\d\d)`)
		checkRegexp(&formats[0], 3, `(4\d\d)`)
		checkRegexp(&formats[0], 4, `(5\d\d)`)

		// check other fields
		if !cmp.Equal(formats[0].CapGroups, correctCapGroups) {
			t.Errorf("got %v, want %v", formats[0].CapGroups, correctCapGroups)
		}
	})

	configDataBadYAML := `
formats:
  test:
  pattern: bad
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadYAML)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestFormatsInitBadYAML", func(t *testing.T) {
		_, err := initLogFormats(config)
		if err == nil {
			t.Errorf("InitLogFormats() should have failed")
		}
	})

	configDataBadPattern := `
formats:
  test:
    - pattern: 'd{1,3}(\.\d{1,3}){3}'
      fg: "#f5ce42"
    - pattern: '[^ ]+'
      bg: "#764a9e"
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadPattern)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestFormatsInitBadPattern", func(t *testing.T) {
		_, err := initLogFormats(config)
		if err == nil {
			t.Errorf("InitLogFormats() should have failed")
		}
	})
}

func TestLogFormatCheckCapGroups(t *testing.T) {
	tests := []struct {
		err string
		lf  LogFormat
	}{
		{
			"%!s(<nil>)",
			LogFormat{
				"testNoErr", CapGroupList{
					{`(\d+:)`, "", "", "", CapGroupList{}, nil},
					{`(\d+:)`, "", "", "bold", CapGroupList{}, nil},
					{`(\d+:)`, "", "#ff00ff", "", CapGroupList{}, nil},
					{`(\d+:)`, "", "#ff0000", "underline", CapGroupList{}, nil},
					{`(\d+:)`, "#0f0f0f", "", "", CapGroupList{}, nil},
					{`(\d+:)`, "#0f0f0f", "", "faint", CapGroupList{}, nil},
					{`(\d+:)`, "#0f0f0f", "#ff00ff", "", CapGroupList{}, nil},
					{`(\d+:)`, "#0f0f0f", "#ff0000", "italic", CapGroupList{}, nil},
					{`(\d+:)`, "#0f0f0f", "1", "overline", CapGroupList{}, nil},
					{`(\d+:)`, "37", "#ff0000", "crossout", CapGroupList{}, nil},
					{`(\d+:)`, "214", "15", "reverse", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			fmt.Sprintf(`[log format: testPatternErr1] capture group pattern () doesn't match %s pattern`, capGroupRegexp),
			LogFormat{
				"testPatternErr1", CapGroupList{
					{`()`, "", "", "", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			`[log format: testPatternErr2] empty patterns are not allowed`,
			LogFormat{
				"testPatternErr2", CapGroupList{
					{``, "", "", "", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			fmt.Sprintf(`[log format: testPatternErr3] capture group pattern ) doesn't match %s pattern`, capGroupRegexp),
			LogFormat{
				"testPatternErr3", CapGroupList{
					{`)`, "", "", "", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			fmt.Sprintf(`[log format: testPatternErr4] capture group pattern (\d\d-\d\d-\d\d doesn't match %s pattern`, capGroupRegexp),
			LogFormat{
				"testPatternErr4", CapGroupList{
					{`(\d\d-\d\d-\d\d`, "", "", "", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			fmt.Sprintf(`[log format: testForegroundErr1] [capture group: (\d+)] foreground color ff00df doesn't match %s pattern`, colorRegexp),
			LogFormat{
				"testForegroundErr1", CapGroupList{
					{`(\d+)`, "ff00df", "", "", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			fmt.Sprintf(`[log format: testBackgroundErr1] [capture group: (\d+)] background color 7000 doesn't match %s pattern`, colorRegexp),
			LogFormat{
				"testBackgroundErr1", CapGroupList{
					{`(\d+)`, "", "7000", "", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			fmt.Sprintf(`[log format: testStyleErr1] [capture group: (\d+)] style NotAStyle doesn't match %s pattern`, styleRegexp),
			LogFormat{
				"testStyleErr1", CapGroupList{
					{`(\d+)`, "", "", "NotAStyle", CapGroupList{}, nil},
				}, nil,
			},
		},
	}

	colorProfile = termenv.TrueColor

	for _, tt := range tests {
		testname := tt.lf.Name
		t.Run(testname, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.lf.checkCapGroups()); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}

func TestLogFormatBuildRegexp(t *testing.T) {
	tests := []struct {
		correctRegexp string
		logFormat     LogFormat
	}{
		{
			`^(?P<capGroup0>\d+:)(?P<capGroup1>[^\[\]]*)(?P<capGroup2>\[test\])$`,
			LogFormat{
				"test1", CapGroupList{
					{`(\d+:)`, "#ff0000", "", "", CapGroupList{}, nil},
					{`([^\[\]]*)`, "164", "", "underline", CapGroupList{}, nil},
					{`(\[test\])`, "#00ff00", "#ffff00", "", CapGroupList{}, nil},
				}, nil,
			},
		},
		{
			`^(?P<capGroup0>\d+:)(?P<capGroup1>hello)$`,
			LogFormat{
				"test2", CapGroupList{
					{`(\d+:)`, "", "", "", CapGroupList{}, nil},
					{`(hello)`, "", "", "underline", CapGroupList{}, nil},
				}, nil,
			},
		},
	}

	colorProfile = termenv.TrueColor

	for _, tt := range tests {
		testname := tt.logFormat.Name
		t.Run(testname, func(t *testing.T) {
			reStr := tt.logFormat.buildRegexp().String()
			if reStr != tt.correctRegexp {
				t.Errorf("got %s, want %s", reStr, tt.correctRegexp)
			}
		})
	}
}

func TestLogFormatHighlight(t *testing.T) {
	configData := `
formats:
  test:
    - pattern: (\d{1,3}(\.\d{1,3}){3} )
      fg: "#f5ce42"
    - pattern: ([^ ]+ )
      bg: "#764a9e"
    - pattern: (\[.+\] )
      style: bold
    - pattern: ("[^"]+")
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
		formats, err := initLogFormats(config)
		if err != nil {
			t.Errorf("InitLogFormats() failed with this error: %s", err)
		}
		t.Run(testname, func(t *testing.T) {
			colored := formats[0].highlight(tt.plain)
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}
