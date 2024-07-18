package logalize

import (
	"bufio"
	"embed"
	"io"

	"github.com/aaaton/golem/v4"
	"github.com/knadh/koanf/v2"
)

func Run(reader io.Reader, writer io.StringWriter, config *koanf.Koanf, builtins embed.FS, lemmatizer *golem.Lemmatizer) error {
	if err := initPatterns(config); err != nil {
		return err
	}

	if err := initWords(config, lemmatizer); err != nil {
		return err
	}

	if err := initLogFormats(config); err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// try one of the log formats
		formatDetected := false
		for _, logFormat := range LogFormats {
			if logFormat.Regexp.MatchString(line) {
				_, err := writer.WriteString(logFormat.highlight(line) + "\n")
				if err != nil {
					return err
				}
				formatDetected = true
				break
			}
		}
		// highlight patterns and words if log format wasn't detected
		if !formatDetected {
			_, err := writer.WriteString(Patterns.highlight(line, true) + "\n")
			if err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
