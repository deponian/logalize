package highlighter

import (
	"embed"
	"fmt"
	"regexp"
	"testing"

	"github.com/deponian/logalize/internal/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

func compareHighlighters(hl1, hl2 Highlighter) error {
	if err := compareFormatLists(hl1.formats, hl2.formats); err != nil {
		return fmt.Errorf("formats are different: %v", err)
	}

	if err := comparePatternLists(hl1.patterns, hl2.patterns); err != nil {
		return fmt.Errorf("patterns are different: %v", err)
	}

	if err := compareWordGroups(hl1.words, hl2.words); err != nil {
		return fmt.Errorf("words are different: %v", err)
	}

	return nil
}

func TestHighlighterNewGood(t *testing.T) {
	correctFormats := formatList{
		{
			"test", &capGroupList{
				[]capGroup{
					{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "#f5ce42", "", "", nil, nil},
					{"two", `([^ ]+ )`, "", "#764a9e", "", nil, nil},
					{"three", `(\[.+\] )`, "", "", "bold", nil, nil},
					{"four", `("[^"]+")`, "#9daf99", "#76fb99", "underline", nil, nil},
					{
						"five",
						`(\d\d\d)`, "", "", "",
						[]capGroup{
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
			},
		},
	}

	correctPatterns := patternList{
		{"string", 500, &capGroupList{
			[]capGroup{
				{
					"", `("[^"]+"|'[^']+')`, "#00ff00", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:"[^"]+"|'[^']+'))`),
		}},
		{"ipv4-address", 0, &capGroupList{
			[]capGroup{
				{
					"one", `(\d\d\d(\.\d\d\d){3})`, "#ffc777", "", "", nil, nil,
				},
				{
					"two", `((:\d{1,5})?)`, "#ff966c", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d\d\d(\.\d\d\d){3}))(?P<capGroup1>(?:(:\d{1,5})?))`),
		}},
		{"number", 0, &capGroupList{
			[]capGroup{
				{
					"", `(\d+)`, "", "#00ffff", "bold", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d+))`),
		}},
	}

	correctWords := wordGroups{
		Good: wordGroup{"good", []string{"true"}, "#52fa8a", "", "bold"},
		Bad:  wordGroup{"bad", []string{"fail", "fatal"}, "", "#f06c62", ""},
		Other: []wordGroup{
			{"foes", []string{"argus", "cletus"}, "#120fbb", "", "underline"},
			{"friends", []string{"toni", "wenzel"}, "#f834b2", "", "underline"},
		},
		Lemmatizer: nil,
	}

	correctHighlighter := Highlighter{
		formats:  correctFormats,
		patterns: correctPatterns,
		words:    correctWords,
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/NewHighlighter/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	t.Run("TestHighlighterNewGood", func(t *testing.T) {
		if err := compareHighlighters(hl, correctHighlighter); err != nil {
			t.Errorf("highlighters are different: %v", err)
		}
	})
}

func TestHighlighterNewBadFormats(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/NewHighlighter/02_bad_formats.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	t.Run("TestHighlighterNewBadFormats", func(t *testing.T) {
		if _, err := NewHighlighter(settings); err == nil {
			t.Error("NewHighlighter() should have failed")
		}
	})
}

func TestHighlighterNewBadPatterns(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/NewHighlighter/03_bad_patterns.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings := config.Settings{Config: cfg, ColorProfile: termenv.TrueColor}

	t.Run("TestHighlighterNewBadPatterns", func(t *testing.T) {
		if _, err := NewHighlighter(settings); err == nil {
			t.Error("NewHighlighter() should have failed")
		}
	})
}

func TestHighlighterNewBadWords(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/NewHighlighter/04_bad_words.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings := config.Settings{Config: cfg, ColorProfile: termenv.TrueColor}

	t.Run("TestHighlighterNewBadWords", func(t *testing.T) {
		if _, err := NewHighlighter(settings); err == nil {
			t.Error("NewHighlighter() should have failed")
		}
	})
}

func TestHighlighterNewHighlightOnlyFormats(t *testing.T) {
	correctFormats := formatList{
		{
			"test", &capGroupList{
				[]capGroup{
					{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "#f5ce42", "", "", nil, nil},
					{"two", `([^ ]+ )`, "", "#764a9e", "", nil, nil},
					{"three", `(\[.+\] )`, "", "", "bold", nil, nil},
					{"four", `("[^"]+")`, "#9daf99", "#76fb99", "underline", nil, nil},
					{
						"five",
						`(\d\d\d)`, "", "", "",
						[]capGroup{
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
			},
		},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/NewHighlighter/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.only-formats", true)
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	t.Run("TestHighlighterNewHighlightOnlyFormats", func(t *testing.T) {
		if err := compareFormatLists(hl.formats, correctFormats); err != nil {
			t.Errorf("error: %v", err)
		}
		if err := comparePatternLists(hl.patterns, patternList{}); err != nil {
			t.Errorf("patterns have to be empty: %v", err)
		}
		if err := compareWordGroups(hl.words, wordGroups{}); err != nil {
			t.Errorf("words have to be empty: %v", err)
		}
	})
}

func TestHighlighterNewHighlightOnlyPatterns(t *testing.T) {
	correctPatterns := patternList{
		{"string", 500, &capGroupList{
			[]capGroup{
				{
					"", `("[^"]+"|'[^']+')`, "#00ff00", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:"[^"]+"|'[^']+'))`),
		}},
		{"ipv4-address", 0, &capGroupList{
			[]capGroup{
				{
					"one", `(\d\d\d(\.\d\d\d){3})`, "#ffc777", "", "", nil, nil,
				},
				{
					"two", `((:\d{1,5})?)`, "#ff966c", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d\d\d(\.\d\d\d){3}))(?P<capGroup1>(?:(:\d{1,5})?))`),
		}},
		{"number", 0, &capGroupList{
			[]capGroup{
				{
					"", `(\d+)`, "", "#00ffff", "bold", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d+))`),
		}},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/NewHighlighter/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.only-patterns", true)
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	t.Run("TestHighlighterNewHighlightOnlyPatterns", func(t *testing.T) {
		if err := compareFormatLists(hl.formats, formatList{}); err != nil {
			t.Errorf("formats have to be empty: %v", err)
		}
		if err := comparePatternLists(hl.patterns, correctPatterns); err != nil {
			t.Errorf("error: %v", err)
		}
		if err := compareWordGroups(hl.words, wordGroups{}); err != nil {
			t.Errorf("words have to be empty: %v", err)
		}
	})
}

func TestHighlighterNewHighlightOnlyWords(t *testing.T) {
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
	err := cfg.Load(file.Provider("./testdata/highlighter/NewHighlighter/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.only-words", true)
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	t.Run("TestHighlighterNewHighlightOnlyWords", func(t *testing.T) {
		if err := compareFormatLists(hl.formats, formatList{}); err != nil {
			t.Errorf("patterns have to be empty: %v", err)
		}
		if err := comparePatternLists(hl.patterns, patternList{}); err != nil {
			t.Errorf("patterns have to be empty: %v", err)
		}
		if err := compareWordGroups(hl.words, correctWords); err != nil {
			t.Errorf("error: %v", err)
		}
	})
}

// normal tests ("test" theme)
func TestHighlighterColorizeNormal(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		// formats
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m"},
		{`0 - hello bye`, "\x1b[38;2;245;206;65m0 - \x1b[0mhello bye"},
		{`1 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;206;65m1 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m true toni rufus"},
		{`22 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;17;65m22 - \x1b[0m777 hello 1.1.1.1 \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus"},
		{`333 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;17;206;65m333 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus"},
		{`4444 - "777 hello 1.1.1.1 true toni rufus" "777 hello 1.1.1.1 true toni rufus" «777 hello 1.1.1.1 true toni rufus»`, "\x1b[38;2;80;206;255m4444 - \x1b[0m\x1b[38;2;0;255;0m\"777 hello 1.1.1.1 true toni rufus\"\x1b[0m \"777 hello 1.1.1.1 \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus\" «\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus»"},

		// patterns
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{`42`, "\x1b[48;2;0;80;80m42\x1b[0m"},
		{`127.0.0.1`, "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[38;2;80;80;80m101\x1b[0m"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status 103`, "status \x1b[38;2;80;80;80m103\x1b[0m"},
		{`status 200`, "status \x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status 302`, "status \x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status 404`, "status \x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status 503`, "status \x1b[38;2;255;0;255m503\x1b[0m"},
		{`status 700`, "status \x1b[38;2;255;255;255m700\x1b[0m"},

		// words
		{`untrue`, "untrue"},
		{`true`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{`fail`, "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{`failed`, "\x1b[48;2;240;108;97mfailed\x1b[0m"},
		{`wenzel`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`argus`, "\x1b[38;2;18;15;187;4margus\x1b[0m"},

		{`not true`, "\x1b[48;2;240;108;97mnot true\x1b[0m"},
		{`Not true`, "\x1b[48;2;240;108;97mNot true\x1b[0m"},
		{`wasn't true`, "\x1b[48;2;240;108;97mwasn't true\x1b[0m"},
		{`won't true`, "\x1b[48;2;240;108;97mwon't true\x1b[0m"},
		{`cannot complete`, "\x1b[48;2;240;108;97mcannot complete\x1b[0m"},
		{`won't be completed`, "\x1b[48;2;240;108;97mwon't be completed\x1b[0m"},
		{`cannot be completed`, "\x1b[48;2;240;108;97mcannot be completed\x1b[0m"},
		{`should not be completed`, "\x1b[48;2;240;108;97mshould not be completed\x1b[0m"},

		{`not false`, "\x1b[38;2;81;250;138;1mnot false\x1b[0m"},
		{`Not false`, "\x1b[38;2;81;250;138;1mNot false\x1b[0m"},
		{`wasn't false`, "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m"},
		{`won't false`, "\x1b[38;2;81;250;138;1mwon't false\x1b[0m"},
		{`cannot fail`, "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m"},
		{`won't be failed`, "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m"},
		{`cannot be failed`, "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m"},
		{`should not be failed`, "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m"},

		{`not toni`, "not \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{`Not wenzel`, "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`wasn't argus`, "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{`won't cletus`, "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
		{`cannot toni`, "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{`won't be wenzel`, "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`cannot be argus`, "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{`should not be cletus`, "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m"},

		// patterns and words
		{`true bad fail 7.7.7.7`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[48;2;240;108;97mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m"},
		{`"true" and true`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m and \x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{`wenzel failed 127 times`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m \x1b[48;2;240;108;97mfailed\x1b[0m \x1b[38;2;80;80;80m127\x1b[0m times"},

		// colored input (ANSI escape sequences should be successfully stripped)
		{"127.0.0.1 - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \"testing\"", "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{"\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m", "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{"\x1b[1;31mtrue\x1b[0m", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{"\x1b[3;32mfail\x1b[0m", "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{"\x1b[38:2:81:250:138mtrue\x1b[0m", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{"\x1b[38:5:100mfail\x1b[0m", "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{"true \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m fail 7.7.7.7", "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[48;2;240;108;97mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/Colorize/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.theme", "test")
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestHighlighterColorizeNormal"+tt.plain, func(t *testing.T) {
			colored := hl.Colorize(tt.plain)

			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}

// tests with default color (test-with-default-color theme)
func TestHighlighterColorizeWithDefaultColor(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		// formats
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m"},
		{`0 - hello bye`, "\x1b[38;2;245;206;65m0 - \x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0mhello bye\x1b[0m"},
		{`1 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;206;65m1 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m hello \x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m true toni rufus\x1b[0m"},
		{`22 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;17;65m22 - \x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m777 hello 1.1.1.1 \x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m rufus\x1b[0m"},
		{`333 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;17;206;65m333 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m hello \x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m rufus\x1b[0m"},
		{`4444 - "777 hello 1.1.1.1 true toni rufus" "777 hello 1.1.1.1 true toni rufus" «777 hello 1.1.1.1 true toni rufus»`, "\x1b[38;2;80;206;255m4444 - \x1b[0m\x1b[38;2;0;255;0m\"777 hello 1.1.1.1 true toni rufus\"\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m\"777 hello 1.1.1.1 \x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m rufus\" \x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m«\x1b[0m\x1b[38;2;255;255;255m777\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m hello \x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m rufus»\x1b[0m"},

		// patterns
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{`42`, "\x1b[48;2;0;80;80m42\x1b[0m"},
		{`127.0.0.1`, "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m: \x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m hello \x1b[0m\x1b[38;2;80;80;80m101\x1b[0m"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status 103`, "\x1b[38;2;255;0;0;48;2;0;255;0mstatus \x1b[0m\x1b[38;2;80;80;80m103\x1b[0m"},
		{`status 200`, "\x1b[38;2;255;0;0;48;2;0;255;0mstatus \x1b[0m\x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status 302`, "\x1b[38;2;255;0;0;48;2;0;255;0mstatus \x1b[0m\x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status 404`, "\x1b[38;2;255;0;0;48;2;0;255;0mstatus \x1b[0m\x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status 503`, "\x1b[38;2;255;0;0;48;2;0;255;0mstatus \x1b[0m\x1b[38;2;255;0;255m503\x1b[0m"},
		{`status 700`, "\x1b[38;2;255;0;0;48;2;0;255;0mstatus \x1b[0m\x1b[38;2;255;255;255m700\x1b[0m"},

		// words
		{`untrue`, "\x1b[38;2;255;0;0;48;2;0;255;0muntrue\x1b[0m"},
		{`true`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{`fail`, "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{`failed`, "\x1b[48;2;240;108;97mfailed\x1b[0m"},
		{`wenzel`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`argus`, "\x1b[38;2;18;15;187;4margus\x1b[0m"},

		{`not true`, "\x1b[48;2;240;108;97mnot true\x1b[0m"},
		{`Not true`, "\x1b[48;2;240;108;97mNot true\x1b[0m"},
		{`wasn't true`, "\x1b[48;2;240;108;97mwasn't true\x1b[0m"},
		{`won't true`, "\x1b[48;2;240;108;97mwon't true\x1b[0m"},
		{`cannot complete`, "\x1b[48;2;240;108;97mcannot complete\x1b[0m"},
		{`won't be completed`, "\x1b[48;2;240;108;97mwon't be completed\x1b[0m"},
		{`cannot be completed`, "\x1b[48;2;240;108;97mcannot be completed\x1b[0m"},
		{`should not be completed`, "\x1b[48;2;240;108;97mshould not be completed\x1b[0m"},

		{`not false`, "\x1b[38;2;81;250;138;1mnot false\x1b[0m"},
		{`Not false`, "\x1b[38;2;81;250;138;1mNot false\x1b[0m"},
		{`wasn't false`, "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m"},
		{`won't false`, "\x1b[38;2;81;250;138;1mwon't false\x1b[0m"},
		{`cannot fail`, "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m"},
		{`won't be failed`, "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m"},
		{`cannot be failed`, "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m"},
		{`should not be failed`, "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m"},

		{`not toni`, "\x1b[38;2;255;0;0;48;2;0;255;0mnot \x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{`Not wenzel`, "\x1b[38;2;255;0;0;48;2;0;255;0mNot \x1b[0m\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`wasn't argus`, "\x1b[38;2;255;0;0;48;2;0;255;0mwasn't \x1b[0m\x1b[38;2;18;15;187;4margus\x1b[0m"},
		{`won't cletus`, "\x1b[38;2;255;0;0;48;2;0;255;0mwon't \x1b[0m\x1b[38;2;18;15;187;4mcletus\x1b[0m"},
		{`cannot toni`, "\x1b[38;2;255;0;0;48;2;0;255;0mcannot \x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{`won't be wenzel`, "\x1b[38;2;255;0;0;48;2;0;255;0mwon't be \x1b[0m\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`cannot be argus`, "\x1b[38;2;255;0;0;48;2;0;255;0mcannot be \x1b[0m\x1b[38;2;18;15;187;4margus\x1b[0m"},
		{`should not be cletus`, "\x1b[38;2;255;0;0;48;2;0;255;0mshould not be \x1b[0m\x1b[38;2;18;15;187;4mcletus\x1b[0m"},

		// patterns and words
		{`true bad fail 7.7.7.7`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m bad \x1b[0m\x1b[48;2;240;108;97mfail\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m"},
		{`"true" and true`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m and \x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{`wenzel failed 127 times`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[48;2;240;108;97mfailed\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;80;80;80m127\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m times\x1b[0m"},

		// colored input (ANSI escape sequences should be successfully stripped)
		{"127.0.0.1 - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \"testing\"", "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{"\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m", "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{"\x1b[1;31mtrue\x1b[0m", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{"\x1b[3;32mfail\x1b[0m", "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{"\x1b[38:2:81:250:138mtrue\x1b[0m", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{"\x1b[38:5:100mfail\x1b[0m", "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{"true \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m fail 7.7.7.7", "\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m bad \x1b[0m\x1b[48;2;240;108;97mfail\x1b[0m\x1b[38;2;255;0;0;48;2;0;255;0m \x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/Colorize/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.theme", "test-with-default-color")
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestHighlighterColorizeWithDefaultColor"+tt.plain, func(t *testing.T) {
			colored := hl.Colorize(tt.plain)

			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}

// test "dry-run" option
func TestHighlighterColorizeDryRun(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		// formats
		{`127.0.0.1 - [test] "testing"`, `127.0.0.1 - [test] "testing"`},
		{`127.0.0.2 test [test hello] "testing again"`, `127.0.0.2 test [test hello] "testing again"`},
		{`127.0.0.3 ___ [.] "_"`, `127.0.0.3 ___ [.] "_"`},
		{`0 - hello bye`, `0 - hello bye`},
		{`1 - 777 hello 1.1.1.1 true toni rufus`, `1 - 777 hello 1.1.1.1 true toni rufus`},
		{`22 - 777 hello 1.1.1.1 true toni rufus`, `22 - 777 hello 1.1.1.1 true toni rufus`},
		{`333 - 777 hello 1.1.1.1 true toni rufus`, `333 - 777 hello 1.1.1.1 true toni rufus`},
		{`4444 - "777 hello 1.1.1.1 true toni rufus" "777 hello 1.1.1.1 true toni rufus" «777 hello 1.1.1.1 true toni rufus»`, `4444 - "777 hello 1.1.1.1 true toni rufus" "777 hello 1.1.1.1 true toni rufus" «777 hello 1.1.1.1 true toni rufus»`},

		// patterns
		{`"string"`, `"string"`},
		{`42`, `42`},
		{`127.0.0.1`, `127.0.0.1`},
		{`"test": 127.7.7.7 hello 101`, `"test": 127.7.7.7 hello 101`},
		{`"true"`, `"true"`},
		{`"42"`, `"42"`},
		{`"237.7.7.7"`, `"237.7.7.7"`},
		{`status 103`, `status 103`},
		{`status 200`, `status 200`},
		{`status 302`, `status 302`},
		{`status 404`, `status 404`},
		{`status 503`, `status 503`},
		{`status 700`, `status 700`},

		// words
		{`untrue`, `untrue`},
		{`true`, `true`},
		{`fail`, `fail`},
		{`failed`, `failed`},
		{`wenzel`, `wenzel`},
		{`argus`, `argus`},

		{`not true`, `not true`},
		{`Not true`, `Not true`},
		{`wasn't true`, `wasn't true`},
		{`won't true`, `won't true`},
		{`cannot complete`, `cannot complete`},
		{`won't be completed`, `won't be completed`},
		{`cannot be completed`, `cannot be completed`},
		{`should not be completed`, `should not be completed`},

		{`not false`, `not false`},
		{`Not false`, `Not false`},
		{`wasn't false`, `wasn't false`},
		{`won't false`, `won't false`},
		{`cannot fail`, `cannot fail`},
		{`won't be failed`, `won't be failed`},
		{`cannot be failed`, `cannot be failed`},
		{`should not be failed`, `should not be failed`},

		{`not toni`, `not toni`},
		{`Not wenzel`, `Not wenzel`},
		{`wasn't argus`, `wasn't argus`},
		{`won't cletus`, `won't cletus`},
		{`cannot toni`, `cannot toni`},
		{`won't be wenzel`, `won't be wenzel`},
		{`cannot be argus`, `cannot be argus`},
		{`should not be cletus`, `should not be cletus`},

		// patterns and words
		{`true bad fail 7.7.7.7`, `true bad fail 7.7.7.7`},
		{`"true" and true`, `"true" and true`},
		{`wenzel failed 127 times`, `wenzel failed 127 times`},

		// colored input (ANSI escape sequences should be preserved)
		{"127.0.0.1 - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \"testing\"", "127.0.0.1 - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \"testing\""},
		{"\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m", "\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m"},
		{"\x1b[1;31mtrue\x1b[0m", "\x1b[1;31mtrue\x1b[0m"},
		{"\x1b[3;32mfail\x1b[0m", "\x1b[3;32mfail\x1b[0m"},
		{"\x1b[38:2:81:250:138mtrue\x1b[0m", "\x1b[38:2:81:250:138mtrue\x1b[0m"},
		{"\x1b[38:5:100mfail\x1b[0m", "\x1b[38:5:100mfail\x1b[0m"},
		{"true \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m fail 7.7.7.7", "true \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m fail 7.7.7.7"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/Colorize/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.theme", "test")
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.dry-run", true)
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestHighlighterColorizeDryRun"+tt.plain, func(t *testing.T) {
			colored := hl.Colorize(tt.plain)

			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}

// test "no-ansi-escape-sequences-stripping" option
func TestHighlighterColorizeNoANSIEscapeSequencesStripping(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		// formats
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m"},
		{`0 - hello bye`, "\x1b[38;2;245;206;65m0 - \x1b[0mhello bye"},
		{`1 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;206;65m1 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m true toni rufus"},
		{`22 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;245;17;65m22 - \x1b[0m777 hello 1.1.1.1 \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus"},
		{`333 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[38;2;17;206;65m333 - \x1b[0m\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus"},
		{`4444 - "777 hello 1.1.1.1 true toni rufus" "777 hello 1.1.1.1 true toni rufus" «777 hello 1.1.1.1 true toni rufus»`, "\x1b[38;2;80;206;255m4444 - \x1b[0m\x1b[38;2;0;255;0m\"777 hello 1.1.1.1 true toni rufus\"\x1b[0m \"777 hello 1.1.1.1 \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus\" «\x1b[38;2;255;255;255m777\x1b[0m hello \x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m \x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[38;2;248;52;178;4mtoni\x1b[0m rufus»"},

		// patterns
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{`42`, "\x1b[48;2;0;80;80m42\x1b[0m"},
		{`127.0.0.1`, "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[38;2;80;80;80m101\x1b[0m"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status 103`, "status \x1b[38;2;80;80;80m103\x1b[0m"},
		{`status 200`, "status \x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status 302`, "status \x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status 404`, "status \x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status 503`, "status \x1b[38;2;255;0;255m503\x1b[0m"},
		{`status 700`, "status \x1b[38;2;255;255;255m700\x1b[0m"},

		// words
		{`untrue`, "untrue"},
		{`true`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{`fail`, "\x1b[48;2;240;108;97mfail\x1b[0m"},
		{`failed`, "\x1b[48;2;240;108;97mfailed\x1b[0m"},
		{`wenzel`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`argus`, "\x1b[38;2;18;15;187;4margus\x1b[0m"},

		{`not true`, "\x1b[48;2;240;108;97mnot true\x1b[0m"},
		{`Not true`, "\x1b[48;2;240;108;97mNot true\x1b[0m"},
		{`wasn't true`, "\x1b[48;2;240;108;97mwasn't true\x1b[0m"},
		{`won't true`, "\x1b[48;2;240;108;97mwon't true\x1b[0m"},
		{`cannot complete`, "\x1b[48;2;240;108;97mcannot complete\x1b[0m"},
		{`won't be completed`, "\x1b[48;2;240;108;97mwon't be completed\x1b[0m"},
		{`cannot be completed`, "\x1b[48;2;240;108;97mcannot be completed\x1b[0m"},
		{`should not be completed`, "\x1b[48;2;240;108;97mshould not be completed\x1b[0m"},

		{`not false`, "\x1b[38;2;81;250;138;1mnot false\x1b[0m"},
		{`Not false`, "\x1b[38;2;81;250;138;1mNot false\x1b[0m"},
		{`wasn't false`, "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m"},
		{`won't false`, "\x1b[38;2;81;250;138;1mwon't false\x1b[0m"},
		{`cannot fail`, "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m"},
		{`won't be failed`, "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m"},
		{`cannot be failed`, "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m"},
		{`should not be failed`, "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m"},

		{`not toni`, "not \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{`Not wenzel`, "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`wasn't argus`, "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{`won't cletus`, "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
		{`cannot toni`, "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{`won't be wenzel`, "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{`cannot be argus`, "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{`should not be cletus`, "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m"},

		// patterns and words
		{`true bad fail 7.7.7.7`, "\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[48;2;240;108;97mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m"},
		{`"true" and true`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m and \x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{`wenzel failed 127 times`, "\x1b[38;2;248;52;178;4mwenzel\x1b[0m \x1b[48;2;240;108;97mfailed\x1b[0m \x1b[38;2;80;80;80m127\x1b[0m times"},

		// colored input (ANSI escape sequences should be preserved)
		{"127.0.0.1 - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \"testing\"", "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \x1b[38;2;0;255;0m\"testing\"\x1b[0m"},
		{"\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m", "\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m"},
		{"\x1b[1;31mtrue\x1b[0m", "\x1b[1;31mtrue\x1b[0m"},
		{"\x1b[3;32mfail\x1b[0m", "\x1b[3;32mfail\x1b[0m"},
		{"\x1b[38:2:81:250:138mtrue\x1b[0m", "\x1b[38:2:81:250:138mtrue\x1b[0m"},
		{"\x1b[38:5:100mfail\x1b[0m", "\x1b[38:5:100mfail\x1b[0m"},
		{"true \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m fail 7.7.7.7", "\x1b[38;2;81;250;138;1mtrue\x1b[0m \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m \x1b[48;2;240;108;97mfail\x1b[0m \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/Colorize/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.theme", "test")
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.no-ansi-escape-sequences-stripping", true)
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestHighlighterColorizeNoANSIEscapeSequencesStripping"+tt.plain, func(t *testing.T) {
			colored := hl.Colorize(tt.plain)

			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}

// test "debug" option
func TestHighlighterColorizeDebug(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		// formats
		{`127.0.0.1 - [test] "testing"`, "\x1b[7m[f(menetekel)]\x1b[0m\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m\x1b[7m[f(/menetekel)]\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[7m[f(menetekel)]\x1b[0m\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m\x1b[7m[f(/menetekel)]\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[7m[f(menetekel)]\x1b[0m\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m\x1b[7m[f(/menetekel)]\x1b[0m"},
		{`0 - hello bye`, "\x1b[7m[f(portafisco-patterns)]\x1b[0m\x1b[38;2;245;206;65m0 - \x1b[0mhello bye\x1b[7m[f(/portafisco-patterns)]\x1b[0m"},
		{`1 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[7m[f(portafisco-patterns)]\x1b[0m\x1b[38;2;245;206;65m1 - \x1b[0m\x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;255;255;255m777\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m hello \x1b[7m[p(ipv4-address)]\x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m\x1b[7m[p(/ipv4-address)]\x1b[0m true toni rufus\x1b[7m[f(/portafisco-patterns)]\x1b[0m"},
		{`22 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[7m[f(portafisco-words)]\x1b[0m\x1b[38;2;245;17;65m22 - \x1b[0m777 hello 1.1.1.1 \x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[7m[w(/friends)]\x1b[0m rufus\x1b[7m[f(/portafisco-words)]\x1b[0m"},
		{`333 - 777 hello 1.1.1.1 true toni rufus`, "\x1b[7m[f(portafisco-patterns-and-words)]\x1b[0m\x1b[38;2;17;206;65m333 - \x1b[0m\x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;255;255;255m777\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m hello \x1b[7m[p(ipv4-address)]\x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m\x1b[7m[p(/ipv4-address)]\x1b[0m \x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[7m[w(/friends)]\x1b[0m rufus\x1b[7m[f(/portafisco-patterns-and-words)]\x1b[0m"},
		{`4444 - "777 hello 1.1.1.1 true toni rufus" "777 hello 1.1.1.1 true toni rufus" «777 hello 1.1.1.1 true toni rufus»`, "\x1b[7m[f(portafisco-combined)]\x1b[0m\x1b[38;2;80;206;255m4444 - \x1b[0m\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"777 hello 1.1.1.1 true toni rufus\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m \"777 hello 1.1.1.1 \x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[7m[w(/friends)]\x1b[0m rufus\" «\x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;255;255;255m777\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m hello \x1b[7m[p(ipv4-address)]\x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m1.1.1.1\x1b[0m\x1b[7m[p(/ipv4-address)]\x1b[0m \x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[7m[w(/friends)]\x1b[0m rufus»\x1b[7m[f(/portafisco-combined)]\x1b[0m"},

		// patterns
		{`"string"`, "\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"string\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m"},
		{`42`, "\x1b[7m[p(number)]\x1b[0m\x1b[48;2;0;80;80m42\x1b[0m\x1b[7m[p(/number)]\x1b[0m"},
		{`127.0.0.1`, "\x1b[7m[p(ipv4-address)]\x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m\x1b[7m[p(/ipv4-address)]\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"test\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m: \x1b[7m[p(ipv4-address)]\x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m\x1b[7m[p(/ipv4-address)]\x1b[0m hello \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;80;80;80m101\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m"},
		{`"true"`, "\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"true\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m"},
		{`"42"`, "\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"42\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m"},
		{`status 103`, "status \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;80;80;80m103\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m"},
		{`status 200`, "status \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;0;255;0;53m200\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m"},
		{`status 302`, "status \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;0;255;255;9m302\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m"},
		{`status 404`, "status \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;255;0;0;7m404\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m"},
		{`status 503`, "status \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;255;0;255m503\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m"},
		{`status 700`, "status \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;255;255;255m700\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m"},

		// words
		{`untrue`, "untrue"},
		{`true`, "\x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`fail`, "\x1b[7m[w(bad)]\x1b[0m\x1b[48;2;240;108;97mfail\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`failed`, "\x1b[7m[w(bad)]\x1b[0m\x1b[48;2;240;108;97mfailed\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`wenzel`, "\x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mwenzel\x1b[0m\x1b[7m[w(/friends)]\x1b[0m"},
		{`argus`, "\x1b[7m[w(foes)]\x1b[0m\x1b[38;2;18;15;187;4margus\x1b[0m\x1b[7m[w(/foes)]\x1b[0m"},

		{`not true`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mnot true\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`Not true`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mNot true\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`wasn't true`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mwasn't true\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`won't true`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mwon't true\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`cannot complete`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mcannot complete\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`won't be completed`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mwon't be completed\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`cannot be completed`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mcannot be completed\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`should not be completed`, "\x1b[7m[w(good)]\x1b[0m\x1b[48;2;240;108;97mshould not be completed\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},

		{`not false`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mnot false\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`Not false`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mNot false\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`wasn't false`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mwasn't false\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`won't false`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mwon't false\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`cannot fail`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mcannot fail\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`won't be failed`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`cannot be failed`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{`should not be failed`, "\x1b[7m[w(bad)]\x1b[0m\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},

		{`not toni`, "not \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[7m[w(/friends)]\x1b[0m"},
		{`Not wenzel`, "Not \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mwenzel\x1b[0m\x1b[7m[w(/friends)]\x1b[0m"},
		{`wasn't argus`, "wasn't \x1b[7m[w(foes)]\x1b[0m\x1b[38;2;18;15;187;4margus\x1b[0m\x1b[7m[w(/foes)]\x1b[0m"},
		{`won't cletus`, "won't \x1b[7m[w(foes)]\x1b[0m\x1b[38;2;18;15;187;4mcletus\x1b[0m\x1b[7m[w(/foes)]\x1b[0m"},
		{`cannot toni`, "cannot \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mtoni\x1b[0m\x1b[7m[w(/friends)]\x1b[0m"},
		{`won't be wenzel`, "won't be \x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mwenzel\x1b[0m\x1b[7m[w(/friends)]\x1b[0m"},
		{`cannot be argus`, "cannot be \x1b[7m[w(foes)]\x1b[0m\x1b[38;2;18;15;187;4margus\x1b[0m\x1b[7m[w(/foes)]\x1b[0m"},
		{`should not be cletus`, "should not be \x1b[7m[w(foes)]\x1b[0m\x1b[38;2;18;15;187;4mcletus\x1b[0m\x1b[7m[w(/foes)]\x1b[0m"},

		// patterns and words
		{`true bad fail 7.7.7.7`, "\x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m bad \x1b[7m[w(bad)]\x1b[0m\x1b[48;2;240;108;97mfail\x1b[0m\x1b[7m[w(/bad)]\x1b[0m \x1b[7m[p(ipv4-address)]\x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m\x1b[7m[p(/ipv4-address)]\x1b[0m"},
		{`"true" and true`, "\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"true\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m and \x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{`wenzel failed 127 times`, "\x1b[7m[w(friends)]\x1b[0m\x1b[38;2;248;52;178;4mwenzel\x1b[0m\x1b[7m[w(/friends)]\x1b[0m \x1b[7m[w(bad)]\x1b[0m\x1b[48;2;240;108;97mfailed\x1b[0m\x1b[7m[w(/bad)]\x1b[0m \x1b[7m[p(http-status-code)]\x1b[0m\x1b[38;2;80;80;80m127\x1b[0m\x1b[7m[p(/http-status-code)]\x1b[0m times"},

		// colored input (ANSI escape sequences should be successfully stripped)
		{"127.0.0.1 - \x1b[0m\x1b[1m\x1b[31m[test]\x1b[0m \"testing\"", "\x1b[7m[f(menetekel)]\x1b[0m\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m\x1b[7m[f(/menetekel)]\x1b[0m"},
		{"\x1b[0m\x1b[1m\x1b[31m\"string\"\x1b[0m", "\x1b[7m[p(string)]\x1b[0m\x1b[38;2;0;255;0m\"string\"\x1b[0m\x1b[7m[p(/string)]\x1b[0m"},
		{"\x1b[1;31mtrue\x1b[0m", "\x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{"\x1b[3;32mfail\x1b[0m", "\x1b[7m[w(bad)]\x1b[0m\x1b[48;2;240;108;97mfail\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{"\x1b[38:2:81:250:138mtrue\x1b[0m", "\x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m"},
		{"\x1b[38:5:100mfail\x1b[0m", "\x1b[7m[w(bad)]\x1b[0m\x1b[48;2;240;108;97mfail\x1b[0m\x1b[7m[w(/bad)]\x1b[0m"},
		{"true \x1b[0m\x1b[1m\x1b[31mbad\x1b[0m fail 7.7.7.7", "\x1b[7m[w(good)]\x1b[0m\x1b[38;2;81;250;138;1mtrue\x1b[0m\x1b[7m[w(/good)]\x1b[0m bad \x1b[7m[w(bad)]\x1b[0m\x1b[48;2;240;108;97mfail\x1b[0m\x1b[7m[w(/bad)]\x1b[0m \x1b[7m[p(ipv4-address)]\x1b[0m\x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m\x1b[7m[p(/ipv4-address)]\x1b[0m"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/highlighter/Colorize/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.theme", "test")
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}
	err = cfg.Set("settings.debug", true)
	if err != nil {
		t.Fatalf("cfg.Set(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(embed.FS{}, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestHighlighterColorizeDebug"+tt.plain, func(t *testing.T) {
			colored := hl.Colorize(tt.plain)

			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}
