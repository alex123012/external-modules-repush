
CGO_CFLAGS = -Wno-deprecated-declarations
CGO_LDFLAGS = -mmacosx-version-min=13.0
FLAGS ?=

.PHONY: run-macos
run-macos:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go run . $(FLAGS)

.PHONY: build-macos
build-macos:
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) go build -o bin/external-modules-transfer main.go

.PHONY: build
build:
	go build -o bin/external-modules-transfer main.go
