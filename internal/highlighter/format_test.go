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
		format  string
		labeled string
	}{
		// combined-log-format
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 100 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:01:01 +0000] </timestamp><request>"GET / HTTP/1.1" </request><status:1xx>100 </status:1xx><body-bytes-sent>162 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"</http-user-agent>`,
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 200 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:01:01 +0000] </timestamp><request>"GET / HTTP/1.1" </request><status:2xx>200 </status:2xx><body-bytes-sent>162 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"</http-user-agent>`,
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 302 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:01:01 +0000] </timestamp><request>"GET / HTTP/1.1" </request><status:3xx>302 </status:3xx><body-bytes-sent>162 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"</http-user-agent>`,
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 404 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:01:01 +0000] </timestamp><request>"GET / HTTP/1.1" </request><status:4xx>404 </status:4xx><body-bytes-sent>162 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"</http-user-agent>`,
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 503 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:01:01 +0000] </timestamp><request>"GET / HTTP/1.1" </request><status:5xx>503 </status:5xx><body-bytes-sent>162 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"</http-user-agent>`,
		},
		{
			`127.0.0.1 - frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08 [en] (Win98; I ;Nav)"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>frank </remote-user><timestamp>[10/Oct/2000:13:55:36 -0700] </timestamp><request>"GET /apache_pb.gif HTTP/1.0" </request><status:2xx>200 </status:2xx><body-bytes-sent>2326 </body-bytes-sent><http-referer>"http://www.example.com/start.html" </http-referer><http-user-agent>"Mozilla/4.08 [en] (Win98; I ;Nav)"</http-user-agent>`,
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:12:46 +0000] "GET /robots.txt HTTP/1.1" 304 - "-" "curl/8.4.0"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:12:46 +0000] </timestamp><request>"GET /robots.txt HTTP/1.1" </request><status:3xx>304 </status:3xx><body-bytes-sent>- </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"curl/8.4.0"</http-user-agent>`,
		},
		{
			`2001:db8:85a3::8a2e:370:7334 - - [16/Feb/2024:00:13:02 +0000] "GET /index.html HTTP/2.0" 200 8193 "https://example.com/" "curl/8.4.0"`,
			"combined-log-format",
			`<remote-addr>2001:db8:85a3::8a2e:370:7334 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:13:02 +0000] </timestamp><request>"GET /index.html HTTP/2.0" </request><status:2xx>200 </status:2xx><body-bytes-sent>8193 </body-bytes-sent><http-referer>"https://example.com/" </http-referer><http-user-agent>"curl/8.4.0"</http-user-agent>`,
		},
		{
			`::1 - - [16/Feb/2024:00:13:44 +0000] "OPTIONS * HTTP/1.0" 200 - "-" "Apache/2.4.58 (Unix) (internal dummy connection)"`,
			"combined-log-format",
			`<remote-addr>::1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:13:44 +0000] </timestamp><request>"OPTIONS * HTTP/1.0" </request><status:2xx>200 </status:2xx><body-bytes-sent>- </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Apache/2.4.58 (Unix) (internal dummy connection)"</http-user-agent>`,
		},
		{
			`client.example.com - - [16/Feb/2024:00:14:07 +0000] "POST /api/v1/upload HTTP/1.1" 503 617 "-" "-"`,
			"combined-log-format",
			`<remote-addr>client.example.com </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:14:07 +0000] </timestamp><request>"POST /api/v1/upload HTTP/1.1" </request><status:5xx>503 </status:5xx><body-bytes-sent>617 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"-"</http-user-agent>`,
		},
		{
			`198.51.100.42 - - [16/Feb/2024:00:15:21 +0000] "GET /search?q=\"quoted\" HTTP/1.1" 400 226 "-" "python-requests/2.31.0"`,
			"combined-log-format",
			`<remote-addr>198.51.100.42 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:15:21 +0000] </timestamp><request>"GET /search?q=\"quoted\" HTTP/1.1" </request><status:4xx>400 </status:4xx><body-bytes-sent>226 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"python-requests/2.31.0"</http-user-agent>`,
		},
		{
			`127.0.0.1 - - [16/Feb/2024:00:16:03 +0000] "GET / HTTP/1.1" 000 0 "-" "-"`,
			"combined-log-format",
			`<remote-addr>127.0.0.1 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:00:16:03 +0000] </timestamp><request>"GET / HTTP/1.1" </request><status>000 </status><body-bytes-sent>0 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"-"</http-user-agent>`,
		},

		// traefik-common
		{
			`192.168.1.7 - - [16/Feb/2024:09:01:12 +0000] "GET /healthz HTTP/1.1" 100 2 "-" "kube-probe/1.29" 1 "web@docker" "http://172.17.0.3:80" 0ms`,
			"traefik-common",
			`<remote-addr>192.168.1.7 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:09:01:12 +0000] </timestamp><request>"GET /healthz HTTP/1.1" </request><status:1xx>100 </status:1xx><content-length>2 </content-length><referer>"-" </referer><user-agent>"kube-probe/1.29" </user-agent><request-count>1 </request-count><router-name>"web@docker" </router-name><server-url>"http://172.17.0.3:80" </server-url><duration>0ms</duration>`,
		},
		{
			`192.168.1.7 - - [16/Feb/2024:09:01:12 +0000] "GET /healthz HTTP/1.1" 200 2 "-" "kube-probe/1.29" 1 "web@docker" "http://172.17.0.3:80" 0ms`,
			"traefik-common",
			`<remote-addr>192.168.1.7 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:09:01:12 +0000] </timestamp><request>"GET /healthz HTTP/1.1" </request><status:2xx>200 </status:2xx><content-length>2 </content-length><referer>"-" </referer><user-agent>"kube-probe/1.29" </user-agent><request-count>1 </request-count><router-name>"web@docker" </router-name><server-url>"http://172.17.0.3:80" </server-url><duration>0ms</duration>`,
		},
		{
			`192.168.1.7 - - [16/Feb/2024:09:01:12 +0000] "GET /healthz HTTP/1.1" 302 2 "-" "kube-probe/1.29" 1 "web@docker" "http://172.17.0.3:80" 0ms`,
			"traefik-common",
			`<remote-addr>192.168.1.7 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:09:01:12 +0000] </timestamp><request>"GET /healthz HTTP/1.1" </request><status:3xx>302 </status:3xx><content-length>2 </content-length><referer>"-" </referer><user-agent>"kube-probe/1.29" </user-agent><request-count>1 </request-count><router-name>"web@docker" </router-name><server-url>"http://172.17.0.3:80" </server-url><duration>0ms</duration>`,
		},
		{
			`192.168.1.7 - - [16/Feb/2024:09:01:12 +0000] "GET /healthz HTTP/1.1" 404 2 "-" "kube-probe/1.29" 1 "web@docker" "http://172.17.0.3:80" 0ms`,
			"traefik-common",
			`<remote-addr>192.168.1.7 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:09:01:12 +0000] </timestamp><request>"GET /healthz HTTP/1.1" </request><status:4xx>404 </status:4xx><content-length>2 </content-length><referer>"-" </referer><user-agent>"kube-probe/1.29" </user-agent><request-count>1 </request-count><router-name>"web@docker" </router-name><server-url>"http://172.17.0.3:80" </server-url><duration>0ms</duration>`,
		},
		{
			`192.168.1.7 - - [16/Feb/2024:09:01:12 +0000] "GET /healthz HTTP/1.1" 503 2 "-" "kube-probe/1.29" 1 "web@docker" "http://172.17.0.3:80" 0ms`,
			"traefik-common",
			`<remote-addr>192.168.1.7 </remote-addr><ident>- </ident><remote-user>- </remote-user><timestamp>[16/Feb/2024:09:01:12 +0000] </timestamp><request>"GET /healthz HTTP/1.1" </request><status:5xx>503 </status:5xx><content-length>2 </content-length><referer>"-" </referer><user-agent>"kube-probe/1.29" </user-agent><request-count>1 </request-count><router-name>"web@docker" </router-name><server-url>"http://172.17.0.3:80" </server-url><duration>0ms</duration>`,
		},
		{
			`10.0.0.4 - admin [16/Feb/2024:09:02:44 +0000] "POST /api/v1/orders HTTP/2.0" 201 512 "https://shop.example.com/cart" "Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/123.0" 137 "orders-router@kubernetes" "http://10.42.1.19:8000" 43.271ms`,
			"traefik-common",
			`<remote-addr>10.0.0.4 </remote-addr><ident>- </ident><remote-user>admin </remote-user><timestamp>[16/Feb/2024:09:02:44 +0000] </timestamp><request>"POST /api/v1/orders HTTP/2.0" </request><status:2xx>201 </status:2xx><content-length>512 </content-length><referer>"https://shop.example.com/cart" </referer><user-agent>"Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/123.0" </user-agent><request-count>137 </request-count><router-name>"orders-router@kubernetes" </router-name><server-url>"http://10.42.1.19:8000" </server-url><duration>43.271ms</duration>`,
		},

		// ingress-nginx-controller
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 100 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:1xx>100 </status:1xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:4xx>403 </upstream-status:4xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 200 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:2xx>200 </status:2xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:4xx>403 </upstream-status:4xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 302 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:3xx>302 </status:3xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:4xx>403 </upstream-status:4xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 404 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:4xx>404 </status:4xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:4xx>403 </upstream-status:4xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 503 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:5xx>503 </status:5xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:4xx>403 </upstream-status:4xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 100 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:4xx>403 </status:4xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:1xx>100 </upstream-status:1xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 200 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:4xx>403 </status:4xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:2xx>200 </upstream-status:2xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 302 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:4xx>403 </status:4xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:3xx>302 </upstream-status:3xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 404 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:4xx>403 </status:4xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:4xx>404 </upstream-status:4xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 503 07d2cd60741517a6d8222f40757b94c4`,
			"ingress-nginx-controller",
			`<remote-addr>127.0.0.102 </remote-addr><dash>- </dash><remote-user>- </remote-user><time-local>[27/Jun/2023:07:13:16 +0000] </time-local><request>"GET /language/en-GB/en-GB.xml HTTP/1.1" </request><status:4xx>403 </status:4xx><body-bytes-sent>9 </body-bytes-sent><http-referer>"-" </http-referer><http-user-agent>"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" </http-user-agent><request-length>619 </request-length><request-time>0.003 </request-time><proxy-upstream-name>[imgproxy-imgproxy-imgproxy-80] </proxy-upstream-name><proxy-alternative-upstream-name>[] </proxy-alternative-upstream-name><upstream-addr>10.64.6.9:8080 </upstream-addr><upstream-response-length>9 </upstream-response-length><upstream-response-time>0.003 </upstream-response-time><upstream-status:5xx>503 </upstream-status:5xx><req-id>07d2cd60741517a6d8222f40757b94c4</req-id>`,
		},

		// klog
		{
			`I0410 23:18:43.650599       1 controller.go:175] "starting healthz server" logger="cert-manager.controller" address="[::]:9403"`,
			"klog",
			`<log-level:info>I0410 </log-level:info><time>23:18:43.650599</time><thread-id>       1 </thread-id><filename>controller.go</filename><line-number>:175</line-number><bracket>] </bracket>"<w:good>starting</w:good> healthz server"<p:logfmt-string:key> logger</p:logfmt-string:key><p:logfmt-string:equal-sign>=</p:logfmt-string:equal-sign><p:logfmt-string:opening-quotation-mark>"</p:logfmt-string:opening-quotation-mark>cert-manager.controller<p:logfmt-string:closing-quotation-mark>"</p:logfmt-string:closing-quotation-mark><p:logfmt-string:key> address</p:logfmt-string:key><p:logfmt-string:equal-sign>=</p:logfmt-string:equal-sign><p:logfmt-string:opening-quotation-mark>"</p:logfmt-string:opening-quotation-mark><p:ipv6-address:opening-bracket>[</p:ipv6-address:opening-bracket><p:ipv6-address:address>::</p:ipv6-address:address><p:ipv6-address:closing-bracket>]</p:ipv6-address:closing-bracket><p:ipv6-address:port>:9403</p:ipv6-address:port><p:logfmt-string:closing-quotation-mark>"</p:logfmt-string:closing-quotation-mark>`,
		},
		{
			`W0704 20:01:06.932182       1 warnings.go:70] annotation "kubernetes.io/ingress.class" is deprecated, please use 'spec.ingressClassName' instead`,
			"klog",
			`<log-level:warning>W0704 </log-level:warning><time>20:01:06.932182</time><thread-id>       1 </thread-id><filename>warnings.go</filename><line-number>:70</line-number><bracket>] </bracket>annotation "kubernetes.io/ingress.class" is deprecated, please use 'spec.ingressClassName' instead`,
		},
		{
			`E0714 16:12:36.594249       1 controller.go:104] "Unhandled Error" err="ingress 'menetekel/main' in work queue no longer exists" logger="UnhandledError"`,
			"klog",
			`<log-level:error>E0714 </log-level:error><time>16:12:36.594249</time><thread-id>       1 </thread-id><filename>controller.go</filename><line-number>:104</line-number><bracket>] </bracket>"Unhandled <w:bad>Error</w:bad>"<p:logfmt-string:key> err</p:logfmt-string:key><p:logfmt-string:equal-sign>=</p:logfmt-string:equal-sign><p:logfmt-string:opening-quotation-mark>"</p:logfmt-string:opening-quotation-mark>ingress 'menetekel/main' in work queue no longer exists<p:logfmt-string:closing-quotation-mark>"</p:logfmt-string:closing-quotation-mark><p:logfmt-string:key> logger</p:logfmt-string:key><p:logfmt-string:equal-sign>=</p:logfmt-string:equal-sign><p:logfmt-string:opening-quotation-mark>"</p:logfmt-string:opening-quotation-mark>UnhandledError<p:logfmt-string:closing-quotation-mark>"</p:logfmt-string:closing-quotation-mark>`,
		},
		{
			`F0123 00:12:34.567890       1 controller.go:4] "Fatal Error" err="fatal error"`,
			"klog",
			`<log-level:fatal>F0123 </log-level:fatal><time>00:12:34.567890</time><thread-id>       1 </thread-id><filename>controller.go</filename><line-number>:4</line-number><bracket>] </bracket>"<w:bad>Fatal</w:bad> <w:bad>Error</w:bad>"<p:logfmt-string:key> err</p:logfmt-string:key><p:logfmt-string:equal-sign>=</p:logfmt-string:equal-sign><p:logfmt-string:opening-quotation-mark>"</p:logfmt-string:opening-quotation-mark><w:bad>fatal</w:bad> <w:bad>error</w:bad><p:logfmt-string:closing-quotation-mark>"</p:logfmt-string:closing-quotation-mark>`,
		},

		// redis
		{
			`1:M 01 Feb 2024 19:41:07.226 # monotonic clock: POSIX clock_gettime`,
			"redis",
			`<pid>1</pid><colon>:</colon><role:master>M </role:master><date>01 Feb 2024 </date><time>19:41:07.226 </time><log-level:warning># </log-level:warning>monotonic clock: POSIX clock_gettime`,
		},
		{
			`22:S 17 Feb 2024 00:39:12.500 * Starting automatic rewriting of AOF on 3886% growth`,
			"redis",
			`<pid>22</pid><colon>:</colon><role:replica>S </role:replica><date>17 Feb 2024 </date><time>00:39:12.500 </time><log-level:notice>* </log-level:notice><w:good>Starting</w:good> automatic rewriting of AOF on 3886% growth`,
		},
		{
			`375:X 20 Jun 2025 13:27:11.773 - Sentinel ID is 2814dfe0610f4b8a99b4c6076693ed87d032af23`,
			"redis",
			`<pid>375</pid><colon>:</colon><role:sentinel>X </role:sentinel><date>20 Jun 2025 </date><time>13:27:11.773 </time><log-level:info>- </log-level:info>Sentinel ID is<p:duration:start> </p:duration:start><p:duration:number>2814</p:duration:number><p:duration:unit>d</p:duration:unit>fe0610f4b8a99b4c6076693ed87d032af23`,
		},
		{
			`8792:C 01 Feb 2024 19:41:07.224 . oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo`,
			"redis",
			`<pid>8792</pid><colon>:</colon><role:rdb-aof-writing-child>C </role:rdb-aof-writing-child><date>01 Feb 2024 </date><time>19:41:07.224 </time><log-level:debug>. </log-level:debug>oO0OoO0OoO0Oo Redis is <w:good>starting</w:good> oO0OoO0OoO0Oo`,
		},

		// syslog-rfc3164
		{
			`Jul  3 08:27:19 menetekel systemd[1]: Condition check resulted in MD array scrubbing - continuation being skipped.`,
			"syslog-rfc3164",
			`<priority></priority><date>Jul  3 </date><time>08:27:19 </time><hostname>menetekel </hostname><program>systemd</program><pid>[1]</pid><colon>: </colon>Condition check resulted in MD array scrubbing - continuation being <w:warning>skipped</w:warning>.`,
		},
		{
			`Jul  3 09:17:01 menetekel CRON[1185749]: (root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
			"syslog-rfc3164",
			`<priority></priority><date>Jul  3 </date><time>09:17:01 </time><hostname>menetekel </hostname><program>CRON</program><pid>[1185749]</pid><colon>: </colon>(root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
		},
		{
			`Jul 13 10:17:02 menetekel CRON[1190762]: (root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
			"syslog-rfc3164",
			`<priority></priority><date>Jul 13 </date><time>10:17:02 </time><hostname>menetekel </hostname><program>CRON</program><pid>[1190762]</pid><colon>: </colon>(root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
		},
		{
			`Jul 13 10:20:04 menetekel systemd[1]: Starting Certbot...`,
			"syslog-rfc3164",
			`<priority></priority><date>Jul 13 </date><time>10:20:04 </time><hostname>menetekel </hostname><program>systemd</program><pid>[1]</pid><colon>: </colon><w:good>Starting</w:good> Certbot...`,
		},
		{
			`Jul 13 10:17:02 menetekel CRON: (root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
			"syslog-rfc3164",
			`<priority></priority><date>Jul 13 </date><time>10:17:02 </time><hostname>menetekel </hostname><program>CRON</program><pid></pid><colon>: </colon>(root) CMD (   cd / && run-parts --report /etc/cron.hourly)`,
		},
		{
			`Jul 13 10:20:04 menetekel systemd: Starting Certbot...`,
			"syslog-rfc3164",
			`<priority></priority><date>Jul 13 </date><time>10:20:04 </time><hostname>menetekel </hostname><program>systemd</program><pid></pid><colon>: </colon><w:good>Starting</w:good> Certbot...`,
		},
		{
			`<25>Jul 13 10:20:04 menetekel systemd[1]: certbot.service: Deactivated successfully.`,
			"syslog-rfc3164",
			`<priority><25></priority><date>Jul 13 </date><time>10:20:04 </time><hostname>menetekel </hostname><program>systemd</program><pid>[1]</pid><colon>: </colon>certbot.service: Deactivated <w:good>successfully</w:good>.`,
		},
		{
			`<123>Jul 13 10:20:04 menetekel systemd[1]: Finished Certbot.`,
			"syslog-rfc3164",
			`<priority><123></priority><date>Jul 13 </date><time>10:20:04 </time><hostname>menetekel </hostname><program>systemd</program><pid>[1]</pid><colon>: </colon><w:good>Finished</w:good> Certbot.`,
		},
	}

	hl, labels := newLabeledHighlighter(t)

	for _, tt := range tests {
		t.Run("TestFormatsBuiltins"+tt.plain, func(t *testing.T) {
			// formats are tried in alphabetical order with no tie-breaker, so
			// overlapping formats must stay distinguishable by their anchors
			var matched []string
			for _, format := range hl.formats {
				if format.match(tt.plain) {
					matched = append(matched, format.Name)
				}
			}
			if len(matched) != 1 || matched[0] != tt.format {
				t.Errorf("got %v, want exactly one format: [%s]", matched, tt.format)

				return
			}

			if labeled := labels.decode(hl.Colorize(tt.plain)); labeled != tt.labeled {
				t.Errorf("got %v, want %v", labeled, tt.labeled)
			}
		})
	}
}
