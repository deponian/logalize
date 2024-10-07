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

var LogFormats LogFormatList

// InitLogFormats returns list of LogFormats collected
// from *koanf.Koanf configuration
func initLogFormats() error {
	LogFormats = LogFormatList{}
	for _, formatName := range Config.MapKeys("formats") {
		var logFormat LogFormat
		logFormat.Name = formatName
		logFormat.CapGroups = &CapGroupList{}
		if err := Config.Unmarshal("formats."+formatName, &logFormat.CapGroups.Groups); err != nil {
			return err
		}
		LogFormats = append(LogFormats, logFormat)
	}

	for _, format := range LogFormats {
		// set colors and style from the theme
		for i, cg := range format.CapGroups.Groups {
			path := "themes." + Opts.Theme + ".formats." + format.Name + "." + cg.Name
			cgReal := &format.CapGroups.Groups[i]

			if len(cg.Alternatives) > 0 {
				cgReal.Foreground = Config.String(path + ".default.fg")
				cgReal.Background = Config.String(path + ".default.bg")
				cgReal.Style = Config.String(path + ".default.style")
				for j, alt := range cg.Alternatives {
					altReal := &format.CapGroups.Groups[i].Alternatives[j]
					altReal.Foreground = Config.String(path + "." + alt.Name + ".fg")
					altReal.Background = Config.String(path + "." + alt.Name + ".bg")
					altReal.Style = Config.String(path + "." + alt.Name + ".style")
				}
			} else {
				cgReal.Foreground = Config.String(path + ".fg")
				cgReal.Background = Config.String(path + ".bg")
				cgReal.Style = Config.String(path + ".style")
			}
		}

		// init capgroups
		if err := format.CapGroups.init(true); err != nil {
			return fmt.Errorf("[log format: %s] %s", format.Name, err)
		}
	}

	return nil
}

func (lf *LogFormat) highlight(str string) (coloredStr string) {
	return lf.CapGroups.highlight(str)
}
