SHELL := /bin/bash

all: run

update:
	@go get -u
	@go mod tidy

install-protobuf-compiler:
	@sudo dnf install protobuf-compiler
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

run:
	@go run *.go
