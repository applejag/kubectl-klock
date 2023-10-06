# SPDX-FileCopyrightText: 2023 Kalle Fagerberg
#
# SPDX-License-Identifier: CC0-1.0

ifeq ($(OS),Windows_NT)
BINARY := kubectl-klock.exe
else
BINARY := kubectl-klock
endif

GO_FILES=$(shell git ls-files '*.go')

.PHONY: build
build: dist/${BINARY}

dist/${BINARY}: dist cmd/*.go pkg/*/*.go VERSION
	go build -o dist/${BINARY}

dist:
	mkdir dist

.PHONY: clean
clean:
	rm -fv dist/${BINARY}

.PHONY: check
check:
	go test ./... -coverprofile cover.out

.PHONY: deps
deps: deps-go deps-npm deps-pip

.PHONY: deps-go
deps-go:
	go get

.PHONY: deps-npm
deps-npm: node_modules

node_modules:
	npm install

.PHONY: deps-pip
deps-pip:
	python3 -m pip install --upgrade --user reuse

.PHONY: lint
lint: lint-md lint-go lint-license

.PHONY: lint-fix
lint-fix: lint-md-fix lint-go-fix

.PHONY: lint-md
lint-md: node_modules
	npx markdownlint-cli2

.PHONY: lint-md-fix
lint-md-fix: node_modules
	npx markdownlint-cli2 --fix

.PHONY: lint-go
lint-go:
	@echo goimports -d '**/*.go'
	@goimports -d $(GO_FILES)
	revive -formatter stylish -config revive.toml ./...

.PHONY: lint-go-fix
lint-fix-go:
	@echo goimports -d -w '**/*.go'
	@goimports -d -w $(GO_FILES)

.PHONY: lint-license
lint-license:
	reuse lint
