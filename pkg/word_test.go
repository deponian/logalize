package logalize

import (
	"fmt"
	"testing"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

func TestWordsInit(t *testing.T) {
	configDataGood := `
words:
  good:
    fg: "#52fa8a"
    style: bold
    list:
      - "true"
  bad:
    bg: "#f06c62"
    list:
      - "fail"
      - "fatal"
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
	goodGroup := WordGroup{"good", []string{"true"}, "#52fa8a", "", "bold"}
	badGroup := WordGroup{"bad", []string{"fail", "fatal"}, "", "#f06c62", ""}
	otherGroups := []WordGroup{
		{"foes", []string{"argus", "cletus"}, "#120fbb", "", "underline"},
		{"friends", []string{"toni", "wenzel"}, "#f834b2", "", "underline"},
	}

	colorProfile = termenv.TrueColor

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	config := koanf.New(".")
	configRaw := []byte(configDataGood)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestWordsInitGood", func(t *testing.T) {
		if err := initWords(config, lemmatizer); err != nil {
			t.Errorf("InitWords() failed with this error: %s", err)
		}
		if !cmp.Equal(Words.Good, goodGroup) {
			t.Errorf("got %v, want %v", Words.Good, goodGroup)
		}
		if !cmp.Equal(Words.Bad, badGroup) {
			t.Errorf("got %v, want %v", Words.Bad, badGroup)
		}
		if !cmp.Equal(Words.Other, otherGroups) {
			t.Errorf("got %v, want %v", Words.Other, otherGroups)
		}
	})

	configDataBadYAML := `
words:
  good:err: bad
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadYAML)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestWordsInitBadYAML", func(t *testing.T) {
		if err := initWords(config, lemmatizer); err == nil {
			t.Errorf("InitWords() should have failed")
		}
	})

	configDataBadStyle := `
words:
  good:
    fg: "#52fa8a"
    style: patterns
    list:
      - "true"
`
	config = koanf.New(".")
	configRaw = []byte(configDataBadStyle)
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}
	t.Run("TestWordsInitBadStyle", func(t *testing.T) {
		if err := initWords(config, lemmatizer); err == nil {
			t.Errorf("InitWords() should have failed")
		}
	})
}

func TestWordsCheck(t *testing.T) {
	tests := []struct {
		err string
		wg  WordGroup
	}{
		{
			"%!s(<nil>)",
			WordGroup{"testNoErr", []string{"test"}, "#ff0000", "#00ff00", "bold"},
		},
		{
			fmt.Sprintf(`[word group: testForegroundErr] foreground color #ff00xd doesn't match %s pattern`, colorRegexp),
			WordGroup{"testForegroundErr", []string{"test"}, "#ff00xd", "", ""},
		},
		{
			fmt.Sprintf(`[word group: testBackgroundErr] background color hello doesn't match %s pattern`, colorRegexp),
			WordGroup{"testBackgroundErr", []string{"test"}, "", "hello", ""},
		},
		{
			fmt.Sprintf(`[word group: testStyleErr1] style words doesn't match %s pattern`, nonRecursiveStyleRegexp),
			WordGroup{"testStyleErr1", []string{"test"}, "", "", "words"},
		},
		{
			fmt.Sprintf(`[word group: testStyleErr2] style patterns doesn't match %s pattern`, nonRecursiveStyleRegexp),
			WordGroup{"testStyleErr2", []string{"test"}, "", "", "patterns"},
		},
		{
			fmt.Sprintf(`[word group: testStyleErr3] style patterns-and-words doesn't match %s pattern`, nonRecursiveStyleRegexp),
			WordGroup{"testStyleErr3", []string{"test"}, "", "", "patterns-and-words"},
		},
	}

	colorProfile = termenv.TrueColor

	for _, tt := range tests {
		testname := tt.wg.Name
		t.Run(testname, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.wg.check()); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}

func TestWordsHighlightWord(t *testing.T) {
	configData := `
words:
  good:
    fg: "#52fa8a"
    style: bold
    list:
      - "true"
  bad:
    bg: "#f06c62"
    style: underline
    list:
      - "fail"
      - "fatal"
  friends:
    fg: "#f834b2"
    style: faint
    list:
      - "toni"
      - "wenzel"
  foes:
    fg: "#120fbb"
    style: italic
    list:
      - "argus"
      - "cletus"
`
	tests := []struct {
		plain   string
		colored string
	}{
		{"hello", "hello"},
		{"untrue", "untrue"},
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{"fail", "\x1b[48;2;240;108;97;4mfail\x1b[0m"},
		{"failed", "\x1b[48;2;240;108;97;4mfailed\x1b[0m"},
		{"wenzel", "\x1b[38;2;248;52;178;2mwenzel\x1b[0m"},
		{"argus", "\x1b[38;2;18;15;187;3margus\x1b[0m"},
	}

	colorProfile = termenv.TrueColor

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		config := koanf.New(".")
		configRaw := []byte(configData)
		if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
			t.Errorf("Error during config loading: %s", err)
		}
		if err := initWords(config, lemmatizer); err != nil {
			t.Errorf("InitWords() failed with this error: %s", err)
		}
		t.Run(testname, func(t *testing.T) {
			colored := Words.highlightWord(tt.plain)
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}

func TestWordsHighlightNegatedWord(t *testing.T) {
	configData := `
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
		{"not hello", "not hello"},

		{"not true", "\x1b[48;2;240;108;97mnot true\x1b[0m"},
		{"Not true", "\x1b[48;2;240;108;97mNot true\x1b[0m"},
		{"wasn't true", "\x1b[48;2;240;108;97mwasn't true\x1b[0m"},
		{"won't true", "\x1b[48;2;240;108;97mwon't true\x1b[0m"},
		{"cannot complete", "\x1b[48;2;240;108;97mcannot complete\x1b[0m"},
		{"won't be completed", "\x1b[48;2;240;108;97mwon't be completed\x1b[0m"},
		{"cannot be completed", "\x1b[48;2;240;108;97mcannot be completed\x1b[0m"},
		{"should not be completed", "\x1b[48;2;240;108;97mshould not be completed\x1b[0m"},

		{"not false", "\x1b[38;2;81;250;138;1mnot false\x1b[0m"},
		{"Not false", "\x1b[38;2;81;250;138;1mNot false\x1b[0m"},
		{"wasn't false", "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m"},
		{"won't false", "\x1b[38;2;81;250;138;1mwon't false\x1b[0m"},
		{"cannot fail", "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m"},
		{"won't be failed", "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m"},
		{"cannot be failed", "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m"},
		{"should not be failed", "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m"},

		{"not toni", "not \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"Not wenzel", "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"wasn't argus", "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"won't cletus", "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
		{"cannot toni", "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"won't be wenzel", "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"cannot be argus", "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"should not be cletus", "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
	}

	colorProfile = termenv.TrueColor

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		config := koanf.New(".")
		configRaw := []byte(configData)
		if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
			t.Errorf("Error during config loading: %s", err)
		}
		if err := initWords(config, lemmatizer); err != nil {
			t.Errorf("InitWords() failed with this error: %s", err)
		}
		t.Run(testname, func(t *testing.T) {
			m := negatedWordRegexp.FindStringSubmatchIndex(tt.plain)
			colored := Words.highlightNegatedWord(tt.plain[m[0]:m[1]], tt.plain[m[2]:m[3]], tt.plain[m[4]:m[5]])
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}

func TestWordsHighlight(t *testing.T) {
	configData := `
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
		{"hello", "hello"},
		{"untrue", "untrue"},
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{"fail", "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{"failed", "\x1b[48;2;240;108;97mfailed\x1b[0m"},
		{"wenzel", "\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},

		{"argus", "\x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"not true", "\x1b[48;2;240;108;97mnot true\x1b[0m"},
		{"Not true", "\x1b[48;2;240;108;97mNot true\x1b[0m"},
		{"wasn't true", "\x1b[48;2;240;108;97mwasn't true\x1b[0m"},
		{"won't true", "\x1b[48;2;240;108;97mwon't true\x1b[0m"},
		{"cannot complete", "\x1b[48;2;240;108;97mcannot complete\x1b[0m"},
		{"won't be completed", "\x1b[48;2;240;108;97mwon't be completed\x1b[0m"},
		{"cannot be completed", "\x1b[48;2;240;108;97mcannot be completed\x1b[0m"},
		{"should not be completed", "\x1b[48;2;240;108;97mshould not be completed\x1b[0m"},

		{"not false", "\x1b[38;2;81;250;138;1mnot false\x1b[0m"},
		{"Not false", "\x1b[38;2;81;250;138;1mNot false\x1b[0m"},
		{"wasn't false", "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m"},
		{"won't false", "\x1b[38;2;81;250;138;1mwon't false\x1b[0m"},
		{"cannot fail", "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m"},
		{"won't be failed", "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m"},
		{"cannot be failed", "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m"},
		{"should not be failed", "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m"},

		{"not toni", "not \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"Not wenzel", "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"wasn't argus", "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"won't cletus", "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
		{"cannot toni", "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"won't be wenzel", "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"cannot be argus", "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"should not be cletus", "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
	}

	colorProfile = termenv.TrueColor

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	for _, tt := range tests {
		testname := tt.plain
		config := koanf.New(".")
		configRaw := []byte(configData)
		if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
			t.Errorf("Error during config loading: %s", err)
		}
		if err := initWords(config, lemmatizer); err != nil {
			t.Errorf("InitWords() failed with this error: %s", err)
		}
		t.Run(testname, func(t *testing.T) {
			colored := Words.highlight(tt.plain)
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}
