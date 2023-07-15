package logalize

import (
	"fmt"

	arg "github.com/alexflint/go-arg"
)

// Options stores the values of command-line options
type Options struct {
	ConfigPath string `arg:"-c, --config" help:"path to configuration file"`
	NoBuiltins bool   `arg:"-n, --no-builtins" help:"disable built-in log formats and words"`
}

func (Options) Version() string {
	return fmt.Sprintf("%s (%s)", logalizeVersion, logalizeReleaseDate)
}

// ParseOptions parses command-line options
func ParseOptions(args []string, opts interface{}) (*arg.Parser, error) {
	parser, err := arg.NewParser(arg.Config{}, opts)
	if err != nil {
		return parser, err
	}
	err = parser.Parse(args)
	return parser, err
}
