export PATH := /usr/local/go/bin:$(PATH)
export TMPDIR := $(PWD)/tmp
export GOPATH := $(PWD)/../../../../../riemann-gitlab
TEST?=$$(go list ./... | grep -v /vendor/)
DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
VETARGS=-asmdecl -atomic -bool -buildtags -copylocks -methods \
									-nilfunc -printf -rangeloops -shift -structtags -unsafeptr

VERSION=$(shell [ -n "${GO_PIPELINE_COUNTER}" ] && echo "${GO_PIPELINE_COUNTER}" || echo "dev" )

all: test

fmt:
		mkdir -p tmp
		go fmt .

vet: fmt
	  go tool vet  ${VETARGS} $$(ls -d */ | grep -v vendor)

test: vet
	  go test -v -timeout=30s -parallel=4 $(TEST)

cover: test
		go test -v -coverprofile=cover.out $(TEST)
		go tool cover -html=cover.out

bench: test
		go test -v $(TEST) --bench=. -cpuprofile=cpu.pprof
		go tool pprof --pdf result.test cpu.pprof >cpu_stats.pdf
		open cpu_stats.pdf

build: test
		go build -ldflags "-X main.version=${VERSION} -s -w" -o bin/riemann-gitlab  main.go
		rm -rf tmp

.PHONY: all vet test cover bench build fmt

