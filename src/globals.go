package logalize

import (
	"os"
	"regexp"

	"github.com/muesli/termenv"
)

// values from configuration files will be checked using these regular expressions
var (
	capGroupRegexp          = regexp.MustCompile(`^\(.+\)$`)
	colorRegexp             = regexp.MustCompile(`^(#[[:xdigit:]]{6}|[[:digit:]]{1,3})?$`)
	styleRegexp             = regexp.MustCompile(`^(bold|faint|italic|underline|overline|crossout|reverse|words|patterns|patterns-and-words)?$`)
	nonRecursiveStyleRegexp = regexp.MustCompile(`^(bold|faint|italic|underline|overline|crossout|reverse)?$`)
	wordRegexp              = regexp.MustCompile(`[A-Za-z]+`)
	negatedWordRegexp       = regexp.MustCompile(`(` +
		// complex negation
		// can't be, shouldn't be, etc.
		`[A-Za-z]+n't be` +
		`|[Cc]annot be` +
		// should not be, will not be, etc.
		`|[A-Za-z]+ not be` +
		// simple negation
		// wasn't, aren't, won't, etc.
		`|[A-Za-z]+n't` +
		`|[Cc]annot` +
		// just plain "not something"
		`| not` +
		`|^not` +
		`|^Not` +
		`)` +
		// negated word itself
		` ([A-Za-z]+)`,
	)
)

// global color profile for all colorization
var colorProfile = termenv.NewOutput(os.Stdout, termenv.WithUnsafe()).EnvColorProfile()

// where to find default configuration files
var defaultConfigPaths = getDefaultConfigPaths()

func getDefaultConfigPaths() []string {
	homeDir, _ := os.UserHomeDir()
	return []string{
		"/etc/logalize/logalize.yaml",
		homeDir + "/.config/logalize/logalize.yaml",
	}
}
