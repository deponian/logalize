package logalize

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/muesli/termenv"
)

// CapGroup represents one capture group in a config file
type CapGroup struct {
	RegExpStr    string     `koanf:"regexp"`
	Foreground   string     `koanf:"fg"`
	Background   string     `koanf:"bg"`
	Style        string     `koanf:"style"`
	Alternatives []CapGroup `koanf:"alternatives"`
	RegExp       *regexp.Regexp
}

// CapGroupList represents a list of capture groups
// that will be parsed as one big regular expression
type CapGroupList struct {
	Groups     []CapGroup
	FullRegExp *regexp.Regexp
}

func (cgl *CapGroupList) init(entireLineRegExp bool) error {
	for _, group := range cgl.Groups {
		// check that all regexps are valid regular expressions
		if err := group.check(); err != nil {
			return err
		}

		// build regexp for whole list
		var format string
		for i, cg := range cgl.Groups {
			// add name for the capture group
			format += fmt.Sprintf("(?P<capGroup%d>(?:%s))", i, cg.RegExpStr[1:len(cg.RegExpStr)-1])
		}
		if entireLineRegExp {
			format = "^" + format + "$"
		}
		cgl.FullRegExp = regexp.MustCompile(format)

		// build regexps for capture groups' alternatives
		for _, cg := range cgl.Groups {
			if len(cg.Alternatives) > 0 {
				for i, alt := range cg.Alternatives {
					cg.Alternatives[i].RegExp = regexp.MustCompile(alt.RegExpStr)
				}
			}
		}
	}
	return nil
}

// highlight colorizes string and applies a style
func highlight(str, fg, bg, style string) string {
	if style == "patterns-and-words" {
		return Patterns.highlight(str, true)
	}
	if style == "patterns" {
		return Patterns.highlight(str, false)
	}
	if style == "words" {
		return Words.highlight(str)
	}

	coloredStr := termenv.String(str)
	if fg != "" {
		coloredStr = coloredStr.Foreground(colorProfile.Color(fg))
	}
	if bg != "" {
		coloredStr = coloredStr.Background(colorProfile.Color(bg))
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

// highlight colorizes string and applies a style
func (cg *CapGroup) highlight(str string) string {
	if len(cg.Alternatives) > 0 {
		for _, alt := range cg.Alternatives {
			if alt.RegExp.MatchString(str) {
				return highlight(str, alt.Foreground, alt.Background, alt.Style)
			}
		}
	}

	return highlight(str, cg.Foreground, cg.Background, cg.Style)
}

// check checks one capture group's fields match corresponding patterns
func (cg *CapGroup) check() error {
	// check regexp
	if cg.RegExpStr == "" {
		return fmt.Errorf("empty regexps are not allowed")
	}
	if !capGroupRegexp.MatchString(cg.RegExpStr) {
		return fmt.Errorf(
			"[capture group: %s] regexp %s must start with ( and end with )",
			cg.RegExpStr, cg.RegExpStr)
	} else {
		if _, err := regexp.Compile(cg.RegExpStr[1 : len(cg.RegExpStr)-1]); err != nil {
			return fmt.Errorf(
				"%s\nCheck that the \"regexp\" starts with an opening bracket ( and ends with a paired closing bracket )\nThat is, your \"regexp\" must be within one large capture group and contain a valid regular expression",
				err)
		}
	}

	// check foreground
	if !colorRegexp.MatchString(cg.Foreground) {
		return fmt.Errorf(
			"[capture group: %s] foreground color %s doesn't match %s regexp",
			cg.RegExpStr, cg.Foreground, colorRegexp)
	}

	// check background
	if !colorRegexp.MatchString(cg.Background) {
		return fmt.Errorf(
			"[capture group: %s] background color %s doesn't match %s regexp",
			cg.RegExpStr, cg.Background, colorRegexp)
	}

	// check style
	if !styleRegexp.MatchString(cg.Style) {
		return fmt.Errorf(
			"[capture group: %s] style %s doesn't match %s regexp",
			cg.RegExpStr, cg.Style, styleRegexp)
	}

	// check alternatives
	if len(cg.Alternatives) > 0 {
		for _, alt := range cg.Alternatives {
			if err := alt.check(); err != nil {
				return fmt.Errorf("[capture group: %s] %s", cg.RegExpStr, err)
			}
		}
	}
	return nil
}

func (cgl *CapGroupList) highlight(str string) (coloredStr string) {
	matches := cgl.FullRegExp.FindStringSubmatch(str)
	for i, cg := range cgl.Groups {
		match := matches[cgl.FullRegExp.SubexpIndex("capGroup"+strconv.Itoa(i))]
		coloredStr += cg.highlight(match)
	}
	return coloredStr
}

func (cgl *CapGroupList) check() error {
	for _, cg := range cgl.Groups {
		if err := cg.check(); err != nil {
			return err
		}
	}
	return nil
}
