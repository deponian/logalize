<picture>
  <source media="(prefers-color-scheme: dark)" srcset="images/avif/logo-dark.avif">
  <source media="(prefers-color-scheme: light)" srcset="images/avif/logo-light.avif">
  <img alt="Screenshot" src="images/avif/logo-light.avif">
</picture>

93% of all logs are not colored[^1]. It's sad. Maybe even illegal. It's time to logalize them. **Logalize** is a log colorizer like [colorize](https://github.com/raszi/colorize) and [ccze](https://github.com/cornet/ccze). But it's faster and, much more importantly, it's extensible. No more hardcoded templates for logs and keywords. Logalize is fully customizable via `logalize.yaml` where you can define your log formats, keyword patterns and more.

<p align="center">
  <a href="https://github.com/deponian/logalize/actions"><img src="https://github.com/deponian/logalize/actions/workflows/tests.yml/badge.svg" alt="Build Status"></a>
  <a href="https://codecov.io/gh/deponian/logalize"><img src="https://codecov.io/gh/deponian/logalize/graph/badge.svg?token=8NJ4ZC8COT"/></a>
  <a href="https://goreportcard.com/report/github.com/deponian/logalize"><img src="https://goreportcard.com/badge/github.com/deponian/logalize" alt="Go Report Card"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/deponian/logalize/releases/latest"><img src="https://img.shields.io/github/v/release/deponian/logalize" alt="Github Release"></a>
</p>

Usage
-----

```sh
cat /path/to/logs/file.log | logalize
```

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="images/avif/screenshot-dark.avif">
  <source media="(prefers-color-scheme: light)" srcset="images/avif/screenshot-light.avif">
  <img alt="Screenshot" src="images/avif/screenshot-light.avif">
</picture>

Installation
------------

Download DEB, RPM and Arch Linux packages or the binary for your architecture from [releases](https://github.com/deponian/logalize/releases/latest).

**Ubuntu/Debian:**
```sh
sudo dpkg -i logalize_X.X.X_linux_amd64.deb
```

**Fedora/Red Hat Enterprise Linux/CentOS:**
```sh
sudo rpm -i logalize_X.X.X_linux_amd64.rpm
```

**Arch Linux/Manjaro:**
```sh
sudo pacman -U logalize_X.X.X_linux_amd64.pkg.tar.zst
```
or install from AUR:
```sh
# to install precompiled binary
yay -S logalize-bin

# to compile it on your machine
yay -S logalize
```

Use `go install` if you already have `$GOPATH/bin` in your `$PATH`:

```sh
go install github.com/deponian/logalize@latest
```

How it works
------------

Logalize reads one line from stdin at a time and then checks if it matches one of log formats (`formats`), general regular expressions (`patterns`) or plain English words and their [inflected](https://en.wikipedia.org/wiki/Inflection) forms (`words`). See configuration below for more details.

Simplified version of the main loop:
1. Read a line from stdin
2. If the entire line matches one of the `formats`, print colored line and go to step 1, otherwise go to step 3
3. Find and color all `patterns` in the line and go to step 4
4. Find and color all `words`, print colored line and go to step 1

Configuration
-------------

Logalize looks for configuration files in these places:
- `/etc/logalize/logalize.yaml`
- `~/.config/logalize/logalize.yaml`
- `.logalize.yaml`
- path from `-c/--config` option

If more than one configuration file is found, they are merged. The lower the file in the list, the higher its priority.

A configuration file can contain three top-level keys: `formats`, `patterns` and `words`.

### Log formats

Configuration example:

```yaml
formats:
  kuvaq:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      fg: "#f5ce42"
    - regexp: (- )
      bg: "#807e7a"
      style: bold
    - regexp: ("[^"]+" )
      fg: "#9ddb56"
      bg: "#f5ce42"
    - regexp: (\d\d\d)
      fg: "#ffffff"
      alternatives:
        - regexp: (2\d\d)
          fg: "#00ff00"
          style: bold
        - regexp: (3\d\d)
          fg: "#00ffff"
          style: bold
```

`formats` describe complete log formats. A line must match a format completely to be colored. For example, the full regular expression for the "kuvaq" format above is `^(\d{1,3}(\.\d{1,3}){3} )(- )("[^"]+" )(\d\d\d)$`. Only lines below will match this format:
- `127.0.0.1 - "menetekel" 777`
- `7.7.7.7 - "m" 000`

But not these:
- `127.0.0.1 - "menetekel" 777 lower ascension station`
- `Upper ascension station 127.0.0.1 - "menetekel" 777`
- `127.0.0.1 - "menetekel" 777000`

For an overview of regular expression syntax, see the [regexp/syntax](https://pkg.go.dev/regexp/syntax) package.

Full log format example using all available fields:

```yaml
formats:
  # name of a log format
  elysium:
    # regexp must begin with an opening parenthesis `(`
    # and it must end with a paired closing parenthesis `)`
    # regexp can't be empty `()`
    # that is, your regexp must be within one large capture group
    # and contain a valid regular expression
    - regexp: (\d\d\d )
      # color can be a hex value like #ff0000
      # or a number between 0 and 255 for ANSI colors
      fg: "#00ff00"
      bg: "#0000ff"
      # available regular styles:
      #  bold, faint, italic, underline,
      #  overline, crossout, reverse
      # there are also three special styles:
      #  patterns - use highlighting from "patterns" section (see below)
      #  words - use highlighting from "words" section (see below)
      #  patterns-and-words - use highlighting from "patterns" and "words" sections
      style: bold
      # alternatives are useful when you have general regular expression
      # but you want different colors for some specific subset of cases
      # within this regular expression
      # a common example is HTTP status code
      alternatives:
        # every regexp here has the same "fg", "bg" and "style" fields
        # but no "alternatives" field
        - regexp: (2\d\d )
          fg: "#00ff00"
          bg: "#0000ff"
          style: bold
        - regexp: (4\d\d )
          fg: "#ff0000"
          bg: "#0000ff"
          style: underline
    # each next regexp is added to the previous one
    # and together they form a complete regexp for the whole string
    - regexp: (--- )
      # . . . . .
    - regexp: ([[:xdigit:]]{32})
      # . . . . .
    # full regexp for this whole example is:
    # ^(\d\d\d )(--- )([[:xdigit:]]{32})$
```

You can find built-in `formats` [here](builtins/logformats). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`.

### Patterns

Configuration example:

```yaml
patterns:
  # simple patterns
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')
    fg: "#00ff00"

  number:
    regexp: (\d+)
    bg: "#00ffff"
    style: bold

  http-status-code:
    regexp: (\d\d\d)
    fg: "#ffffff"
    alternatives:
      - regexp: (2\d\d)
        fg: "#00ff00"
      - regexp: (3\d\d)
        fg: "#00ffff"
      - regexp: (4\d\d)
        fg: "#ff0000"
      - regexp: (5\d\d)
        fg: "#ff00ff"

  # complex pattern
  ipv4-address-with-port:
    regexps:
      - regexp: (\d{1,3}(\.\d{1,3}){3})
        fg: "#ffc777"
      - regexp: ((:\d{1,5})?)
        fg: "#ff966c"

```

`patterns` are standard regular expressions. You can highlight any sequence of characters in a string that matches a regular expression. It may consist of several parts (see `ipv4-address-with-port` above). This is convenient if you want different parts of a pattern to have different colors or styles. Think of these complex patterns as little log formats that can be found in any part of a string.

Patterns have priority. Ones with higher priority will be painted earlier. Default priority is 0.

Full pattern example using all available fields:

```yaml
patterns:
  # simple pattern (when you use only "regexp" field)
  # name of a pattern
  ipv4-address:
    # the same fields are used here as in log formats (see above)
    priority: 10
    regexp: (\d{1,3}(\.\d{1,3}){3})
    fg: "#00ff00"
    bg: "#0000ff"
    style: bold
    alternatives:
      - regexp: (1\d{1,2}(\.\d{1,3}){3})
        fg: "#00ff00"
        bg: "#0000ff"
        style: bold
      - regexp: (2\d{1,2}(\.\d{1,3}){3})
        fg: "#ff0000"
        bg: "#0000ff"
        style: underline

  # complex pattern (when you use "regexps" field)
  # the same fields are used here as in log formats (see above)
  # complex pattern are formed from all regexps in the "regexps" list
  # e.g. pattern below will be rendered as (\d{1,3}(\.\d{1,3}){3})((:\d{1,5})?)
  # the main difference from a simple pattern is that you can control
  # the style of the individual parts of the pattern
  ipv4-address-with-port:
    regexps:
      - regexp: (\d{1,3}(\.\d{1,3}){3})
        fg: "#ffc777"
        bg: "#0000ff"
        style: bold
      - regexp: ((:\d{1,5})?)
        fg: "#ff966c"
        bg: "#00ffff"
        style: underline

  # complex patterns are mainly used when you want to build a pattern
  # that builds on other patterns. for example, you want to make a highlighter
  # for the "logfmt" format. an example of "logfmt" log line:
  # ts=2024-02-16T23:00:02.953Z caller=db.go:1619 level=info component=tsdb msg="Deleting..."
  # you can't use log formats (see above) because the structure of "logfmt" is impermanent
  # in such a case, you can describe the base "logfmt" element (xxx=xxx) and look for other
  # existing patterns (date, time, IP address, etc.) on the right side of the equals sign
  logfmt:
    regexps:
      - regexp: ( [^=]+)
        fg: "#ff0000"
      - regexp: (=)
        fg: "#00ff00"
      - regexp: ([^ ]+)
        style: patterns-and-words
```

You can find built-in `patterns` [here](builtins/patterns). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`.

### Words

Configuration example:

```yaml
words:
  good:
    fg: "#52fa8a"
    style: bold
    list:
      - "complete"
      - "enable"
      - "online"
      - "succeed"
      - "success"
      - "successful"
      - "successfully"
      - "true"
      - "valid"

  bad:
    bg: "#f06c62"
    style: underline
    list:
      - "block"
      - "critical"
      - "deny"
      - "disable"
      - "error"
      - "fail"
      - "false"
      - "fatal"
      - "invalid"

  your-word-group:
    bg: "#0b78f1"
    list:
      - "lonzo"
      - "gizmo"
      - "lotek"
      - "toni"
```

`words` are just lists of words that will be colored using values from `fg`, `bg` and `style` fields (see more details about these fields above under [Log formats](#log-formats)). `words` could have been implemented using patterns, if it weren't for one feature.

Words from these lists are used not only literally, but also as [lemmas](https://en.wikipedia.org/wiki/Lemma_(morphology)). It means that by listing the word "complete", you will also highlight the words "completes", "completed" and "completing" in any line. Similarly, if you add the word "sing" to a list, the words "sang" and "sung" will also be highlighted. It works only for the English language.

There are two special word groups: `good` and `bad`. The negation of a word from `good` group will be colored using values from `bad` group and vice versa. For example, if `good` group has the word "complete" then "not completed", "wasn't completed", "cannot be completed" and other negative forms will be colored using values from `bad` word group.

You can find built-in `words` [here](builtins/words). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`.

Acknowledgements
----------------

Thanks to my brother [@emptysad](https://github.com/emptysad) for coming up with the name Logalize and for the logo idea.

Thanks to [@ekivoka](https://github.com/ekivoka) for the help with choosing the design and testing the logo. 

Thanks to the authors of these awesome libraries:

- [cobra](https://github.com/spf13/cobra)
- [golem](https://github.com/aaaton/golem)
- [koanf](https://github.com/knadh/koanf)
- [termenv](https://github.com/muesli/termenv)

[^1]: I made that up
