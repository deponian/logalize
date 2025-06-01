package logalize

import (
	"bufio"
	"bytes"
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

	bufReader := bufio.NewReader(reader)
	var buffer bytes.Buffer

	for {
		b, readErr := bufReader.ReadByte()
		if readErr != nil && readErr != io.EOF {
			return readErr
		}

		if b == '\r' || b == '\n' || readErr == io.EOF {
			// b is equal to 0 when we reach the end of the file
			// so we won't add it to the end of the string
			lastCharacter := ""
			if b == '\r' || b == '\n' {
				lastCharacter = string(b)
			}

			colored := colorize(buffer.String())

			_, err := writer.WriteString(colored + lastCharacter)
			if err != nil {
				return err
			}

			buffer.Reset()

			if readErr == io.EOF {
				break
			} else {
				continue
			}
		}

		buffer.WriteByte(b)
	}

	return nil
}

// colorize detects log formats, patterns and words in the input string
// and returns colored result string
func colorize(line string) string {
	// don't alter the input in any way if user set --dry-run flag
	if Opts.DryRun {
		return line
	}

	// remove all ANSI escape sequences from the input by default
	if !Opts.NoANSIEscapeSequencesStripping {
		line = StripANSIEscapeSequences(line)
	}

	// try one of the log formats
	for _, logFormat := range LogFormats {
		if logFormat.CapGroups.FullRegExp.MatchString(line) {
			return logFormat.highlight(line)
		}
	}

	// highlight patterns and words if log format wasn't detected
	return Patterns.highlight(line, true)
}
