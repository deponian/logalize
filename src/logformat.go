package logalize

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

// representation of one capture group in a config file
type CapGroup struct {
	Pattern      string       `koanf:"pattern"`
	Foreground   string       `koanf:"fg"`
	Background   string       `koanf:"bg"`
	Style        string       `koanf:"style"`
	Alternatives CapGroupList `koanf:"alternatives"`
	Regexp       *regexp.Regexp
}

// representation of a list of capture groups
type CapGroupList []CapGroup

// representation of log format
type LogFormat struct {
	Name      string
	CapGroups CapGroupList
	Regexp    *regexp.Regexp
}

// InitLogFormats returns list of LogFormats collected
// from *koanf.Koanf configuration
func initLogFormats(config *koanf.Koanf) ([]LogFormat, error) {
	var logFormats []LogFormat

	for _, formatName := range config.MapKeys("formats") {
		var logFormat LogFormat
		logFormat.Name = formatName
		if err := config.Unmarshal("formats."+formatName, &logFormat.CapGroups); err != nil {
			return nil, err
		}
		logFormats = append(logFormats, logFormat)
	}

	for i, format := range logFormats {
		// check that all patterns are valid regular expressions
		if err := format.checkCapGroups(); err != nil {
			return nil, err
		}

		// build regexp for whole log format line
		logFormats[i].Regexp = format.buildRegexp()

		// build regexps for capture groups' alternatives
		for _, cg := range format.CapGroups {
			if len(cg.Alternatives) > 0 {
				for k, alt := range cg.Alternatives {
					cg.Alternatives[k].Regexp = regexp.MustCompile(alt.Pattern)
				}
			}
		}
	}
	return logFormats, nil
}

// highlight colorizes string and applies a style
func highlight(str, fg, bg, style string) string {
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
			if alt.Regexp.MatchString(str) {
				return highlight(str, alt.Foreground, alt.Background, alt.Style)
			}
		}
	}

	return highlight(str, cg.Foreground, cg.Background, cg.Style)
}

// check checks one capture group's fields match corresponding patterns
func (cg *CapGroup) check() error {
	// check pattern
	if cg.Pattern == "" {
		return fmt.Errorf("empty patterns are not allowed")
	}
	if !capGroupRegexp.MatchString(cg.Pattern) {
		return fmt.Errorf(
			"capture group pattern %s doesn't match %s pattern",
			cg.Pattern, capGroupRegexp)
	}

	// check foreground
	if !colorRegexp.MatchString(cg.Foreground) {
		return fmt.Errorf(
			"[capture group: %s] foreground color %s doesn't match %s pattern",
			cg.Pattern, cg.Foreground, colorRegexp)
	}

	// check background
	if !colorRegexp.MatchString(cg.Background) {
		return fmt.Errorf(
			"[capture group: %s] background color %s doesn't match %s pattern",
			cg.Pattern, cg.Background, colorRegexp)
	}

	// check style
	if !styleRegexp.MatchString(cg.Style) {
		return fmt.Errorf(
			"[capture group: %s] style %s doesn't match %s pattern",
			cg.Pattern, cg.Style, styleRegexp)
	}

	// check alternatives
	if len(cg.Alternatives) > 0 {
		return cg.Alternatives.check()
	}
	return nil
}

// check checks that capture groups' fields match corresponding patterns
func (cgl *CapGroupList) check() error {
	for _, cg := range *cgl {
		if err := cg.check(); err != nil {
			return err
		}
	}
	return nil
}

// checkCapGroups checks that all capture groups' fields match corresponding patterns
func (lf *LogFormat) checkCapGroups() error {
	if err := lf.CapGroups.check(); err != nil {
		return fmt.Errorf("[log format: %s] %s", lf.Name, err)
	}
	return nil
}

// buildRegexp builds full regexp string from the list of capture groups
func (lf *LogFormat) buildRegexp() (formatRegexp *regexp.Regexp) {
	var format string
	for i, cg := range lf.CapGroups {
		// add name for the capture group
		format += fmt.Sprintf("(?P<capGroup%d>", i) + cg.Pattern[1:]
	}
	format = "^" + format + "$"
	return regexp.MustCompile(format)
}

func (lf *LogFormat) highlight(str string) (coloredStr string) {
	matches := lf.Regexp.FindStringSubmatch(str)
	for i, cg := range lf.CapGroups {
		match := matches[lf.Regexp.SubexpIndex("capGroup"+strconv.Itoa(i))]
		coloredStr += cg.highlight(match)
	}
	return coloredStr
}
