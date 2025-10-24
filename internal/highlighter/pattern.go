package highlighter

import (
	"fmt"
	"sort"

	"github.com/knadh/koanf/v2"
)

// pattern represents a pattern
type pattern struct {
	Name      string
	Priority  int
	CapGroups *capGroupList
}

// patternList represents a list of pattern
type patternList []pattern

// newPatterns initializes global list of patterns collected
// from *koanf.Koanf configuration using a theme from the second argument
func newPatterns(config *koanf.Koanf, theme string) (patternList, error) {
	if config == nil {
		return patternList{}, nil
	}

	patterns, err := collectPatterns(config)
	if err != nil {
		return nil, err
	}

	for i := range patterns {
		if err := initPattern(&patterns[i], config, theme); err != nil {
			return nil, err
		}
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Priority > patterns[j].Priority
	})

	return patterns, nil
}

func collectPatterns(config *koanf.Koanf) (patternList, error) {
	var patterns patternList

	// collect list of patterns
	for _, patternName := range config.MapKeys("patterns") {
		var pattern pattern
		pattern.Name = patternName
		pattern.Priority = config.Int("patterns." + patternName + ".priority")
		pattern.CapGroups = &capGroupList{}
		if config.Exists("patterns." + patternName + ".regexps") {
			if err := config.Unmarshal("patterns."+patternName+".regexps", &pattern.CapGroups.groups); err != nil {
				return nil, err
			}
		} else {
			if err := config.Unmarshal("patterns."+patternName, &pattern.CapGroups.groups); err != nil {
				return nil, err
			}
		}
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

func initPattern(p *pattern, config *koanf.Koanf, theme string) error {
	// set colors and style from the theme
	for i, cg := range p.CapGroups.groups {
		cgReal := &p.CapGroups.groups[i]
		// simple CapGroupLists don't have a name (see "uuid" pattern)
		// so we need a second level of nesting only for the complex ones (those with "regexps" field)
		var path string
		if config.Exists("patterns." + p.Name + ".regexps") {
			path = "themes." + theme + ".patterns." + p.Name + "." + cg.Name
		} else {
			path = "themes." + theme + ".patterns." + p.Name

			// if we have a pattern with one regexp then it should get its name
			cgReal.Name = p.Name
		}

		if len(cg.Alternatives) > 0 {
			cgReal.Foreground = config.String(path + ".default.fg")
			cgReal.Background = config.String(path + ".default.bg")
			cgReal.Style = config.String(path + ".default.style")

			for j, alt := range cg.Alternatives {
				altReal := &p.CapGroups.groups[i].Alternatives[j]
				altReal.Foreground = config.String(path + "." + alt.Name + ".fg")
				altReal.Background = config.String(path + "." + alt.Name + ".bg")
				altReal.Style = config.String(path + "." + alt.Name + ".style")
			}
		} else {
			cgReal.Foreground = config.String(path + ".fg")
			cgReal.Background = config.String(path + ".bg")
			cgReal.Style = config.String(path + ".style")
		}

		cgReal.LinkTo = config.String(path + ".link-to")
	}

	// init capturing groups
	if err := p.CapGroups.init(false); err != nil {
		return fmt.Errorf("[pattern: %s] %s", p.Name, err)
	}

	return nil
}

// highlight colorizes various patterns like IP address, date, HTTP response code, etc.
// It doesn't touch already colored parts of the input.
func (patterns patternList) highlight(str string, h Highlighter) string {
	return walkNonSGR(str, func(part string) string {
		if part == "" {
			return part
		}
		for _, pattern := range patterns {
			matches := pattern.CapGroups.fullRegExp.FindStringSubmatchIndex(part)
			if matches != nil {
				leftPart := patterns.highlight(part[0:matches[0]], h)
				match := pattern.CapGroups.highlight(part[matches[0]:matches[1]], h)
				rightPart := patterns.highlight(part[matches[1]:], h)
				if h.settings.Opts.Debug {
					match = h.addDebugInfo(match, pattern)
				}

				return leftPart + match + rightPart
			}
		}

		return part
	})
}
