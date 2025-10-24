package highlighter

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func compareCapGroups(group1, group2 capGroup) error {
	if group1.Name != group2.Name {
		return fmt.Errorf("names %s and %s are different", group1.Name, group2.Name)
	}
	if group1.RegExpStr != group2.RegExpStr ||
		group1.Foreground != group2.Foreground ||
		group1.Background != group2.Background ||
		group1.Style != group2.Style {
		return fmt.Errorf("foregrounds, backgrounds, styles or regexps aren't equal")
	}
	if len(group1.Alternatives) != len(group2.Alternatives) {
		return fmt.Errorf("alternatives have different length")
	}
	if group1.LinkTo != group2.LinkTo {
		return fmt.Errorf("link-to %s and %s are different", group1.LinkTo, group2.LinkTo)
	}
	if group1.Alternatives != nil &&
		group2.Alternatives != nil &&
		len(group1.Alternatives) == len(group2.Alternatives) {
		for i := range len(group1.Alternatives) {
			if err := compareCapGroups(group1.Alternatives[i], group2.Alternatives[i]); err != nil {
				return err
			}
		}
	}
	if group1.RegExp == nil && group2.RegExp != nil {
		return fmt.Errorf("first regexp is nil and second regexp is %s", group2.RegExp.String())
	}
	if group1.RegExp != nil && group2.RegExp == nil {
		return fmt.Errorf("first regexp is %s and second regexp is nil", group1.RegExp.String())
	}
	if group1.RegExp != nil && group2.RegExp != nil &&
		group1.RegExp.String() != group2.RegExp.String() {
		return fmt.Errorf("first regexp %s != second regexp %s", group1.RegExp.String(), group2.RegExp.String())
	}

	return nil
}

func compareCapGroupLists(list1, list2 capGroupList) error {
	if list1.groups == nil || list2.groups == nil ||
		len(list1.groups) != len(list2.groups) {
		return fmt.Errorf("groups are empty or have different length")
	}

	for i := range list1.groups {
		if err := compareCapGroups(list1.groups[i], list2.groups[i]); err != nil {
			return fmt.Errorf("[capgroup1: %s, capgroup2: %s]: %s", list1.groups[i].Name, list2.groups[i].Name, err)
		}
	}

	if list1.fullRegExp.String() != list2.fullRegExp.String() {
		return fmt.Errorf("first regexp %s != second regexp %s", list1.fullRegExp.String(), list2.fullRegExp.String())
	}

	if !cmp.Equal(list1.index, list2.index) {
		return fmt.Errorf("first index map %v != second index map %v", list1.index, list2.index)
	}

	return nil
}

func TestCapGroupsListInitGood(t *testing.T) {
	formatCapGroupList := &capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "", "", "", "", nil, nil},
			{"two", `([^ ]+ )`, "", "", "", "one", nil, nil},
			{"three", `(\[.+\] )`, "", "", "", "four", nil, nil},
			{"four", `("[^"]+")`, "", "", "", "five", nil, nil},
			{
				"five",
				`(\d\d\d)`, "", "", "", "",
				[]capGroup{
					{"alt1", `(1\d\d)`, "", "", "", "", nil, nil},
					{"alt2", `(2\d\d)`, "", "", "", "", nil, nil},
					{"alt3", `(3\d\d)`, "", "", "", "", nil, nil},
					{"alt4", `(4\d\d)`, "", "", "", "", nil, nil},
					{"alt5", `(5\d\d)`, "", "", "", "", nil, nil},
				},
				nil,
			},
		},
		nil,
		nil,
	}

	correctFormatCapGroupList := capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "", "", "", "", nil, nil},
			{"two", `([^ ]+ )`, "", "", "", "one", nil, nil},
			{"three", `(\[.+\] )`, "", "", "", "four", nil, nil},
			{"four", `("[^"]+")`, "", "", "", "five", nil, nil},
			{
				"five",
				`(\d\d\d)`, "", "", "", "",
				[]capGroup{
					{"alt1", `(1\d\d)`, "", "", "", "", nil, regexp.MustCompile(`(1\d\d)`)},
					{"alt2", `(2\d\d)`, "", "", "", "", nil, regexp.MustCompile(`(2\d\d)`)},
					{"alt3", `(3\d\d)`, "", "", "", "", nil, regexp.MustCompile(`(3\d\d)`)},
					{"alt4", `(4\d\d)`, "", "", "", "", nil, regexp.MustCompile(`(4\d\d)`)},
					{"alt5", `(5\d\d)`, "", "", "", "", nil, regexp.MustCompile(`(5\d\d)`)},
				},
				nil,
			},
		},
		regexp.MustCompile(`^(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3} ))(?P<capGroup1>(?:[^ ]+ ))(?P<capGroup2>(?:\[.+\] ))(?P<capGroup3>(?:"[^"]+"))(?P<capGroup4>(?:\d\d\d))$`),
		map[string]int{"one": 0, "two": 1, "three": 2, "four": 3, "five": 4},
	}

	patternCapGroupList := &capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3})`, "", "", "", "two", nil, nil},
			{"two", `(.*)`, "", "", "", "", nil, nil},
		},
		nil,
		nil,
	}

	correctPatternCapGroupList := capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3})`, "", "", "", "two", nil, nil},
			{"two", `(.*)`, "", "", "", "", nil, nil},
		},
		regexp.MustCompile(`(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3}))(?P<capGroup1>(?:.*))`),
		map[string]int{"one": 0, "two": 1},
	}

	t.Run("TestCapGroupsListInitGood", func(t *testing.T) {
		if err := formatCapGroupList.init(true); err != nil {
			t.Errorf("formatCapGroupList.init(\"\", true) failed with this error: %s", err)
		}

		if err := compareCapGroupLists(*formatCapGroupList, correctFormatCapGroupList); err != nil {
			t.Errorf("%s", err)
		}

		if err := patternCapGroupList.init(false); err != nil {
			t.Errorf("patternCapGroupList.init(\"\", false) failed with this error: %s", err)
		}

		if err := compareCapGroupLists(*patternCapGroupList, correctPatternCapGroupList); err != nil {
			t.Errorf("%s", err)
		}
	})
}

func TestCapGroupsListInitBad(t *testing.T) {
	tests := []struct {
		err string
		cgl capGroupList
	}{
		{
			`[capture group: one] link-to "hello" refers to unknown capture group`,
			capGroupList{
				[]capGroup{
					{"one", `(\d+:)`, "", "", "", "hello", []capGroup{}, nil},
					{"two", `(\d+:)`, "", "", "", "two", []capGroup{}, nil},
					{"three", `(\d+:)`, "", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run("TestCapGroupsListInitBad"+tt.cgl.groups[0].Name, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.cgl.init(false)); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}

func TestCapGroupsValidateLinkTo(t *testing.T) {
	tests := []struct {
		err string
		cgl capGroupList
	}{
		{
			`[capture group: one] link-to "hello" refers to unknown capture group`,
			capGroupList{
				[]capGroup{
					{"one", `(\d+:)`, "", "", "", "hello", []capGroup{}, nil},
					{"two", `(\d+:)`, "", "", "", "two", []capGroup{}, nil},
					{"three", `(\d+:)`, "", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`[capture group: one] cyclic link-to detected`,
			capGroupList{
				[]capGroup{
					{"one", `(\d+:)`, "", "", "", "two", []capGroup{}, nil},
					{"two", `(\d+:)`, "", "", "", "one", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`[capture group: one] cyclic link-to detected`,
			capGroupList{
				[]capGroup{
					{"one", `(\d+:)`, "", "", "", "two", []capGroup{}, nil},
					{"two", `(\d+:)`, "", "", "", "three", []capGroup{}, nil},
					{"three", `(\d+:)`, "", "", "", "one", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run("TestCapGroupsValidateLinkTo"+tt.cgl.groups[0].Name, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.cgl.validateLinkTo()); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}

func TestCapGroupsLinkedStyle(t *testing.T) {
	tests := []struct {
		str string

		cgIndex int

		fg    string
		bg    string
		style string
		ok    bool
	}{
		{
			`hello --- 100`,
			0,
			"#ff0000",
			"",
			"",
			true,
		},
		{
			`hello --- 200`,
			0,
			"",
			"#00ff00",
			"",
			true,
		},
		{
			`hello --- 300`,
			0,
			"",
			"",
			"bold",
			true,
		},
		{
			`hello --- 777`,
			0,
			"#ffffff",
			"",
			"",
			true,
		},

		{
			`hello --- 100`,
			1,
			"#ff0000",
			"",
			"",
			true,
		},
		{
			`hello --- 200`,
			1,
			"",
			"#00ff00",
			"",
			true,
		},
		{
			`hello --- 300`,
			1,
			"",
			"",
			"bold",
			true,
		},
		{
			`hello --- 777`,
			1,
			"#ffffff",
			"",
			"",
			true,
		},
	}

	cgl := &capGroupList{
		[]capGroup{
			{"one", `(hello )`, "", "", "", "three", nil, nil},
			{"two", `(--- )`, "", "", "", "one", nil, nil},
			{
				"three",
				`(\d\d\d)`, "#ffffff", "", "", "",
				[]capGroup{
					{"alt1", `(1\d\d)`, "#ff0000", "", "", "", nil, nil},
					{"alt2", `(2\d\d)`, "", "#00ff00", "", "", nil, nil},
					{"alt3", `(3\d\d)`, "", "", "bold", "", nil, nil},
				},
				nil,
			},
		},
		nil,
		nil,
	}

	if err := cgl.init(false); err != nil {
		t.Fatalf("cgl.init(...) failed with this error: %s", err)
	}

	for _, tt := range tests {
		t.Run("TestCapGroupsLinkedStyle"+tt.str, func(t *testing.T) {
			matches := cgl.fullRegExp.FindStringSubmatch(tt.str)
			fg, bg, style, ok := cgl.linkedStyle(matches, cgl.groups[tt.cgIndex])
			if fg != tt.fg || bg != tt.bg || style != tt.style || ok != tt.ok {
				t.Errorf("got (%s, %s, %s, %v), want (%s, %s, %s, %v)", fg, bg, style, ok, tt.fg, tt.bg, tt.style, tt.ok)
			}
		})
	}
}

func TestCapGroupsListCheck(t *testing.T) {
	tests := []struct {
		err string
		cgl capGroupList
	}{
		{
			"%!s(<nil>)",
			capGroupList{
				[]capGroup{
					{"1", `(\d+:)`, "", "", "", "", []capGroup{}, nil},
					{"2", `(\d+:)`, "", "", "bold", "", []capGroup{}, nil},
					{"3", `(\d+:)`, "", "#ff00ff", "", "", []capGroup{}, nil},
					{"4", `(\d+:)`, "", "#ff0000", "underline", "", []capGroup{}, nil},
					{"5", `(\d+:)`, "#0f0f0f", "", "", "", []capGroup{}, nil},
					{"6", `(\d+:)`, "#0f0f0f", "", "faint", "", []capGroup{}, nil},
					{"7", `(\d+:)`, "#0f0f0f", "#ff00ff", "", "", []capGroup{}, nil},
					{"8", `(\d+:)`, "#0f0f0f", "#ff0000", "italic", "", []capGroup{}, nil},
					{"9", `(\d+:)`, "#0f0f0f", "1", "overline", "", []capGroup{}, nil},
					{"10", `(\d+:)`, "37", "#ff0000", "crossout", "", []capGroup{}, nil},
					{"11", `(\d+:)`, "214", "15", "reverse", "", []capGroup{}, nil},
					{"12", `(\d+:)`, "#0f0f0f", "#ff0000", "patterns", "", []capGroup{}, nil},
					{"13", `(\d+:)`, "#0f0f0f", "#ff0000", "words", "", []capGroup{}, nil},
					{"14", `(\d+:)`, "#0f0f0f", "#ff0000", "patterns-and-words", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`capturing group can't have empty "name" field`,
			capGroupList{
				[]capGroup{
					{"", `(.*)`, "", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`[capturing group: one] regexp () must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `()`, "", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`[capturing group: one] empty "regexp" field`,
			capGroupList{
				[]capGroup{
					{"one", ``, "", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`[capturing group: one] regexp ) must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `)`, "", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`[capturing group: one] regexp (\d\d-\d\d-\d\d must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `(\d\d-\d\d-\d\d`, "", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			fmt.Sprintf(`[capturing group: one] foreground color ff00df doesn't match %s regexp`, colorRegexp),
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "ff00df", "", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			fmt.Sprintf(`[capturing group: one] background color 7000 doesn't match %s regexp`, colorRegexp),
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "", "7000", "", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			fmt.Sprintf(`[capturing group: one] style NotAStyle doesn't match %s regexp`, styleRegexp),
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "", "", "NotAStyle", "", []capGroup{}, nil},
				},
				nil,
				nil,
			},
		},
		{
			`[capturing group: one] [capturing group: alt1] regexp hello must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "", "", "", "", []capGroup{{"alt1", "hello", "", "", "", "", nil, nil}}, nil},
				},
				nil,
				nil,
			},
		},
		{
			"[capturing group: one] error parsing regexp: unexpected ): `\\d+)(\\d+`\nCheck that the \"regexp\" starts with an opening bracket ( and ends with a paired closing bracket )\nThat is, your \"regexp\" must be within one large capturing group and contain a valid regular expression",
			capGroupList{
				[]capGroup{
					{"one", `(\d+)(\d+)`, "", "", "", "", nil, nil},
				},
				nil,
				nil,
			},
		},
	}

	for _, tt := range tests {
		testname := tt.cgl.groups[0].RegExpStr
		t.Run(testname, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.cgl.check()); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}
