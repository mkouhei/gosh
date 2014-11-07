#!/usr/bin/make -f
# -*- makefile -*-

BIN := gosh
SRC := *.go
GOPKG := github.com/mkouhei/gosh/
GOPATH := $(CURDIR)/_build
export GOPATH


all: precheck clean test format build

precheck:
	@if [ -d .git ]; then \
	set -e; \
	diff -u .git/hooks/pre-commit utils/pre-commit.txt ;\
	[ -x .git/hooks/pre-commit ] ;\
	fi

prebuild:
	go get -d -v ./...
	install -d $(CURDIR)/_build/src/$(GOPKG)
	cp -a $(CURDIR)/*.go $(CURDIR)/_build/src/$(GOPKG)


build: prebuild
	go build -ldflags "-X main.version $(shell git describe --always)" -o _build/$(BIN)

build-only:
	go build -ldflags "-X main.version $(shell git describe --always)" -o _build/$(BIN)

clean:
	@rm -f _build/$(BIN)

format:
	for src in $(SRC); do \
		gofmt $$src > $$src.tmp ;\
		goimports -w $$src.tmp ;\
		mv -f $$src.tmp $$src ;\
		rm -f $$src.tmp ;\
	done


test: prebuild
	go test -v -coverprofile=c.out $(GOPKG)
	go tool cover -func=c.out
	unlink c.out
