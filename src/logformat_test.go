package logalize

import (
	"bytes"
	"embed"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
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

// Below are the tests for all built-in log formats

// nginx-combined
func TestLogFormatNginxCombined(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 100 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"\x1b[38;2;238;204;159m127.0.0.1 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[16/Feb/2024:00:01:01 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET / HTTP/1.1\" \x1b[0m\x1b[38;2;0;0;255;1m100 \x1b[0m\x1b[38;2;99;109;166m162 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1\"\x1b[0m",
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 200 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"\x1b[38;2;238;204;159m127.0.0.1 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[16/Feb/2024:00:01:01 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET / HTTP/1.1\" \x1b[0m\x1b[38;2;0;255;0;1m200 \x1b[0m\x1b[38;2;99;109;166m162 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1\"\x1b[0m",
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 302 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"\x1b[38;2;238;204;159m127.0.0.1 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[16/Feb/2024:00:01:01 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET / HTTP/1.1\" \x1b[0m\x1b[38;2;0;255;255;1m302 \x1b[0m\x1b[38;2;99;109;166m162 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1\"\x1b[0m",
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 404 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"\x1b[38;2;238;204;159m127.0.0.1 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[16/Feb/2024:00:01:01 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET / HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m404 \x1b[0m\x1b[38;2;99;109;166m162 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1\"\x1b[0m",
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 503 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"\x1b[38;2;238;204;159m127.0.0.1 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[16/Feb/2024:00:01:01 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET / HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;255;1m503 \x1b[0m\x1b[38;2;99;109;166m162 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1\"\x1b[0m",
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

			result := strings.TrimSuffix(output.String(), "\n")

			if result != tt.colored {
				t.Errorf("got %v, want %v", result, tt.colored)
			}
		})
	}
}

// nginx-ingress-controller
func TestLogFormatNginxIngressController(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
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

			result := strings.TrimSuffix(output.String(), "\n")

			if result != tt.colored {
				t.Errorf("got %v, want %v", result, tt.colored)
			}
		})
	}
}

// klog
func TestLogFormatKlog(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{
			`I0410 13:18:43.650599       1 controller.go:175] "starting healthz server" logger="cert-manager.controller" address="[::]:9403"`,
			"\x1b[38;2;130;170;255;1mI0410 \x1b[0m\x1b[38;2;252;167;234m13:18:43.650599\x1b[0m\x1b[38;2;99;109;166m       1 \x1b[0m\x1b[38;2;137;221;255mcontroller.go\x1b[0m\x1b[38;2;99;109;166m:175\x1b[0m\x1b[38;2;255;150;108m] \x1b[0m\"\x1b[38;2;81;250;138;1mstarting\x1b[0m healthz server\"\x1b[38;2;154;173;236m logger\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0mcert-manager.controller\x1b[38;2;154;173;236m\"\x1b[0m\x1b[38;2;154;173;236m address\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255m::\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:9403\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m",
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

			result := strings.TrimSuffix(output.String(), "\n")

			if result != tt.colored {
				t.Errorf("got %v, want %v", result, tt.colored)
			}
		})
	}
}

// redis
func TestLogFormatRedis(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
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

			result := strings.TrimSuffix(output.String(), "\n")

			if result != tt.colored {
				t.Errorf("got %v, want %v", result, tt.colored)
			}
		})
	}
}
