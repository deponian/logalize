package logalize

import (
	"bufio"
	"io"

	"github.com/aaaton/golem/v4"
)

func Run(reader io.Reader, writer io.StringWriter, lemmatizer *golem.Lemmatizer) error {
	if err := initPatterns(); err != nil {
		return err
	}

	if err := initWords(lemmatizer); err != nil {
		return err
	}

	if err := initLogFormats(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// remove all ANSI escape sequences from the input by default
		if !Opts.NoANSIEscapeSequencesStripping {
			line = StripANSIEscapeSequences(line)
		}
		// try one of the log formats
		formatDetected := false
		for _, logFormat := range LogFormats {
			if logFormat.CapGroups.FullRegExp.MatchString(line) {
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
