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

// InitPatterns initializes global list of patterns collected
// from *koanf.Koanf configuration
func initPatterns(settings Settings) (PatternList, error) {
	var patterns PatternList

	// collect list of patterns
	for _, patternName := range settings.Config.MapKeys("patterns") {
		var pattern Pattern
		pattern.Name = patternName
		pattern.Priority = settings.Config.Int("patterns." + patternName + ".priority")
		pattern.CapGroups = &CapGroupList{}
		if settings.Config.Exists("patterns." + patternName + ".regexps") {
			if err := settings.Config.Unmarshal("patterns."+patternName+".regexps", &pattern.CapGroups.Groups); err != nil {
				return nil, err
			}
		} else {
			if err := settings.Config.Unmarshal("patterns."+patternName, &pattern.CapGroups.Groups); err != nil {
				return nil, err
			}
		}
		patterns = append(patterns, pattern)
	}

	// init patterns
	for _, pattern := range patterns {
		// set colors and style from the theme
		for i, cg := range pattern.CapGroups.Groups {
			cgReal := &pattern.CapGroups.Groups[i]
			// simple CapGroupLists don't have a name (see "uuid" pattern)
			// so we need a second level of nesting only for the complex ones (those with "regexps" field)
			var path string
			if settings.Config.Exists("patterns." + pattern.Name + ".regexps") {
				path = "themes." + settings.Opts.Theme + ".patterns." + pattern.Name + "." + cg.Name
			} else {
				path = "themes." + settings.Opts.Theme + ".patterns." + pattern.Name
			}
			if len(cg.Alternatives) > 0 {
				cgReal.Foreground = settings.Config.String(path + ".default.fg")
				cgReal.Background = settings.Config.String(path + ".default.bg")
				cgReal.Style = settings.Config.String(path + "default.style")

				for j, alt := range cg.Alternatives {
					altReal := &pattern.CapGroups.Groups[i].Alternatives[j]
					altReal.Foreground = settings.Config.String(path + "." + alt.Name + ".fg")
					altReal.Background = settings.Config.String(path + "." + alt.Name + ".bg")
					altReal.Style = settings.Config.String(path + "." + alt.Name + ".style")
				}
			} else {
				cgReal.Foreground = settings.Config.String(path + ".fg")
				cgReal.Background = settings.Config.String(path + ".bg")
				cgReal.Style = settings.Config.String(path + ".style")
			}
		}

		// init capture groups
		if err := pattern.CapGroups.init(false); err != nil {
			return nil, fmt.Errorf("[pattern: %s] %s", pattern.Name, err)
		}
	}

	// sort by priority
	sort.Slice(patterns, func(i, j int) bool {
		iv, jv := patterns[i], patterns[j]
		return iv.Priority > jv.Priority
	})

	return patterns, nil
}

// highlight colorizes various patterns
// like IP address, date, HTTP response code, etc.
func (patterns PatternList) highlight(str string, h Highlighter) string {
	if str == "" {
		return str
	}

	// skip already colored parts of the string
	matches := sgrANSIEscapeSequenceRegexp.FindStringSubmatchIndex(str)
	if matches != nil {
		leftPart := patterns.highlight(str[0:matches[0]], h)
		alreadyColored := str[matches[0]:matches[1]]
		rightPart := patterns.highlight(str[matches[1]:], h)
		return leftPart + alreadyColored + rightPart
	}

	// color patterns
	for _, pattern := range patterns {
		matches := pattern.CapGroups.FullRegExp.FindStringSubmatchIndex(str)
		if matches != nil {
			leftPart := patterns.highlight(str[0:matches[0]], h)
			match := pattern.CapGroups.highlight(str[matches[0]:matches[1]], h)
			rightPart := patterns.highlight(str[matches[1]:], h)
			if h.settings.Opts.Debug {
				match = h.addDebugInfo(match, pattern)
			}
			return leftPart + match + rightPart
		}
	}

	return str
}
