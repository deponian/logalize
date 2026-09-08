package highlighter

import (
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/deponian/logalize/internal/config"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

// repoRoot returns the repository root as an fs.FS.
func repoRoot() fs.FS {
	return os.DirFS("../..")
}

// newLabeledHighlighter builds a Highlighter over the real built-in rules,
// themed so that its output decodes into <label>text</label> markers.
func newLabeledHighlighter(t *testing.T) (Highlighter, *labelTheme) {
	t.Helper()

	settings, err := config.NewSettings(repoRoot(), nil, nil, true)
	if err != nil {
		t.Fatalf("config.NewSettings(...) failed with this error: %s", err)
	}
	settings.ColorProfile = termenv.TrueColor

	labels, err := newLabelTheme(settings.Config, settings.ColorProfile)
	if err != nil {
		t.Fatalf("newLabelTheme(...) failed with this error: %s", err)
	}
	settings.Opts.Theme = labelThemeName

	hl, err := NewHighlighter(settings)
	if err != nil {
		t.Fatalf("NewHighlighter() failed with this error: %s", err)
	}

	return hl, labels
}

// labelThemeName is the name of the synthetic theme newLabelTheme adds to a
// configuration.
const labelThemeName = "label"

// labelTheme is a theme generated from the rules themselves, in which every
// capturing group, alternative and word group is given its own color. The colors
// carry no meaning beyond being distinct, which makes them reversible using decode()
type labelTheme struct {
	profile termenv.Profile
	// counter is used for generating distinct colors
	counter int
	// labelBySequence maps the SGR attributes a label's color actually
	// produces back to that label
	labelBySequence map[string]string
}

// colorable is one thing in the rules that a theme can put a color on:
// a capturing group, one of a group's alternatives or a word group.
type colorable struct {
	// path locates it under a theme, e.g. "formats.klog.log-level.info"
	path string
	// label is the marker decode writes around the text it colors,
	// e.g. "log-level:info"
	label string
}

// colorables lists everything in the rules that a theme can color. The three
// kinds of rule are nested differently:
//
//	formats.<format>.<group>[.<alternative>]      ->  <group>[:<alternative>]
//	patterns.<pattern>.<group>[.<alternative>]    ->  p:<pattern>:<group>[:<alternative>]
//	words.<group>                                 ->  w:<group>
//
// with two exceptions: a pattern written as a single regexp has no group level
// at all, so its colors sit where a pattern with a "regexps" list keeps its
// groups (see initPattern), and word groups have neither groups nor
// alternatives.
func colorables(config *koanf.Koanf) ([]colorable, error) {
	var all []colorable

	// add records one capturing group followed by each of its alternatives
	add := func(path, label string, cg capGroup) {
		all = append(all, colorable{path, label})
		for _, alt := range cg.Alternatives {
			all = append(all, colorable{path + "." + alt.Name, label + ":" + alt.Name})
		}
	}

	formats, err := collectFormats(config)
	if err != nil {
		return nil, err
	}
	for _, format := range formats {
		for _, cg := range format.CapGroups.groups {
			add("formats."+format.Name+"."+cg.Name, cg.Name, cg)
		}
	}

	patterns, err := collectPatterns(config)
	if err != nil {
		return nil, err
	}
	for _, pattern := range patterns {
		hasGroups := config.Exists("patterns." + pattern.Name + ".regexps")
		for _, cg := range pattern.CapGroups.groups {
			path, label := "patterns."+pattern.Name, "p:"+pattern.Name
			if hasGroups {
				path, label = path+"."+cg.Name, label+":"+cg.Name
			}
			add(path, label, cg)
		}
	}

	for _, name := range config.MapKeys("words") {
		all = append(all, colorable{"words." + name, "w:" + name})
	}

	return all, nil
}

// newLabelTheme adds a theme named "label" to config, giving every colorable a
// color that no other colorable has.
func newLabelTheme(config *koanf.Koanf, profile termenv.Profile) (*labelTheme, error) {
	lt := &labelTheme{profile: profile, labelBySequence: map[string]string{}}

	all, err := colorables(config)
	if err != nil {
		return nil, err
	}
	// Which fields delegate is a property of the log format rather than of a
	// palette, so the shipped themes agree about it and any one of them can be read for all.
	// TestThemesAreComplete() tests it.
	reference := config.MapKeys("themes")[0]

	for _, c := range all {
		path := "themes." + labelThemeName + "." + c.path

		style := delegatingStyle(config, reference, c.path)

		// a delegating style is structural rather than decorative: it hands the
		// matched text back to the pattern and word passes instead of coloring
		// it, so the generated theme has to carry it over as it is
		if style != "" {
			err = config.Set(path+".style", style)
		} else {
			err = config.Set(path+".fg", lt.colorFor(c.label))
		}
		if err != nil {
			return nil, err
		}
	}

	return lt, nil
}

// delegatingStyle returns the delegating style theme puts on path or "" when
// it gives path an ordinary color instead.
func delegatingStyle(config *koanf.Koanf, theme, path string) string {
	style := config.String("themes." + theme + "." + path + ".style")
	if !(style == "patterns" || style == "words" || style == "patterns-and-words") {
		return ""
	}

	return style
}

// colorFor hands label the next color that no other label has taken and returns
// it as a hex string.
//
// termenv puts hex colors through floating point on their way to an escape
// sequence, so the color a theme asks for is not always the color that comes
// out, and two hex values can end up as the same sequence. Reserving colors by
// the sequence they actually produce, and stepping over the ones already taken,
// keeps the mapping one-to-one by construction.
func (lt *labelTheme) colorFor(label string) string {
	for {
		lt.counter++
		i := lt.counter
		hex := fmt.Sprintf("#%02x%02x%02x", i>>16&0xff, i>>8&0xff, i&0xff)
		sequence := lt.profile.Color(hex).Sequence(false)
		if _, taken := lt.labelBySequence[sequence]; taken {
			continue
		}
		lt.labelBySequence[sequence] = label

		return hex
	}
}

// decode replaces every colored span with a <label>text</label> marker.
// Spans whose color isn't in the table are marked <?attributes> so that a
// stray color shows up in the diff instead of being silently accepted.
func (lt *labelTheme) decode(str string) string {
	return sgrANSIEscapeSequenceRegExp.ReplaceAllStringFunc(str, func(span string) string {
		match := sgrANSIEscapeSequenceRegExp.FindStringSubmatch(span)
		label, ok := lt.labelBySequence[match[1]]
		if !ok {
			label = "?" + match[1]
		}

		return "<" + label + ">" + match[2] + "</" + label + ">"
	})
}
