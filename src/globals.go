package logalize

import (
	"fmt"
	"os"
	"regexp"

	"github.com/muesli/termenv"
)

var (
	// values from configuration files will be checked using these regular expressions
	capGroupRegexp          = regexp.MustCompile(`^\(.+\)$`)
	colorRegexp             = regexp.MustCompile(`^(#[[:xdigit:]]{6}|[[:digit:]]{1,3})?$`)
	styleRegexp             = regexp.MustCompile(`^(bold|faint|italic|underline|overline|crossout|reverse|words|patterns|patterns-and-words)?$`)
	nonRecursiveStyleRegexp = regexp.MustCompile(`^(bold|faint|italic|underline|overline|crossout|reverse)?$`)

	// "words" will be deletected using these regular expressions
	wordRegexp        = regexp.MustCompile(`[A-Za-z]+`)
	negatedWordRegexp = regexp.MustCompile(`(` +
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
		`|[Nn]ot` +
		`)` +
		// negated word itself
		` ([A-Za-z]+)`,
	)

	// based on https://github.com/chalk/ansi-regex
	// with the addition of ":"-separated colors like '\x1B[38:5:185mTEST\e[0m'
	// match (or try to match) all ANSI escape sequences
	allANSIEscapeSequencesRegexp = regexp.MustCompile(`` +
		`[\x1B\x9B]` +
		`[[\]()#;?]*` +
		`(?:(?:(?:(?:;[-a-zA-Z\d\\/#&.:=?%@~_]+)*|[a-zA-Z\d]+(?:;[-a-zA-Z\d\\/#&.:=?%@~_]*)*)?(?:\x07|\x1B\x5C|\x9C))` +
		`|` +
		`(?:(?:\d{1,4}(?:[;:]\d{0,4})*)?[\dA-PR-TZcf-nq-uy=><~]))`,
	)

	// match only Select Graphic Rendition sequences
	sgrANSIEscapeSequenceRegexp = regexp.MustCompile(`` +
		// CSI (Control Sequence Introducer)
		`(?:\x1B\[|\x9B)` +
		// opening sequence (attributes that set color or text style)
		`\d{1,4}(?:[;:]\d{0,4})*m` +
		// text that will be displayed in according to attributes from the opening sequence above
		`.*?` +
		// closing sequence
		`(?:\x1B\[|\x9B)0?m`,
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

func applyDefaultColor(str string) string {
	if str == "" {
		return str
	}

	// skip already colored parts of the string
	matches := sgrANSIEscapeSequenceRegexp.FindStringSubmatchIndex(str)
	if matches != nil {
		leftPart := applyDefaultColor(str[0:matches[0]])
		alreadyColored := str[matches[0]:matches[1]]
		rightPart := applyDefaultColor(str[matches[1]:])
		return leftPart + alreadyColored + rightPart
	}

	defaultColor := Config.StringMap("themes." + Opts.Theme + ".default")
	return highlight(str, defaultColor["fg"], defaultColor["bg"], defaultColor["style"])
}

func addDebugInfo(str string, kind any) string {
	opening := ""
	closing := ""

	switch k := kind.(type) {
	case LogFormat:
		opening = fmt.Sprintf("[lf(%s)]", k.Name)
		closing = fmt.Sprintf("[lf(/%s)]", k.Name)
	case Pattern:
		opening = fmt.Sprintf("[p(%s)]", k.Name)
		closing = fmt.Sprintf("[p(/%s)]", k.Name)
	case WordGroup:
		opening = fmt.Sprintf("[w(%s)]", k.Name)
		closing = fmt.Sprintf("[w(/%s)]", k.Name)
	}

	opening = highlight(opening, "", "", "reverse")
	closing = highlight(closing, "", "", "reverse")

	return opening + str + closing
}
