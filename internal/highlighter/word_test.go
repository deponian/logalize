package highlighter

import (
	"fmt"
	"testing"

	"github.com/deponian/logalize/internal/config"
	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

func compareWordGroups(wg1, wg2 wordGroups) error {
	if !cmp.Equal(wg1.Good, wg2.Good) {
		return fmt.Errorf("good groups %v and %v are different", wg1.Good, wg2.Good)
	}
	if !cmp.Equal(wg1.Bad, wg2.Bad) {
		return fmt.Errorf("bad groups %v and %v are different", wg1.Bad, wg2.Bad)
	}
	if !cmp.Equal(wg1.Other, wg2.Other) {
		return fmt.Errorf("other groups %v and %v are different", wg1.Other, wg2.Other)
	}

	return nil
}

func TestWordsNewGood(t *testing.T) {
	correctWords := wordGroups{
		Good: wordGroup{"good", []string{"true"}, "#52fa8a", "", "bold"},
		Bad:  wordGroup{"bad", []string{"fail", "fatal"}, "", "#f06c62", ""},
		Other: []wordGroup{
			{"foes", []string{"argus", "cletus"}, "#120fbb", "", "underline"},
			{"friends", []string{"toni", "wenzel"}, "#f834b2", "", "underline"},
		},
		Lemmatizer: nil,
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/words/newWords/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestWordsNewGood", func(t *testing.T) {
		words, err := newWords(cfg, "test")
		if err != nil {
			t.Errorf("newWords() failed with this error: %s", err)
		}
		err = compareWordGroups(words, correctWords)
		if err != nil {
			t.Errorf("wordGroups are different: %s", err)
		}
	})
}

func TestWordsNewBadYAML(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/words/newWords/02_bad_yaml.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestWordsNewBadYAML", func(t *testing.T) {
		_, err := newWords(cfg, "test")
		if err == nil {
			t.Errorf("newWords() should have failed")
		}
	})
}

func TestWordsNewBadStyle(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/words/newWords/03_bad_style.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestWordsNewBadStyle", func(t *testing.T) {
		_, err := newWords(cfg, "test")
		if err == nil {
			t.Errorf("newWords() should have failed")
		}
	})
}

func TestWordsCheck(t *testing.T) {
	tests := []struct {
		err string
		wg  wordGroup
	}{
		{
			"%!s(<nil>)",
			wordGroup{"testNoErr", []string{"test"}, "#ff0000", "#00ff00", "bold"},
		},
		{
			fmt.Sprintf(`[word group: testForegroundErr] foreground color #ff00xd doesn't match %s pattern`, colorRegExp),
			wordGroup{"testForegroundErr", []string{"test"}, "#ff00xd", "", ""},
		},
		{
			fmt.Sprintf(`[word group: testBackgroundErr] background color hello doesn't match %s pattern`, colorRegExp),
			wordGroup{"testBackgroundErr", []string{"test"}, "", "hello", ""},
		},
		{
			fmt.Sprintf(`[word group: testStyleErr1] style words doesn't match %s pattern`, nonRecursiveStyleRegExp),
			wordGroup{"testStyleErr1", []string{"test"}, "", "", "words"},
		},
		{
			fmt.Sprintf(`[word group: testStyleErr2] style patterns doesn't match %s pattern`, nonRecursiveStyleRegExp),
			wordGroup{"testStyleErr2", []string{"test"}, "", "", "patterns"},
		},
		{
			fmt.Sprintf(`[word group: testStyleErr3] style patterns-and-words doesn't match %s pattern`, nonRecursiveStyleRegExp),
			wordGroup{"testStyleErr3", []string{"test"}, "", "", "patterns-and-words"},
		},
	}

	for _, tt := range tests {
		t.Run("TestWordsCheck"+tt.wg.Name, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.wg.check()); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}

func TestWordsHighlightWord(t *testing.T) {
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

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/words/highlightWord/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings := config.Settings{Config: cfg, ColorProfile: termenv.TrueColor}

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	words, err := newWords(settings.Config, "test")
	if err != nil {
		t.Errorf("newWords() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestWordsHighlightWord"+tt.plain, func(t *testing.T) {
			colored := words.highlightWord(tt.plain, hl)
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}

func TestWordsHighlightNegatedWord(t *testing.T) {
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

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/words/highlightNegatedWord/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings := config.Settings{Config: cfg, ColorProfile: termenv.TrueColor}

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	words, err := newWords(settings.Config, "test")
	if err != nil {
		t.Errorf("newWords() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestWordsHighlightNegatedWord"+tt.plain, func(t *testing.T) {
			m := negatedWordRegExp.FindStringSubmatchIndex(tt.plain)
			colored := words.highlightNegatedWord(tt.plain[m[0]:m[1]], tt.plain[m[2]:m[3]], tt.plain[m[4]:m[5]], hl)
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}

func TestWordsHighlight(t *testing.T) {
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

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/words/highlight/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings := config.Settings{Config: cfg, ColorProfile: termenv.TrueColor}

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	words, err := newWords(settings.Config, "test")
	if err != nil {
		t.Errorf("newWords() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestWordsHighlight"+tt.plain, func(t *testing.T) {
			colored := words.highlight(tt.plain, hl)
			if colored != tt.colored {
				t.Errorf("got %s, want %s", colored, tt.colored)
			}
		})
	}
}
