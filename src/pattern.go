package logalize

import (
	"fmt"
	"sort"
)

// Pattern represents a pattern
type Pattern struct {
	Name      string
	Priority  int
	CapGroups *CapGroupList
}

// PatternList represents a list of pattern
type PatternList []Pattern

var Patterns PatternList

// InitPatterns initializes global list of patterns collected
// from *koanf.Koanf configuration
func initPatterns() error {
	Patterns = PatternList{}

	// collect list of patterns
	for _, patternName := range Config.MapKeys("patterns") {
		var pattern Pattern
		pattern.Name = patternName
		pattern.Priority = Config.Int("patterns." + patternName + ".priority")
		pattern.CapGroups = &CapGroupList{}
		if Config.Exists("patterns." + patternName + ".regexps") {
			if err := Config.Unmarshal("patterns."+patternName+".regexps", &pattern.CapGroups.Groups); err != nil {
				return err
			}
		} else {
			if err := Config.Unmarshal("patterns."+patternName, &pattern.CapGroups.Groups); err != nil {
				return err
			}
		}
		Patterns = append(Patterns, pattern)
	}

	// init patterns
	for _, pattern := range Patterns {
		// set colors and style from the theme
		for i, cg := range pattern.CapGroups.Groups {
			cgReal := &pattern.CapGroups.Groups[i]
			// simple CapGroupLists don't have a name (see "uuid" pattern)
			// so we need a second level of nesting only for the complex ones (those with "regexps" field)
			var path string
			if Config.Exists("patterns." + pattern.Name + ".regexps") {
				path = "themes." + Opts.Theme + ".patterns." + pattern.Name + "." + cg.Name
			} else {
				path = "themes." + Opts.Theme + ".patterns." + pattern.Name
			}
			if len(cg.Alternatives) > 0 {
				cgReal.Foreground = Config.String(path + ".default.fg")
				cgReal.Background = Config.String(path + ".default.bg")
				cgReal.Style = Config.String(path + "default.style")

				for j, alt := range cg.Alternatives {
					altReal := &pattern.CapGroups.Groups[i].Alternatives[j]
					altReal.Foreground = Config.String(path + "." + alt.Name + ".fg")
					altReal.Background = Config.String(path + "." + alt.Name + ".bg")
					altReal.Style = Config.String(path + "." + alt.Name + ".style")
				}
			} else {
				cgReal.Foreground = Config.String(path + ".fg")
				cgReal.Background = Config.String(path + ".bg")
				cgReal.Style = Config.String(path + ".style")
			}
		}

		// init capture groups
		if err := pattern.CapGroups.init(false); err != nil {
			return fmt.Errorf("[pattern: %s] %s", pattern.Name, err)
		}
	}

	// sort by priority
	sort.Slice(Patterns, func(i, j int) bool {
		iv, jv := Patterns[i], Patterns[j]
		return iv.Priority > jv.Priority
	})

	return nil
}

// highlight colorizes various patterns
// like IP address, date, HTTP response code and (optionally) special words
func (patterns PatternList) highlight(str string, highlightWords bool) string {
	if str == "" {
		return str
	}

	// patterns
	for _, pattern := range patterns {
		matches := pattern.CapGroups.FullRegExp.FindStringSubmatchIndex(str)
		if matches != nil {
			leftPart := patterns.highlight(str[0:matches[0]], highlightWords)
			match := pattern.CapGroups.highlight(str[matches[0]:matches[1]])
			rightPart := patterns.highlight(str[matches[1]:], highlightWords)
			return leftPart + match + rightPart
		}
	}

	// words
	if highlightWords {
		return Words.highlight(str)
	} else {
		// at this point we know that str doesn't contain any patterns and
		// we don't want to highlight words, so we can apply default color here
		defaultColor := Config.StringMap("themes." + Opts.Theme + ".default")
		return highlight(str, defaultColor["fg"], defaultColor["bg"], defaultColor["style"])
	}
}
