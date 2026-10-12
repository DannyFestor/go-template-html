# Captured before any include, so `help` parses this file and not .env.
MAKEFILE_PATH := $(lastword $(MAKEFILE_LIST))

-include .env

.DEFAULT_GOAL := help

GOLANGCI_LINT := go tool -modfile=tools/golangci-lint.mod golangci-lint
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

# golangci-lint fails on a directory whose Go files were all deleted, so only directories still holding some are kept.
go-package-dirs = $(sort $(dir $(wildcard $(addsuffix *.go,$(sort $(dir $(1)))))))
# Empty when FILES only names deleted packages, which leaves nothing to lint.
LINT_TARGETS = $(if $(FILES),$(call go-package-dirs,$(FILES)),./...)

tmp:
	mkdir -p $@

.PHONY: refuse-commit-on-main
refuse-commit-on-main:
	scripts/refuse-commit-on-ref.sh $(PROTECTED_REF)

.PHONY: refuse-push-to-main
refuse-push-to-main:
	scripts/refuse-push-to-ref.sh $(PROTECTED_REF)

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
fmt: ## Format Go code, or only FILES
	$(GOLANGCI_LINT) fmt $(FILES)

.PHONY: fmt-check
fmt-check: ## Fail if Go code, or only FILES, is not formatted
	$(GOLANGCI_LINT) fmt --diff $(FILES)

.PHONY: lint
lint: ## Lint Go code with golangci-lint, or only the packages holding FILES
	$(if $(LINT_TARGETS),$(GOLANGCI_LINT) run $(LINT_TARGETS))

.PHONY: fmt-js
fmt-js: bin/vp ## Format TypeScript, CSS, JSON and YAML with oxfmt, or only FILES
	$(VP) fmt $(FILES)

.PHONY: lint-js
lint-js: bin/vp ## Lint TypeScript with oxlint
	$(VP) lint

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

.PHONY: arch-graph
arch-graph: | tmp ## Draw the component dependency graph into tmp/arch-graph.svg
	$(GO_ARCH_LINT) graph --out tmp/arch-graph.svg

.PHONY: check
check: fmt fmt-js lint lint-js typecheck-js arch test ## Format, lint, type-check, check the architecture and run the tests

##@ Test

.PHONY: test
test: test-go test-js ## Run the Go and JS tests that need no Docker

.PHONY: test-go
test-go: ## Run the Go unit, architecture and tooling tests
	go test -race -shuffle=on ./...

.PHONY: test-js
test-js: bin/vp ## Run the JS unit tests with Vitest
	$(VP) test

##@ Build

.PHONY: build
build: ## Build cmd/web into bin/
	go build -o bin/web ./cmd/web

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
