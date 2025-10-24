// Package highlighter is responsible for everything related to colorization
package highlighter

import (
	"fmt"
	"strings"

	"github.com/deponian/logalize/internal/config"
	"github.com/muesli/termenv"
)

// Highlighter applies colorization.
//
// It detects formats, patterns and word groups defined in
// configuration, then renders the result using the terminal's color profile.
type Highlighter struct {
	settings config.Settings

	defaultFg    string
	defaultBg    string
	defaultStyle string

	formats  formatList
	patterns patternList
	words    wordGroups
}

// NewHighlighter creates a Highlighter configured from the provided settings.
//
// The constructor determines the effective terminal color profile and loads
// formats, patterns, and word groups for the selected theme. If the
// settings restrict highlighting to only certain categories (formats,
// patterns, or words), the other categories are initialized empty.
// It returns an error if configuration cannot be loaded or parsed.
func NewHighlighter(settings config.Settings) (Highlighter, error) {
	h := Highlighter{settings: settings}

	formats, err := newFormats(settings.Config, settings.Opts.Theme)
	if err != nil {
		return Highlighter{}, err
	}

	patterns, err := newPatterns(settings.Config, settings.Opts.Theme)
	if err != nil {
		return Highlighter{}, err
	}

	words, err := newWords(settings.Config, settings.Opts.Theme)
	if err != nil {
		return Highlighter{}, err
	}

	// keep in the highlighter only things we want to colorize
	if settings.Opts.HighlightOnlyFormats || settings.Opts.HighlightOnlyPatterns || settings.Opts.HighlightOnlyWords {
		// init with the empty config all the things we don't need
		if settings.Opts.HighlightOnlyFormats {
			h.formats = formats
		} else {
			h.formats, _ = newFormats(nil, "")
		}

		if settings.Opts.HighlightOnlyPatterns {
			h.patterns = patterns
		} else {
			h.patterns, _ = newPatterns(nil, "")
		}

		if settings.Opts.HighlightOnlyWords {
			h.words = words
		} else {
			h.words, _ = newWords(nil, "")
		}
	} else {
		h.formats = formats
		h.patterns = patterns
		h.words = words
	}

	// set default color
	if settings.Config != nil {
		defaultColor := settings.Config.StringMap("themes." + settings.Opts.Theme + ".default")
		h.defaultFg = defaultColor["fg"]
		h.defaultBg = defaultColor["bg"]
		h.defaultStyle = defaultColor["style"]
	}

	return h, nil
}

// Colorize detects formats, patterns and words in the input string
// and returns colored result string.
func (h Highlighter) Colorize(line string) string {
	// don't alter the input in any way if user set --dry-run flag
	if h.settings.Opts.DryRun {
		return line
	}

	// remove all ANSI escape sequences from the input by default
	if !h.settings.Opts.NoANSIEscapeSequencesStripping {
		line = allANSIEscapeSequencesRegexp.ReplaceAllString(line, "")
	}

	// try one of the formats
	for _, format := range h.formats {
		if format.match(line) {
			return format.highlight(line, h)
		}
	}

	// if format wasn't detected highlight patterns and words
	// and then apply default color to the rest
	line = h.patterns.highlight(line, h)
	line = h.words.highlight(line, h)
	line = h.applyDefaultColor(line)

	return line
}

// highlight colorizes string and applies a style.
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
		coloredStr = coloredStr.Foreground(h.settings.ColorProfile.Color(fg))
	}
	if bg != "" {
		coloredStr = coloredStr.Background(h.settings.ColorProfile.Color(bg))
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

// applyDefaultColor applies default color to all non-colored parts of the input.
func (h Highlighter) applyDefaultColor(str string) string {
	return walkNonSGR(str, func(part string) string {
		if part == "" {
			return part
		}

		return h.highlight(part, h.defaultFg, h.defaultBg, h.defaultStyle)
	})
}

func (h Highlighter) addDebugInfo(str string, kind any) string {
	opening := ""
	closing := ""

	switch k := kind.(type) {
	case format:
		opening = fmt.Sprintf("[f(%s)]", k.Name)
		closing = fmt.Sprintf("[f(/%s)]", k.Name)
	case pattern:
		opening = fmt.Sprintf("[p(%s)]", k.Name)
		closing = fmt.Sprintf("[p(/%s)]", k.Name)
	case wordGroup:
		opening = fmt.Sprintf("[w(%s)]", k.Name)
		closing = fmt.Sprintf("[w(/%s)]", k.Name)
	}

	opening = h.highlight(opening, "", "", "reverse")
	closing = h.highlight(closing, "", "", "reverse")

	return opening + str + closing
}

// walkNonSGR applies f to every non-colored part and keeps SGR segments untouched
func walkNonSGR(str string, f func(string) string) string {
	if str == "" {
		return str
	}
	out := strings.Builder{}
	for {
		loc := sgrANSIEscapeSequenceRegexp.FindStringIndex(str)
		if loc == nil {
			out.WriteString(f(str))

			return out.String()
		}
		// color left part
		out.WriteString(f(str[:loc[0]]))
		// copy already colored part verbatim
		out.WriteString(str[loc[0]:loc[1]])
		// process right part
		str = str[loc[1]:]
	}
}
