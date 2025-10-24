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

func compareFormats(format1, format2 format) error {
	if format1.Name != format2.Name {
		return fmt.Errorf("[format1: %s, format2: %s] names aren't equal", format1.Name, format2.Name)
	}

	if err := compareCapGroupLists(*format1.CapGroups, *format2.CapGroups); err != nil {
		return fmt.Errorf("[format1: %s, format2: %s] %s", format1.Name, format2.Name, err)
	}

	return nil
}

func compareFormatLists(fl1, fl2 formatList) error {
	if len(fl1) != len(fl2) {
		return fmt.Errorf("format lists have differenet length")
	}

	for i := range fl1 {
		if err := compareFormats(fl1[i], fl2[i]); err != nil {
			return err
		}
	}

	return nil
}

func TestFormatsNewGood(t *testing.T) {
	correctFormat := format{
		"test", &capGroupList{
			[]capGroup{
				{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "#f5ce42", "", "", "", nil, nil},
				{"two", `([^ ]+ )`, "", "#764a9e", "", "", nil, nil},
				{"three", `(\[.+\] )`, "", "", "bold", "", nil, nil},
				{"four", `("[^"]+")`, "#9daf99", "#76fb99", "underline", "", nil, nil},
				{
					"five",
					`(\d\d\d)`, "", "", "", "",
					[]capGroup{
						{"1", `(1\d\d)`, "#505050", "", "", "", nil, regexp.MustCompile(`(1\d\d)`)},
						{"2", `(2\d\d)`, "#00ff00", "", "overline", "", nil, regexp.MustCompile(`(2\d\d)`)},
						{"3", `(3\d\d)`, "#00ffff", "", "crossout", "", nil, regexp.MustCompile(`(3\d\d)`)},
						{"4", `(4\d\d)`, "#ff0000", "", "reverse", "", nil, regexp.MustCompile(`(4\d\d)`)},
						{"5", `(5\d\d)`, "#ff00ff", "", "", "", nil, regexp.MustCompile(`(5\d\d)`)},
					},
					nil,
				},
			},
			regexp.MustCompile(`^(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3} ))(?P<capGroup1>(?:[^ ]+ ))(?P<capGroup2>(?:\[.+\] ))(?P<capGroup3>(?:"[^"]+"))(?P<capGroup4>(?:\d\d\d))$`),
			map[string]int{"one": 0, "two": 1, "three": 2, "four": 3, "five": 4},
		},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/formats/newFormats/01_good.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestFormatsNewGood", func(t *testing.T) {
		formats, err := newFormats(cfg, "test")
		if err != nil {
			t.Errorf("newFormats() failed with this error: %s", err)
		}

		if err := compareFormats(formats[0], correctFormat); err != nil {
			t.Errorf("%s", err)
		}
	})
}

func TestFormatsNewBadYAML(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/formats/newFormats/02_bad_yaml.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestFormatsNewBadYAML", func(t *testing.T) {
		if _, err := newFormats(cfg, "test"); err == nil {
			t.Errorf("newFormats() should have failed")
		}
	})
}

func TestFormatsNewBadRegExp(t *testing.T) {
	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/formats/newFormats/03_bad_regexp.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	t.Run("TestFormatsNewBadRegExp", func(t *testing.T) {
		if _, err := newFormats(cfg, "test"); err == nil {
			t.Errorf("newFormats() should have failed")
		}
	})
}

func TestFormatsHighlight(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m"},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/formats/highlight/01_main.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings := config.Settings{Config: cfg, ColorProfile: termenv.TrueColor}

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Errorf("NewHighlighter() failed with this error: %s", err)
	}

	formats, err := newFormats(settings.Config, "test")
	if err != nil {
		t.Errorf("newWords() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestPatternsHighlight"+tt.plain, func(t *testing.T) {
			colored := formats[0].highlight(tt.plain, hl)
			if colored != tt.colored {
				t.Errorf("got %v, want %v", colored, tt.colored)
			}
		})
	}
}

// Below are the tests for all built-in formats
func TestFormatsBuiltins(t *testing.T) {
	tests := []struct {
		plain   string
		colored string
	}{
		// nginx-combined
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

		// nginx-ingress-controller
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 100 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;0;0;255;1m100 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 200 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;0;255;0;1m200 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 302 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;0;255;255;1m302 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 404 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m404 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 503 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;255;1m503 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 100 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;0;0;255;1m100 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 200 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;0;255;0;1m200 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 302 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;0;255;255;1m302 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 404 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;0;1m404 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 503 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;238;204;159m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;238;204;159m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;100;198;213m0.003 \x1b[0m\x1b[38;2;255;0;255;1m503 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},

		// klog
		{
			`I0410 23:18:43.650599       1 controller.go:175] "starting healthz server" logger="cert-manager.controller" address="[::]:9403"`,
			"\x1b[38;2;130;170;255;1mI0410 \x1b[0m\x1b[38;2;252;167;234m23:18:43.650599\x1b[0m\x1b[38;2;99;109;166m       1 \x1b[0m\x1b[38;2;137;221;255mcontroller.go\x1b[0m\x1b[38;2;99;109;166m:175\x1b[0m\x1b[38;2;255;150;108m] \x1b[0m\"starting healthz server\" logger=\"cert-manager.controller\" address=\"[::]:9403\"",
		},
		{
			`W0704 20:01:06.932182       1 warnings.go:70] annotation "kubernetes.io/ingress.class" is deprecated, please use 'spec.ingressClassName' instead`,
			"\x1b[38;2;255;199;119;1mW0704 \x1b[0m\x1b[38;2;252;167;234m20:01:06.932182\x1b[0m\x1b[38;2;99;109;166m       1 \x1b[0m\x1b[38;2;137;221;255mwarnings.go\x1b[0m\x1b[38;2;99;109;166m:70\x1b[0m\x1b[38;2;255;150;108m] \x1b[0mannotation \"kubernetes.io/ingress.class\" is deprecated, please use 'spec.ingressClassName' instead",
		},
		{
			`E0714 16:12:36.594249       1 controller.go:104] "Unhandled Error" err="ingress 'menetekel/main' in work queue no longer exists" logger="UnhandledError"`,
			"\x1b[38;2;255;117;127;1mE0714 \x1b[0m\x1b[38;2;252;167;234m16:12:36.594249\x1b[0m\x1b[38;2;99;109;166m       1 \x1b[0m\x1b[38;2;137;221;255mcontroller.go\x1b[0m\x1b[38;2;99;109;166m:104\x1b[0m\x1b[38;2;255;150;108m] \x1b[0m\"Unhandled Error\" err=\"ingress 'menetekel/main' in work queue no longer exists\" logger=\"UnhandledError\"",
		},
		{
			`F0123 00:12:34.567890       1 controller.go:4] "Fatal Error" err="fatal error"`,
			"\x1b[38;2;197;59;83;1mF0123 \x1b[0m\x1b[38;2;252;167;234m00:12:34.567890\x1b[0m\x1b[38;2;99;109;166m       1 \x1b[0m\x1b[38;2;137;221;255mcontroller.go\x1b[0m\x1b[38;2;99;109;166m:4\x1b[0m\x1b[38;2;255;150;108m] \x1b[0m\"Fatal Error\" err=\"fatal error\"",
		},

		// redis
		{
			`1:M 01 Feb 2024 19:41:07.226 # monotonic clock: POSIX clock_gettime`,
			"\x1b[38;2;154;173;236m1\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;255;117;127;1mM \x1b[0m\x1b[38;2;192;153;255m01 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m19:41:07.226 \x1b[0m\x1b[38;2;255;199;119;1m# \x1b[0mmonotonic clock: POSIX clock_gettime",
		},
		{
			`22:S 17 Feb 2024 00:39:12.500 * Starting automatic rewriting of AOF on 3886% growth`,
			"\x1b[38;2;154;173;236m22\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;130;170;255;1mS \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.500 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mStarting automatic rewriting of AOF on 3886% growth",
		},
		{
			`375:X 20 Jun 2025 13:27:11.773 - Sentinel ID is 2814dfe0610f4b8a99b4c6076693ed87d032af23`,
			"\x1b[38;2;154;173;236m375\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;255;199;119;1mX \x1b[0m\x1b[38;2;192;153;255m20 Jun 2025 \x1b[0m\x1b[38;2;252;167;234m13:27:11.773 \x1b[0m\x1b[38;2;130;170;255;1m- \x1b[0mSentinel ID is 2814dfe0610f4b8a99b4c6076693ed87d032af23",
		},
		{
			`8792:C 01 Feb 2024 19:41:07.224 . oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo`,
			"\x1b[38;2;154;173;236m8792\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m01 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m19:41:07.224 \x1b[0m\x1b[38;2;184;219;135;1m. \x1b[0moO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo",
		},

		// syslog-rfc3164
		{
			`Jul  3 08:27:19 menetekel systemd[1]: Condition check resulted in MD array scrubbing - continuation being skipped.`,
			"\x1b[38;2;65;166;181m\x1b[0m\x1b[38;2;192;153;255mJul  3 \x1b[0m\x1b[38;2;252;167;234m08:27:19 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255msystemd\x1b[0m\x1b[38;2;238;204;159m[1]\x1b[0m\x1b[38;2;99;109;166m: \x1b[0mCondition check resulted in MD array scrubbing - continuation being skipped.",
		},
		{
			`Jul  3 09:17:01 menetekel CRON[1185749]: (root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
			"\x1b[38;2;65;166;181m\x1b[0m\x1b[38;2;192;153;255mJul  3 \x1b[0m\x1b[38;2;252;167;234m09:17:01 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255mCRON\x1b[0m\x1b[38;2;238;204;159m[1185749]\x1b[0m\x1b[38;2;99;109;166m: \x1b[0m(root) CMD (   cd / && run-parts --report /etc/cron.hourly)",
		},
		{
			`Jul 13 10:17:02 menetekel CRON[1190762]: (root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
			"\x1b[38;2;65;166;181m\x1b[0m\x1b[38;2;192;153;255mJul 13 \x1b[0m\x1b[38;2;252;167;234m10:17:02 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255mCRON\x1b[0m\x1b[38;2;238;204;159m[1190762]\x1b[0m\x1b[38;2;99;109;166m: \x1b[0m(root) CMD (   cd / && run-parts --report /etc/cron.hourly)",
		},
		{
			`Jul 13 10:20:04 menetekel systemd[1]: Starting Certbot...`,
			"\x1b[38;2;65;166;181m\x1b[0m\x1b[38;2;192;153;255mJul 13 \x1b[0m\x1b[38;2;252;167;234m10:20:04 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255msystemd\x1b[0m\x1b[38;2;238;204;159m[1]\x1b[0m\x1b[38;2;99;109;166m: \x1b[0mStarting Certbot...",
		},
		{
			`Jul 13 10:17:02 menetekel CRON: (root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
			"\x1b[38;2;65;166;181m\x1b[0m\x1b[38;2;192;153;255mJul 13 \x1b[0m\x1b[38;2;252;167;234m10:17:02 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255mCRON\x1b[0m\x1b[38;2;238;204;159m\x1b[0m\x1b[38;2;99;109;166m: \x1b[0m(root) CMD (   cd / && run-parts --report /etc/cron.hourly)",
		},
		{
			`Jul 13 10:20:04 menetekel systemd: Starting Certbot...`,
			"\x1b[38;2;65;166;181m\x1b[0m\x1b[38;2;192;153;255mJul 13 \x1b[0m\x1b[38;2;252;167;234m10:20:04 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255msystemd\x1b[0m\x1b[38;2;238;204;159m\x1b[0m\x1b[38;2;99;109;166m: \x1b[0mStarting Certbot...",
		},
		{
			`<25>Jul 13 10:20:04 menetekel systemd[1]: certbot.service: Deactivated successfully.`,
			"\x1b[38;2;65;166;181m<25>\x1b[0m\x1b[38;2;192;153;255mJul 13 \x1b[0m\x1b[38;2;252;167;234m10:20:04 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255msystemd\x1b[0m\x1b[38;2;238;204;159m[1]\x1b[0m\x1b[38;2;99;109;166m: \x1b[0mcertbot.service: Deactivated successfully.",
		},
		{
			`<123>Jul 13 10:20:04 menetekel systemd[1]: Finished Certbot.`,
			"\x1b[38;2;65;166;181m<123>\x1b[0m\x1b[38;2;192;153;255mJul 13 \x1b[0m\x1b[38;2;252;167;234m10:20:04 \x1b[0m\x1b[38;2;137;221;255mmenetekel \x1b[0m\x1b[38;2;130;170;255msystemd\x1b[0m\x1b[38;2;238;204;159m[1]\x1b[0m\x1b[38;2;99;109;166m: \x1b[0mFinished Certbot.",
		},
	}

	cfg := koanf.New(".")
	err := cfg.Load(file.Provider("./testdata/formats/builtins/theme.yaml"), yaml.Parser())
	if err != nil {
		t.Fatalf("cfg.Load(...) failed with this error: %s", err)
	}

	settings, err := config.NewSettings(builtins, cfg, nil)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Fatalf("NewHighlighter() failed with this error: %s", err)
	}

	formats, err := newFormats(settings.Config, "test")
	if err != nil {
		t.Fatalf("newWords() failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestFormatsHighlight"+tt.plain, func(t *testing.T) {
			for _, lf := range formats {
				if lf.match(tt.plain) {
					colored := lf.highlight(tt.plain, hl)
					if colored != tt.colored {
						t.Errorf("got %v, want %v", colored, tt.colored)
					}
				}
			}
		})
	}
}
