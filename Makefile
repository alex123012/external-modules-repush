GOOS ?= $(shell go env GOOS)

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

FLAGS ?=

CGO_CFLAGS = -Wno-deprecated-declarations
ifeq ($(GOOS),darwin)
	CGO_LDFLAGS = -mmacosx-version-min=$(shell sw_vers -productVersion | sed 's/10//')
else
	CGO_LDFLAGS =
endif

.PHONY: run-macos
run-macos:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go run . $(FLAGS)

.PHONY: build-macos
build-macos:
	CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go build -o bin/external-modules-transfer main.go

.PHONY: build
build:
	go build -o bin/external-modules-transfer main.go
