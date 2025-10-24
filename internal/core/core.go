// Package core is the main entrypoint to the whole application
package core

import (
	"bufio"
	"bytes"
	"io"

	"github.com/deponian/logalize/internal/config"
	"github.com/deponian/logalize/internal/highlighter"
)

// Run reads lines from the reader, colorizes them based on the settings
// and writes the lines to the writer
func Run(reader io.Reader, writer io.Writer, settings config.Settings) error {
	hl, err := highlighter.NewHighlighter(settings)
	if err != nil {
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
			var lastCharacter string
			if b == '\r' || b == '\n' {
				lastCharacter = string(b)
			}

			colored := hl.Colorize(buffer.String())

			_, err := writer.Write([]byte(colored + lastCharacter))
			if err != nil {
				return err
			}

			buffer.Reset()

			if readErr == io.EOF {
				break
			}

			continue
		}

		buffer.WriteByte(b)
	}

	return nil
}
