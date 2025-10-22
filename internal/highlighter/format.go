package highlighter

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

type format struct {
	Name      string
	CapGroups *capGroupList
}

type formatList []format

// newFormats returns list of formats collected
// from *koanf.Koanf configuration
func newFormats(config *koanf.Koanf, theme string) (formatList, error) {
	if config == nil {
		return formatList{}, nil
	}

	formats, err := collectFormats(config)
	if err != nil {
		return nil, err
	}

	for i := range formats {
		if err := initFormat(&formats[i], config, theme); err != nil {
			return nil, err
		}
	}

	return formats, nil
}

func collectFormats(config *koanf.Koanf) (formatList, error) {
	var formats formatList

	for _, formatName := range config.MapKeys("formats") {
		var format format
		format.Name = formatName
		format.CapGroups = &capGroupList{}
		if err := config.Unmarshal("formats."+formatName, &format.CapGroups.Groups); err != nil {
			return nil, err
		}
		formats = append(formats, format)
	}

	return formats, nil
}

func initFormat(lf *format, config *koanf.Koanf, theme string) error {
	// set colors and style from the theme
	for i, cg := range lf.CapGroups.Groups {
		path := "themes." + theme + ".formats." + lf.Name + "." + cg.Name
		cgReal := &lf.CapGroups.Groups[i]

		if len(cg.Alternatives) > 0 {
			cgReal.Foreground = config.String(path + ".default.fg")
			cgReal.Background = config.String(path + ".default.bg")
			cgReal.Style = config.String(path + ".default.style")
			for j, alt := range cg.Alternatives {
				altReal := &lf.CapGroups.Groups[i].Alternatives[j]
				altReal.Foreground = config.String(path + "." + alt.Name + ".fg")
				altReal.Background = config.String(path + "." + alt.Name + ".bg")
				altReal.Style = config.String(path + "." + alt.Name + ".style")
			}
		} else {
			cgReal.Foreground = config.String(path + ".fg")
			cgReal.Background = config.String(path + ".bg")
			cgReal.Style = config.String(path + ".style")
		}
	}

	// init capgroups
	if err := lf.CapGroups.init(true); err != nil {
		return fmt.Errorf("[format: %s] %s", lf.Name, err)
	}

	return nil
}

func (lf format) highlight(str string, h Highlighter) (coloredStr string) {
	str = lf.CapGroups.highlight(str, h)
	if h.settings.Opts.Debug {
		str = h.addDebugInfo(str, lf)
	}

	return str
}

func (lf format) match(str string) bool {
	return lf.CapGroups.FullRegExp.MatchString(str)
}
