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

```sh
go install github.com/deponian/logalize@latest
```

How it works
------------

Logalize reads one line at a time and then checks if it matches one of log formats (`formats`), general regular expressions (`patterns`) or plain English words and their [inflected](https://en.wikipedia.org/wiki/Inflection) forms (`words`). See configuration below for more details.

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
    - pattern: (\d{1,3}(\.\d{1,3}){3} )
      fg: "#f5ce42"
    - pattern: (- )
      bg: "#807e7a"
      style: bold
    - pattern: ("[^"]+" )
      fg: "#9ddb56"
      bg: "#f5ce42"
    - pattern: (\d\d\d)
      fg: "#ffffff"
      alternatives:
        - pattern: (2\d\d)
          fg: "#00ff00"
          style: bold
        - pattern: (3\d\d)
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

For an overview of the pattern syntax, see the [regexp/syntax](https://pkg.go.dev/regexp/syntax) package.

Full log format example using all available fields:

```yaml
formats:
  # name of a log format
  elysium:
    # pattern must begin with an opening parenthesis `(`
    # and it must end with a closing parenthesis `)`
    # pattern can't be empty `()`
    - pattern: (\d\d\d )
      # color can be a hex value like #ff0000
      # or a number between 0 and 255 for ANSI colors
      fg: "#00ff00"
      bg: "#0000ff"
      # available styles are bold, faint, italic,
      # underline, overline, crossout, reverse
      style: bold
      # alternatives are useful when you have general regular expression
      # but you want different colors for some specific subset of cases
      # within this regular expression
      # a common example is HTTP status code
      alternatives:
        # every pattern here has the same "fg", "bg" and "style" fields
        # but no "alternatives" field
        - pattern: (2\d\d )
          fg: "#00ff00"
          bg: "#0000ff"
          style: bold
        - pattern: (4\d\d )
          fg: "#ff0000"
          bg: "#0000ff"
          style: underline
    # each next pattern is added to the previous one
    # and together they form a complete pattern for the whole string
    - pattern: (--- )
      # . . . . .
    - pattern: ([[:xdigit:]]{32})
      # . . . . .
    # full pattern for this whole example is:
    # ^(\d\d\d )(--- )([[:xdigit:]]{32})$
```

You can find built-in `formats` [here](builtins/logformats). If you want to customize them or turn them off completely, overwrite the corresponding values in your `logalize.yaml`.

### Patterns

Configuration example:

```yaml
patterns:
  string:
    priority: 500
    pattern: ("[^"]+"|'[^']+')
    fg: "#00ff00"

  number:
    pattern: (\d+)
    bg: "#00ffff"
    style: bold

  http-status-code:
    pattern: (\d\d\d)
    fg: "#ffffff"
    alternatives:
      - pattern: (2\d\d)
        fg: "#00ff00"
      - pattern: (3\d\d)
        fg: "#00ffff"
      - pattern: (4\d\d)
        fg: "#ff0000"
      - pattern: (5\d\d)
        fg: "#ff00ff"
```

`patterns` are standard regular expressions. You can highlight any sequence of characters in a string that matches a regular expression. The same `fg`, `bg`, `style` and `alternatives` fields are used here, see more details above under [Log formats](#log-formats). The only new field here is `priority`. Patterns with higher priority will be painted earlier. Default priority is 0.

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

- [go-arg](https://github.com/alexflint/go-arg)
- [golem](https://github.com/aaaton/golem)
- [koanf](https://github.com/knadh/koanf)
- [termenv](https://github.com/muesli/termenv)

[^1]: I made that up
