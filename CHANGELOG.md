# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0] - 2025-07-23

### Added

- Add "syslog-rfc3164" log format
- Add ability to repeat -c/--config flag

### Changed

- Update "nginx-combined" log format
- Update "nginx-ingress-controller" log format

## [0.5.0] - 2025-07-12

### Added

- Add ability to save original colors of input data (see -s/--no-ansi-escape-sequences-stripping flag and https://github.com/deponian/logalize/issues/6)
- Add debug output mode (see -d/--debug flag)

### Fixed

- Update regular expression for negated words

## [0.4.8] - 2025-06-01

### Fixed

- Properly handle \r in the input data (see https://github.com/deponian/logalize/issues/7)

## [0.4.7] - 2025-05-28

### Fixed

- Strip ANSI escape sequences with ":"-separated colors (see https://github.com/deponian/logalize/issues/6)

## [0.4.6] - 2025-05-26

### Fixed

- Don't alter the input in any way when user set --dry-run flag
- Reflect changes from v0.4.5 in the man page

## [0.4.5] - 2025-05-26

### Changed

- Strip ANSI escape sequences form the input by default

### Added

- Add -s/--no-ansi-escape-sequences-stripping flag to get pre v0.4.5 behavior

## [0.4.4] - 2025-05-20

### Added

- Add "debug" to builtin words
- Add default color option

## [0.4.3] - 2025-01-20

### Added

- Add a clarification to the example configuration file

### Fixed

- Improve error handling in regular expressions

## [0.4.2] - 2024-11-24

### Added

- Add "gruvbox-dark" theme

### Changed

- Rename "tokyonight" theme to "tokyonight-dark"
- Rename main directory from "pkg" to "src"
- Respect NO_COLOR environmental variable

### Fixed

- Colorize output even for non TTYs (like pipes)

## [0.4.1] - 2024-10-23

### Added

- Add "crit" to the list of bad words

### Changed

- Tune colors in "tokyonight" theme
- Switch to go v1.23.2 and update modules

### Fixed

- Bring back --dry-run and --config flags

## [0.4.0] - 2024-10-22

### Added

- (!) Add theme support (see the new configuration in README.md)
- Add -T/--list-themes flag
- Add -C/--print-config flag
- Add "failure" to the list of "bad" words
- Add mask detection for IPv4 address pattern

### Changed

- Split config and options; add "settings" section
- Replace embed.FS with fs.FS to simplify testing

### Fixed

- Check theme availability during config initialization
- Split default paths and .logalize.yaml in current directory

## [0.3.0] - 2024-08-17

### Added

- Add complex patterns (`regexps` key)
- Add "redis" log format
- Add "logfmt" pattern
- Add "duration" pattern
- Add -L, -P, -W, -N, -l, -p, -w, -n flags; -b flag replaced old -p flag

### Changed

- Tune the colors of built-in log formats and patterns
- Use non-capturing groups for built-in logformats and patterns
- Colorize quotation marks in "logfmt-string" pattern
- Update modules and go to v1.22.5
- Add :port to "ipv6-address" pattern
- Update git-cliff configuration

### Fixed

- Improve datetime and network patterns
- Don't detect duration-like sequences inside words

## [0.2.0] - 2024-07-25

### Added

- Add "--print-builtins" flag
- Add "info" word group
- Add "klog" log format
- Add "patterns", "words" and "patterns-and-words" color styles
- Add datetime patterns
- Add MAC address and UUID patterns

### Changed

- Change cert-manager example log
- Use global variables in init* functions instead of returning the value
- Load built-in configs recursively

### Fixed

- Check lowercased words
- Change colors for nginx-combined and nginx-ingress-controller logging formats
- Change checking process for log formats and word groups
- Rectify "print-builtins" after changes in d5fa7fd

## [0.1.2] - 2024-06-19

### Added

- Add man pages generator
- Add man pages to deb, rpm and Arch Linux packages
- Add completions generator
- Add Makefile
- Add "changelog" target to Makefile

### Changed

- Use cobra instead of go-arg
- Update mangen dependencies
- Update social preview image
- Update test and coverage targets in Makefile

### Fixed

- Don't add newline at the end of man page
- Return an error if logalize is started with an argument
- Fix comments on exported types
- Don't reset EXTRA_LDFLAGS and EXTRA_GOFLAGS env variables in Makefile

## [0.1.1] - 2024-05-19

### Added

- Add git-cliff config
- Add .logalize.yaml as configuration example

### Fixed

- Print version as "version (commit) date"

[0.6.0]: https://github.com/deponian/logalize/compare/v0.5.0..v0.6.0
[0.5.0]: https://github.com/deponian/logalize/compare/v0.4.8..v0.5.0
[0.4.8]: https://github.com/deponian/logalize/compare/v0.4.7..v0.4.8
[0.4.7]: https://github.com/deponian/logalize/compare/v0.4.6..v0.4.7
[0.4.6]: https://github.com/deponian/logalize/compare/v0.4.5..v0.4.6
[0.4.5]: https://github.com/deponian/logalize/compare/v0.4.4..v0.4.5
[0.4.4]: https://github.com/deponian/logalize/compare/v0.4.3..v0.4.4
[0.4.3]: https://github.com/deponian/logalize/compare/v0.4.2..v0.4.3
[0.4.2]: https://github.com/deponian/logalize/compare/v0.4.1..v0.4.2
[0.4.1]: https://github.com/deponian/logalize/compare/v0.4.0..v0.4.1
[0.4.0]: https://github.com/deponian/logalize/compare/v0.3.0..v0.4.0
[0.3.0]: https://github.com/deponian/logalize/compare/v0.2.0..v0.3.0
[0.2.0]: https://github.com/deponian/logalize/compare/v0.1.2..v0.2.0
[0.1.2]: https://github.com/deponian/logalize/compare/v0.1.1..v0.1.2
[0.1.1]: https://github.com/deponian/logalize/compare/v0.1.0..v0.1.1

<!-- generated by git-cliff -->
