package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	cmd "github.com/deponian/logalize/cmd/logalize"
	"github.com/muesli/mango"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
)

func buildCommand(m mango.ManPage, doc mango.Builder, c mango.Command) {
	if len(c.Flags) > 0 {
		if c.Name == m.Root.Name {
			doc.Section("Options")
			doc.TaggedParagraph(-1)
		} else {
			doc.TaggedParagraph(-1)
			doc.TextBold("OPTIONS")
			doc.Indent(4)
		}
		keys := make([]string, 0, len(c.Flags))
		for k := range c.Flags {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			opt := c.Flags[k]
			if i > 0 {
				doc.TaggedParagraph(-1)
			}

			prefix := "-"
			if opt.PFlag {
				prefix = "--"
			}

			if opt.Short != "" {
				doc.TextBold(fmt.Sprintf("-%[2]s, %[1]s%[3]s", prefix, opt.Short, opt.Name))
			} else {
				doc.TextBold(prefix + opt.Name)
			}
			doc.EndSection()
			doc.Text(strings.ReplaceAll(opt.Usage, "\n", " "))
		}

		if c.Name == m.Root.Name {
		} else {
			doc.IndentEnd()
		}
	}

	if len(c.Commands) > 0 {
		if c.Name == m.Root.Name {
			doc.Section("Commands")
			doc.TaggedParagraph(-1)
		} else {
			doc.TaggedParagraph(-1)
			doc.TextBold("COMMANDS")
			doc.Indent(4)
		}
		keys := make([]string, 0, len(c.Commands))
		for k := range c.Commands {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			opt := c.Commands[k]
			if i > 0 {
				doc.TaggedParagraph(-1)
			}

			doc.TextBold(opt.Name)
			if opt.Usage != "" {
				doc.Text(strings.TrimPrefix(opt.Usage, opt.Name))
			}
			doc.Indent(4)
			doc.Text(strings.ReplaceAll(opt.Short, "\n", " "))
			doc.IndentEnd()

			buildCommand(m, doc, *opt)
		}

		if c.Name == m.Root.Name {
		} else {
			doc.IndentEnd()
		}
	}

	if c.Example != "" {
		if c.Name == m.Root.Name {
			doc.Section("Examples")
			doc.TaggedParagraph(-1)
		} else {
			doc.TaggedParagraph(-1)
			doc.TextBold("EXAMPLES")
			doc.Indent(4)
		}
		doc.Text(c.Example)

		if c.Name == m.Root.Name {
			doc.EndSection()
		} else {
			doc.IndentEnd()
		}
	}
}

func main() {
	version := os.Args[1]
	date := time.Now().Format("2006-01-02")
	var section uint = 1

	logalizeCmd := cmd.NewCommand(embed.FS{})

	name := logalizeCmd.Name()
	description := logalizeCmd.Short

	manPage, err := mcobra.NewManPage(section, logalizeCmd)
	if err != nil {
		log.Fatal(err)
	}
	doc := roff.NewDocument()

	doc.Section("Name")
	doc.Text(name + " - " + description)
	doc.EndSection()

	doc.Section("Synopsis")
	doc.TextBold(name)
	doc.Text(" [")
	doc.TextItalic("options...")
	doc.Text("]")
	doc.EndSection()

	doc.Section("Description")
	doc.Text(`Logalize is a log colorizer. It's extensible alternative to ccze and colorize.

Logalize reads one line from stdin at a time and then checks if it matches one of formats ("formats" key in configuration file), general regular expressions ("patterns" key in configuration file) or plain English words and their inflected forms ("words" key in configuration file).

Simplified version of the main loop:`)

	doc.Indent(7)
	doc.Text("1. Read a line from stdin")
	doc.IndentEnd()
	doc.Indent(7)
	doc.Text("2. Strip all ANSI escape sequences (see --no-ansi-escape-sequences-stripping flag below)")
	doc.IndentEnd()
	doc.Indent(7)
	doc.Text("3. If the entire line matches one of the formats, print colored line and go to step 1, otherwise go to step 3")
	doc.IndentEnd()
	doc.Indent(7)
	doc.Text("4. Find and color all patterns in the line and go to step 4")
	doc.IndentEnd()
	doc.Indent(7)
	doc.Text("5. Find and color all words, print colored line and go to step 1")
	doc.IndentEnd()
	doc.EndSection()

	buildCommand(*manPage, doc, manPage.Root)

	doc.Section("Files")
	doc.TextBold("/etc/logalize/logalize.yaml")
	doc.Indent(7)
	doc.Text("System-wide configuration file.")
	doc.IndentEnd()
	doc.Paragraph()
	doc.TextBold("$HOME/.config/logalize/logalize.yaml")
	doc.Indent(7)
	doc.Text("User configuration file.")
	doc.IndentEnd()
	doc.Paragraph()
	doc.TextBold("./.logalize.yaml")
	doc.Indent(7)
	doc.Text("Local per-project configuration file.")
	doc.IndentEnd()
	doc.EndSection()

	doc.Section("Bugs")
	doc.Text("Bugs can be reported on GitHub: https://github.com/deponian/logalize/issues")
	doc.EndSection()

	doc.Section("See also")
	doc.Text("ccze(1), colorize(1)")
	doc.EndSection()

	doc.Section("Authors")
	doc.Text("Rufus Deponian (rufus@deponian.com)")
	doc.EndSection()

	doc.Section("Copyright")
	doc.Text("Released under MIT license")
	doc.EndSection()

	fmt.Printf(
		".TH %s %v \"%s\" \"%s %s\" \"%s\"",
		strings.ToUpper(name), section, date, name, version, description,
	)
	fmt.Print(doc.String())
}
