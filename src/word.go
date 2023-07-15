package logalize

import (
	"slices"

	"github.com/aaaton/golem/v4"
	"github.com/knadh/koanf/v2"
)

type WordGroup struct {
	Name       string   `koanf:"name"`
	List       []string `koanf:"list"`
	Foreground string   `koanf:"fg"`
	Background string   `koanf:"bg"`
	Style      string   `koanf:"style"`
}

type Words struct {
	Good       WordGroup
	Bad        WordGroup
	Other      []WordGroup
	Lemmatizer *golem.Lemmatizer
}

// InitWords returns global list of words collected
// from *koanf.Koanf configuration
func initWords(config *koanf.Koanf, lemmatizer *golem.Lemmatizer) (Words, error) {
	var words Words

	for _, wordGroupName := range config.MapKeys("words") {
		var wordGroup WordGroup
		wordGroup.Name = wordGroupName
		if err := config.Unmarshal("words."+wordGroupName, &wordGroup); err != nil {
			return Words{}, err
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

	words.Lemmatizer = lemmatizer

	return words, nil
}

// highlightWord colors single word in a string
func (words *Words) highlightWord(word string) string {
	allWordGroups := append(words.Other, words.Good, words.Bad)
	for _, wordGroup := range allWordGroups {
		lemma := words.Lemmatizer.Lemma(word)
		if slices.Contains(wordGroup.List, lemma) || slices.Contains(wordGroup.List, word) {
			word = highlight(word, wordGroup.Foreground, wordGroup.Background, wordGroup.Style)
			break
		}
	}

	return word
}

// highlightNegated colors a phrase with negated word in a string
// if the word is good, then color the whole phrase as bad and vice versa
// if the word is neither good nor bad, then don't color the phrase
func (words *Words) highlightNegatedWord(phrase, negator, word string) string {
	lemma := words.Lemmatizer.Lemma(word)
	// good
	if slices.Contains(words.Good.List, lemma) || slices.Contains(words.Good.List, word) {
		return highlight(phrase, words.Bad.Foreground, words.Bad.Background, words.Bad.Style)
	}
	// bad
	if slices.Contains(words.Bad.List, lemma) || slices.Contains(words.Bad.List, word) {
		return highlight(phrase, words.Good.Foreground, words.Good.Background, words.Good.Style)
	}
	// other
	for _, wordGroup := range words.Other {
		if slices.Contains(wordGroup.List, lemma) || slices.Contains(wordGroup.List, word) {
			return negator + " " + highlight(word, wordGroup.Foreground, wordGroup.Background, wordGroup.Style)
		}
	}

	return phrase
}

// highlight colors all words in a string
func (words *Words) highlight(str string) string {
	if str == "" {
		return str
	}

	for {
		if m := negatedWordRegexp.FindStringSubmatchIndex(str); m != nil {
			leftPart := words.highlight(str[0:m[0]])
			match := words.highlightNegatedWord(str[m[0]:m[1]], str[m[2]:m[3]], str[m[4]:m[5]])
			rightPart := words.highlight(str[m[1]:])
			return leftPart + match + rightPart
		} else if m := wordRegexp.FindStringIndex(str); m != nil {
			leftPart := words.highlight(str[0:m[0]])
			match := words.highlightWord(str[m[0]:m[1]])
			rightPart := words.highlight(str[m[1]:])
			return leftPart + match + rightPart
		} else {
			return str
		}
	}
}
