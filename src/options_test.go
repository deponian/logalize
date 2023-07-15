package logalize

import (
	"strings"
	"testing"

	arg "github.com/alexflint/go-arg"
	"github.com/google/go-cmp/cmp"
)

func TestOptionsVersion(t *testing.T) {
	SetGlobals("0.0.0", "1970-01-01")
	options := Options{
		ConfigPath: "",
		NoBuiltins: true,
	}
	t.Run("TestOptionsVersion", func(t *testing.T) {
		if options.Version() != "0.0.0 (1970-01-01)" {
			t.Errorf("Options.Version() failed: got %v, want %v", options.Version(), "0.0.0 (test) 1970-01-01")
		}
	})
}

func TestOptionsParse(t *testing.T) {
	tests := []struct {
		args    []string
		options Options
	}{
		{[]string{}, Options{ConfigPath: "", NoBuiltins: false}},
		{[]string{"-c", "logalize.yaml"}, Options{ConfigPath: "logalize.yaml", NoBuiltins: false}},
		{[]string{"-n"}, Options{ConfigPath: "", NoBuiltins: true}},
		{[]string{"-c", "logalize.yaml", "-n"}, Options{ConfigPath: "logalize.yaml", NoBuiltins: true}},
		{[]string{"-n", "-c", "logalize.yaml"}, Options{ConfigPath: "logalize.yaml", NoBuiltins: true}},
	}
	for _, tt := range tests {
		testname := strings.Join(tt.args, "_")

		t.Run(testname, func(t *testing.T) {
			options := Options{}
			_, err := ParseOptions(tt.args, &options)
			if err != nil {
				t.Errorf("ParseOptions() failed with error: %v", err)
			}
			if !cmp.Equal(options, tt.options) {
				t.Errorf("got %v, want %v", options, tt.options)
			}
		})
	}

	t.Run("TestOptionsParseHelp", func(t *testing.T) {
		options := Options{}
		_, err := ParseOptions([]string{"-h"}, &options)
		if err != arg.ErrHelp {
			t.Errorf("ParseOptions() should have failed with error %v, got %v", arg.ErrHelp, err)
		}
	})

	t.Run("TestOptionsParseVersion", func(t *testing.T) {
		options := Options{}
		_, err := ParseOptions([]string{"--version"}, &options)
		if err != arg.ErrVersion {
			t.Errorf("ParseOptions() should have failed with error %v, got %v", arg.ErrVersion, err)
		}
	})

	t.Run("TestOptionsParseWrongOpts", func(t *testing.T) {
		options := Pattern{}
		_, err := ParseOptions([]string{}, &options)
		if err.Error() != "Pattern.CapGroup: *logalize.CapGroup fields are not supported" {
			t.Errorf("ParseOptions() should have failed with error *errors.errorString, got [%T] %v", err, err)
		}
	})
}
