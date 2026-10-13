# Captured before any include, so `help` parses this file and not .env.
MAKEFILE_PATH := $(lastword $(MAKEFILE_LIST))

-include .env

.DEFAULT_GOAL := help

# The Go module the Go targets work in, relative to the checkout. FILES are relative to it.
GO_MODULE := .
GO := go -C $(GO_MODULE)
GOLANGCI_LINT := $(GO) tool -modfile=$(CURDIR)/tools/golangci-lint.mod golangci-lint
GO_ARCH_LINT := go tool -modfile=tools/go-arch-lint.mod go-arch-lint
GITLEAKS := go tool -modfile=tools/gitleaks.mod gitleaks
LEFTHOOK := go tool -modfile=tools/tools.mod lefthook
SHELL_SCRIPTS := $(shell git ls-files '*.sh')

HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH := $(shell uname -m)
HOST_ARCH_GNU := $(subst arm64,aarch64,$(HOST_ARCH))
INSTALL_PINNED_BINARY := scripts/install-pinned-binary.sh

pinned-version = $(shell cat tools/$(1).version)

# $(1) tool name, $(2) release asset URL, $(3) binary path inside the asset (empty if the asset is the binary)
define pinned-binary
bin/$(1): tools/$(1).version tools/$(1).sha256 $(INSTALL_PINNED_BINARY)
	$(INSTALL_PINNED_BINARY) $$@ $(2) tools/$(1).sha256 $(3)
endef

SHELLCHECK_VERSION := $(call pinned-version,shellcheck)
SHELLCHECK_ASSET := shellcheck-$(SHELLCHECK_VERSION).$(HOST_OS).$(HOST_ARCH_GNU).tar.gz
SHELLCHECK_URL := https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/$(SHELLCHECK_ASSET)
$(eval $(call pinned-binary,shellcheck,$(SHELLCHECK_URL),shellcheck-$(SHELLCHECK_VERSION)/shellcheck))

RUST_TARGET_VENDOR_OS_darwin := apple-darwin
RUST_TARGET_VENDOR_OS_linux := unknown-linux-gnu
RUST_TARGET := $(HOST_ARCH_GNU)-$(RUST_TARGET_VENDOR_OS_$(HOST_OS))

COG_VERSION := $(call pinned-version,cog)
# Upstream ships Linux x86_64 only as musl.
COG_TARGET := $(subst x86_64-unknown-linux-gnu,x86_64-unknown-linux-musl,$(RUST_TARGET))
COG_URL := https://github.com/cocogitto/cocogitto/releases/download/$(COG_VERSION)/cocogitto-$(COG_VERSION)-$(COG_TARGET).tar.gz
$(eval $(call pinned-binary,cog,$(COG_URL),$(COG_TARGET)/cog))

VP_REPO := voidzero-dev/vite-plus
# Deferred to the recipe, because `vp-upgrade` rewrites tools/vp.version after this file was read.
VP_URL = https://github.com/$(VP_REPO)/releases/download/$(call pinned-version,vp)/vp-$(RUST_TARGET).tar.gz
$(eval $(call pinned-binary,vp,$$(VP_URL),vp))
# Without its own home, bin/vp sets itself up in ~/.vite-plus and repoints a global vp's shims at itself.
VP := VP_HOME=$(CURDIR)/.vite-plus bin/vp

PROTECTED_REF := refs/heads/main

# Kept in variables, because a literal comma would split the $(call) and $(if) arguments they are used in.
TEST_TAGS := integration,feature,e2e
LINT_WIP_DISABLED := unused,unparam

# golangci-lint fails on a directory whose Go files were all deleted, so only directories still holding some are kept.
# Each keeps its ./ prefix, which the go command needs and which a module's own root directory would otherwise lose.
go-package-dirs = $(patsubst $(GO_MODULE)/%,./%,$(sort $(dir $(wildcard $(addsuffix *.go,$(sort $(dir $(addprefix $(GO_MODULE)/,$(1)))))))))
# Empty when FILES only names deleted packages, which leaves nothing to check.
GO_PACKAGES = $(if $(FILES),$(call go-package-dirs,$(FILES)),./...)
on-go-packages = $(if $(GO_PACKAGES),$(1) $(GO_PACKAGES))

AGENTHOOK_MODULE := tools/agenthook
AGENTHOOK_SOURCES := $(shell find $(AGENTHOOK_MODULE) -name '*.go' ! -name '*_test.go' ! -path '*/testdata/*') \
	$(AGENTHOOK_MODULE)/go.mod $(AGENTHOOK_MODULE)/go.sum
AGENTHOOK_EVENTS := $(addprefix agenthook-,pre-tool-use post-tool-use stop)

tmp:
	mkdir -p $@

.PHONY: refuse-commit-on-main
refuse-commit-on-main:
	scripts/refuse-commit-on-ref.sh $(PROTECTED_REF)

.PHONY: refuse-push-to-main
refuse-push-to-main:
	scripts/refuse-push-to-ref.sh $(PROTECTED_REF)

bin/agenthook: $(AGENTHOOK_SOURCES)
	go -C $(AGENTHOOK_MODULE) build -o $(CURDIR)/$@ .

# .claude/settings.json pipes each Claude Code hook event's JSON through these.
.PHONY: $(AGENTHOOK_EVENTS)
$(AGENTHOOK_EVENTS): agenthook-%: bin/agenthook
	bin/agenthook $*

##@ General

.PHONY: help
help: ## Show this help
	@awk -f $(dir $(MAKEFILE_PATH))scripts/make-help.awk $(MAKEFILE_PATH)

##@ Setup

.PHONY: setup
setup: go-mod-download js-install hooks-install ## Download the Go and JS dependencies and install the git hooks

.PHONY: go-mod-download
go-mod-download:
	go mod download

.PHONY: js-install
js-install: bin/vp
	$(VP) install --frozen-lockfile

.PHONY: hooks-install
hooks-install:
	$(LEFTHOOK) install

##@ Code

.PHONY: fmt
fmt: ## Format Go code, or only FILES, in GO_MODULE
	$(GOLANGCI_LINT) fmt $(FILES)

.PHONY: fmt-check
fmt-check: ## Fail if Go code, or only FILES, in GO_MODULE is not formatted
	$(GOLANGCI_LINT) fmt --diff $(FILES)

.PHONY: lint
lint: ## Lint Go code in GO_MODULE with golangci-lint, or only the packages holding FILES
	$(call on-go-packages,$(GOLANGCI_LINT) run)

# Agents get this after each edit: the caller of code they just wrote often does not exist yet.
.PHONY: lint-wip
lint-wip:
	$(call on-go-packages,$(GOLANGCI_LINT) run --disable $(LINT_WIP_DISABLED))

.PHONY: lint-agenthook
lint-agenthook: ## Lint the agent hooks in tools/agenthook
	$(MAKE) lint GO_MODULE=$(AGENTHOOK_MODULE) FILES=

.PHONY: vet
vet: ## Vet Go code in GO_MODULE with every test tag, or only the packages holding FILES
	$(call on-go-packages,$(GO) vet -tags=$(TEST_TAGS))

.PHONY: fmt-js
fmt-js: bin/vp ## Format TypeScript, CSS, JSON and YAML with oxfmt, or only FILES
	$(VP) fmt $(FILES)

.PHONY: lint-js
lint-js: bin/vp ## Lint TypeScript with oxlint, or only FILES
	$(VP) lint $(FILES)

.PHONY: typecheck-js
typecheck-js: bin/vp ## Type-check TypeScript with tsc
	$(VP) exec tsc --noEmit

.PHONY: lint-shell
lint-shell: bin/shellcheck ## Lint every tracked shell script with shellcheck, or only FILES
	bin/shellcheck -x $(or $(FILES),$(SHELL_SCRIPTS))

.PHONY: lint-secrets
lint-secrets: ## Scan the staged changes for secrets with gitleaks
	$(GITLEAKS) git --staged --no-banner --redact --verbose

.PHONY: lint-commit-msg
lint-commit-msg: bin/cog ## Check that the message in COMMIT_MSG_FILE is a Conventional Commit
	bin/cog verify --ignore-fixup-commits --ignore-merge-commits --file $(COMMIT_MSG_FILE)

.PHONY: arch
arch: ## Check package dependencies against .go-arch-lint.yml
	$(GO_ARCH_LINT) check

# The agents' end-of-turn gate reads the report to keep only the warnings on the files it checks.
.PHONY: arch-json
arch-json:
	$(GO_ARCH_LINT) check --json

.PHONY: arch-graph
arch-graph: | tmp ## Draw the component dependency graph into tmp/arch-graph.svg
	$(GO_ARCH_LINT) graph --out tmp/arch-graph.svg

.PHONY: check
check: fmt fmt-js lint lint-agenthook lint-js typecheck-js arch test ## Format, lint, type-check, check the architecture and run the tests

##@ Test

.PHONY: test
test: test-go test-agenthook test-js ## Run the Go and JS tests that need no Docker

.PHONY: test-go
test-go: ## Run the Go unit, architecture and tooling tests in GO_MODULE, or only the packages holding FILES
	$(call on-go-packages,$(GO) test -race -shuffle=on)

.PHONY: test-agenthook
test-agenthook: ## Run the agent hook tests in tools/agenthook
	$(MAKE) test-go GO_MODULE=$(AGENTHOOK_MODULE) FILES=

.PHONY: test-js
test-js: bin/vp ## Run the JS unit tests with Vitest
	$(VP) test

##@ Build

.PHONY: build
build: ## Build cmd/web into bin/
	go build -o bin/web ./cmd/web

.PHONY: build-packages
build-packages:
	$(call on-go-packages,$(GO) build)

.PHONY: vite-build
vite-build: bin/vp ## Build the frontend assets into internal/asset/dist/
	$(VP) build

.PHONY: watch-vite
watch-vite: bin/vp ## Rebuild the frontend assets into internal/asset/dist/ on every change
	$(VP) build --watch

##@ Upgrade

.PHONY: vp-upgrade
vp-upgrade: vp-pin-latest vite-plus-pin vp-migrate ## Bump bin/vp, vite-plus and its vite override to the latest release

.PHONY: vp-pin-latest
vp-pin-latest:
	scripts/pin-latest-release.sh $(VP_REPO) vp-checksums.txt tools/vp.version tools/vp.sha256

# `vp migrate` runs the project's own vite-plus, which only aligns the rest once it is on bin/vp's version.
.PHONY: vite-plus-pin
vite-plus-pin: bin/vp
	$(VP) add -D vite-plus@$(patsubst v%,%,$(call pinned-version,vp))

.PHONY: vp-migrate
vp-migrate: bin/vp
	$(VP) migrate
