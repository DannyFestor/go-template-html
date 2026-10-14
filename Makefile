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
GOVULNCHECK := $(GO) tool -modfile=$(CURDIR)/tools/govulncheck.mod govulncheck
LEFTHOOK := go tool -modfile=tools/tools.mod lefthook
AIR := go tool -modfile=tools/tools.mod air
# Only the root module has templ files, so templ runs there whatever GO_MODULE names.
TEMPL := go tool templ
# templ fmt and generate -f act on their first file only, so FILES gets one run per file, or the whole tree without it.
# $(1) templ arguments before each file, $(2) templ arguments for the whole tree
templ-on-files = $(if $(FILES),$(foreach file,$(FILES),$(TEMPL) $(1) $(file) &&) true,$(TEMPL) $(2))
# Deferred, because the image's build context has no .git for git ls-files to read.
SHELL_SCRIPTS = $(shell git ls-files '*.sh')

# The image's build context has no .git either, so `image` passes the version in as a build arg.
VERSION ?= $(shell git describe --always --dirty)
GO_LDFLAGS = -s -w -X github.com/DannyFestor/go-template-html/internal/buildinfo.Version=$(VERSION)
IMAGE_NAME := go-template-html

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

# The templ proxy forwards to the port the app listens on, which HTTP_ADDR in .env moves away from the app's default.
HTTP_ADDR ?= :8080
DEV_APP_URL = http://localhost:$(lastword $(subst :, ,$(HTTP_ADDR)))

# Kept in variables, because a literal comma would split the $(call) and $(if) arguments they are used in.
TEST_TAGS := integration,feature,e2e
LINT_WIP_DISABLED := unused,unparam

# golangci-lint fails on a directory whose Go files were all deleted, so only directories still holding some are kept.
go-files-beside = $(wildcard $(addsuffix *.go,$(sort $(dir $(addprefix $(GO_MODULE)/,$(1))))))
# Each keeps its ./ prefix, which the go command needs and which a module's own root directory would otherwise lose.
go-package-dirs = $(patsubst $(GO_MODULE)/%,./%,$(sort $(dir $(1))))
# Empty when FILES only names deleted packages, which leaves nothing to check.
GO_PACKAGES = $(if $(FILES),$(call go-package-dirs,$(call go-files-beside,$(FILES))),./...)
# go build refuses a package that only holds tests, which test-go compiles anyway.
GO_BUILD_PACKAGES = $(if $(FILES),$(call go-package-dirs,$(filter-out %_test.go,$(call go-files-beside,$(FILES)))),./...)
# These hold only tagged tests, which go test refuses to build when named without their tag. ./... skips them.
TAGGED_TEST_DIRS := ./tests/feature/ ./tests/e2e/
UNTAGGED_TEST_PACKAGES = $(filter-out $(TAGGED_TEST_DIRS),$(GO_PACKAGES))
on-packages = $(if $(2),$(1) $(2))

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

.PHONY: github-setup
github-setup: ## Create or update the GitHub repo's main ruleset and merge settings with gh, which needs admin rights
	scripts/github-setup.sh .github/rulesets/main.json .github/merge-settings.json

##@ Services

.PHONY: services
services: ## Start the services in compose.yaml in Docker and wait until all are healthy
	docker compose up -d --wait

.PHONY: down
down: ## Stop and remove the service containers, keeping their data
	docker compose down

.PHONY: reset
reset: ## Stop and remove the service containers and delete the Postgres and Valkey data
	docker compose down -v

##@ Dev

# bin/vp is fetched up front, so its download does not interleave with the output of the parallel watchers.
.PHONY: dev
dev: bin/vp services generate ## Start the services, then rebuild, restart and reload the browser on every change
	$(MAKE) -j watch-templ watch-vite watch-assets

.PHONY: watch-templ
watch-templ: ## Regenerate templ files, rerun cmd/web on Go changes and reload the browser through the templ proxy
	$(TEMPL) generate --watch --proxy=$(DEV_APP_URL) --cmd="go run ./cmd/web"

.PHONY: watch-vite
watch-vite: bin/vp ## Rebuild the frontend assets into internal/asset/dist/ on every change
	$(VP) build --watch

# air only has to tell the templ proxy to reload, so `true` stands in for the binary it expects to run after each build.
.PHONY: watch-assets
watch-assets: ## Reload the browser through the templ proxy whenever internal/asset/dist/ changes
	$(AIR) --build.cmd "$(TEMPL) generate --notify-proxy" --build.entrypoint true \
		--build.include_dir internal/asset/dist --build.include_ext css,js,json

##@ Code

.PHONY: generate
generate: generate-templ generate-go ## Regenerate the generated code

.PHONY: generate-templ
generate-templ: ## Regenerate the Go code of every templ file, or only of the templ files in FILES
	$(call templ-on-files,generate -f,generate)

.PHONY: generate-go
generate-go: ## Run go generate in GO_MODULE
	$(GO) generate ./...

.PHONY: drift-check
drift-check: ## Fail if make generate changed or added a file
	scripts/fail-on-drift.sh

.PHONY: fmt
fmt: fmt-go fmt-templ ## Format Go and templ code

.PHONY: fmt-go
fmt-go: ## Format Go code, or only FILES, in GO_MODULE
	$(GOLANGCI_LINT) fmt $(FILES)

.PHONY: fmt-templ
fmt-templ: ## Format templ files, or only FILES
	$(call templ-on-files,fmt,fmt .)

.PHONY: fmt-check
fmt-check: fmt-check-go fmt-check-templ fmt-check-js ## Fail if Go, templ or frontend code is not formatted

.PHONY: fmt-check-go
fmt-check-go: ## Fail if Go code, or only FILES, in GO_MODULE is not formatted
	$(GOLANGCI_LINT) fmt --diff $(FILES)

# templ has no dry run: -fail formats before it fails, and -stdout reports every file as changed.
.PHONY: fmt-check-templ
fmt-check-templ: ## Fail if templ files, or only FILES, are not formatted, formatting them in place
	$(call templ-on-files,fmt -fail,fmt -fail .)

.PHONY: fmt-check-js
fmt-check-js: bin/vp ## Fail if TypeScript, CSS, JSON or YAML, or only FILES, is not formatted
	$(VP) fmt --check $(FILES)

.PHONY: lint
lint: ## Lint Go code in GO_MODULE with golangci-lint, or only the packages holding FILES
	$(call on-packages,$(GOLANGCI_LINT) run --build-tags $(TEST_TAGS),$(GO_PACKAGES))

# Agents get this after each edit: the caller of code they just wrote often does not exist yet.
.PHONY: lint-wip
lint-wip:
	$(call on-packages,$(GOLANGCI_LINT) run --build-tags $(TEST_TAGS) --disable $(LINT_WIP_DISABLED),$(GO_PACKAGES))

.PHONY: lint-agenthook
lint-agenthook: ## Lint the agent hooks in tools/agenthook
	$(MAKE) lint GO_MODULE=$(AGENTHOOK_MODULE) FILES=

.PHONY: vet
vet: ## Vet Go code in GO_MODULE with every test tag, or only the packages holding FILES
	$(call on-packages,$(GO) vet -tags=$(TEST_TAGS),$(GO_PACKAGES))

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

# The title comes from the environment, so a crafted title can never reach the shell as code. A squash merge turns
# it into the commit on main, so fixup! and merge titles fail too.
.PHONY: lint-pr-title
lint-pr-title: bin/cog ## Check that the pull request title in PR_TITLE is a Conventional Commit
	bin/cog verify "$$PR_TITLE"

.PHONY: vuln
vuln: ## Check GO_MODULE and its dependencies for known vulnerabilities with govulncheck
	$(GOVULNCHECK) ./...

.PHONY: vuln-agenthook
vuln-agenthook: ## Check the agent hooks in tools/agenthook for known vulnerabilities
	$(MAKE) vuln GO_MODULE=$(AGENTHOOK_MODULE) FILES=

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
	$(call on-packages,$(GO) test -race -shuffle=on,$(UNTAGGED_TEST_PACKAGES))

.PHONY: test-agenthook
test-agenthook: ## Run the agent hook tests in tools/agenthook
	$(MAKE) test-go GO_MODULE=$(AGENTHOOK_MODULE) FILES=

.PHONY: test-js
test-js: bin/vp ## Run the JS unit tests with Vitest
	$(VP) test

# The app under test embeds internal/asset/dist, so `make vite-build` has to run first.
.PHONY: test-feature
test-feature: ## Run the feature tests in tests/feature against the last `make vite-build`
	$(GO) test -race -shuffle=on -tags feature ./tests/feature/...

##@ Build

.PHONY: build
build: js-install vite-build build-web ## Build the frontend assets, then cmd/web into bin/ stamped with VERSION

.PHONY: build-web
build-web:
	CGO_ENABLED=0 go build -trimpath -ldflags='$(GO_LDFLAGS)' -o bin/web ./cmd/web

.PHONY: build-packages
build-packages:
	$(call on-packages,$(GO) build,$(GO_BUILD_PACKAGES))

.PHONY: vite-build
vite-build: bin/vp ## Build the frontend assets into internal/asset/dist/
	$(VP) build

.PHONY: image
image: ## Build the production image for the host platform, tagged IMAGE_NAME:VERSION and IMAGE_NAME:latest
	docker build --build-arg VERSION=$(VERSION) --tag $(IMAGE_NAME):$(VERSION) --tag $(IMAGE_NAME):latest .

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
