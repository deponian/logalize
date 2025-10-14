package highlighter

import (
	"fmt"

	"github.com/knadh/koanf/v2"
)

// logFormat represents a log format
type logFormat struct {
	Name      string
	CapGroups *capGroupList
}

// logFormatList represents a list of log formats
type logFormatList []logFormat

// InitLogFormats returns list of LogFormats collected
// from *koanf.Koanf configuration
func newLogFormats(config *koanf.Koanf) (logFormatList, error) {
	var logFormats logFormatList
	for _, formatName := range config.MapKeys("formats") {
		var logFormat logFormat
		logFormat.Name = formatName
		logFormat.CapGroups = &capGroupList{}
		if err := config.Unmarshal("formats."+formatName, &logFormat.CapGroups.Groups); err != nil {
			return nil, err
		}
		logFormats = append(logFormats, logFormat)
	}

	for _, format := range logFormats {
		// set colors and style from the theme
		for i, cg := range format.CapGroups.Groups {
			path := "theme.formats." + format.Name + "." + cg.Name
			cgReal := &format.CapGroups.Groups[i]

			if len(cg.Alternatives) > 0 {
				cgReal.Foreground = config.String(path + ".default.fg")
				cgReal.Background = config.String(path + ".default.bg")
				cgReal.Style = config.String(path + ".default.style")
				for j, alt := range cg.Alternatives {
					altReal := &format.CapGroups.Groups[i].Alternatives[j]
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
		if err := format.CapGroups.init(true); err != nil {
			return nil, fmt.Errorf("[log format: %s] %s", format.Name, err)
		}
	}

	return logFormats, nil
}

func (lf *logFormat) highlight(str string, h Highlighter) (coloredStr string) {
	str = lf.CapGroups.highlight(str, h)
	if h.settings.Opts.Debug {
		str = h.addDebugInfo(str, *lf)
	}
	return str
}
