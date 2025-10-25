package highlighter

import (
	"fmt"
	"regexp"
	"testing"
)

func compareCapGroups(group1, group2 capGroup) error {
	if group1.Name != group2.Name {
		return fmt.Errorf("names %q and %q are different", group1.Name, group2.Name)
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
	if list1.Groups == nil || list2.Groups == nil ||
		len(list1.Groups) != len(list2.Groups) {
		return fmt.Errorf("groups are empty or have different length")
	}

	for i := range list1.Groups {
		if err := compareCapGroups(list1.Groups[i], list2.Groups[i]); err != nil {
			return fmt.Errorf("[capgroup1: %s, capgroup2: %s]: %s", list1.Groups[i].Name, list2.Groups[i].Name, err)
		}
	}

	if list1.FullRegExp.String() != list2.FullRegExp.String() {
		return fmt.Errorf("first regexp %s != second regexp %s", list1.FullRegExp.String(), list2.FullRegExp.String())
	}

	return nil
}

func TestCapGroupsListInit(t *testing.T) {
	formatCapGroupList := &capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "", "", "", nil, nil},
			{"two", `([^ ]+ )`, "", "", "", nil, nil},
			{"three", `(\[.+\] )`, "", "", "", nil, nil},
			{"four", `("[^"]+")`, "", "", "", nil, nil},
			{
				"five",
				`(\d\d\d)`, "", "", "",
				[]capGroup{
					{"alt1", `(1\d\d)`, "", "", "", nil, nil},
					{"alt2", `(2\d\d)`, "", "", "", nil, nil},
					{"alt3", `(3\d\d)`, "", "", "", nil, nil},
					{"alt4", `(4\d\d)`, "", "", "", nil, nil},
					{"alt5", `(5\d\d)`, "", "", "", nil, nil},
				},
				nil,
			},
		},
		nil,
	}

	correctFormatCapGroupList := capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3} )`, "", "", "", nil, nil},
			{"two", `([^ ]+ )`, "", "", "", nil, nil},
			{"three", `(\[.+\] )`, "", "", "", nil, nil},
			{"four", `("[^"]+")`, "", "", "", nil, nil},
			{
				"five",
				`(\d\d\d)`, "", "", "",
				[]capGroup{
					{"alt1", `(1\d\d)`, "", "", "", nil, regexp.MustCompile(`(1\d\d)`)},
					{"alt2", `(2\d\d)`, "", "", "", nil, regexp.MustCompile(`(2\d\d)`)},
					{"alt3", `(3\d\d)`, "", "", "", nil, regexp.MustCompile(`(3\d\d)`)},
					{"alt4", `(4\d\d)`, "", "", "", nil, regexp.MustCompile(`(4\d\d)`)},
					{"alt5", `(5\d\d)`, "", "", "", nil, regexp.MustCompile(`(5\d\d)`)},
				},
				nil,
			},
		},
		regexp.MustCompile(`^(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3} ))(?P<capGroup1>(?:[^ ]+ ))(?P<capGroup2>(?:\[.+\] ))(?P<capGroup3>(?:"[^"]+"))(?P<capGroup4>(?:\d\d\d))$`),
	}

	patternCapGroupList := &capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3})`, "", "", "", nil, nil},
		},
		nil,
	}

	correctPatternCapGroupList := capGroupList{
		[]capGroup{
			{"one", `(\d{1,3}(\.\d{1,3}){3})`, "", "", "", nil, nil},
		},
		regexp.MustCompile(`(?P<capGroup0>(?:\d{1,3}(\.\d{1,3}){3}))`),
	}

	t.Run("TestCapGroupsListInit", func(t *testing.T) {
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

func TestCapGroupsListCheck(t *testing.T) {
	tests := []struct {
		err string
		cgl capGroupList
	}{
		{
			"%!s(<nil>)",
			capGroupList{
				[]capGroup{
					{"1", `(\d+:)`, "", "", "", []capGroup{}, nil},
					{"2", `(\d+:)`, "", "", "bold", []capGroup{}, nil},
					{"3", `(\d+:)`, "", "#ff00ff", "", []capGroup{}, nil},
					{"4", `(\d+:)`, "", "#ff0000", "underline", []capGroup{}, nil},
					{"5", `(\d+:)`, "#0f0f0f", "", "", []capGroup{}, nil},
					{"6", `(\d+:)`, "#0f0f0f", "", "faint", []capGroup{}, nil},
					{"7", `(\d+:)`, "#0f0f0f", "#ff00ff", "", []capGroup{}, nil},
					{"8", `(\d+:)`, "#0f0f0f", "#ff0000", "italic", []capGroup{}, nil},
					{"9", `(\d+:)`, "#0f0f0f", "1", "overline", []capGroup{}, nil},
					{"10", `(\d+:)`, "37", "#ff0000", "crossout", []capGroup{}, nil},
					{"11", `(\d+:)`, "214", "15", "reverse", []capGroup{}, nil},
					{"12", `(\d+:)`, "#0f0f0f", "#ff0000", "patterns", []capGroup{}, nil},
					{"13", `(\d+:)`, "#0f0f0f", "#ff0000", "words", []capGroup{}, nil},
					{"14", `(\d+:)`, "#0f0f0f", "#ff0000", "patterns-and-words", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			`capturing group can't have empty "name" field`,
			capGroupList{
				[]capGroup{
					{"", `(.*)`, "", "", "", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capturing group: one] regexp () must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `()`, "", "", "", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capturing group: one] empty "regexp" field`,
			capGroupList{
				[]capGroup{
					{"one", ``, "", "", "", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capturing group: one] regexp ) must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `)`, "", "", "", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capturing group: one] regexp (\d\d-\d\d-\d\d must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `(\d\d-\d\d-\d\d`, "", "", "", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			fmt.Sprintf(`[capturing group: one] foreground color ff00df doesn't match %s regexp`, colorRegexp),
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "ff00df", "", "", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			fmt.Sprintf(`[capturing group: one] background color 7000 doesn't match %s regexp`, colorRegexp),
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "", "7000", "", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			fmt.Sprintf(`[capturing group: one] style NotAStyle doesn't match %s regexp`, styleRegexp),
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "", "", "NotAStyle", []capGroup{}, nil},
				},
				nil,
			},
		},
		{
			`[capturing group: one] [capturing group: alt1] regexp hello must start with ( and end with )`,
			capGroupList{
				[]capGroup{
					{"one", `(\d+)`, "", "", "", []capGroup{{"alt1", "hello", "", "", "", nil, nil}}, nil},
				},
				nil,
			},
		},
		{
			"[capturing group: one] error parsing regexp: unexpected ): `\\d+)(\\d+`\nCheck that the \"regexp\" starts with an opening bracket ( and ends with a paired closing bracket )\nThat is, your \"regexp\" must be within one large capturing group and contain a valid regular expression",
			capGroupList{
				[]capGroup{
					{"one", `(\d+)(\d+)`, "", "", "", nil, nil},
				},
				nil,
			},
		},
	}

	for _, tt := range tests {
		testname := tt.cgl.Groups[0].RegExpStr
		t.Run(testname, func(t *testing.T) {
			if err := fmt.Sprintf("%s", tt.cgl.check()); err != tt.err {
				t.Errorf("got %s, want %s", err, tt.err)
			}
		})
	}
}
