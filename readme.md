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
- path from `-c/--config` flag
- `.logalize.yaml` in current directory

If more than one configuration file is found, they are merged. The lower the file in the list, the higher its priority.

A configuration file can contain five top-level keys: `formats`, `patterns`, `words`, `themes` and `settings`. In the first three you define what you want to catch and in `themes` you describe how you want to colorize it. `settings` is a way to set some options if you don't want to pass them as flags.

### Log formats

Configuration example:

```yaml
formats:
  kuvaq:
    - regexp: (\d{1,3}(\.\d{1,3}){3} )
      name: ip-address
    - regexp: (- )
      name: dash
    - regexp: ("[^"]+" )
      name: string
    - regexp: (\d\d\d)
      name: http-status-code
      alternatives:
        - regexp: (2\d\d)
          name: 2xx
        - regexp: (3\d\d)
          name: 3xx
        - regexp: (4\d\d)
          name: 4xx
        - regexp: (5\d\d)
          name: 5xx
```

`formats` describe complete log formats. A line must match a format completely to be colored. For example, the full regular expression for the "kuvaq" format above will look like this:\
`^(\d{1,3}(\.\d{1,3}){3} )(- )("[^"]+" )(\d\d\d)$`

Only lines below will match this format:
- `127.0.0.1 - "menetekel" 200`
- `7.7.7.7 - "m" 404`

But not these:
- `127.0.0.1 - "menetekel" 503 lower ascension station`
- `Upper ascension station 127.0.0.1 - "menetekel" 403`
- `127.0.0.1 - "menetekel" 404000`

For an overview of regular expression syntax, see the [regexp/syntax](https://pkg.go.dev/regexp/syntax) package.

Full log format example using all available fields:

```yaml
formats:
  # name of a log format
  elysium:
    # regexp must begin with an opening parenthesis `(`
    # and it must end with a paired closing parenthesis `)`
    # regexp can't be empty `()`
    # that is, your regexp must be within one capture group
    # and contain a valid regular expression
    - regexp: (\d\d\d )
      # name of the capture group
      # it will be used to assign colors and style
      # later in "themes" section (see below)
      name: capture-group-name
      # alternatives are useful when you have general regular expression
      # but you want different colors for some specific subset of cases
      # within this regular expression
      # a common example is HTTP status code
      alternatives:
        # every regexp here has the same "name" field
        # but no "alternatives" field
        - regexp: (2\d\d )
          name: 2xx
        - regexp: (4\d\d )
          name: 4xx
    # each next regexp is added to the previous one
    # and together they form a complete regexp for the whole string
    - regexp: (--- )
      name: dashes
    - regexp: ([[:xdigit:]]{32})
      name: hash
    # full regexp for this whole example is:
    # ^(\d\d\d )(--- )([[:xdigit:]]{32})$
```

You can find built-in `formats` [here](builtins/logformats). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`. See [Customization](#customization) section below for more details.

### Patterns

Configuration example:

```yaml
patterns:
  # simple patterns (one regexp)
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')

  number:
    regexp: (\d+)

  # complex pattern (built from a list of regexps)
  ipv4-address-with-port:
    regexps:
      - regexp: (\d{1,3}(\.\d{1,3}){3})
        name: address
      - regexp: ((:\d{1,5})?)
        name: port

```

`patterns` are standard regular expressions. You can highlight any sequence of characters in a string that matches a regular expression. It may consist of several parts (see `ipv4-address-with-port` above). This is convenient if you want different parts of a pattern to have different colors or styles. Think of these complex patterns as little log formats that can be found in any part of a string.

Patterns have priority. Ones with higher priority will be painted earlier. Default priority is 0. The priorities of the built-in patterns are between -100 and 100.

Full pattern example using all available fields:

```yaml
patterns:
  # simple pattern (when you use only "regexp" field)
  # name of a pattern
  http-status-code:
    # patterns with higher priority will be painted earlier
    # default priority is 0
    priority: 10
    # the same fields are used here as in log formats (see above)
    regexp: (\d\d\d)
    # patterns can have alternatives just like in log formats
    alternatives:
      - regexp: (2\d\d)
        name: 2xx
      - regexp: (3\d\d)
        name: 3xx
      - regexp: (4\d\d)
        name: 4xx
      - regexp: (5\d\d)
        name: 5xx

  # complex pattern (when you use "regexps" field)
  # the same fields are used here as in log formats (see above)
  # complex pattern are formed from all regexps in the "regexps" list
  # e.g. pattern below will be rendered as (\d{1,3}(\.\d{1,3}){3})((:\d{1,5})?)
  # the main difference from a simple pattern is that you can control
  # the style of the individual parts of the pattern
  ipv4-address-with-port:
    regexps:
      - regexp: (\d{1,3}(\.\d{1,3}){3})
        name: address
      - regexp: ((:\d{1,5})?)
        name: port

  # complex patterns are mainly used when you want to build a pattern
  # that builds on other patterns. for example, you want to make a highlighter
  # for the "logfmt" format. an example of "logfmt" log line:
  # ts=2024-02-16T23:00:02.953Z caller=db.go:16 level=info component=tsdb msg="Deleting..."
  # you can't use log formats (see above) because the structure of "logfmt" is variable.
  # in such a case, you can describe the base "logfmt" element (xxx=xxx) and look for other
  # existing patterns (date, time, IP address, etc.) on the right side of the equals sign
  # (see how to accomplish this below in the "themes" section)
  logfmt:
    regexps:
      - regexp: ( [^=]+)
        name: key
      - regexp: (=)
        name: equal-sign
      - regexp: ([^ ]+)
        name: value
```

You can find built-in `patterns` [here](builtins/patterns). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`. See [Customization](#customization) section below for more details.

### Words

Configuration example:

```yaml
words:
  good:
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
    - "lonzo"
    - "gizmo"
    - "lotek"
    - "toni"
```

`words` are just lists of words that will be colored according to your theme (see below). `words` could have been implemented using patterns, if it weren't for one feature.

Words from these lists are used not only literally, but also as [lemmas](https://en.wikipedia.org/wiki/Lemma_(morphology)). It means that by listing the word "complete", you will also highlight the words "completes", "completed" and "completing" in any line. Similarly, if you add the word "sing" to a list, the words "sang" and "sung" will also be highlighted. It works only for the English language.

There are two special word groups: `good` and `bad`. The negation of a word from `good` group will be colored using values from `bad` group and vice versa. For example, if `good` group has the word "complete" then "not completed", "wasn't completed", "cannot be completed" and other negative forms will be colored using values from `bad` word group.

You can find built-in `words` [here](builtins/words). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`. See [Customization](#customization) section below for more details.

### Themes

Configuration example:

```yaml
themes:
  # name of a theme
  utopia:
    formats:
      kuvaq:
        ip-address:
          fg: "#f5ce42"
        dash:
          bg: "#807e7a"
          style: bold
        string:
          fg: "#9ddb56"
          bg: "#f5ce42"
        http-status-code:
          default:
            fg: "#ffffff"
          2xx:
            fg: "#00ff00"
            style: bold
          3xx:
            fg: "#00ffff"
            style: bold
          4xx:
            fg: "#ff0000"
            style: bold
          5xx:
            fg: "#ff00ff"
            style: bold

      elysium:
        capture-group-name:
          default:
            fg: "#ffffff"
          2xx:
            fg: "#00ff00"
            style: bold
          4xx:
            fg: "#ff0000"
            style: bold
        dashes:
          bg: "#807e7a"
          style: bold
        hash:
          fg: "#9ddb56"
          bg: "#f5ce42"

    patterns:
      string:
        fg: "#00ff00"

      number:
        bg: "#00ffff"
        style: bold

      http-status-code:
        default:
          fg: "#ffffff"
        2xx:
          fg: "#00ff00"
        3xx:
          fg: "#00ffff"
        4xx:
          fg: "#ff0000"
        5xx:
          fg: "#ff00ff"

      ipv4-address-with-port:
        address:
          fg: "#ffc777"
        port:
          fg: "#ff966c"

      logfmt:
        key:
          fg: "#ff0000"
        equal-sign:
          fg: "#00ff00"
        value:
          style: patterns-and-words

    words:
      good:
        fg: "#52fa8a"
        style: bold

      bad:
        fg: "#f06c62"
        style: bold

      your-word-group:
        bg: "#0b78f1"

  # another theme
  menetekel:
    formats:
      # . . .
    patterns:
      # . . .
    words:
      # . . .
```

`themes` is the place where you apply colors and style to log formats, patterns and word groups you defined earlier. Every capture group can be colorized using `fg`, `bg` and `style` fields.

`fg` and `bg` are foreground and background colors correspondingly. They can be a hex value like `#ff0000` or a number between 0 and 255 for ANSI colors.

`style` field can be set to one of 7 regular styles: `bold`, `faint`, `italic`, `underline`, `overline`, `crossout` and `reverse`. There are also three special styles:
- `patterns` - use highlighting from `patterns` section (see above)
- `words` - use highlighting from `words` section (see above)
- `patterns-and-words` - use highlighting from `patterns` and `words` sections

You can get a list of all available themes with `-T/--list-themes` flag and set it with `-t/--theme` flag or `theme` key in `settings` section (see below).

### Settings

Configuration example:

```yaml
settings:
  theme: "utopia"

  no-builtin-logformats: false
  no-builtin-patterns: false
  no-builtin-words: false
  no-builtins: true

  only-logformats: false
  only-patterns: false
  only-words: false
```

Here you can set some options that are equivalent to command line flags. `theme` is the same as `--theme` flag and so on. Only the flags from the example above are supported.

Customization
-------------

#### I want to change one of the colors in one of the built-in themes

1. Suppose you don't like the color of the "uuid" pattern in the "tokyonight" theme
2. Just override it in your `logalize.yaml` like this:
```yaml
# . . .
themes:
  tokyonight:
    patterns:
      uuid:
        fg: "#ff0000"
        bg: "#00ff00"
        style: bold
# . . .
```

#### I want to define and use my own theme

1. Get current configuration to use it as an example:
```sh
logalize --print-config > example.logalize.yaml
```
2. Copy one of the built-in themes from `example.logalize.yaml` to your `logalize.yaml`, rename it and change it the way you like it:
```yaml
# . . .
themes:
  your-theme-name:
    formats:
      # . . .
    patterns:
      # . . .
    words:
      # . . .
# . . .
```
3. Set the theme in your `logalize.yaml` in `settings` section:
```yaml
# . . .
settings:
  theme: your-theme-name
  # . . .
# . . .
```
4. ... or set the theme using `--theme` flag:
```sh
cat logs | logalize --theme "your-theme-name"
```

#### I want to disable all builtins and use only data from my own `logalize.yaml`

1. Define any log formats, patterns, word groups and themes in your `logalize.yaml`
2. Disable builtins in `settings` section:
```yaml
# . . .
settings:
  # . . .
  no-builtins: true
  # . . .
# . . .
```
3. ... or use `-N/--no-builtins` flag:
```sh
cat logs | logalize -N
```

Acknowledgements
----------------

Thanks to my brother [@emptysad](https://github.com/emptysad) for coming up with the name Logalize and for the logo idea.

Thanks to [@ekivoka](https://github.com/ekivoka) for the help with choosing the design and testing the logo. 

Thanks to the authors of these awesome libraries:

- [spf13/cobra](https://github.com/spf13/cobra)
- [aaaton/golem](https://github.com/aaaton/golem)
- [knadh/koanf](https://github.com/knadh/koanf)
- [muesli/termenv](https://github.com/muesli/termenv)
- [goccy/go-yaml](https://github.com/goccy/go-yaml)

[^1]: I made that up
