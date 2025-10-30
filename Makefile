SHELL := /bin/bash

all: run

update:
	@go get -u
	@go mod tidy

run:
	@go run *.go

test:
	@go test -v ./...

lines:
	@find . -name '*.go' | xargs wc -l

build:
	@GOOS=linux GOARCH=amd64 go build -o dist/linux/meshbot *.go
	@cp config.json dist/linux/

	@GOOS=windows GOARCH=amd64 go build -o dist/windows/meshbot.exe *.go
	@cp config.json dist/windows/

	@GOOS=linux GOARCH=arm64 go build -o dist/raspberry-pi/meshbot *.go
	@cp config.json dist/raspberry-pi/

	@GOOS=darwin GOARCH=amd64 go build -o dist/macos-intel/meshbot *.go
	@cp config.json dist/macos-intel/

	@GOOS=darwin GOARCH=arm64 go build -o dist/macos-apple-silicon/meshbot *.go
	@cp config.json dist/macos-apple-silicon/

	@docker build --quiet -t timendus/meshbot:latest .
	@mkdir -p dist/docker
	@docker save timendus/meshbot:latest | gzip > dist/docker/meshbot.tar.gz
