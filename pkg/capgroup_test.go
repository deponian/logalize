package logalize

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/muesli/termenv"
)

func compareCapGroups(group1, group2 CapGroup) error {
	if group1.RegExpStr != group2.RegExpStr ||
		group1.Foreground != group2.Foreground ||
		group1.Background != group2.Background ||
		group1.Style != group2.Style {
		return fmt.Errorf("foregrounds, backgrounds, styles or regexps aren't equal")
	}
	if len(group1.Alternatives) != len(group2.Alternatives) {
		return fmt.Errorf("alternatives have different length")
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

func compareCapGroupLists(list1, list2 CapGroupList) error {
	if list1.Groups == nil || list2.Groups == nil ||
		len(list1.Groups) != len(list2.Groups) {
		return fmt.Errorf("groups are empty or have different length")
	}

	for i := range list1.Groups {
		if err := compareCapGroups(list1.Groups[i], list2.Groups[i]); err != nil {
			return fmt.Errorf("[capgroup1: %s, capgroup2: %s]: %s", list1.Groups[i].RegExpStr, list2.Groups[i].RegExpStr, err)
		}
	}

	if list1.FullRegExp.String() != list2.FullRegExp.String() {
		return fmt.Errorf("first regexp %s != second regexp %s", list1.FullRegExp.String(), list2.FullRegExp.String())
	}
	return nil
}

func TestCapGroupsListInit(t *testing.T) {
	logFormatCapGroupList := &CapGroupList{
		[]CapGroup{
			{"", `(\d{1,3}(\.\d{1,3}){3} )`, "", "", "", nil, nil},
			{"", `([^ ]+ )`, "", "", "", nil, nil},
			{"", `(\[.+\] )`, "", "", "", nil, nil},
			{"", `("[^"]+")`, "", "", "", nil, nil},
			{
				"",
				`(\d\d\d)`, "", "", "",
				[]CapGroup{
					{"", `(1\d\d)`, "", "", "", nil, nil},
					{"", `(2\d\d)`, "", "", "", nil, nil},
					{"", `(3\d\d)`, "", "", "", nil, nil},
					{"", `(4\d\d)`, "", "", "", nil, nil},
					{"", `(5\d\d)`, "", "", "", nil, nil},
				},
				nil,
			},
		},
		nil,
	}

	correctLogFormatCapGroupList := CapGroupList{
		[]CapGroup{
			{"", `(\d{1,3}(\.\d{1,3}){3} )`, "", "", "", nil, nil},
			{"", `([^ ]+ )`, "", "", "", nil, nil},
			{"", `(\[.+\] )`, "", "", "", nil, nil},
			{"", `("[^"]+")`, "", "", "", nil, nil},
			{
				"",
				`(\d\d\d)`, "", "", "",
				[]CapGroup{
					{"", `(1\d\d)`, "", "", "", nil, regexp.MustCompile(`(1\d\d)`)},
					{"", `(2\d\d)`, "", "", "", nil, regexp.MustCompile(`(2\d\d)`)},
					{"", `(3\d\d)`, "", "", "", nil, regexp.MustCompile(`(3\d\d)`)},
					{"", `(4\d\d)`, "", "", "", nil, regexp.MustCompile(`(4\d\d)`)},
					{"", `(5\d\d)`, "", "", "", nil, regexp.MustCompile(`(5\d\d)`)},
				},
				nil,
			},
		},
		regexp.MustCompile(`^(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3} ))(?P<capGroup1>(?:[^ ]+ ))(?P<capGroup2>(?:\[.+\] ))(?P<capGroup3>(?:"[^"]+"))(?P<capGroup4>(?:\d\d\d))$`),
	}

	patternCapGroupList := &CapGroupList{
		[]CapGroup{
			{"", `(\d{1,3}(\.\d{1,3}){3})`, "", "", "", nil, nil},
		},
		nil,
	}

	correctPatternCapGroupList := CapGroupList{
		[]CapGroup{
			{"", `(\d{1,3}(\.\d{1,3}){3})`, "", "", "", nil, nil},
		},
		regexp.MustCompile(`(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3}))`),
	}

	t.Run("TestCapGroupsListInit", func(t *testing.T) {
		if err := logFormatCapGroupList.init(true); err != nil {
			t.Errorf("logFormatCapGroupList.init(\"\", true) failed with this error: %s", err)
		}

		if err := compareCapGroupLists(*logFormatCapGroupList, correctLogFormatCapGroupList); err != nil {
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

func TestCapGroupsListCheck(t *testing.T) {
	tests := []struct {
		err string
		cgl CapGroupList
	}{
		{
			"%!s(<nil>)",
			CapGroupList{
				[]CapGroup{
					{"", `(\d+:)`, "", "", "", []CapGroup{}, nil},
					{"", `(\d+:)`, "", "", "bold", []CapGroup{}, nil},
					{"", `(\d+:)`, "", "#ff00ff", "", []CapGroup{}, nil},
					{"", `(\d+:)`, "", "#ff0000", "underline", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "", "", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "", "faint", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "#ff00ff", "", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "#ff0000", "italic", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "1", "overline", []CapGroup{}, nil},
					{"", `(\d+:)`, "37", "#ff0000", "crossout", []CapGroup{}, nil},
					{"", `(\d+:)`, "214", "15", "reverse", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "#ff0000", "patterns", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "#ff0000", "words", []CapGroup{}, nil},
					{"", `(\d+:)`, "#0f0f0f", "#ff0000", "patterns-and-words", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capture group: ()] regexp () must start with ( and end with )`,
			CapGroupList{
				[]CapGroup{
					{"", `()`, "", "", "", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			`empty regexps are not allowed`,
			CapGroupList{
				[]CapGroup{
					{"", ``, "", "", "", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capture group: )] regexp ) must start with ( and end with )`,
			CapGroupList{
				[]CapGroup{
					{"", `)`, "", "", "", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capture group: (\d\d-\d\d-\d\d] regexp (\d\d-\d\d-\d\d must start with ( and end with )`,
			CapGroupList{
				[]CapGroup{
					{"", `(\d\d-\d\d-\d\d`, "", "", "", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			fmt.Sprintf(`[capture group: (\d+)] foreground color ff00df doesn't match %s regexp`, colorRegexp),
			CapGroupList{
				[]CapGroup{
					{"", `(\d+)`, "ff00df", "", "", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			fmt.Sprintf(`[capture group: (\d+)] background color 7000 doesn't match %s regexp`, colorRegexp),
			CapGroupList{
				[]CapGroup{
					{"", `(\d+)`, "", "7000", "", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			fmt.Sprintf(`[capture group: (\d+)] style NotAStyle doesn't match %s regexp`, styleRegexp),
			CapGroupList{
				[]CapGroup{
					{"", `(\d+)`, "", "", "NotAStyle", []CapGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capture group: (\d+)] [capture group: hello] regexp hello must start with ( and end with )`,
			CapGroupList{
				[]CapGroup{
					{"", `(\d+)`, "", "", "", []CapGroup{{"", "hello", "", "", "", nil, nil}}, nil},
				},
				nil,
			},
		},
		{
			"error parsing regexp: unexpected ): `\\d+)(\\d+`\nCheck that the \"regexp\" starts with an opening bracket ( and ends with a paired closing bracket )\nThat is, your \"regexp\" must be within one large capture group and contain a valid regular expression",
			CapGroupList{
				[]CapGroup{
					{"", `(\d+)(\d+)`, "", "", "", nil, nil},
				},
				nil,
			},
		},
	}

	colorProfile = termenv.TrueColor

	for _, tt := range tests {
		testname := tt.cgl.Groups[0].RegExpStr
		t.Run(testname, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.cgl.check()); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}
