# Captured before any include, so `help` parses this file and not .env.
MAKEFILE_PATH := $(lastword $(MAKEFILE_LIST))

-include .env

.DEFAULT_GOAL := help

GOLANGCI_LINT := go tool -modfile=tools/golangci-lint.mod golangci-lint
GO_ARCH_LINT := go tool -modfile=tools/go-arch-lint.mod go-arch-lint
SHELL_SCRIPTS := $(shell git ls-files '*.sh')

HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH := $(shell uname -m)
INSTALL_PINNED_BINARY := scripts/install-pinned-binary.sh

pinned-version = $(shell cat tools/$(1).version)

# $(1) tool name, $(2) release asset URL, $(3) binary path inside the asset (empty if the asset is the binary)
define pinned-binary
bin/$(1): tools/$(1).version tools/$(1).sha256 $(INSTALL_PINNED_BINARY)
	$(INSTALL_PINNED_BINARY) $$@ $(2) tools/$(1).sha256 $(3)
endef

SHELLCHECK_VERSION := $(call pinned-version,shellcheck)
SHELLCHECK_ARCH := $(subst arm64,aarch64,$(HOST_ARCH))
SHELLCHECK_ASSET := shellcheck-$(SHELLCHECK_VERSION).$(HOST_OS).$(SHELLCHECK_ARCH).tar.gz
SHELLCHECK_URL := https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/$(SHELLCHECK_ASSET)
$(eval $(call pinned-binary,shellcheck,$(SHELLCHECK_URL),shellcheck-$(SHELLCHECK_VERSION)/shellcheck))

tmp:
	mkdir -p $@

##@ General

.PHONY: help
help: ## Show this help
	@awk -f $(dir $(MAKEFILE_PATH))scripts/make-help.awk $(MAKEFILE_PATH)

##@ Setup

.PHONY: setup
setup: ## Download Go module dependencies
	go mod download

##@ Code

.PHONY: fmt
fmt: ## Format Go code
	$(GOLANGCI_LINT) fmt

.PHONY: fmt-check
fmt-check: ## Fail if Go code is not formatted
	$(GOLANGCI_LINT) fmt --diff

.PHONY: lint
lint: ## Lint Go code with golangci-lint
	$(GOLANGCI_LINT) run

.PHONY: lint-shell
lint-shell: bin/shellcheck ## Lint every tracked shell script with shellcheck
	bin/shellcheck $(SHELL_SCRIPTS)

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
test: ## Run the unit and architecture tests, no Docker needed
	go test -race -shuffle=on ./...

##@ Build

.PHONY: build
build: ## Build cmd/web into bin/
	go build -o bin/web ./cmd/web
