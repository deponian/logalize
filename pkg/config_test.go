package logalize

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/muesli/termenv"
)

//go:embed builtins/logformats/good.yaml
//go:embed builtins/patterns/good.yaml
//go:embed builtins/words/good.yaml
var builtinsAllGood embed.FS

func TestConfigLoadBuiltinGood(t *testing.T) {
	colorProfile = termenv.TrueColor
	configData := `
formats:
  menetekel:
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

patterns:
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')
    fg: "#00ff00"

  ipv4-address:
    priority: 400
    regexp: (\d{1,3}(\.\d{1,3}){3})
    fg: "#ff0000"
    bg: "#ffff00"
    style: bold

  number:
    regexp: (\d+)
    bg: "#005050"

  http-status-code:
    priority: 300
    regexp: (\d\d\d)
    fg: "#ffffff"
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

words:
  friends:
    fg: "#f834b2"
    style: underline
    list:
      - "toni"
      - "wenzel"
  foes:
    fg: "#120fbb"
    style: underline
    list:
      - "argus"
      - "cletus"
`
	tests := []struct {
		plain   string
		colored string
	}{
		// log format
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m\n"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m\n"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m\n"},
		{`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 301 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[38;2;128;126;121m- \x1b[0m\x1b[38;2;118;73;158m- \x1b[0m\x1b[38;2;20;141;217m[16/Feb/2024:00:01:01 +0000] \x1b[0m\x1b[38;2;157;219;86m\"GET / HTTP/1.1\" \x1b[0m\x1b[38;2;0;255;255;1m301 \x1b[0m\x1b[38;2;125;125;125m162 \x1b[0m\x1b[38;2;58;225;240m\"-\" \x1b[0m\x1b[38;2;170;125;209m\"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1\"\x1b[0m\n"},

		// pattern
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m\n"},
		{"42", "\x1b[48;2;0;80;80m42\x1b[0m\n"},
		{"127.0.0.1", "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m\n"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[38;2;80;80;80m101\x1b[0m\n"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m\n"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m\n"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m\n"},
		{`status 103`, "status \x1b[38;2;80;80;80m103\x1b[0m\n"},
		{`status 200`, "status \x1b[38;2;0;255;0;53m200\x1b[0m\n"},
		{`status 302`, "status \x1b[38;2;0;255;255;9m302\x1b[0m\n"},
		{`status 404`, "status \x1b[38;2;255;0;0;7m404\x1b[0m\n"},
		{`status 503`, "status \x1b[38;2;255;0;255m503\x1b[0m\n"},
		{`status 700`, "status \x1b[38;2;255;255;255m700\x1b[0m\n"},

		// words
		{"untrue", "untrue\n"},
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m\n"},
		{"fail", "\x1b[38;2;240;108;97;1mfail\x1b[0m\n"},
		{"failed", "\x1b[38;2;240;108;97;1mfailed\x1b[0m\n"},
		{"wenzel", "\x1b[38;2;248;52;178;4mwenzel\x1b[0m\n"},
		{"argus", "\x1b[38;2;18;15;187;4margus\x1b[0m\n"},

		{"not true", "\x1b[38;2;240;108;97;1mnot true\x1b[0m\n"},
		{"Not true", "\x1b[38;2;240;108;97;1mNot true\x1b[0m\n"},
		{"wasn't true", "\x1b[38;2;240;108;97;1mwasn't true\x1b[0m\n"},
		{"won't true", "\x1b[38;2;240;108;97;1mwon't true\x1b[0m\n"},
		{"cannot complete", "\x1b[38;2;240;108;97;1mcannot complete\x1b[0m\n"},
		{"won't be completed", "\x1b[38;2;240;108;97;1mwon't be completed\x1b[0m\n"},
		{"cannot be completed", "\x1b[38;2;240;108;97;1mcannot be completed\x1b[0m\n"},
		{"should not be completed", "\x1b[38;2;240;108;97;1mshould not be completed\x1b[0m\n"},

		{"not false", "\x1b[38;2;81;250;138;1mnot false\x1b[0m\n"},
		{"Not false", "\x1b[38;2;81;250;138;1mNot false\x1b[0m\n"},
		{"wasn't false", "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m\n"},
		{"won't false", "\x1b[38;2;81;250;138;1mwon't false\x1b[0m\n"},
		{"cannot fail", "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m\n"},
		{"won't be failed", "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m\n"},
		{"cannot be failed", "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m\n"},
		{"should not be failed", "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m\n"},

		{"not toni", "not \x1b[38;2;248;52;178;4mtoni\x1b[0m\n"},
		{"Not wenzel", "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m\n"},
		{"wasn't argus", "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m\n"},
		{"won't cletus", "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m\n"},
		{"cannot toni", "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m\n"},
		{"won't be wenzel", "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m\n"},
		{"cannot be argus", "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m\n"},
		{"should not be cletus", "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m\n"},

		// patterns and words
		{`true bad fail 7.7.7.7`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m\n"},
		{`"true" and true`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m and \x1b[38;2;81;250;138;1mtrue\x1b[0m\n"},
		{`wenzel failed 127 times`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m \x1b[38;2;240;108;97;1mfailed\x1b[0m \x1b[38;2;80;80;80m127\x1b[0m times\n"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	options := Options{
		ConfigPath: "",
		NoBuiltins: false,
	}

	config, err := InitConfig(options, builtinsAllGood)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}
	configRaw := []byte(configData)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			Run(input, &output, config, lemmatizer)

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

//go:embed builtins/logformats/bad.yaml
var builtinsLogformatsBad embed.FS

//go:embed builtins/words/bad.yaml
var builtinsWordsBad embed.FS

//go:embed builtins/patterns/bad.yaml
var builtinsPatternsBad embed.FS

func TestConfigLoadBuiltinBad(t *testing.T) {
	colorProfile = termenv.TrueColor

	options := Options{
		ConfigPath: "",
		NoBuiltins: false,
	}

	t.Run("TestConfigLoadBuiltinLogformatsBadYAML", func(t *testing.T) {
		_, err := InitConfig(options, builtinsLogformatsBad)
		if err.Error() != "yaml: line 3: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	t.Run("TestConfigLoadBuiltinWordsBadYAML", func(t *testing.T) {
		_, err := InitConfig(options, builtinsWordsBad)
		if err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	t.Run("TestConfigLoadBuiltinWordsBadYAML", func(t *testing.T) {
		_, err := InitConfig(options, builtinsPatternsBad)
		if err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})
}

func TestConfigLoadUserDefinedGood(t *testing.T) {
	colorProfile = termenv.TrueColor
	configData := `
formats:
  menetekel:
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

patterns:
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')
    fg: "#00ff00"

  ipv4-address:
    priority: 400
    regexp: (\d{1,3}(\.\d{1,3}){3})
    fg: "#ff0000"
    bg: "#ffff00"
    style: bold

  number:
    regexp: (\d+)
    bg: "#005050"

  http-status-code:
    priority: 300
    regexp: (\d\d\d)
    fg: "#ffffff"
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

words:
  friends:
    fg: "#f834b2"
    style: underline
    list:
      - "toni"
      - "wenzel"
  foes:
    fg: "#120fbb"
    style: underline
    list:
      - "argus"
      - "cletus"
`
	tests := []struct {
		plain   string
		colored string
	}{
		// log format
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m\n"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m\n"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m\n"},

		// pattern
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m\n"},
		{"42", "\x1b[48;2;0;80;80m42\x1b[0m\n"},
		{"127.0.0.1", "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m\n"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[38;2;80;80;80m101\x1b[0m\n"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m\n"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m\n"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m\n"},
		{`status 103`, "status \x1b[38;2;80;80;80m103\x1b[0m\n"},
		{`status 200`, "status \x1b[38;2;0;255;0;53m200\x1b[0m\n"},
		{`status 302`, "status \x1b[38;2;0;255;255;9m302\x1b[0m\n"},
		{`status 404`, "status \x1b[38;2;255;0;0;7m404\x1b[0m\n"},
		{`status 503`, "status \x1b[38;2;255;0;255m503\x1b[0m\n"},
		{`status 700`, "status \x1b[38;2;255;255;255m700\x1b[0m\n"},

		// words
		{"untrue", "untrue\n"},
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m\n"},
		{"fail", "\x1b[38;2;240;108;97;1mfail\x1b[0m\n"},
		{"failed", "\x1b[38;2;240;108;97;1mfailed\x1b[0m\n"},
		{"wenzel", "\x1b[38;2;248;52;178;4mwenzel\x1b[0m\n"},
		{"argus", "\x1b[38;2;18;15;187;4margus\x1b[0m\n"},

		{"not true", "\x1b[38;2;240;108;97;1mnot true\x1b[0m\n"},
		{"Not true", "\x1b[38;2;240;108;97;1mNot true\x1b[0m\n"},
		{"wasn't true", "\x1b[38;2;240;108;97;1mwasn't true\x1b[0m\n"},
		{"won't true", "\x1b[38;2;240;108;97;1mwon't true\x1b[0m\n"},
		{"cannot complete", "\x1b[38;2;240;108;97;1mcannot complete\x1b[0m\n"},
		{"won't be completed", "\x1b[38;2;240;108;97;1mwon't be completed\x1b[0m\n"},
		{"cannot be completed", "\x1b[38;2;240;108;97;1mcannot be completed\x1b[0m\n"},
		{"should not be completed", "\x1b[38;2;240;108;97;1mshould not be completed\x1b[0m\n"},

		{"not false", "\x1b[38;2;81;250;138;1mnot false\x1b[0m\n"},
		{"Not false", "\x1b[38;2;81;250;138;1mNot false\x1b[0m\n"},
		{"wasn't false", "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m\n"},
		{"won't false", "\x1b[38;2;81;250;138;1mwon't false\x1b[0m\n"},
		{"cannot fail", "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m\n"},
		{"won't be failed", "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m\n"},
		{"cannot be failed", "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m\n"},
		{"should not be failed", "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m\n"},

		{"not toni", "not \x1b[38;2;248;52;178;4mtoni\x1b[0m\n"},
		{"Not wenzel", "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m\n"},
		{"wasn't argus", "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m\n"},
		{"won't cletus", "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m\n"},
		{"cannot toni", "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m\n"},
		{"won't be wenzel", "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m\n"},
		{"cannot be argus", "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m\n"},
		{"should not be cletus", "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m\n"},

		// patterns and words
		{`true bad fail 7.7.7.7`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m\n"},
		{`"true" and true`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m and \x1b[38;2;81;250;138;1mtrue\x1b[0m\n"},
		{`wenzel failed 127 times`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m \x1b[38;2;240;108;97;1mfailed\x1b[0m \x1b[38;2;80;80;80m127\x1b[0m times\n"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	userConfig := t.TempDir() + "/userConfig.yaml"
	configRaw := []byte(configData)
	err = os.WriteFile(userConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", userConfig, err)
	}

	options := Options{
		ConfigPath: userConfig,
		NoBuiltins: false,
	}

	config, err := InitConfig(options, builtinsAllGood)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		input := strings.NewReader(tt.plain)
		output := bytes.Buffer{}

		t.Run(testname, func(t *testing.T) {
			Run(input, &output, config, lemmatizer)

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

	options := Options{
		ConfigPath: userConfig,
		NoBuiltins: true,
	}

	t.Run("TestConfigLoadUserDefinedBadYAML", func(t *testing.T) {
		_, err := InitConfig(options, builtinsAllGood)
		if err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	options = Options{
		ConfigPath: userConfig + "error",
		NoBuiltins: true,
	}

	t.Run("TestConfigLoadUserDefinedFileDoesntExist", func(t *testing.T) {
		_, err := InitConfig(options, builtinsAllGood)
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

	options = Options{
		ConfigPath: userConfigReadOnly,
		NoBuiltins: false,
	}

	t.Run("TestConfigLoadUserDefinedReadOnly", func(t *testing.T) {
		_, err := InitConfig(options, builtinsAllGood)
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

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() failed with this error: %s", err)
	}
	defaultConfig := wd + "/.logalize.yaml"
	configRaw := []byte(configDataBadYAML)
	if ok, err := checkFileIsReadable(defaultConfig); ok {
		if err != nil {
			t.Errorf("checkFileIsReadable() failed with this error: %s", err)
		}
		err = os.Remove(defaultConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", defaultConfig, err)
		}
	}

	err = os.WriteFile(defaultConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", defaultConfig, err)
	}

	t.Cleanup(func() {
		err = os.Remove(defaultConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", defaultConfig, err)
		}
	})

	options := Options{
		ConfigPath: "",
		NoBuiltins: true,
	}

	t.Run("TestConfigLoadDefaultBadYAML", func(t *testing.T) {
		_, err := InitConfig(options, builtinsAllGood)
		if err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	err = os.Chmod(defaultConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", defaultConfig, err)
	}

	t.Run("TestConfigLoadDefaultReadOnly", func(t *testing.T) {
		_, err := InitConfig(options, builtinsAllGood)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitConfig() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}
