<picture>
  <source media="(prefers-color-scheme: dark)" srcset="images/avif/logo-dark.avif">
  <source media="(prefers-color-scheme: light)" srcset="images/avif/logo-light.avif">
  <img alt="Screenshot" src="images/avif/logo-light.avif">
</picture>

93% of all logs are not colored[^1]. It's sad. Maybe even illegal. It's time to logalize them. **Logalize** is a log colorizer like [colorize](https://github.com/raszi/colorize) and [ccze](https://github.com/cornet/ccze). But it's faster and, much more importantly, it's extensible. No more hardcoded templates for logs and keywords. Logalize is fully customizable via `logalize.yaml` where you can define your formats, keyword patterns and more.

<p align="center">
  <a href="https://github.com/deponian/logalize/actions"><img src="https://github.com/deponian/logalize/actions/workflows/tests.yml/badge.svg" alt="Build Status"></a>
  <a href="https://codecov.io/gh/deponian/logalize"><img src="https://codecov.io/gh/deponian/logalize/graph/badge.svg?token=8NJ4ZC8COT"/></a>
  <a href="https://goreportcard.com/report/github.com/deponian/logalize"><img src="https://goreportcard.com/badge/github.com/deponian/logalize" alt="Go Report Card"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/deponian/logalize/releases/latest"><img src="https://img.shields.io/github/v/release/deponian/logalize" alt="Github Release"></a>
  <a href="https://ko-fi.com/M4M41I8ECW"><img src="https://img.shields.io/badge/Buy_me_a_coffee-FF5E5B?logo=kofi&logoColor=white" alt="Ko-fi"></a>
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

Download DEB, RPM, Arch Linux packages, or the binary for your architecture, from [releases](https://github.com/deponian/logalize/releases/latest).

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
or install from the AUR:
```sh
# to install the precompiled binary
yay -S logalize-bin

# to compile it on your machine
yay -S logalize
```

**macOS:**

```sh
brew install deponian/tap/logalize
```

**OS-agnostic:**

Use `go install` if you already have `$GOPATH/bin` in your `$PATH`:

```sh
go install github.com/deponian/logalize@latest
```

How it works
------------

Logalize reads one line from stdin at a time and then checks if it matches one of the formats (`formats`), general regular expressions (`patterns`), or plain English words and their [inflected](https://en.wikipedia.org/wiki/Inflection) forms (`words`). See configuration below for more details.

Simplified version of the main loop:
1. Read a line from stdin.
2. Strip all [ANSI escape sequences](https://en.wikipedia.org/wiki/ANSI_escape_code) (see the `settings` section below).
3. If the entire line matches one of the `formats`, print the colored line and go to step 1; otherwise, go to step 4.
4. Find and color all `patterns` in the line, then go to step 5.
5. Find and color all `words`, print the colored line, and go to step 1.

Configuration
-------------

Logalize looks for configuration files in these places:
- `/etc/logalize/logalize.yaml`
- `~/.config/logalize/logalize.yaml`
- `.logalize.yaml` in the current directory
- path(s) from `-c/--config` flag (can be repeated)

If more than one configuration file is found, they are merged. The lower the file in the list, the higher its priority.

A configuration file can contain five top-level keys: `formats`, `patterns`, `words`, `themes`, and `settings`. In the first three, you define what you want to match, and in `themes` you describe how you want to colorize them. `settings` lets you set options if you don't want to pass them as flags.

### Formats

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

`formats` describe complete formats. A line must match a format completely to be colored. For example, the full regular expression for the "kuvaq" format above looks like this:\
`^(\d{1,3}(\.\d{1,3}){3} )(- )("[^"]+" )(\d\d\d)$`

Only the lines below will match this format:
- `127.0.0.1 - "menetekel" 200`
- `7.7.7.7 - "m" 404`

But not these:
- `127.0.0.1 - "menetekel" 503 lower ascension station`
- `Upper ascension station 127.0.0.1 - "menetekel" 403`
- `127.0.0.1 - "menetekel" 404000`

For an overview of regular expression syntax, see the [regexp/syntax](https://pkg.go.dev/regexp/syntax) package.

Full format example using all available fields:

```yaml
formats:
  # Name of a format
  elysium:
    # Regexp must begin with an opening parenthesis `(`
    # and it must end with a paired closing parenthesis `)`
    # Regexp can't be empty `()`
    # That is, your regexp must be within one capturing group
    # and contain a valid regular expression.
    - regexp: (\d\d\d )
      # Name of the capturing group.
      # It will be used to assign colors and style
      # later in the "themes" section (see below).
      name: capgroup-name
      # Alternatives are useful when you have a general regular expression
      # but want different colors for a specific subset of cases
      # within this regular expression.
      # A common example is an HTTP status code.
      alternatives:
        # Every regexp here has the same "name" field
        # and no "alternatives" field.
        - regexp: (2\d\d )
          name: 2xx
        - regexp: (4\d\d )
          name: 4xx
    # Each subsequent regexp is added to the previous one,
    # and together they form a complete regexp for the whole string
    - regexp: (--- )
      name: dashes
    - regexp: ([[:xdigit:]]{32})
      name: hash
    # Full regexp for this whole example:
    # ^(\d\d\d )(--- )([[:xdigit:]]{32})$
```

You can find built-in `formats` [here](builtins/formats). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`. See the [Customization](#customization) section below for more details.

### Patterns

Configuration example:

```yaml
patterns:
  # Simple patterns (one regexp)
  string:
    priority: 500
    regexp: ("[^"]+"|'[^']+')

  number:
    regexp: (\d+)

  # Complex pattern (built from a list of regexps)
  ipv4-address-with-port:
    regexps:
      - regexp: (\d{1,3}(\.\d{1,3}){3})
        name: address
      - regexp: ((:\d{1,5})?)
        name: port

```

`patterns` are standard regular expressions. You can highlight any sequence of characters in a string that matches a regular expression. A pattern may consist of several parts (see `ipv4-address-with-port` above). This is convenient if you want different parts of a pattern to have different colors or styles. Think of these complex patterns as small formats that can be found in any part of a string.

Patterns have priority. Those with higher priority will be painted earlier. The default priority is 0. The priorities of the built-in patterns are between -100 and 100.

Full pattern example using all available fields:

```yaml
patterns:
  # Simple pattern (when you use only "regexp" field)
  # Name of a pattern
  http-status-code:
    # Patterns with higher priority will be painted earlier;
    # the default priority is 0.
    priority: 10
    # The same fields are used here as in formats (see above).
    regexp: (\d\d\d)
    # Patterns can have alternatives just like in formats.
    alternatives:
      - regexp: (2\d\d)
        name: 2xx
      - regexp: (3\d\d)
        name: 3xx
      - regexp: (4\d\d)
        name: 4xx
      - regexp: (5\d\d)
        name: 5xx

  # Complex pattern (when you use the "regexps" field)
  # The same fields are used here as in formats (see above).
  # Complex patterns are formed from all regexps in the "regexps" list.
  # For example, the pattern below will be rendered as (\d{1,3}(\.\d{1,3}){3})((:\d{1,5})?)
  # The main difference from a simple pattern is that you can control
  # the style of the individual parts of the pattern.
  ipv4-address-with-port:
    regexps:
      - regexp: (\d{1,3}(\.\d{1,3}){3})
        name: address
      - regexp: ((:\d{1,5})?)
        name: port

  # Complex patterns are mainly used when you want to build a pattern
  # that builds on other patterns. For example, you may want to make a highlighter
  # for the "logfmt" format. An example of a "logfmt" log line:
  # ts=2024-02-16T23:00:02.953Z caller=db.go:16 level=info component=tsdb msg="Deleting..."
  # You can't use formats (see above) because the structure of "logfmt" is variable.
  # In such a case, you can describe the base "logfmt" element (xxx=xxx) and look for other
  # existing patterns (date, time, IP address, etc.) on the right side of the equals sign
  # (see how to accomplish this below in the "themes" section).
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

`words` are just lists of words that will be colored according to your theme (see below). `words` could have been implemented using patterns, if not for one feature.

Words from these lists are used not only literally but also as [lemmas](https://en.wikipedia.org/wiki/Lemma_(morphology)). This means that by listing the word "complete", you will also highlight the words "completes", "completed", and "completing" in any line. Similarly, if you add the word "sing" to a list, the words "sang" and "sung" will also be highlighted. This works only for the English language.

There are two special word groups: `good` and `bad`. The negation of a word from the `good` group will be colored using values from the `bad` group, and vice versa. For example, if the `good` group has the word "complete", then "not completed", "wasn't completed", "cannot be completed", and other negative forms will be colored using values from the `bad` group.

You can find built-in `words` [here](builtins/words). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`. See [Customization](#customization) section below for more details.

### Themes

Configuration example:

```yaml
themes:
  # Name of a theme
  utopia:
    # The default color is used for anything that does not fall into
    # a format, pattern, or word. If you don't specify a default color,
    # the normal color of your terminal will be used.
    #default:
    #  fg: "#ff0000"
    #  bg: "#00ff00"
    #  style: "bold"

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
          # default colors and style if none of the alternatives are matched
          fg: "#ffffff"
          bg: "#f5ce42"
          style: bold

          # alternatives
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
        capgroup-name:
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

`themes` is the place where you apply colors and style to formats, patterns, and word groups you defined earlier (or to the built-in ones). Every capturing group can be colorized using the `fg`, `bg`, and `style` fields. There is also a special field called `link-to`. See the next section for details.

`fg` and `bg` are foreground and background colors, respectively. They can be hex values like `#ff0000` or numbers between 0 and 255 for ANSI colors.

The `style` field can be set to one of seven regular styles: `bold`, `faint`, `italic`, `underline`, `overline`, `crossout`, and `reverse`. There are also three special styles:
- `patterns` - use highlighting from the `patterns` section (see above)
- `words` - use highlighting from the `words` section (see above)
- `patterns-and-words` - use highlighting from both the `patterns` and `words` sections

You can get a list of all available themes with the `-T/--list-themes` flag and set it with the `-t/--theme` flag or the `theme` key in the `settings` section (see below).

#### Linking styles between capturing groups (`link-to`)

Use `link-to` to reuse the exact color and style of another capturing group in the same format or in the same complex pattern. The link is resolved at runtime for every line:

- If the target group has alternatives and one of them matched, the linked group inherits that alternative’s `fg`/`bg`/`style`.
- Otherwise, it inherits the target group’s default `fg`/`bg`/`style`.

Configuration example:

```yaml
formats:
  organon:
    - regexp: ([IWEF][0-9]{4} )
      name: log-level
      alternatives:
        - regexp: (I[0-9]{4} )
          name: info
        - regexp: (W[0-9]{4} )
          name: warning
        - regexp: (E[0-9]{4} )
          name: error
        - regexp: (F[0-9]{4} )
          name: fatal
    - regexp: (--- )
      name: separator
    - regexp: ([0-9]+ )
      name: thread-id
    - regexp: (.*)
      name: message

themes:
  paradox:
    formats:
      organon:
        log-level:
          info:
            fg: "#82aaff"
            style: bold
          warning:
            fg: "#ffc777"
            style: bold
          error:
            fg: "#ff757f"
            style: bold
          fatal:
            fg: "#c53b53"
            style: bold
        separator:
          fg: "#456789"
        thread-id:
          link-to: log-level   # inherits the effective style of "log-level"
        message:
          style: patterns-and-words
```

In the example above, when `log-level` matches `warning`, `thread-id` is colored with the same foreground and style as the `warning` alternative of `log-level` (here: `#ffc777` + `bold`). The inheritance is dynamic per line.

**Notes**
- `link-to` works within a single format or within a single complex pattern (it cannot link across different formats or patterns).
- Links can be chained (`A -> B -> C`); cycles are invalid and will be rejected during initialization.
- When `link-to` is present, the linked style takes precedence over any `fg`/`bg`/`style` set directly on that group.

### Settings

Configuration example:

```yaml
settings:
  theme: "utopia"

  no-builtin-formats: false
  no-builtin-patterns: false
  no-builtin-words: false
  no-builtins: true

  only-formats: false
  only-patterns: false
  only-words: false

  no-ansi-escape-sequences-stripping: false

  debug: false
  dry-run: false
```

Here you can set options equivalent to command-line flags. `theme` is the same as the `--theme` flag, and so on. Only the flags from the example above are supported.

Customization
-------------

#### I want to change one of the colors in one of the built-in themes

1. Suppose you don't like the color of the `uuid` pattern in the `tokyonight-dark` theme.
2. Just override it in your `logalize.yaml` like this:
```yaml
# . . .
themes:
  tokyonight-dark:
    patterns:
      uuid:
        fg: "#ff0000"
        bg: "#00ff00"
        style: bold
# . . .
```

#### I want to change the color for everything that doesn't fall into a format, pattern, or word (i.e., just plain text)

1. Suppose you use the `tokyonight-dark` theme.
2. Just set the default color in your `logalize.yaml` like this:
```yaml
# . . .
themes:
  tokyonight-dark:
    default:
      fg: "#ff0000"
      bg: "#00ff00"
      style: bold
# . . .
```

#### I want to define and use my own theme

1. Use one of the existing themes as an example. Pick one [here](themes/).
2. Copy it to your `logalize.yaml`, rename it, and change it the way you like:
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
3. Test it using logs from the [testlogs](testlogs/) directory:
```sh
cat testlogs/* | logalize --theme your-theme-name
```
4. Set the theme in your `logalize.yaml` in the `settings` section when it's ready:
```yaml
# . . .
settings:
  theme: your-theme-name
  # . . .
# . . .
```
5. If it's a well-known theme and you think others might benefit from it, feel free to open a PR asking to add that theme as one of the built-in themes.

#### I want to disable all builtins and use only data from my own `logalize.yaml`

1. Define any formats, patterns, word groups, and themes in your `logalize.yaml`.
2. Disable builtins in the `settings` section:
```yaml
# . . .
settings:
  # . . .
  no-builtins: true
  # . . .
# . . .
```
3. ... or use the `-N/--no-builtins` flag:
```sh
cat logs | logalize -N
```

Acknowledgements
----------------

Thanks to my brother [@emptysad](https://github.com/emptysad) for coming up with the name Logalize and for the logo idea.

Thanks to [@ekivoka](https://github.com/ekivoka) for her help with choosing the design and testing the logo.

Thanks to [@antiflasher](https://github.com/antiflasher) and [@romashamin](https://github.com/romashamin) for their help with AVIF image conversion.

Thanks to the authors of these awesome projects:

- [spf13/cobra](https://github.com/spf13/cobra)
- [aaaton/golem](https://github.com/aaaton/golem)
- [knadh/koanf](https://github.com/knadh/koanf)
- [muesli/termenv](https://github.com/muesli/termenv)
- [goccy/go-yaml](https://github.com/goccy/go-yaml)
- [chalk/ansi-regex](https://github.com/chalk/ansi-regex)

[^1]: I made that up
