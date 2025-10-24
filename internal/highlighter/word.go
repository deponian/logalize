package highlighter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	"github.com/knadh/koanf/v2"
)

type wordGroup struct {
	Name       string
	List       []string
	Foreground string
	Background string
	Style      string
}

type wordGroups struct {
	Good       wordGroup
	Bad        wordGroup
	Other      []wordGroup
	Lemmatizer *golem.Lemmatizer
}

// newWords initializes global list of words collected
// from *koanf.Koanf configuration
func newWords(config *koanf.Koanf, theme string) (wordGroups, error) {
	if config == nil {
		return wordGroups{}, nil
	}

	var words wordGroups

	for _, wordGroupName := range config.MapKeys("words") {
		var wordGroup wordGroup

		wordGroup.Name = wordGroupName

		path := "themes." + theme + ".words." + wordGroupName + "."
		wordGroup.Foreground = config.String(path + "fg")
		wordGroup.Background = config.String(path + "bg")
		wordGroup.Style = config.String(path + "style")

		if err := config.Unmarshal("words."+wordGroupName, &wordGroup.List); err != nil {
			return wordGroups{}, err
		}

		if err := wordGroup.check(); err != nil {
			return wordGroups{}, err
		}

		switch wordGroupName {
		case "good":
			words.Good = wordGroup
		case "bad":
			words.Bad = wordGroup
		default:
			words.Other = append(words.Other, wordGroup)
		}
	}

	words.Lemmatizer, _ = golem.New(en.New())

	return words, nil
}

// highlight colors all words in a string.
// It doesn't touch already colored parts of the input.
func (words wordGroups) highlight(str string, h Highlighter) string {
	return walkNonSGR(str, func(part string) string {
		if part == "" {
			return part
		}

		if m := negatedWordRegexp.FindStringSubmatchIndex(part); m != nil {
			leftPart := words.highlight(part[0:m[0]], h)
			match := words.highlightNegatedWord(part[m[0]:m[1]], part[m[2]:m[3]], part[m[4]:m[5]], h)
			rightPart := words.highlight(part[m[1]:], h)

			return leftPart + match + rightPart
		}

		if m := wordRegexp.FindStringIndex(part); m != nil {
			leftPart := words.highlight(part[0:m[0]], h)
			match := words.highlightWord(part[m[0]:m[1]], h)
			rightPart := words.highlight(part[m[1]:], h)

			return leftPart + match + rightPart
		}

		return part
	})
}

// highlightWord colors single word in a string
func (words wordGroups) highlightWord(word string, h Highlighter) string {
	// search in all word groups
	for _, wordGroup := range append(words.Other, words.Good, words.Bad) {
		lemma := words.Lemmatizer.Lemma(word)
		if slices.Contains(wordGroup.List, lemma) ||
			slices.Contains(wordGroup.List, word) ||
			slices.Contains(wordGroup.List, strings.ToLower(word)) {
			word = h.highlight(word, wordGroup.Foreground, wordGroup.Background, wordGroup.Style)
			if h.settings.Opts.Debug {
				word = h.addDebugInfo(word, wordGroup)
			}

			break
		}
	}

	return word
}

// highlightNegated colors a phrase with negated word in a string
// if the word is good, then color the whole phrase as bad and vice versa
// if the word is neither good nor bad, then don't color the phrase
func (words wordGroups) highlightNegatedWord(phrase, negator, word string, h Highlighter) string {
	lemma := words.Lemmatizer.Lemma(word)
	// good
	if slices.Contains(words.Good.List, lemma) ||
		slices.Contains(words.Good.List, word) ||
		slices.Contains(words.Good.List, strings.ToLower(word)) {
		phrase = h.highlight(phrase, words.Bad.Foreground, words.Bad.Background, words.Bad.Style)
		if h.settings.Opts.Debug {
			phrase = h.addDebugInfo(phrase, words.Good)
		}

		return phrase
	}
	// bad
	if slices.Contains(words.Bad.List, lemma) ||
		slices.Contains(words.Bad.List, word) ||
		slices.Contains(words.Bad.List, strings.ToLower(word)) {
		phrase = h.highlight(phrase, words.Good.Foreground, words.Good.Background, words.Good.Style)
		if h.settings.Opts.Debug {
			phrase = h.addDebugInfo(phrase, words.Bad)
		}

		return phrase
	}
	// other
	for _, wordGroup := range words.Other {
		if slices.Contains(wordGroup.List, lemma) ||
			slices.Contains(wordGroup.List, word) ||
			slices.Contains(wordGroup.List, strings.ToLower(word)) {
			word = h.highlight(word, wordGroup.Foreground, wordGroup.Background, wordGroup.Style)
			if h.settings.Opts.Debug {
				word = h.addDebugInfo(word, wordGroup)
			}

			return negator + " " + word
		}
	}

	return phrase
}

func (wg wordGroup) check() error {
	// check foreground
	if !colorRegexp.MatchString(wg.Foreground) {
		return fmt.Errorf(
			"[word group: %s] foreground color %s doesn't match %s pattern",
			wg.Name, wg.Foreground, colorRegexp,
		)
	}

	// check background
	if !colorRegexp.MatchString(wg.Background) {
		return fmt.Errorf(
			"[word group: %s] background color %s doesn't match %s pattern",
			wg.Name, wg.Background, colorRegexp,
		)
	}

	// check style
	if !nonRecursiveStyleRegexp.MatchString(wg.Style) {
		return fmt.Errorf(
			"[word group: %s] style %s doesn't match %s pattern",
			wg.Name, wg.Style, nonRecursiveStyleRegexp,
		)
	}

	return nil
}
