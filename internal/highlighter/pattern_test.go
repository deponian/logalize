package highlighter

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/deponian/logalize/internal/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

func comparePatterns(pattern1, pattern2 pattern) error {
	if pattern1.Name != pattern2.Name || pattern1.Priority != pattern2.Priority {
		return fmt.Errorf("[pattern1: %s, pattern2: %s] names or priorities aren't equal", pattern1.Name, pattern2.Name)
	}

	if err := compareCapGroupLists(*pattern1.CapGroups, *pattern2.CapGroups); err != nil {
		return fmt.Errorf("[pattern1: %s, pattern2: %s] %s", pattern1.Name, pattern2.Name, err)
	}

	return nil
}

func comparePatternLists(pl1, pl2 patternList) error {
	if len(pl1) != len(pl2) {
		return fmt.Errorf("pattern lists have differenet length")
	}

	for i := range pl1 {
		if err := comparePatterns(pl1[i], pl2[i]); err != nil {
			return err
		}
	}

	return nil
}

func TestPatternsNewGood(t *testing.T) {
	correctPatterns := []pattern{
		{"string", 500, &capGroupList{
			[]capGroup{
				{
					"string", `("[^"]+"|'[^']+')`, "#00ff00", "", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:"[^"]+"|'[^']+'))`),
			map[string]int{"string": 0},
		}},
		{"ipv4-address", 0, &capGroupList{
			[]capGroup{
				{
					"one", `(\d\d\d(\.\d\d\d){3})`, "#ffc777", "", "", "", nil, nil,
				},
				{
					"two", `((:\d{1,5})?)`, "#ff966c", "", "", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d\d\d(\.\d\d\d){3}))(?P<capGroup1>(?:(:\d{1,5})?))`),
			map[string]int{"one": 0, "two": 1},
		}},
		{"number", 0, &capGroupList{
			[]capGroup{
				{
					"number", `(\d+)`, "", "#00ffff", "bold", "", nil, nil,
				},
			},
			regexp.MustCompile(`(?P<capGroup0>(?:\d+))`),
			map[string]int{"number": 0},
		}},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/patterns/newPatterns/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestPatternsNewGood", func(t *testing.T) {
		patterns, err := newPatterns(cfg, "test")
		if err != nil {
			t.Errorf("newPatterns() failed with this error: %s", err)
		}
		for i, pattern := range patterns {
			if err := comparePatterns(pattern, correctPatterns[i]); err != nil {
				t.Errorf("%s", err)
			}
		}
	})
}

func TestPatternsNewBadYAML1(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/patterns/newPatterns/02_bad_yaml_1.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestPatternsNewBadYAML1", func(t *testing.T) {
		if _, err := newPatterns(cfg, "test"); err == nil {
			t.Errorf("newPatterns() should have failed")
		}
	})
}

func TestPatternsNewBadYAML2(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/patterns/newPatterns/03_bad_yaml_2.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestPatternsNewBadYAML2", func(t *testing.T) {
		if _, err := newPatterns(cfg, "test"); err == nil {
			t.Errorf("newPatterns() should have failed")
		}
	})
}

func TestPatternsNewBadRegExp(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/patterns/newPatterns/04_bad_regexp.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestPatternsNewBadRegExp", func(t *testing.T) {
		if _, err := newPatterns(cfg, "test"); err == nil {
			t.Errorf("newPatterns() should have failed")
		}
	})
}

func TestPatternsNewBadStyle(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/patterns/newPatterns/05_bad_style.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestPatternsNewBadStyle", func(t *testing.T) {
		if _, err := newPatterns(cfg, "test"); err == nil {
			t.Errorf("newPatterns() should have failed")
		}
	})
}

func TestPatternsHighlight(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		{"hello", "hello"},
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{"42", "\x1b[48;2;0;80;80m42\x1b[0m"},
		{"127.0.0.1", "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[48;2;0;80;80m101\x1b[0m"},
		{`"true"`, "\x1b[38;2;0;255;0m\"true\"\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status:103`, "status:\x1b[38;2;80;80;80m103\x1b[0m"},
		{`status:200`, "status:\x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status:302`, "status:\x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status:404`, "status:\x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status:503`, "status:\x1b[38;2;255;0;255m503\x1b[0m"},
		{`status:700`, "status:\x1b[38;2;255;255;255m700\x1b[0m"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/patterns/highlight/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings := config.Settings{Config: cfg, ColorProfile: termenv.TrueColor}

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	patterns, err := newPatterns(settings.Config, "test")
	if err != nil {
		t.Errorf("newWords() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestPatternsHighlight"+tt.plain, func(t *testing.T) {
			colored := patterns.highlight(tt.plain, hl)
			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}

// Below are the tests for all built-in patterns
func TestPatternsBuiltins(t *testing.T) {
	tests := []struct {
		plain   string
		labeled string
	}{
		// rfc3339
		{
			`2024-02-17T06:56:10Z`,
			`<p:rfc3339:date>2024-02-17</p:rfc3339:date><p:rfc3339:t-delimiter>T</p:rfc3339:t-delimiter><p:rfc3339:time>06:56:10</p:rfc3339:time><p:rfc3339:offset>Z</p:rfc3339:offset>`,
		},
		{
			`2024-02-17T06:56:10+05:00`,
			`<p:rfc3339:date>2024-02-17</p:rfc3339:date><p:rfc3339:t-delimiter>T</p:rfc3339:t-delimiter><p:rfc3339:time>06:56:10</p:rfc3339:time><p:rfc3339:offset>+05:00</p:rfc3339:offset>`,
		},
		{
			`2024-02-17T06:56:10.636960544-01:00`,
			`<p:rfc3339:date>2024-02-17</p:rfc3339:date><p:rfc3339:t-delimiter>T</p:rfc3339:t-delimiter><p:rfc3339:time>06:56:10.636960544</p:rfc3339:time><p:rfc3339:offset>-01:00</p:rfc3339:offset>`,
		},
		{
			`2024-02-17t06:56:10z`,
			`<p:rfc3339:date>2024-02-17</p:rfc3339:date><p:rfc3339:t-delimiter>t</p:rfc3339:t-delimiter><p:rfc3339:time>06:56:10</p:rfc3339:time><p:rfc3339:offset>z</p:rfc3339:offset>`,
		},
		{
			`2024-02-17t06:56:10+05:00`,
			`<p:rfc3339:date>2024-02-17</p:rfc3339:date><p:rfc3339:t-delimiter>t</p:rfc3339:t-delimiter><p:rfc3339:time>06:56:10</p:rfc3339:time><p:rfc3339:offset>+05:00</p:rfc3339:offset>`,
		},
		{
			`2024-02-17t06:56:10.636960544-01:00`,
			`<p:rfc3339:date>2024-02-17</p:rfc3339:date><p:rfc3339:t-delimiter>t</p:rfc3339:t-delimiter><p:rfc3339:time>06:56:10.636960544</p:rfc3339:time><p:rfc3339:offset>-01:00</p:rfc3339:offset>`,
		},

		// time
		{
			`23:42:12`,
			`<p:time>23:42:12</p:time>`,
		},
		{
			`01:37:59.743`,
			`<p:time>01:37:59.743</p:time>`,
		},
		{
			`17:49:37.034123`,
			`<p:time>17:49:37.034123</p:time>`,
		},

		// date-1
		{
			`1999-07-10`,
			`<p:date-1>1999-07-10</p:date-1>`,
		},
		{
			`1999/07/10`,
			`<p:date-1>1999/07/10</p:date-1>`,
		},
		{
			`07-10-1999`,
			`<p:date-1>07-10-1999</p:date-1>`,
		},
		{
			`07/10/1999`,
			`<p:date-1>07/10/1999</p:date-1>`,
		},

		// date-2
		{
			`27 Jan`,
			`<p:date-2>27 Jan</p:date-2>`,
		},
		{
			`27 January`,
			`<p:date-2>27 January</p:date-2>`,
		},
		{
			`27 Jan 2023`,
			`<p:date-2>27 Jan 2023</p:date-2>`,
		},
		{
			`27 August 2023`,
			`<p:date-2>27 August 2023</p:date-2>`,
		},
		{
			`27-Jan-2023`,
			`<p:date-2>27-Jan-2023</p:date-2>`,
		},
		{
			`27-August-2023`,
			`<p:date-2>27-August-2023</p:date-2>`,
		},
		{
			`27/Jan/2023`,
			`<p:date-2>27/Jan/2023</p:date-2>`,
		},
		{
			`27/August/2023`,
			`<p:date-2>27/August/2023</p:date-2>`,
		},

		// date-3
		{
			`Jan 27`,
			`<p:date-3>Jan 27</p:date-3>`,
		},
		{
			`January 27`,
			`<p:date-3>January 27</p:date-3>`,
		},
		{
			`Jan 27 2023`,
			`<p:date-3>Jan 27 2023</p:date-3>`,
		},
		{
			`August 27 2023`,
			`<p:date-3>August 27 2023</p:date-3>`,
		},
		{
			`Jan-27-2023`,
			`<p:date-3>Jan-27-2023</p:date-3>`,
		},
		{
			`August-27-2023`,
			`<p:date-3>August-27-2023</p:date-3>`,
		},
		{
			`Jan/27/2023`,
			`<p:date-3>Jan/27/2023</p:date-3>`,
		},
		{
			`August/27/2023`,
			`<p:date-3>August/27/2023</p:date-3>`,
		},

		// date-4
		{
			`Mon 17`,
			`<p:date-4>Mon 17</p:date-4>`,
		},
		{
			`Sunday 3`,
			`<p:date-4>Sunday 3</p:date-4>`,
		},

		// duration
		{
			`75.984854ms`,
			`<p:duration:start></p:duration:start><p:duration:number>75.984854</p:duration:number><p:duration:unit>ms</p:duration:unit>`,
		},
		{
			`5s`,
			`<p:duration:start></p:duration:start><p:duration:number>5</p:duration:number><p:duration:unit>s</p:duration:unit>`,
		},
		{
			`784m`,
			`<p:duration:start></p:duration:start><p:duration:number>784</p:duration:number><p:duration:unit>m</p:duration:unit>`,
		},
		{
			`7.5h`,
			`<p:duration:start></p:duration:start><p:duration:number>7.5</p:duration:number><p:duration:unit>h</p:duration:unit>`,
		},
		{
			`25d`,
			`<p:duration:start></p:duration:start><p:duration:number>25</p:duration:number><p:duration:unit>d</p:duration:unit>`,
		},

		// logfmt-general
		{
			`key=value`,
			`<p:logfmt-general:key>key</p:logfmt-general:key><p:logfmt-general:equal-sign>=</p:logfmt-general:equal-sign>value`,
		},
		{
			`key=5s`,
			`<p:logfmt-general:key>key</p:logfmt-general:key><p:logfmt-general:equal-sign>=</p:logfmt-general:equal-sign><p:duration:start></p:duration:start><p:duration:number>5</p:duration:number><p:duration:unit>s</p:duration:unit>`,
		},

		// logfmt-string
		{
			`key="value"`,
			`<p:logfmt-string:key>key</p:logfmt-string:key><p:logfmt-string:equal-sign>=</p:logfmt-string:equal-sign><p:logfmt-string:opening-quotation-mark>"</p:logfmt-string:opening-quotation-mark>value<p:logfmt-string:closing-quotation-mark>"</p:logfmt-string:closing-quotation-mark>`,
		},
		{
			`key="5s"`,
			`<p:logfmt-string:key>key</p:logfmt-string:key><p:logfmt-string:equal-sign>=</p:logfmt-string:equal-sign><p:logfmt-string:opening-quotation-mark>"</p:logfmt-string:opening-quotation-mark><p:duration:start></p:duration:start><p:duration:number>5</p:duration:number><p:duration:unit>s</p:duration:unit><p:logfmt-string:closing-quotation-mark>"</p:logfmt-string:closing-quotation-mark>`,
		},

		// ipv4-address
		{
			`127.0.0.1`,
			`<p:ipv4-address:address>127.0.0.1</p:ipv4-address:address><p:ipv4-address:mask-or-port></p:ipv4-address:mask-or-port>`,
		},
		{
			`12.34.56.78`,
			`<p:ipv4-address:address>12.34.56.78</p:ipv4-address:address><p:ipv4-address:mask-or-port></p:ipv4-address:mask-or-port>`,
		},
		{
			`255.255.255.255`,
			`<p:ipv4-address:address>255.255.255.255</p:ipv4-address:address><p:ipv4-address:mask-or-port></p:ipv4-address:mask-or-port>`,
		},
		{
			`0.0.0.0`,
			`<p:ipv4-address:address>0.0.0.0</p:ipv4-address:address><p:ipv4-address:mask-or-port></p:ipv4-address:mask-or-port>`,
		},
		{
			`10.0.0.200/16`,
			`<p:ipv4-address:address>10.0.0.200</p:ipv4-address:address><p:ipv4-address:mask-or-port>/16</p:ipv4-address:mask-or-port>`,
		},
		{
			`10.0.0.0/8`,
			`<p:ipv4-address:address>10.0.0.0</p:ipv4-address:address><p:ipv4-address:mask-or-port>/8</p:ipv4-address:mask-or-port>`,
		},
		{
			`10.0.7.107:80`,
			`<p:ipv4-address:address>10.0.7.107</p:ipv4-address:address><p:ipv4-address:mask-or-port>:80</p:ipv4-address:mask-or-port>`,
		},
		{
			`8.9.10.237:8080`,
			`<p:ipv4-address:address>8.9.10.237</p:ipv4-address:address><p:ipv4-address:mask-or-port>:8080</p:ipv4-address:mask-or-port>`,
		},
		{
			`1.2.3.4:17846`,
			`<p:ipv4-address:address>1.2.3.4</p:ipv4-address:address><p:ipv4-address:mask-or-port>:17846</p:ipv4-address:mask-or-port>`,
		},

		// ipv6-address
		{
			`2001:db8:4006:812::200e`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:db8:4006:812::200e</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:0db8:0000:cd30:0000:0000:0000:0000`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:0db8:0000:cd30:0000:0000:0000:0000</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:0db8::cd30:0:0:0:0`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:0db8::cd30:0:0:0:0</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:0db8:0:cd30::`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:0db8:0:cd30::</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`ff02:0:0:0:0:1:ff00:0000`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>ff02:0:0:0:0:1:ff00:0000</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`ff02:0:0:0:0:1:ffff:ffff`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>ff02:0:0:0:0:1:ffff:ffff</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:db8::1234:5678`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:db8::1234:5678</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`ff02:0:0:0:0:0:0:2`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>ff02:0:0:0:0:0:0:2</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`fdf8:f53b:82e4::53`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>fdf8:f53b:82e4::53</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`fe80::200:5aee:feaa:20a2`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>fe80::200:5aee:feaa:20a2</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:0000:4136:e378:`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:0000:4136:e378:</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`8000:63bf:3fff:fdd2`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>8000:63bf:3fff:fdd2</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:db8::`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:db8::</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::1234:5678`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::1234:5678</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2000::`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2000::</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:db8:a0b:12f0::1`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:db8:a0b:12f0::1</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:4:112:cd:65a:753:0:a1`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:4:112:cd:65a:753:0:a1</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:0002:6c::430`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:0002:6c::430</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:5::`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:5::</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`fe08::7:8`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>fe08::7:8</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2002:cb0a:3cdd:1::1`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2002:cb0a:3cdd:1::1</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:db8:8:4::2`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:db8:8:4::2</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`ff01:0:0:0:0:0:0:2`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>ff01:0:0:0:0:0:0:2</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::ffff:0:0`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::ffff:0:0</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:0000::`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:0000::</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::ffff:192.0.2.47`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::ffff:192.0.2.47</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::ffff:0.0.0.0`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::ffff:0.0.0.0</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::ffff:255.255.255.255`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::ffff:255.255.255.255</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::ffff:10.0.0.3`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::ffff:10.0.0.3</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::192.168.0.1`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::192.168.0.1</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::255.255.255.255`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::255.255.255.255</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`2001:db8:122:344::192.0.2.33`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:db8:122:344::192.0.2.33</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`0:0:0:0:0:0:13.1.68.3`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>0:0:0:0:0:0:13.1.68.3</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`0:0:0:0:0:ffff:129.144.52.3`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>0:0:0:0:0:ffff:129.144.52.3</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::13.1.68.3`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::13.1.68.3</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`::ffff:129.144.52.38`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>::ffff:129.144.52.38</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`59fb:0:0:0:0:1005:cc57:6571`,
			`<p:ipv6-address:opening-bracket></p:ipv6-address:opening-bracket><p:ipv6-address:address>59fb:0:0:0:0:1005:cc57:6571</p:ipv6-address:address><p:ipv6-address:closing-bracket></p:ipv6-address:closing-bracket><p:ipv6-address:port></p:ipv6-address:port>`,
		},
		{
			`[2001:5::]:22`,
			`<p:ipv6-address:opening-bracket>[</p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:5::</p:ipv6-address:address><p:ipv6-address:closing-bracket>]</p:ipv6-address:closing-bracket><p:ipv6-address:port>:22</p:ipv6-address:port>`,
		},
		{
			`[2001:db8:4006:812::200e]:8080`,
			`<p:ipv6-address:opening-bracket>[</p:ipv6-address:opening-bracket><p:ipv6-address:address>2001:db8:4006:812::200e</p:ipv6-address:address><p:ipv6-address:closing-bracket>]</p:ipv6-address:closing-bracket><p:ipv6-address:port>:8080</p:ipv6-address:port>`,
		},
		{
			`[ff02:0:0:0:0:1:ffff:ffff]:23456`,
			`<p:ipv6-address:opening-bracket>[</p:ipv6-address:opening-bracket><p:ipv6-address:address>ff02:0:0:0:0:1:ffff:ffff</p:ipv6-address:address><p:ipv6-address:closing-bracket>]</p:ipv6-address:closing-bracket><p:ipv6-address:port>:23456</p:ipv6-address:port>`,
		},

		// mac-address
		{
			`3D:F2:C9:A6:B3:4F`,
			`<p:mac-address>3D:F2:C9:A6:B3:4F</p:mac-address>`,
		},
		{
			`3D-F2-C9-A6-B3-4F`,
			`<p:mac-address>3D-F2-C9-A6-B3-4F</p:mac-address>`,
		},
		{
			`3d:f2:c9:a6:b3:4f`,
			`<p:mac-address>3d:f2:c9:a6:b3:4f</p:mac-address>`,
		},
		{
			`3d-f2-c9-a6-b3-4f`,
			`<p:mac-address>3d-f2-c9-a6-b3-4f</p:mac-address>`,
		},

		// uuid
		{
			`0a99af43-0ad4-4237-b9cd-064966eb2803`,
			`<p:uuid>0a99af43-0ad4-4237-b9cd-064966eb2803</p:uuid>`,
		},
	}

	hl, labels := newLabeledHighlighter(t)

	for _, tt := range tests {
		t.Run("TestPatternsBuiltins"+tt.plain, func(t *testing.T) {
			if labeled := labels.decode(hl.patterns.highlight(tt.plain, hl)); labeled != tt.labeled {
				t.Errorf("got %v, want %v", labeled, tt.labeled)
			}
		})
	}
}
