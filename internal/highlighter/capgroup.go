package highlighter

import (
	"fmt"
	"regexp"
	"strconv"
)

// capGroup represents one capturing group in a config file
type capGroup struct {
	Name         string `koanf:"name"`
	RegExpStr    string `koanf:"regexp"`
	Foreground   string
	Background   string
	Style        string
	Alternatives []capGroup     `koanf:"alternatives"`
	RegExp       *regexp.Regexp `koanf:"-"`
}

// capGroupList represents a list of capturing groups
// that will be parsed as one big regular expression
type capGroupList struct {
	Groups     []capGroup
	FullRegExp *regexp.Regexp
}

func (cgl *capGroupList) init(isFormat bool) error {
	for _, group := range cgl.Groups {
		// check that all regexps are valid regular expressions
		if err := group.check(); err != nil {
			return err
		}
	}

	// build regexp for the whole list
	var format string
	for i, cg := range cgl.Groups {
		// add name for the capturing group
		format += fmt.Sprintf("(?P<capGroup%d>(?:%s))", i, cg.RegExpStr[1:len(cg.RegExpStr)-1])
	}
	if isFormat {
		format = "^" + format + "$"
	}
	cgl.FullRegExp = regexp.MustCompile(format)

	// build regexps for capturing groups' alternatives
	for i, cg := range cgl.Groups {
		if len(cg.Alternatives) > 0 {
			for j, alt := range cg.Alternatives {
				cgl.Groups[i].Alternatives[j].RegExp = regexp.MustCompile(alt.RegExpStr)
			}
		}
	}

	return nil
}

func (cgl *capGroupList) highlight(str string, h Highlighter) (coloredStr string) {
	matches := cgl.FullRegExp.FindStringSubmatch(str)
	for i, cg := range cgl.Groups {
		match := matches[cgl.FullRegExp.SubexpIndex("capGroup"+strconv.Itoa(i))]
		coloredStr += cg.highlight(match, h)
	}

	return coloredStr
}

// highlight colorizes string and applies a style
func (cg *capGroup) highlight(str string, h Highlighter) string {
	if len(cg.Alternatives) > 0 {
		for _, alt := range cg.Alternatives {
			if alt.RegExp.MatchString(str) {
				return h.highlight(str, alt.Foreground, alt.Background, alt.Style)
			}
		}
	}

	return h.highlight(str, cg.Foreground, cg.Background, cg.Style)
}

func (cgl *capGroupList) check() error {
	for _, cg := range cgl.Groups {
		if err := cg.check(); err != nil {
			return err
		}
	}

	return nil
}

// check checks one capturing group's fields match corresponding patterns
func (cg *capGroup) check() error {
	// check name
	if cg.Name == "" {
		return fmt.Errorf("capturing group can't have empty \"name\" field")
	}

	// check regexp
	if cg.RegExpStr == "" {
		return fmt.Errorf("[capturing group: %s] empty \"regexp\" field", cg.Name)
	}
	if !capGroupRegexp.MatchString(cg.RegExpStr) {
		return fmt.Errorf(
			"[capturing group: %s] regexp %s must start with ( and end with )",
			cg.Name, cg.RegExpStr)
	}
	if _, err := regexp.Compile(cg.RegExpStr[1 : len(cg.RegExpStr)-1]); err != nil {
		return fmt.Errorf(
			"[capturing group: %s] %s\nCheck that the \"regexp\" starts with an opening bracket ( and "+
				"ends with a paired closing bracket )\nThat is, your \"regexp\" must be "+
				"within one large capturing group and contain a valid regular expression",
			cg.Name,
			err)
	}

	// check foreground
	if !colorRegexp.MatchString(cg.Foreground) {
		return fmt.Errorf(
			"[capturing group: %s] foreground color %s doesn't match %s regexp",
			cg.Name, cg.Foreground, colorRegexp)
	}

	// check background
	if !colorRegexp.MatchString(cg.Background) {
		return fmt.Errorf(
			"[capturing group: %s] background color %s doesn't match %s regexp",
			cg.Name, cg.Background, colorRegexp)
	}

	// check style
	if !styleRegexp.MatchString(cg.Style) {
		return fmt.Errorf(
			"[capturing group: %s] style %s doesn't match %s regexp",
			cg.Name, cg.Style, styleRegexp)
	}

	// check alternatives
	if len(cg.Alternatives) > 0 {
		for _, alt := range cg.Alternatives {
			if err := alt.check(); err != nil {
				return fmt.Errorf("[capturing group: %s] %s", cg.Name, err)
			}
		}
	}

	return nil
}
