# general vars
app := logalize

ifdef VERSION
  COMMIT := $(shell git rev-list -1 --abbrev-commit v$(VERSION))
else
  VERSION := latest
  COMMIT := $(shell git rev-parse --short HEAD)
endif

DATE := $(shell git show -s --date=format:'%Y-%m-%d' --format=%cd $(COMMIT))

# building vars
EXTRA_LDFLAGS ?=
ldflags       := -s -w
ldflags       += -X github.com/deponian/logalize/cmd/logalize.version=$(VERSION)
ldflags       += -X github.com/deponian/logalize/cmd/logalize.commit=$(COMMIT)
ldflags       += -X github.com/deponian/logalize/cmd/logalize.date=$(DATE)
ldflags       += $(EXTRA_LDFLAGS)

EXTRA_GOFLAGS ?=
goflags       := -trimpath
goflags       += $(EXTRA_GOFLAGS)

CGO_ENABLED ?= 1

src := $(shell find . -type f -name '*.go' -print) go.mod go.sum

build_bindir  := ./dist
build_compdir := ./completions
build_mandir  := ./manpages

# installation vars
DESTDIR :=
PREFIX  := /usr/local
install_bindir  := $(PREFIX)/bin
install_datadir := $(PREFIX)/share
install_mandir  := $(install_datadir)/man

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /'
	@echo 'Default values of variables:'
	@echo "  VERSION: $(VERSION)"
	@echo "  COMMIT: $(COMMIT)"
	@echo "  DATE: $(DATE)"
	@echo "  LDFLAGS: $(ldflags)"
	@echo "  EXTRA_LDFLAGS: $(EXTRA_LDFLAGS)"
	@echo "  GOFLAGS: $(goflags)"
	@echo "  CGO_ENABLED: $(CGO_ENABLED)"
	@echo "  DESTDIR: $(DESTDIR)"
	@echo "  PREFIX: $(PREFIX)"

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./cmd/...
	go fmt ./internal/...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	rm -rf ./internal/highlighter/builtins
	cp -r builtins ./internal/highlighter
	go mod verify
	go vet ./internal/... ./cmd/...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./internal/... ./cmd/...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./internal/... ./cmd/...

## test: run all tests
.PHONY: test
test:
	rm -rf ./internal/highlighter/builtins
	cp -r builtins ./internal/highlighter
	go test -race -coverprofile=coverage.out ./internal/...

## coverage-func: run all tests and display coverage with "-func"
.PHONY: coverage-func
coverage-func: test
	go tool cover -func=coverage.out

## coverage-html: run all tests and display coverage with "-html"
.PHONY: coverage-html
coverage-html: test
	go tool cover -html=coverage.out

## changelog: generate new changelog
.PHONY: changelog
changelog:
	@git cliff -c .cliff.toml --bump -u -s header

## build: build the application
.PHONY: build
build: $(build_bindir)/$(VERSION)/$(app)

$(build_bindir)/$(VERSION)/$(app): $(src)
	mkdir -p $(build_bindir)/$(VERSION)
	CGO_ENABLED=$(CGO_ENABLED) go build $(goflags) -ldflags '$(ldflags)' -o $(build_bindir)/$(VERSION)/$(app)

## clean: delete all compiled/generated files
.PHONY: clean
clean:
	rm -rf $(build_bindir)
	rm -rf $(build_compdir)
	rm -rf $(build_mandir)

## completions: generate bash, fish and zsh completion files
.PHONY: completions
completions:
	mkdir -p $(build_compdir)
	go run cmd/compgen/main.go "bash" > $(build_compdir)/$(app).bash
	go run cmd/compgen/main.go "fish" > $(build_compdir)/$(app).fish
	go run cmd/compgen/main.go "zsh" > $(build_compdir)/$(app).zsh

## manpage: generate manpage
.PHONY: manpage
manpage:
	mkdir -p $(build_mandir)
	go run ./cmd/mangen/main.go "$(VERSION)" | gzip -c -9 > $(build_mandir)/$(app).1.gz

## install: install the binary, manpage and completions (to /usr/local by default)
.PHONY: install
install: build manpage completions
	install -d $(DESTDIR)$(install_bindir)
	install -m755 $(build_bindir)/$(VERSION)/$(app) $(DESTDIR)$(install_bindir)/
	install -d $(DESTDIR)$(install_mandir)/man1
	install -m644 $(build_mandir)/* $(DESTDIR)$(install_mandir)/man1/
	install -d $(DESTDIR)$(install_datadir)/bash-completion/completions
	install -m644 $(build_compdir)/$(app).bash $(DESTDIR)$(install_datadir)/bash-completion/completions/$(app)
	install -d $(DESTDIR)$(install_datadir)/fish/vendor_completions.d
	install -m644 $(build_compdir)/$(app).fish $(DESTDIR)$(install_datadir)/fish/vendor_completions.d/$(app).fish
	install -d $(DESTDIR)$(install_datadir)/zsh/site-functions
	install -m644 $(build_compdir)/$(app).zsh $(DESTDIR)$(install_datadir)/zsh/site-functions/_$(app)

## uninstall: uninstall the binary, manpage and completions
.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(install_bindir)/$(app)
	rm -f $(DESTDIR)$(install_mandir)/man1/$(app).1.gz
	rm -f $(DESTDIR)$(install_datadir)/bash-completion/completions/$(app)
	rm -f $(DESTDIR)$(install_datadir)/fish/vendor_completions.d/$(app).fish
	rm -f $(DESTDIR)$(install_datadir)/zsh/site-functions/_$(app)
