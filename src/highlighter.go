package logalize

import (
	"fmt"
	"os"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/muesli/termenv"
)

type Highlighter struct {
	settings Settings

	colorProfile termenv.Profile

	formats  LogFormatList
	patterns PatternList
	words    WordGroups
}

func NewHighlighter(settings Settings) (Highlighter, error) {
	h := Highlighter{settings: settings}

	h.colorProfile = termenv.NewOutput(os.Stdout, termenv.WithUnsafe()).EnvColorProfile()

	lemmatizer, err := golem.New(en.New())
	if err != nil {
		return Highlighter{}, err
	}
	if h.words, err = initWords(settings.Config, lemmatizer); err != nil {
		return Highlighter{}, err
	}

	if h.formats, err = initLogFormats(settings.Config); err != nil {
		return Highlighter{}, err
	}

	if h.patterns, err = initPatterns(settings.Config); err != nil {
		return Highlighter{}, err
	}

	return h, nil
}

// colorize detects log formats, patterns and words in the input string
// and returns colored result string
func (h Highlighter) colorize(line string) string {
	// don't alter the input in any way if user set --dry-run flag
	if h.settings.Opts.DryRun {
		return line
	}

	// remove all ANSI escape sequences from the input by default
	if !h.settings.Opts.NoANSIEscapeSequencesStripping {
		line = allANSIEscapeSequencesRegexp.ReplaceAllString(line, "")
	}

	// try one of the log formats
	for _, logFormat := range h.formats {
		if logFormat.CapGroups.FullRegExp.MatchString(line) {
			return logFormat.highlight(line, h)
		}
	}

	// if log format wasn't detected highlight patterns and words
	// and then apply default color to the rest
	line = h.patterns.highlight(line, h)
	line = h.words.highlight(line, h)
	line = h.applyDefaultColor(line)
	return line
}

// highlight colorizes string and applies a style
func (h Highlighter) highlight(str, fg, bg, style string) string {
	if style == "patterns-and-words" {
		str = h.patterns.highlight(str, h)
		str = h.words.highlight(str, h)
		str = h.applyDefaultColor(str)
		return str
	}
	if style == "patterns" {
		str = h.patterns.highlight(str, h)
		str = h.applyDefaultColor(str)
		return str
	}
	if style == "words" {
		str = h.words.highlight(str, h)
		str = h.applyDefaultColor(str)
		return str
	}

	coloredStr := termenv.String(str)
	if fg != "" {
		coloredStr = coloredStr.Foreground(h.colorProfile.Color(fg))
	}
	if bg != "" {
		coloredStr = coloredStr.Background(h.colorProfile.Color(bg))
	}
	switch style {
	case "bold":
		coloredStr = coloredStr.Bold()
	case "faint":
		coloredStr = coloredStr.Faint()
	case "italic":
		coloredStr = coloredStr.Italic()
	case "underline":
		coloredStr = coloredStr.Underline()
	case "overline":
		coloredStr = coloredStr.Overline()
	case "crossout":
		coloredStr = coloredStr.CrossOut()
	case "reverse":
		coloredStr = coloredStr.Reverse()
	}
	return coloredStr.String()
}

func (h Highlighter) applyDefaultColor(str string) string {
	if str == "" {
		return str
	}

	// skip already colored parts of the string
	matches := sgrANSIEscapeSequenceRegexp.FindStringSubmatchIndex(str)
	if matches != nil {
		leftPart := h.applyDefaultColor(str[0:matches[0]])
		alreadyColored := str[matches[0]:matches[1]]
		rightPart := h.applyDefaultColor(str[matches[1]:])
		return leftPart + alreadyColored + rightPart
	}

	defaultColor := h.settings.Config.StringMap("themes.default")
	return h.highlight(str, defaultColor["fg"], defaultColor["bg"], defaultColor["style"])
}

func (h Highlighter) addDebugInfo(str string, kind any) string {
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

	opening = h.highlight(opening, "", "", "reverse")
	closing = h.highlight(closing, "", "", "reverse")

	return opening + str + closing
}
