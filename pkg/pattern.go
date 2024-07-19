package logalize

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/knadh/koanf/v2"
)

type Pattern struct {
	Name     string
	Priority int
	CapGroup *CapGroup
}

type PatternList []Pattern

var Patterns PatternList

// InitPatterns initializes global list of patterns collected
// from *koanf.Koanf configuration
func initPatterns(config *koanf.Koanf) error {
	Patterns = PatternList{}
	for _, patternName := range config.MapKeys("patterns") {
		var pattern Pattern
		pattern.Name = patternName
		pattern.Priority = config.Int("patterns." + patternName + ".priority")
		if err := config.Unmarshal("patterns."+patternName, &pattern.CapGroup); err != nil {
			return err
		}
		Patterns = append(Patterns, pattern)
	}

	for _, pattern := range Patterns {
		// validate patterns' capture groups
		if err := pattern.CapGroup.check(false); err != nil {
			return fmt.Errorf("[pattern: %s] %s", pattern.Name, err)
		}

		// build main regexp
		pattern.CapGroup.Regexp = regexp.MustCompile(pattern.CapGroup.Pattern)

		// build regexps for capture groups' alternatives
		if len(pattern.CapGroup.Alternatives) > 0 {
			for k, alt := range pattern.CapGroup.Alternatives {
				pattern.CapGroup.Alternatives[k].Regexp = regexp.MustCompile(alt.Pattern)
			}
		}
	}
	// sort by priority
	sort.Slice(Patterns, func(i, j int) bool {
		iv, jv := Patterns[i], Patterns[j]
		return iv.Priority > jv.Priority
	})

	return nil
}

// HighlightPatternsAndWords colorizes various patterns
// like IP address, date, HTTP response code and (optionally) special words
func (patterns PatternList) highlight(str string, highlightWords bool) string {
	if str == "" {
		return str
	}

	// patterns
	for _, pattern := range patterns {
		matches := pattern.CapGroup.Regexp.FindStringSubmatchIndex(str)
		if matches != nil {
			leftPart := patterns.highlight(str[0:matches[0]], highlightWords)
			match := pattern.CapGroup.highlight(str[matches[0]:matches[1]])
			rightPart := patterns.highlight(str[matches[1]:], highlightWords)
			return leftPart + match + rightPart
		}
	}

	// words
	if highlightWords {
		str = Words.highlight(str)
	}

	return str
}
