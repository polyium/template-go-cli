# ====================================================================================
# Project Specific Globals
# ------------------------------------------------------------------------------------
#
# - It's assumed the $(name) is the same literal as the compiled binary or executable.
# - Override the defaults if not available in a pipeline's environment variables.
#
# - Default GitHub environment variables: https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
#

name := polyforge
ifdef CI_PROJECT_NAME
    override name = $(CI_PROJECT_NAME)
endif

homebrew-tap := polyium/polyforge
ifdef HOMEBREW_TAP
    override homebrew-tap = $(HOMEBREW_TAP)
endif

# homebrew-tap-repository := gitlab.com:example-organization/group-1/group-2/homebrew-taps.git
homebrew-tap-repository := https://github.com/polyium/homebrew-taps
ifdef HOMEBREW_TAP_REPOSITORY
    override homebrew-tap-repository = $(HOMEBREW_TAP_REPOSITORY)
endif

type = patch
ifdef RELEASE
    override type = $(RELEASE)
endif

type-title = $(shell printf "%s" "$(shell tr '[:lower:]' '[:upper:]' <<< "$(type)")")

ifeq (,$(shell go env GOBIN))
    GOBIN=$(shell go env GOPATH)/bin
else
    GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail

.SHELLFLAGS = -ec

# ====================================================================================
# Colors
# ------------------------------------------------------------------------------------

black        := $(shell printf "\033[30m")
black-bold   := $(shell printf "\033[30;1m")
red          := $(shell printf "\033[31m")
red-bold     := $(shell printf "\033[31;1m")
green        := $(shell printf "\033[32m")
green-bold   := $(shell printf "\033[32;1m")
yellow       := $(shell printf "\033[33m")
yellow-bold  := $(shell printf "\033[33;1m")
blue         := $(shell printf "\033[34m")
blue-bold    := $(shell printf "\033[34;1m")
magenta      := $(shell printf "\033[35m")
magenta-bold := $(shell printf "\033[35;1m")
cyan         := $(shell printf "\033[36m")
cyan-bold    := $(shell printf "\033[36;1m")
white        := $(shell printf "\033[37m")
white-bold   := $(shell printf "\033[37;1m")

faint         := $(shell printf "\033[2m")
italic         := $(shell printf "\033[3m")
reset        := $(shell printf "\033[0m")

# ====================================================================================
# Logger
# ------------------------------------------------------------------------------------
#
# - Variables are declared with $@_ prefix that makes them "local" to the rule.
#

define format
    $(eval $@_COLOR = $(1))
    $(eval $@_RESET = $(2))
    $(eval $@_MESSAGE = $(3))

    @echo "${$@_COLOR}${$@_MESSAGE}${$@_RESET}"
endef

define info
    @$(call format,"$(blue-bold)","$(reset)","$(1)")
endef

define trace
    @$(call format,"$(faint)","$(reset)","$(1)")
endef

define step
    @$(call trace," - $(1)")
endef

# ====================================================================================
# Utility Command(s)
# ------------------------------------------------------------------------------------

url = $(shell git config --get remote.origin.url | sed -r 's/.*(\@|\/\/)(.*)(\:|\/)([^:\/]*)\/([^\/\.]*)\.git/https:\/\/\2\/\4\/\5/')

repository = $(shell basename -s .git $(shell git config --get remote.origin.url))
organization = $(shell git remote -v | grep "(fetch)" | sed 's/.*\/\([^ ]*\)\/.*/\1/')
package = $(shell git remote -v | grep "(fetch)" | sed 's/^origin[[:space:]]*//; s/[[:space:]]*(fetch)$$//' | sed 's/https:\/\///; s/git@//; s/\.git$$//; s/:/\//' | sed -E 's|^ssh/+||')

version = $(shell [ -f VERSION ] && head VERSION || echo "0.0.0")

major-version      		= $(shell echo $(version) | sed "s/^\([0-9]*\).*/\1/")
minor-version      		= $(shell echo $(version) | sed "s/[0-9]*\.\([0-9]*\).*/\1/")
patch-version      		= $(shell echo $(version) | sed "s/[0-9]*\.[0-9]*\.\([0-9]*\).*/\1/")

zero = $(shell printf "%s" "0")

major-upgrade 	= $(shell expr $(major-version) + 1).$(zero).$(zero)
minor-upgrade 	= $(major-version).$(shell expr $(minor-version) + 1).$(zero)
patch-upgrade 	= $(major-version).$(minor-version).$(shell expr $(patch-version) + 1)

dirty = $(shell git diff --quiet)
dirty-contents 			= $(shell git diff --shortstat 2>/dev/null 2>/dev/null | tail -n1)

# ====================================================================================
# Build Command(s)
# ------------------------------------------------------------------------------------

compile = go build --mod "vendor" --ldflags "-s -w -X=main.version=$(tag) -X=main.date=$(shell date +%Y-%m-%d:%H-%M-%S) -X=main.source=false" -o "./build/$(name)-$(GOOS)-$(GOARCH)/$(name)"
compile-windows = go build --mod "vendor" --ldflags "-s -w -X=main.version=$(tag) -X=main.date=$(shell date +%Y-%m-%d:%H-%M-%S) -X=main.source=false" -o "./build/$(name)-$(GOOS)-$(GOARCH)/$(name).exe"

archive = tar -czvf "$(name)-$(GOOS)-$(GOARCH).tar.gz" -C "./build/$(name)-$(GOOS)-$(GOARCH)" .
archive-windows = cd "./build/$(name)-$(GOOS)-$(GOARCH)" && zip -r "../../$(name)-$(GOOS)-$(GOARCH).zip" "." && cd -

distribute = mkdir -p distribution && mv *.tar.gz distribution

# ====================================================================================
# Default
# ------------------------------------------------------------------------------------

all: pre-requisites test-release $(type)-release

# ====================================================================================
# Pre-Requisites
# ------------------------------------------------------------------------------------

.PHONY: pre-requisites
pre-requisites:
	@echo "$(blue-bold)Checking Requirements ...$(reset)" && echo
	@command -v brew 2>&1> /dev/null || bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
	@echo "- Verified Homebrew Installation"
	@command -v pre-commit 2>&1> /dev/null || brew install pre-commit && pre-commit install 2>&1> /dev/null && pre-commit install-hooks 2>&1> /dev/null
	@echo "- Verified Pre-Commit Hooks"
	@command -v goreleaser 2>&1> /dev/null || brew install goreleaser
	@echo "- Verified Releaser"
	@command -v go 2>&1> /dev/null || brew install go
	@echo "- Verified Runtime"

# ====================================================================================
# Brew & Installation
# ------------------------------------------------------------------------------------

.PHONY: brew-uninstall
brew-uninstall:
	@echo "$(blue-bold)Uninstalling Package$(reset): ($(name))" && echo
	@rm -rf /opt/homebrew/etc/gitconfig
	@brew uninstall $(name) --force || true
	@brew untap $(homebrew-tap) --force || true

.PHONY: brew-install
brew-install: brew-uninstall
	@echo "$(blue-bold)Installing Package$(reset): ($(name))" && echo
	@brew tap $(homebrew-tap) $(homebrew-tap-repository) --force-auto-update --force
	@brew update
	@brew install $(name)

.PHONY: overwrite-private-homebrew-download-strategy
overwrite-private-homebrew-download-strategy:
	@echo "$(blue-bold)Overwriting Private Homebrew Download Strategy$(reset): ($(name))" && echo
	@rm -rf ./.upstreams
	@sed -i -e "s/using: GitDownloadStrategy/using: GitDownloadStrategy, tag: \"$(tag)\"/g" ./dist/homebrew/Formula/$(name).rb
	@mkdir -p .upstreams
	@git clone $(homebrew-tap-repository) ./.upstreams/homebrew-taps
	@rm -f ./.upstreams/homebrew-taps/Formula/$(name).rb
	@cp -f ./dist/homebrew/Formula/$(name).rb ./.upstreams/homebrew-taps/Formula/$(name).rb
	@cd ./.upstreams/homebrew-taps && git add ./Formula/$(name).rb && git commit -m "[Chore] - Overwrote URL + Tag" && git push -u origin main
	@cd "$(git rev-parse --show-toplevel)"
	@rm -rf ./.upstreams
	@echo "$(green-bold)Successfully Changed Upstream$(reset)" && echo

# ====================================================================================
# CI-CD Build Targets
# ------------------------------------------------------------------------------------

build: build-darwin build-linux build-windows

# (Darwin) Build Targets

build-darwin: build-darwin-amd64 build-darwin-arm64

build-darwin-arm64: export GOOS := darwin
build-darwin-arm64: export GOARCH := arm64
build-darwin-arm64:
	$(compile)
	$(archive)

build-darwin-amd64: export GOOS := darwin
build-darwin-amd64: export GOARCH := amd64
build-darwin-amd64:
	$(compile)
	$(archive)

# (Linux) Build Targets

build-linux: build-linux-amd64 build-linux-arm64 build-linux-386

build-linux-arm64: export GOOS := linux
build-linux-arm64: export GOARCH := arm64
build-linux-arm64:
	$(compile)
	$(archive)

build-linux-amd64: export GOOS := linux
build-linux-amd64: export GOARCH := amd64
build-linux-amd64:
	$(compile)
	$(archive)

build-linux-386: export GOOS := linux
build-linux-386: export GOARCH := 386
build-linux-386:
	$(compile)
	$(archive)

# (Windows) Build Targets

build-windows: build-windows-amd64 build-windows-386

build-windows-amd64: export GOOS := windows
build-windows-amd64: export GOARCH := amd64
build-windows-amd64:
	$(compile-windows)
	$(archive-windows)

build-windows-386: export GOOS := windows
build-windows-386: export GOARCH := 386
build-windows-386:
	$(compile-windows)
	$(archive-windows)

# ====================================================================================
# Testing
# ------------------------------------------------------------------------------------

.PHONY: test
test: tidy
	@go test ./...

.PHONY: test-release
test-release:
	@goreleaser release --snapshot --clean

# ====================================================================================
# Git + Versioning
# ------------------------------------------------------------------------------------

.PHONY: git-check-tree
git-check-tree:
	@echo "$(blue-bold)"Checking Working Tree"$(reset) ..." && echo
	@if ! git diff --quiet --exit-code; then \
    	git status ; \
    	echo "" ; \
		echo "$(red-bold)Dirty Working Tree$(reset) - Commit Changes and Try Again"; \
		echo "" ; \
		exit 1; \
	fi
	@$(call step,"Clean Working Tree") && echo

.PHONY: bump
bump: git-check-tree test
	@echo "$(green-bold)Bumping Version: \"$(yellow-bold)$(package)$(reset)\" - $(white-bold)$(version)$(reset)" && echo
	@echo "$($(type)-upgrade)" > VERSION
	@$(call step,"Updated Version Lock") && echo

.PHONY: commit
commit: bump
	@echo "$(blue-bold)Tag-Release$(reset) ($(type-title)): $(yellow-bold)$(package)$(reset) - $(white-bold)$(version)$(reset)" && echo
	@git add VERSION
	@git commit --message "Chore ($(type-title)) - Tag Release: $(version)"
	@git push --set-upstream origin main
	@git tag "v$(version)"
	@git push origin "v$(version)"

# ====================================================================================
# Utilities
# ------------------------------------------------------------------------------------

.PHONY: clean
clean:
	rm *.tar.gz && rm *.zip

.PHONY: tidy
tidy:
	@go mod tidy && go mod vendor

.PHONY: imports
imports:
	@go tool goimports -format-only -v -w .

.PHONY: escape-hatch
escape-hatch:
	@rm -rf ./.upstreams
	@sed -i -e "s/using: GitDownloadStrategy/using: GitDownloadStrategy, tag: \"$(tag)\"/g" ./dist/homebrew/Formula/$(name).rb
	@mkdir -p .upstreams
	@git clone $(homebrew-tap-repository) ./.upstreams/homebrew-taps
	@rm -f ./.upstreams/homebrew-taps/Formula/$(name).rb
	@cp -f ./dist/homebrew/Formula/$(name).rb ./.upstreams/homebrew-taps/Formula/$(name).rb
	@cd ./.upstreams/homebrew-taps && git add ./Formula/$(name).rb && git commit -m "[Chore] - Overwrote URL + Tag" && git push -u origin main
	@cd "$(git rev-parse --show-toplevel)"
	@rm -rf ./.upstreams

# ====================================================================================
# Release
# ------------------------------------------------------------------------------------

patch: override type = patch
patch release &: commit build

minor: override type = minor
minor release &: commit build

major: override type = major
