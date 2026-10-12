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

COG_VERSION := $(call pinned-version,cog)
# Upstream ships Linux x86_64 only as musl and Linux aarch64 only as gnu.
COG_TARGET_darwin_aarch64 := aarch64-apple-darwin
COG_TARGET_darwin_x86_64 := x86_64-apple-darwin
COG_TARGET_linux_aarch64 := aarch64-unknown-linux-gnu
COG_TARGET_linux_x86_64 := x86_64-unknown-linux-musl
COG_TARGET := $(COG_TARGET_$(HOST_OS)_$(HOST_ARCH_GNU))
COG_URL := https://github.com/cocogitto/cocogitto/releases/download/$(COG_VERSION)/cocogitto-$(COG_VERSION)-$(COG_TARGET).tar.gz
$(eval $(call pinned-binary,cog,$(COG_URL),$(COG_TARGET)/cog))

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
setup: go-mod-download hooks-install ## Download Go module dependencies and install the git hooks

.PHONY: go-mod-download
go-mod-download:
	go mod download

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
check: fmt lint arch test ## Format, lint, check the architecture and run the tests

##@ Test

.PHONY: test
test: ## Run the unit, architecture and tooling tests, no Docker needed
	go test -race -shuffle=on ./...

##@ Build

.PHONY: build
build: ## Build cmd/web into bin/
	go build -o bin/web ./cmd/web
