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

// InitPatterns returns global list of patterns collected
// from *koanf.Koanf configuration
func initPatterns(config *koanf.Koanf) ([]Pattern, error) {
	patterns := []Pattern{}

	for _, patternName := range config.MapKeys("patterns") {
		var pattern Pattern
		pattern.Name = patternName
		pattern.Priority = config.Int("patterns." + patternName + ".priority")
		if err := config.Unmarshal("patterns."+patternName, &pattern.CapGroup); err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}

	for _, pattern := range patterns {
		// validate patterns' capture groups
		if err := pattern.CapGroup.check(); err != nil {
			return nil, fmt.Errorf("[pattern: %s] %s", pattern.Name, err)
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
	sort.Slice(patterns, func(i, j int) bool {
		iv, jv := patterns[i], patterns[j]
		return iv.Priority > jv.Priority
	})
	return patterns, nil
}

// HighlightPatternsAndWords colorizes various patterns
// like IP address, date, HTTP response code and special words
func highlightPatternsAndWords(str string, patterns []Pattern, words Words) string {
	if str == "" {
		return str
	}

	// patterns
	for _, pattern := range patterns {
		matches := pattern.CapGroup.Regexp.FindStringSubmatchIndex(str)
		if matches != nil {
			leftPart := highlightPatternsAndWords(str[0:matches[0]], patterns, words)
			match := pattern.CapGroup.highlight(str[matches[0]:matches[1]])
			rightPart := highlightPatternsAndWords(str[matches[1]:], patterns, words)
			return leftPart + match + rightPart
		}
	}

	// words
	return words.highlight(str)
}
