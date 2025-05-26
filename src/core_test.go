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
	"github.com/muesli/termenv"
)

func TestRunGood(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS
	configData := `
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

  portafisco-patterns:
    - regexp: (\d{1} - )
      name: one
    - regexp: (.*)
      name: two

  portafisco-words:
    - regexp: (\d{2} - )
      name: one
    - regexp: (.*)
      name: two

  portafisco-patterns-and-words:
    - regexp: (\d{3} - )
      name: one
    - regexp: (.*)
      name: two

  portafisco-combined:
    - regexp: (\d{4} - )
      name: one
    - regexp: (".*" )
      name: two
    - regexp: (".*" )
      name: three
    - regexp: («.*»)
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
  good:
    - "true"
    - "complete"
  bad:
    - "false"
    - "fail"
  friends:
    - "toni"
    - "wenzel"
  foes:
    - "argus"
    - "cletus"

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

      portafisco-patterns:
        one:
          fg: "#f5ce42"
        two:
          style: patterns

      portafisco-words:
        one:
          fg: "#f5ce42"
        two:
          style: words

      portafisco-patterns-and-words:
        one:
          fg: "#f5ce42"
        two:
          style: patterns-and-words

      portafisco-combined:
        one:
          fg: "#f5ce42"
        two:
          style: patterns
        three:
          style: words
        four:
          style: patterns-and-words

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
      good:
        fg: "#52fa8a"
        style: bold
      bad:
        bg: "#f06c62"
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

		// colored input (ANSI escape sequences should be successfully stripped)
		{"127.0.0.1 - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \"testing\"", "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m\n"},
		{"\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m", "\x1b[38;2;0;255;0m\"string\"\x1b[0m\n"},
		{"\x1b[0m\x1b[1m\x1b[31mtrue\x1b[0m", "\x1b[38;2;81;250;138;1mtrue\x1b[0m\n"},
		{"\x1b[0m\x1b[1m\x1b[31mfail\x1b[0m", "\x1b[48;2;240;108;97mfail\x1b[0m\n"},
		{"true \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m fail 7.7.7.7", "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[48;2;240;108;97mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m\n"},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	testConfig := t.TempDir() + "/testConfig.yaml"
	configRaw := []byte(configData)
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

func TestRunBadInit(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS

	input := strings.NewReader("test")
	output := bytes.Buffer{}
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	configDataBadFormats := `
formats:
  test:
    regexp: bad
themes:
  test:
`
	testConfig := t.TempDir() + "/testConfig.yaml"
	configRaw := []byte(configDataBadFormats)
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
	t.Run("TestRunBadFormats", func(t *testing.T) {
		err := Run(input, &output, lemmatizer)
		if err == nil || err.Error() != `[log format: test] [capture group: bad] regexp bad must start with ( and end with )` {
			t.Errorf("Run() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	configDataBadRegExps := `
patterns:
  string:priority: 100
themes:
  test:
`
	testConfig = t.TempDir() + "/testConfig.yaml"
	configRaw = []byte(configDataBadRegExps)
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
	t.Run("TestRunBadRegExps", func(t *testing.T) {
		err := Run(input, &output, lemmatizer)
		if err == nil || err.Error() != "decoding failed due to the following error(s):\n\n'[0]' expected a map, got 'int'" {
			t.Errorf("Run() should have failed with *fmt.wrapError, got: [%T] %s", err, err)
		}
	})

	configDataBadWords := `
words:
  good:
    - []
themes:
  test:
`
	testConfig = t.TempDir() + "/testConfig.yaml"
	configRaw = []byte(configDataBadWords)
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
	t.Run("TestRunBadWords", func(t *testing.T) {
		err := Run(input, &output, lemmatizer)
		if err == nil {
			t.Errorf("Run() should have failed with an error")
		}
	})
}

func TestRunBadWriter(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}
	configData := `
formats:
  test:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      name: one
    - regexp: ("[^"]+")
      name: two

themes:
  test:
    formats:
      test:
        one:
          fg: "#f5ce42"
        two:
          fg: "#9daf99"
          bg: "#76fb99"
          style: underline
`
	testConfig := t.TempDir() + "/testConfig.yaml"
	configRaw := []byte(configData)
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
		err := Run(input, file, lemmatizer)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("Run() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	input = strings.NewReader(`127.0.0.1 "testing"`)
	t.Run("TestRunBadWriterLogFormat", func(t *testing.T) {
		err := Run(input, file, lemmatizer)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("Run() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}
