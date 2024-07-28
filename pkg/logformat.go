package logalize

import (
	"fmt"

	"github.com/knadh/koanf/v2"
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
func initLogFormats(config *koanf.Koanf) error {
	LogFormats = LogFormatList{}
	for _, formatName := range config.MapKeys("formats") {
		var logFormat LogFormat
		logFormat.Name = formatName
		logFormat.CapGroups = &CapGroupList{}
		if err := config.Unmarshal("formats."+formatName, &logFormat.CapGroups.Groups); err != nil {
			return err
		}
		LogFormats = append(LogFormats, logFormat)
	}

	for _, format := range LogFormats {
		if err := format.CapGroups.init(true); err != nil {
			return fmt.Errorf("[log format: %s] %s", format.Name, err)
		}
	}

	return nil
}

func (lf *LogFormat) highlight(str string) (coloredStr string) {
	return lf.CapGroups.highlight(str)
}
