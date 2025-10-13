package logalize

import (
	"fmt"
)

// LogFormat represents a log format
type LogFormat struct {
	Name      string
	CapGroups *CapGroupList
}

// LogFormatList represents a list of log formats
type LogFormatList []LogFormat

// InitLogFormats returns list of LogFormats collected
// from *koanf.Koanf configuration
func initLogFormats(settings Settings) (LogFormatList, error) {
	var logFormats LogFormatList
	for _, formatName := range settings.Config.MapKeys("formats") {
		var logFormat LogFormat
		logFormat.Name = formatName
		logFormat.CapGroups = &CapGroupList{}
		if err := settings.Config.Unmarshal("formats."+formatName, &logFormat.CapGroups.Groups); err != nil {
			return nil, err
		}
		logFormats = append(logFormats, logFormat)
	}

	for _, format := range logFormats {
		// set colors and style from the theme
		for i, cg := range format.CapGroups.Groups {
			path := "themes." + settings.Opts.Theme + ".formats." + format.Name + "." + cg.Name
			cgReal := &format.CapGroups.Groups[i]

			if len(cg.Alternatives) > 0 {
				cgReal.Foreground = settings.Config.String(path + ".default.fg")
				cgReal.Background = settings.Config.String(path + ".default.bg")
				cgReal.Style = settings.Config.String(path + ".default.style")
				for j, alt := range cg.Alternatives {
					altReal := &format.CapGroups.Groups[i].Alternatives[j]
					altReal.Foreground = settings.Config.String(path + "." + alt.Name + ".fg")
					altReal.Background = settings.Config.String(path + "." + alt.Name + ".bg")
					altReal.Style = settings.Config.String(path + "." + alt.Name + ".style")
				}
			} else {
				cgReal.Foreground = settings.Config.String(path + ".fg")
				cgReal.Background = settings.Config.String(path + ".bg")
				cgReal.Style = settings.Config.String(path + ".style")
			}
		}

		// init capgroups
		if err := format.CapGroups.init(true); err != nil {
			return nil, fmt.Errorf("[log format: %s] %s", format.Name, err)
		}
	}

	return logFormats, nil
}

func (lf *LogFormat) highlight(str string, h Highlighter) (coloredStr string) {
	str = lf.CapGroups.highlight(str, h)
	if h.settings.Opts.Debug {
		str = h.addDebugInfo(str, *lf)
	}
	return str
}
