package logalize

import (
	"bytes"
	"embed"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/muesli/termenv"
)

//go:embed builtins/*
//go:embed themes/*
var builtinsAllGood embed.FS

func TestConfigLoadBuiltinGood(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		// nginx-combined
		{
			`127.0.0.1 - - [16/Feb/2024:00:01:01 +0000] "GET / HTTP/1.1" 301 162 "-" "Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1"`,
			"\x1b[38;2;255;199;119m127.0.0.1 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[16/Feb/2024:00:01:01 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET / HTTP/1.1\" \x1b[0m\x1b[38;2;0;255;255;1m301 \x1b[0m\x1b[38;2;99;109;166m162 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Mobile/15E148 Safari/604.1\"\x1b[0m",
		},
		// nginx-ingress-controller
		{
			`127.0.0.102 - - [27/Jun/2023:07:13:16 +0000] "GET /language/en-GB/en-GB.xml HTTP/1.1" 403 9 "-" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36" 619 0.003 [imgproxy-imgproxy-imgproxy-80] [] 10.64.6.9:8080 9 0.003 403 07d2cd60741517a6d8222f40757b94c4`,
			"\x1b[38;2;255;199;119m127.0.0.102 \x1b[0m\x1b[38;2;130;139;184m- \x1b[0m\x1b[38;2;79;214;190m- \x1b[0m\x1b[38;2;192;153;255m[27/Jun/2023:07:13:16 +0000] \x1b[0m\x1b[38;2;195;232;141m\"GET /language/en-GB/en-GB.xml HTTP/1.1\" \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m9 \x1b[0m\x1b[38;2;252;167;234m\"-\" \x1b[0m\x1b[38;2;130;170;255m\"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36\" \x1b[0m\x1b[38;2;65;166;181m619 \x1b[0m\x1b[38;2;195;232;141m0.003 \x1b[0m\x1b[38;2;101;188;255m[imgproxy-imgproxy-imgproxy-80] \x1b[0m\x1b[38;2;99;109;166m[] \x1b[0m\x1b[38;2;255;199;119m10.64.6.9:8080 \x1b[0m\x1b[38;2;192;153;255m9 \x1b[0m\x1b[38;2;255;117;127m0.003 \x1b[0m\x1b[38;2;255;0;0;1m403 \x1b[0m\x1b[38;2;99;109;166m07d2cd60741517a6d8222f40757b94c4\x1b[0m",
		},
		// klog
		{
			`I0410 13:18:43.650599       1 controller.go:175] "starting healthz server" logger="cert-manager.controller" address="[::]:9403"`,
			"\x1b[38;2;130;170;255;1mI0410 \x1b[0m\x1b[38;2;252;167;234m13:18:43.650599\x1b[0m\x1b[38;2;99;109;166m       1 \x1b[0m\x1b[38;2;137;221;255mcontroller.go\x1b[0m\x1b[38;2;99;109;166m:175\x1b[0m\x1b[38;2;255;150;108m] \x1b[0m\"\x1b[38;2;81;250;138;1mstarting\x1b[0m healthz server\"\x1b[38;2;154;173;236m logger\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0mcert-manager.controller\x1b[38;2;154;173;236m\"\x1b[0m\x1b[38;2;154;173;236m address\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255m::\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:9403\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m",
		},
		// redis
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
		},

		// patterns
		// rfc3339
		{
			`2024-02-17T06:56:10Z`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mT\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255mZ\x1b[0m",
		},
		{
			`2024-02-17T06:56:10+05:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mT\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255m+05:00\x1b[0m",
		},
		{
			`2024-02-17T06:56:10.636960544-01:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mT\x1b[0m\x1b[38;2;252;167;234m06:56:10.636960544\x1b[0m\x1b[38;2;130;170;255m-01:00\x1b[0m",
		},
		{
			`2024-02-17t06:56:10z`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mt\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255mz\x1b[0m",
		},
		{
			`2024-02-17t06:56:10+05:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mt\x1b[0m\x1b[38;2;252;167;234m06:56:10\x1b[0m\x1b[38;2;130;170;255m+05:00\x1b[0m",
		},
		{
			`2024-02-17t06:56:10.636960544-01:00`,
			"\x1b[38;2;192;153;255m2024-02-17\x1b[0m\x1b[38;2;130;170;255mt\x1b[0m\x1b[38;2;252;167;234m06:56:10.636960544\x1b[0m\x1b[38;2;130;170;255m-01:00\x1b[0m",
		},

		// time
		{`23:42:12`, "\x1b[38;2;252;167;234m23:42:12\x1b[0m"},
		{`01:37:59.743`, "\x1b[38;2;252;167;234m01:37:59.743\x1b[0m"},
		{`17:49:37.034123`, "\x1b[38;2;252;167;234m17:49:37.034123\x1b[0m"},

		// dates
		{`1999-07-10`, "\x1b[38;2;192;153;255m1999-07-10\x1b[0m"},
		{`1999/07/10`, "\x1b[38;2;192;153;255m1999/07/10\x1b[0m"},
		{`07-10-1999`, "\x1b[38;2;192;153;255m07-10-1999\x1b[0m"},
		{`07/10/1999`, "\x1b[38;2;192;153;255m07/10/1999\x1b[0m"},
		{`27 Jan`, "\x1b[38;2;192;153;255m27 Jan\x1b[0m"},
		{`27 January`, "\x1b[38;2;192;153;255m27 January\x1b[0m"},
		{`27 Jan 2023`, "\x1b[38;2;192;153;255m27 Jan 2023\x1b[0m"},
		{`27 August 2023`, "\x1b[38;2;192;153;255m27 August 2023\x1b[0m"},
		{`27-Jan-2023`, "\x1b[38;2;192;153;255m27-Jan-2023\x1b[0m"},
		{`27-August-2023`, "\x1b[38;2;192;153;255m27-August-2023\x1b[0m"},
		{`27/Jan/2023`, "\x1b[38;2;192;153;255m27/Jan/2023\x1b[0m"},
		{`27/August/2023`, "\x1b[38;2;192;153;255m27/August/2023\x1b[0m"},
		{`Jan 27`, "\x1b[38;2;192;153;255mJan 27\x1b[0m"},
		{`January 27`, "\x1b[38;2;192;153;255mJanuary 27\x1b[0m"},
		{`Jan 27 2023`, "\x1b[38;2;192;153;255mJan 27 2023\x1b[0m"},
		{`August 27 2023`, "\x1b[38;2;192;153;255mAugust 27 2023\x1b[0m"},
		{`Jan-27-2023`, "\x1b[38;2;192;153;255mJan-27-2023\x1b[0m"},
		{`August-27-2023`, "\x1b[38;2;192;153;255mAugust-27-2023\x1b[0m"},
		{`Jan/27/2023`, "\x1b[38;2;192;153;255mJan/27/2023\x1b[0m"},
		{`August/27/2023`, "\x1b[38;2;192;153;255mAugust/27/2023\x1b[0m"},
		{`Mon 17`, "\x1b[38;2;192;153;255mMon 17\x1b[0m"},
		{`Sunday 3`, "\x1b[38;2;192;153;255mSunday 3\x1b[0m"},

		// duration
		{`75.984854ms`, "\x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m"},
		{`5s`, "\x1b[38;2;79;214;190m5\x1b[0m\x1b[38;2;65;166;181ms\x1b[0m"},
		{`784m`, "\x1b[38;2;79;214;190m784\x1b[0m\x1b[38;2;65;166;181mm\x1b[0m"},
		{`7.5h`, "\x1b[38;2;79;214;190m7.5\x1b[0m\x1b[38;2;65;166;181mh\x1b[0m"},
		{`25d`, "\x1b[38;2;79;214;190m25\x1b[0m\x1b[38;2;65;166;181md\x1b[0m"},

		// logfmt-general
		{`key=value`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0mvalue"},
		{`key=5s`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;79;214;190m5\x1b[0m\x1b[38;2;65;166;181ms\x1b[0m"},

		// logfmt-string
		{`key="value"`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0mvalue\x1b[38;2;154;173;236m\"\x1b[0m"},
		{`key="5s"`, "\x1b[38;2;154;173;236mkey\x1b[0m\x1b[38;2;99;109;166m=\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m\x1b[38;2;79;214;190m5\x1b[0m\x1b[38;2;65;166;181ms\x1b[0m\x1b[38;2;154;173;236m\"\x1b[0m"},

		// ipv4-address
		{`127.0.0.1`, "\x1b[38;2;118;211;255m127.0.0.1\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`255.255.255.255`, "\x1b[38;2;118;211;255m255.255.255.255\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`0.0.0.0`, "\x1b[38;2;118;211;255m0.0.0.0\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},
		{`10.0.0.200/16`, "\x1b[38;2;118;211;255m10.0.0.200\x1b[0m\x1b[38;2;13;185;215m/16\x1b[0m"},
		{`10.0.0.0/8`, "\x1b[38;2;118;211;255m10.0.0.0\x1b[0m\x1b[38;2;13;185;215m/8\x1b[0m"},
		{`10.0.7.107:80`, "\x1b[38;2;118;211;255m10.0.7.107\x1b[0m\x1b[38;2;13;185;215m:80\x1b[0m"},
		{`8.9.10.237:8080`, "\x1b[38;2;118;211;255m8.9.10.237\x1b[0m\x1b[38;2;13;185;215m:8080\x1b[0m"},
		{`1.2.3.4:17846`, "\x1b[38;2;118;211;255m1.2.3.4\x1b[0m\x1b[38;2;13;185;215m:17846\x1b[0m"},

		// ipv6-address
		{
			`2001:db8:4006:812::200e`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:4006:812::200e\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0db8:0000:cd30:0000:0000:0000:0000`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0db8:0000:cd30:0000:0000:0000:0000\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0db8::cd30:0:0:0:0`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0db8::cd30:0:0:0:0\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0db8:0:cd30::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0db8:0:cd30::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff02:0:0:0:0:1:ff00:0000`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:1:ff00:0000\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff02:0:0:0:0:1:ffff:ffff`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:1:ffff:ffff\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8::1234:5678`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8::1234:5678\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff02:0:0:0:0:0:0:2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:0:0:2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`fdf8:f53b:82e4::53`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mfdf8:f53b:82e4::53\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`fe80::200:5aee:feaa:20a2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mfe80::200:5aee:feaa:20a2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0000:4136:e378:`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0000:4136:e378:\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`8000:63bf:3fff:fdd2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m8000:63bf:3fff:fdd2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::1234:5678`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::1234:5678\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2000::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2000::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8:a0b:12f0::1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:a0b:12f0::1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:4:112:cd:65a:753:0:a1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:4:112:cd:65a:753:0:a1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0002:6c::430`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0002:6c::430\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:5::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:5::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`fe08::7:8`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mfe08::7:8\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2002:cb0a:3cdd:1::1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2002:cb0a:3cdd:1::1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8:8:4::2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:8:4::2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`ff01:0:0:0:0:0:0:2`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255mff01:0:0:0:0:0:0:2\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:0:0`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:0:0\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:0000::`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:0000::\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:192.0.2.47`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:192.0.2.47\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:0.0.0.0`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:0.0.0.0\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:255.255.255.255`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:255.255.255.255\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:10.0.0.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:10.0.0.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::192.168.0.1`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::192.168.0.1\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::255.255.255.255`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::255.255.255.255\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`2001:db8:122:344::192.0.2.33`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m2001:db8:122:344::192.0.2.33\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`0:0:0:0:0:0:13.1.68.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m0:0:0:0:0:0:13.1.68.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`0:0:0:0:0:ffff:129.144.52.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m0:0:0:0:0:ffff:129.144.52.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::13.1.68.3`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::13.1.68.3\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`::ffff:129.144.52.38`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m::ffff:129.144.52.38\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`59fb:0:0:0:0:1005:cc57:6571`,
			"\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;118;211;255m59fb:0:0:0:0:1005:cc57:6571\x1b[0m\x1b[38;2;99;109;166m\x1b[0m\x1b[38;2;13;185;215m\x1b[0m",
		},
		{
			`[2001:5::]:22`,
			"\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255m2001:5::\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:22\x1b[0m",
		},
		{
			`[2001:db8:4006:812::200e]:8080`,
			"\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255m2001:db8:4006:812::200e\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:8080\x1b[0m",
		},
		{
			`[ff02:0:0:0:0:1:ffff:ffff]:23456`,
			"\x1b[38;2;99;109;166m[\x1b[0m\x1b[38;2;118;211;255mff02:0:0:0:0:1:ffff:ffff\x1b[0m\x1b[38;2;99;109;166m]\x1b[0m\x1b[38;2;13;185;215m:23456\x1b[0m",
		},

		// mac-address
		{`3D:F2:C9:A6:B3:4F`, "\x1b[38;2;79;214;190m3D:F2:C9:A6:B3:4F\x1b[0m"},
		{`3D-F2-C9-A6-B3-4F`, "\x1b[38;2;79;214;190m3D-F2-C9-A6-B3-4F\x1b[0m"},
		{`3d:f2:c9:a6:b3:4f`, "\x1b[38;2;79;214;190m3d:f2:c9:a6:b3:4f\x1b[0m"},
		{`3d-f2-c9-a6-b3-4f`, "\x1b[38;2;79;214;190m3d-f2-c9-a6-b3-4f\x1b[0m"},

		// uuid
		{`0a99af43-0ad4-4237-b9cd-064966eb2803`, "\x1b[38;2;134;225;252m0a99af43-0ad4-4237-b9cd-064966eb2803\x1b[0m"},

		// words
		{"untrue", "untrue"},
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},
		{"fail", "\x1b[38;2;240;108;97;1mfail\x1b[0m"},
		{"failed", "\x1b[38;2;240;108;97;1mfailed\x1b[0m"},

		{"not true", "\x1b[38;2;240;108;97;1mnot true\x1b[0m"},
		{"Not true", "\x1b[38;2;240;108;97;1mNot true\x1b[0m"},
		{"wasn't true", "\x1b[38;2;240;108;97;1mwasn't true\x1b[0m"},
		{"won't true", "\x1b[38;2;240;108;97;1mwon't true\x1b[0m"},
		{"cannot complete", "\x1b[38;2;240;108;97;1mcannot complete\x1b[0m"},
		{"won't be completed", "\x1b[38;2;240;108;97;1mwon't be completed\x1b[0m"},
		{"cannot be completed", "\x1b[38;2;240;108;97;1mcannot be completed\x1b[0m"},
		{"should not be completed", "\x1b[38;2;240;108;97;1mshould not be completed\x1b[0m"},

		{"not false", "\x1b[38;2;81;250;138;1mnot false\x1b[0m"},
		{"Not false", "\x1b[38;2;81;250;138;1mNot false\x1b[0m"},
		{"wasn't false", "\x1b[38;2;81;250;138;1mwasn't false\x1b[0m"},
		{"won't false", "\x1b[38;2;81;250;138;1mwon't false\x1b[0m"},
		{"cannot fail", "\x1b[38;2;81;250;138;1mcannot fail\x1b[0m"},
		{"won't be failed", "\x1b[38;2;81;250;138;1mwon't be failed\x1b[0m"},
		{"cannot be failed", "\x1b[38;2;81;250;138;1mcannot be failed\x1b[0m"},
		{"should not be failed", "\x1b[38;2;81;250;138;1mshould not be failed\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m \x1b[38;2;118;211;255m7.7.7.7\x1b[0m\x1b[38;2;13;185;215m\x1b[0m \x1b[38;2;252;167;234m01:37:59.743\x1b[0m \x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		Theme: "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagNoBuiltins(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      false,
		NoBuiltins:          true,
		Theme:               "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagNoBuiltinLogFormats(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C \x1b[38;2;192;153;255m17 Feb 2024\x1b[0m \x1b[38;2;252;167;234m00:39:12.557\x1b[0m * Parent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},

		// words
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m \x1b[38;2;118;211;255m7.7.7.7\x1b[0m\x1b[38;2;13;185;215m\x1b[0m \x1b[38;2;252;167;234m01:37:59.743\x1b[0m \x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: true,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      false,
		NoBuiltins:          false,
		Theme:               "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagNoBuiltinPatterns(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      false,
		NoBuiltins:          false,
		Theme:               "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagNoBuiltinWords(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail \x1b[38;2;118;211;255m7.7.7.7\x1b[0m\x1b[38;2;13;185;215m\x1b[0m \x1b[38;2;252;167;234m01:37:59.743\x1b[0m \x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   false,
		NoBuiltinWords:      true,
		NoBuiltins:          false,
		Theme:               "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagNoBuiltinPatternsAndWords(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		NoBuiltinLogFormats: false,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      true,
		NoBuiltins:          false,
		Theme:               "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagDryRun(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      false,
		DryRun:                  true,
		Theme:                   "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagHighlightOnlyLogFormats(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"\x1b[38;2;154;173;236m4018569\x1b[0m\x1b[38;2;99;109;166m:\x1b[0m\x1b[38;2;184;219;135;1mC \x1b[0m\x1b[38;2;192;153;255m17 Feb 2024 \x1b[0m\x1b[38;2;252;167;234m00:39:12.557 \x1b[0m\x1b[38;2;137;221;255;1m* \x1b[0mParent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: true,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      false,
		DryRun:                  false,
		Theme:                   "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagHighlightOnlyPatterns(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C \x1b[38;2;192;153;255m17 Feb 2024\x1b[0m \x1b[38;2;252;167;234m00:39:12.557\x1b[0m * Parent agreed to stop sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "\x1b[38;2;118;211;255m12.34.56.78\x1b[0m\x1b[38;2;13;185;215m\x1b[0m"},

		// words
		{"true", "true"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail \x1b[38;2;118;211;255m7.7.7.7\x1b[0m\x1b[38;2;13;185;215m\x1b[0m \x1b[38;2;252;167;234m01:37:59.743\x1b[0m \x1b[38;2;79;214;190m75.984854\x1b[0m\x1b[38;2;65;166;181mms\x1b[0m",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   true,
		HighlightOnlyWords:      false,
		DryRun:                  false,
		Theme:                   "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinFlagHighlightOnlyWords(t *testing.T) {
	colorProfile = termenv.TrueColor

	tests := []struct {
		plain   string
		colored string
	}{
		// log formats
		{
			`4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to stop sending diffs. Finalizing AOF...`,
			"4018569:C 17 Feb 2024 00:39:12.557 * Parent agreed to \x1b[38;2;240;108;97;1mstop\x1b[0m sending diffs. Finalizing AOF...",
		},

		// patterns
		{`12.34.56.78`, "12.34.56.78"},

		// words
		{"true", "\x1b[38;2;81;250;138;1mtrue\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"\x1b[38;2;81;250;138;1mtrue\x1b[0m bad \x1b[38;2;240;108;97;1mfail\x1b[0m 7.7.7.7 01:37:59.743 75.984854ms",
		},
	}

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
	}

	Opts = Settings{
		HighlightOnlyLogFormats: false,
		HighlightOnlyPatterns:   false,
		HighlightOnlyWords:      true,
		DryRun:                  false,
		Theme:                   "tokyonight",
	}

	err = InitConfig(builtinsAllGood)
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

func TestConfigLoadBuiltinBad(t *testing.T) {
	colorProfile = termenv.TrueColor

	Opts = Settings{
		Theme: "tokyonight",
	}

	builtinsLogformatsBad := fstest.MapFS{
		"builtins/logformats/bad.yaml": {
			Data: []byte("formats:\n  test:\n  regexp: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinLogformatsBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsLogformatsBad)
		if err == nil || err.Error() != "yaml: line 3: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	builtinsPatternsBad := fstest.MapFS{
		"builtins/patterns/bad.yaml": {
			Data: []byte("patterns:\n  bad: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinPatternsBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsPatternsBad)
		if err == nil || err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	builtinsWordsBad := fstest.MapFS{
		"builtins/words/bad.yaml": {
			Data: []byte("words:\n  bad: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinWordsBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsWordsBad)
		if err == nil || err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	builtinsThemesBad := fstest.MapFS{
		"themes/bad.yaml": {
			Data: []byte("themes:\n  bad: bad:\n"),
		},
	}

	t.Run("TestConfigLoadBuiltinThemesBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsThemesBad)
		if err == nil || err.Error() != "yaml: line 2: mapping values are not allowed in this context" {
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
      name: one
    - regexp: ([^ ]+ )
      name: two
    - regexp: (\[.+\] )
      name: three
    - regexp: ("[^"]+")
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
		{`127.0.0.1 - [test] "testing"`, "\x1b[38;2;245;206;65m127.0.0.1 \x1b[0m\x1b[48;2;118;73;158m- \x1b[0m\x1b[1m[test] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing\"\x1b[0m"},
		{`127.0.0.2 test [test hello] "testing again"`, "\x1b[38;2;245;206;65m127.0.0.2 \x1b[0m\x1b[48;2;118;73;158mtest \x1b[0m\x1b[1m[test hello] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"testing again\"\x1b[0m"},
		{`127.0.0.3 ___ [.] "_"`, "\x1b[38;2;245;206;65m127.0.0.3 \x1b[0m\x1b[48;2;118;73;158m___ \x1b[0m\x1b[1m[.] \x1b[0m\x1b[38;2;157;175;153;48;2;118;251;153;4m\"_\"\x1b[0m"},

		// pattern
		{`"string"`, "\x1b[38;2;0;255;0m\"string\"\x1b[0m"},
		{`42`, "\x1b[48;2;0;80;80m42\x1b[0m"},
		{`127.0.0.1`, "\x1b[38;2;255;0;0;48;2;255;255;0;1m127.0.0.1\x1b[0m"},
		{`"test": 127.7.7.7 hello 101`, "\x1b[38;2;0;255;0m\"test\"\x1b[0m: \x1b[38;2;255;0;0;48;2;255;255;0;1m127.7.7.7\x1b[0m hello \x1b[38;2;80;80;80m101\x1b[0m"},
		{`"42"`, "\x1b[38;2;0;255;0m\"42\"\x1b[0m"},
		{`"237.7.7.7"`, "\x1b[38;2;0;255;0m\"237.7.7.7\"\x1b[0m"},
		{`status 103`, "status \x1b[38;2;80;80;80m103\x1b[0m"},
		{`status 200`, "status \x1b[38;2;0;255;0;53m200\x1b[0m"},
		{`status 302`, "status \x1b[38;2;0;255;255;9m302\x1b[0m"},
		{`status 404`, "status \x1b[38;2;255;0;0;7m404\x1b[0m"},
		{`status 503`, "status \x1b[38;2;255;0;255m503\x1b[0m"},
		{`status 700`, "status \x1b[38;2;255;255;255m700\x1b[0m"},

		// words
		{"wenzel", "\x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"argus", "\x1b[38;2;18;15;187;4margus\x1b[0m"},

		{"not toni", "not \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"Not wenzel", "Not \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"wasn't argus", "wasn't \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"won't cletus", "won't \x1b[38;2;18;15;187;4mcletus\x1b[0m"},
		{"cannot toni", "cannot \x1b[38;2;248;52;178;4mtoni\x1b[0m"},
		{"won't be wenzel", "won't be \x1b[38;2;248;52;178;4mwenzel\x1b[0m"},
		{"cannot be argus", "cannot be \x1b[38;2;18;15;187;4margus\x1b[0m"},
		{"should not be cletus", "should not be \x1b[38;2;18;15;187;4mcletus\x1b[0m"},

		// patterns and words
		{
			`true bad fail 7.7.7.7 01:37:59.743 75.984854ms`,
			"true bad fail \x1b[38;2;255;0;0;48;2;255;255;0;1m7.7.7.7\x1b[0m \x1b[48;2;0;80;80m01\x1b[0m:\x1b[48;2;0;80;80m37\x1b[0m:\x1b[48;2;0;80;80m59\x1b[0m.\x1b[38;2;255;255;255m743\x1b[0m \x1b[48;2;0;80;80m75\x1b[0m.\x1b[38;2;255;255;255m984\x1b[0m\x1b[38;2;255;255;255m854\x1b[0mms",
		},
		{
			`"wenzel" and wenzel`,
			"\x1b[38;2;0;255;0m\"wenzel\"\x1b[0m and \x1b[38;2;248;52;178;4mwenzel\x1b[0m",
		},
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

	Opts = Settings{
		ConfigPath: userConfig,
		NoBuiltins: true,
		Theme:      "test",
	}

	err = InitConfig(builtinsAllGood)
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

	Opts = Settings{
		ConfigPath: userConfig,
		NoBuiltins: true,
		Theme:      "tokyonight",
	}

	t.Run("TestConfigLoadUserDefinedBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsAllGood)
		if err == nil || err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	Opts = Settings{
		ConfigPath: userConfig + "error",
		NoBuiltins: true,
		Theme:      "tokyonight",
	}

	t.Run("TestConfigLoadUserDefinedFileDoesntExist", func(t *testing.T) {
		err := InitConfig(builtinsAllGood)
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

	Opts = Settings{
		ConfigPath: userConfigReadOnly,
		NoBuiltins: false,
		Theme:      "tokyonight",
	}

	t.Run("TestConfigLoadUserDefinedReadOnly", func(t *testing.T) {
		err := InitConfig(builtinsAllGood)
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

	tempDefaultConfig := t.TempDir() + "/tempDefaultConfig.yaml"
	configRaw := []byte(configDataBadYAML)
	err := os.WriteFile(tempDefaultConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", tempDefaultConfig, err)
	}

	t.Cleanup(func() {
		err = os.Remove(tempDefaultConfig)
		if err != nil {
			t.Errorf("Wasn't able to delete %s: %s", tempDefaultConfig, err)
		}
	})

	Opts = Settings{
		ConfigPath: "",
		NoBuiltins: true,
		Theme:      "tokyonight",
	}

	defaultConfigPaths = append(defaultConfigPaths, tempDefaultConfig)

	t.Run("TestConfigLoadDefaultBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsAllGood)
		if err == nil || err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	err = os.Chmod(tempDefaultConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", tempDefaultConfig, err)
	}

	t.Run("TestConfigLoadDefaultReadOnly", func(t *testing.T) {
		err := InitConfig(builtinsAllGood)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitConfig() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})

	defaultConfigPaths = defaultConfigPaths[:len(defaultConfigPaths)-1]
}

func TestConfigLoadDotLogalize(t *testing.T) {
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

	Opts = Settings{
		ConfigPath: "",
		NoBuiltins: true,
		Theme:      "tokyonight",
	}

	t.Run("TestConfigLoadDotLogalizeBadYAML", func(t *testing.T) {
		err := InitConfig(builtinsAllGood)
		if err == nil || err.Error() != "yaml: line 4: mapping values are not allowed in this context" {
			t.Errorf("InitConfig() should have failed with *errors.errorString, got: [%T] %s", err, err)
		}
	})

	err = os.Chmod(defaultConfig, 0200)
	if err != nil {
		t.Errorf("Wasn't able to change mode of %s: %s", defaultConfig, err)
	}

	t.Run("TestConfigLoadDotLogalizeReadOnly", func(t *testing.T) {
		err := InitConfig(builtinsAllGood)
		if _, ok := err.(*fs.PathError); !ok {
			t.Errorf("InitConfig() should have failed with *fs.PathError, got: [%T] %s", err, err)
		}
	})
}

func TestConfigTheme(t *testing.T) {
	colorProfile = termenv.TrueColor
	var builtins embed.FS

	configData := `
themes:
  test: {}
`

	testConfig := t.TempDir() + "/testConfig.yaml"
	configRaw := []byte(configData)
	err := os.WriteFile(testConfig, configRaw, 0644)
	if err != nil {
		t.Errorf("Wasn't able to write test file to %s: %s", testConfig, err)
	}

	// good
	Opts = Settings{
		ConfigPath: testConfig,
		Theme:      "test",
	}

	t.Run("TestConfigThemeIsDefined", func(t *testing.T) {
		err := InitConfig(builtins)
		if err != nil {
			t.Errorf("InitConfig() failed with this error: %s", err)
		}
	})

	// bad
	Opts = Settings{
		ConfigPath: testConfig,
		Theme:      "idontexist",
	}

	t.Run("TestConfigThemeIsNotDefined", func(t *testing.T) {
		err := InitConfig(builtins)
		if err == nil || err.Error() != "Theme \"idontexist\" is not defined. Use -T/--list-themes flag to see the list of all available themes" {
			t.Errorf("InitConfig() should have failed with \"Theme \"idontexist\" is not defined. Use -T/--list-themes flag to see the list of all available themes\", got: [%T] %s", err, err)
		}
	})
}
