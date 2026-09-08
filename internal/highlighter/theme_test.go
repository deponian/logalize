package highlighter

import (
	"testing"

	"github.com/deponian/logalize/internal/config"
)

// deliberatelyUncolored lists the rule paths that every theme is expected to
// leave uncolored with the reason why.
var deliberatelyUncolored = map[string]string{
	"patterns.duration.start": "delimiter group; coloring it would tint the character in front of the duration",

	// A group whose alternatives cover every value its own regexp admits can
	// never fall back on its own color, so an entry for it would be dead config.
	"formats.klog.log-level":  `alternatives cover all of [IWEF]`,
	"formats.redis.role":      `alternatives cover all of [MSXC]`,
	"formats.redis.log-level": `alternatives cover all of [#*.-]`,
}

// TestThemesAreComplete checks that every shipped theme colors every capturing
// group, alternative and word group the built-in rules define.
func TestThemesAreComplete(t *testing.T) {
	settings, err := config.NewSettings(repoRoot(), nil, nil, true)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	config := settings.Config

	all, err := colorables(config)
	if err != nil {
		t.Fatalf("colorables(...) failed with this error: %s", err)
	}

	isColored := func(path string) bool {
		for _, key := range []string{".fg", ".bg", ".style", ".link-to"} {
			if config.String(path+key) != "" {
				return true
			}
		}

		return false
	}

	themes := config.MapKeys("themes")
	for _, theme := range themes {
		for _, c := range all {
			if _, ok := deliberatelyUncolored[c.path]; ok {
				continue
			}
			if isColored("themes." + theme + "." + c.path) {
				continue
			}
			t.Errorf("theme %q gives no color to %s", theme, c.path)
		}
	}

	// An entry in deliberatelyUncolored that every theme colors anyway is stale:
	// it silences a check that has nothing left to silence.
	for path := range deliberatelyUncolored {
		colored := 0
		for _, theme := range themes {
			if isColored("themes." + theme + "." + path) {
				colored++
			}
		}
		if colored == len(themes) {
			t.Errorf("%s is listed in deliberatelyUncolored but every theme colors it", path)
		}
	}

	// A delegating style ("patterns", "words", "patterns-and-words") says that a
	// field holds free-form text to be sent back through the pattern and word
	// passes. That is a property of the log format rather than of a palette, so
	// the themes must not disagree about it.
	for _, c := range all {
		first := delegatingStyle(config, themes[0], c.path)
		for _, theme := range themes[1:] {
			style := delegatingStyle(config, theme, c.path)
			if style == first {
				continue
			}
			t.Errorf("themes disagree about %s: %q gives it style %q, %q gives it style %q",
				c.path, themes[0], first, theme, style)
		}
	}
}
