package highlighter

import (
	"fmt"
	"regexp"
	"strconv"
)

// capGroup represents one capturing group in a config file
type capGroup struct {
	Name      string `koanf:"name"`
	RegExpStr string `koanf:"regexp"`

	Foreground string
	Background string
	Style      string

	// LinkTo makes this group inherit the color/style from another group
	// (the target group is referenced by its Name and must exist in the same list)
	LinkTo string

	Alternatives []capGroup `koanf:"alternatives"`

	RegExp *regexp.Regexp `koanf:"-"`
}

// capGroupList represents a list of capturing groups
// that will be parsed as one big regular expression
type capGroupList struct {
	groups     []capGroup
	fullRegExp *regexp.Regexp
	// index maps group names to their index in Groups for quick lookup
	index map[string]int
}

func (cgl *capGroupList) init(isFormat bool) error {
	// check that all regexps are valid regular expressions
	if err := cgl.validate(); err != nil {
		return err
	}

	// build regexp for the whole list
	var fullRegExp string
	for i, cg := range cgl.groups {
		// add name for the capturing group
		fullRegExp += fmt.Sprintf("(?P<capGroup%d>(?:%s))", i, cg.RegExpStr[1:len(cg.RegExpStr)-1])
	}
	if isFormat {
		fullRegExp = "^" + fullRegExp + "$"
	}
	cgl.fullRegExp = regexp.MustCompile(fullRegExp)

	// build regexps for capturing groups' alternatives
	for i, cg := range cgl.groups {
		if len(cg.Alternatives) > 0 {
			for j, alt := range cg.Alternatives {
				cgl.groups[i].Alternatives[j].RegExp = regexp.MustCompile(alt.RegExpStr)
			}
		}
	}

	if err := cgl.validateLinkTo(); err != nil {
		return err
	}

	return nil
}

// validateLinkTo validates link targets & cycles
func (cgl *capGroupList) validateLinkTo() error {
	// build name to index lookup
	cgl.index = make(map[string]int)
	for i, cg := range cgl.groups {
		cgl.index[cg.Name] = i
	}

	for _, cg := range cgl.groups {
		if cg.LinkTo == "" {
			continue
		}
		// cycle detection (A->B->C->A)
		seen := map[string]bool{}
		for cg.LinkTo != "" {
			if seen[cg.Name] {
				return fmt.Errorf("[capturing group: %s] cyclic link-to detected", cg.Name)
			}
			seen[cg.Name] = true
			idx, ok := cgl.index[cg.LinkTo]
			if !ok {
				return fmt.Errorf("[capturing group: %s] link-to %q refers to unknown capturing group", cg.Name, cg.LinkTo)
			}
			cg = cgl.groups[idx]
		}
	}

	return nil
}

func (cgl *capGroupList) highlight(str string, h Highlighter) (coloredStr string) {
	matches := cgl.fullRegExp.FindStringSubmatch(str)
	for i, cg := range cgl.groups {
		match := matches[cgl.fullRegExp.SubexpIndex("capGroup"+strconv.Itoa(i))]

		// If this group links to another, borrow that group's effective style.
		if fg, bg, style, ok := cgl.linkedStyle(matches, cg); ok {
			coloredStr += h.highlight(match, fg, bg, style)

			continue
		}
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

// linkedStyle returns the effective style of the target group if cg.LinkTo is set:
// - if the target has matching alternatives, return that alt's fg/bg/style
// - else return the target's default fg/bg/style
func (cgl *capGroupList) linkedStyle(matches []string, cg capGroup) (fg, bg, style string, ok bool) {
	if cg.LinkTo == "" || cgl.index == nil {
		return "", "", "", false
	}
	// Follow chains (A->B->C). Stop on first non-link group
	// Don't worry about cycle, they are already handled by validateLinkTo()
	curIdx := cgl.index[cg.LinkTo]
	for {

		target := cgl.groups[curIdx]
		// hop if chained link
		if target.LinkTo != "" {
			curIdx = cgl.index[target.LinkTo]

			continue
		}

		// compute effective style of the terminal target
		match := matches[cgl.fullRegExp.SubexpIndex("capGroup"+strconv.Itoa(curIdx))]

		if len(target.Alternatives) > 0 {
			for _, alt := range target.Alternatives {
				if alt.RegExp.MatchString(match) {
					return alt.Foreground, alt.Background, alt.Style, true
				}
			}
		}

		return target.Foreground, target.Background, target.Style, true
	}
}

func (cgl *capGroupList) validate() error {
	for _, cg := range cgl.groups {
		if err := cg.validate(); err != nil {
			return err
		}
	}

	// check that capgroup names are unique
	seen := make(map[string]bool, len(cgl.groups))
	for _, cg := range cgl.groups {
		if seen[cg.Name] {
			return fmt.Errorf("[capturing group: %s] capturing group names must be unique", cg.Name)
		}
		seen[cg.Name] = true
	}

	return nil
}

// validate checks one capturing group's fields match corresponding patterns
func (cg *capGroup) validate() error {
	// check name
	if cg.Name == "" {
		return fmt.Errorf("capturing group can't have empty \"name\" field")
	}
	if keywordRegExp.MatchString(cg.Name) {
		return fmt.Errorf(
			"[capturing group: %s] capturing group cannot be named \"fg\", \"bg\", \"style\", or \"link-to\"",
			cg.Name)
	}

	// check regexp
	if cg.RegExpStr == "" {
		return fmt.Errorf("[capturing group: %s] empty \"regexp\" field", cg.Name)
	}
	if !capGroupRegExp.MatchString(cg.RegExpStr) {
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
	if !colorRegExp.MatchString(cg.Foreground) {
		return fmt.Errorf(
			"[capturing group: %s] foreground color %s doesn't match %s regexp",
			cg.Name, cg.Foreground, colorRegExp)
	}

	// check background
	if !colorRegExp.MatchString(cg.Background) {
		return fmt.Errorf(
			"[capturing group: %s] background color %s doesn't match %s regexp",
			cg.Name, cg.Background, colorRegExp)
	}

	// check style
	if !styleRegExp.MatchString(cg.Style) {
		return fmt.Errorf(
			"[capturing group: %s] style %s doesn't match %s regexp",
			cg.Name, cg.Style, styleRegExp)
	}

	// check alternatives
	if len(cg.Alternatives) > 0 {
		for _, alt := range cg.Alternatives {
			if err := alt.validate(); err != nil {
				return fmt.Errorf("[capturing group: %s] %s", cg.Name, err)
			}
		}
	}

	return nil
}
