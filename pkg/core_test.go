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

func TestRunGood(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS
	configData := `
formats:
  menetekel:
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

  portafisco-patterns:
    - pattern: (\d{1} - )
      fg: "#f5ce42"
    - pattern: (.*)
      style: patterns

  portafisco-words:
    - pattern: (\d{2} - )
      fg: "#f5ce42"
    - pattern: (.*)
      style: words

  portafisco-patterns-and-words:
    - pattern: (\d{3} - )
      fg: "#f5ce42"
    - pattern: (.*)
      style: patterns-and-words

  portafisco-combined:
    - pattern: (\d{4} - )
      fg: "#f5ce42"
    - pattern: (".*" )
      style: patterns
    - pattern: (".*" )
      style: words
    - pattern: («.*»)
      style: patterns-and-words

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
      - "complete"
  bad:
    bg: "#f06c62"
    list:
      - "false"
      - "fail"
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
		{`0 - hello bye`, "\x1b[38;2;245;206;65m0 - \x1b[0mhello bye\n"},
		{`1 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;206;65m1 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m true toni rufus\n"},
		{`22 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;206;65m22 - \x1b[0m777 hello 1.1.1.1 \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus\n"},
		{`333 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;206;65m333 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus\n"},
		{`4444 - "777 hello 1.1.1.1 true toni rufus" "777 hello 1.1.1.1 true toni rufus" «777 hello 1.1.1.1 true toni rufus»`, "\x1b[38;2;245;206;65m4444 - \x1b[0m\x1b[38;2;0;255;0m\"777 hello 1.1.1.1 true toni rufus\"\x1b[0m \"777 hello 1.1.1.1 \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus\" «\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus»\n"},

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
		{"fail", "\x1b[48;2;240;108;97mfail\x1b[0m\n"},
		{"failed", "\x1b[48;2;240;108;97mfailed\x1b[0m\n"},
		{"wenzel", "\x1b[38;2;248;52;178;4mwenzel\x1b[0m\n"},
		{"argus", "\x1b[38;2;18;15;187;4margus\x1b[0m\n"},

		{"not true", "\x1b[48;2;240;108;97mnot true\x1b[0m\n"},
		{"Not true", "\x1b[48;2;240;108;97mNot true\x1b[0m\n"},
		{"wasn't true", "\x1b[48;2;240;108;97mwasn't true\x1b[0m\n"},
		{"won't true", "\x1b[48;2;240;108;97mwon't true\x1b[0m\n"},
		{"cannot complete", "\x1b[48;2;240;108;97mcannot complete\x1b[0m\n"},
		{"won't be completed", "\x1b[48;2;240;108;97mwon't be completed\x1b[0m\n"},
		{"cannot be completed", "\x1b[48;2;240;108;97mcannot be completed\x1b[0m\n"},
		{"should not be completed", "\x1b[48;2;240;108;97mshould not be completed\x1b[0m\n"},

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
		{`true bad fail 7.7.7.7`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[48;2;240;108;97mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m\n"},
		{`"true" and true`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m and \x1b[38;2;81;250;138;1mtrue\x1b[0m\n"},
		{`wenzel failed 127 times`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m \x1b[48;2;240;108;97mfailed\x1b[0m \x1b[38;2;80;80;80m127\x1b[0m times\n"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	options := Options{
		ConfigPath: "",
		NoBuiltins: true,
	}
	config, err := InitConfig(options, builtins)
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
			Run(input, &output, config, builtins, lemmatizer)

			if output.String() != tt.colored {
				t.Errorf("got %v, want %v", output.String(), tt.colored)
			}
		})
	}
}

func TestRunBadInit(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS
	options := Options{
		ConfigPath: "",
		NoBuiltins: true,
	}
	input := strings.NewReader("test")
	output := bytes.Buffer{}
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	configDataBadFormats := `
formats:
  test:
    pattern: bad
`
	config, err := InitConfig(options, builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}
	configRaw := []byte(configDataBadFormats)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestRunBadFormats", func(t *testing.T) {
		err := Run(input, &output, config, builtins, lemmatizer)
		if err.Error() != `[log format: test] capture group pattern bad doesn't match ^\(.+\)$ pattern` {
			t.Errorf("Run() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	configDataBadPatterns := `
patterns:
  string:priority: 100
`
	config, err = InitConfig(options, builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}
	configRaw = []byte(configDataBadPatterns)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestRunBadPatterns", func(t *testing.T) {
		err := Run(input, &output, config, builtins, lemmatizer)
		if err.Error() != "'' expected a map, got 'int'" {
			t.Errorf("Run() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	configDataBadWords := `
words:
  good:err: bad
`
	config, err = InitConfig(options, builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}
	configRaw = []byte(configDataBadWords)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestRunBadWords", func(t *testing.T) {
		err := Run(input, &output, config, builtins, lemmatizer)
		if err.Error() != "'' expected a map, got 'string'" {
			t.Errorf("Run() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})
}

func TestRunBadWriter(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS
	options := Options{
		ConfigPath: "",
		NoBuiltins: true,
	}
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}
	configData := `
formats:
  test:
    - pattern: (\d{1,3}(\.\d{1,3}){3} )
      fg: "#f5ce42"
    - pattern: ("[^"]+")
      fg: "#9daf99"
      bg: "#76fb99"
      style: underline
`
	config, err := InitConfig(options, builtins)
	if err != nil {
		t.Errorf("InitConfig() failed with this error: %s", err)
	}
	configRaw := []byte(configData)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	filename := t.TempDir() + "/test.yaml"
	file, err := os.Create(filename)
	if err != nil {
		t.Errorf("Wasn't able to create test %s: %s", filename, err)
	}
	err = file.Chmod(0444)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", filename, err)
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Wasn't able to close %s: %s", filename, err)
	}

	input := strings.NewReader("test")
	t.Run("TestRunBadWriterNotLogFormat", func(t *testing.T) {
		err := Run(input, file, config, builtins, lemmatizer)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("Run() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	input = strings.NewReader(`127.0.0.1 "testing"`)
	t.Run("TestRunBadWriterLogFormat", func(t *testing.T) {
		err := Run(input, file, config, builtins, lemmatizer)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("Run() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}
