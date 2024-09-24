SHELL := /bin/bash

all: run

build:
	@echo "Building docker image..."
	@docker build -t timendus/meshbot .

run-image:
	@echo "Running as a docker image..."
	@docker run --name=meshbot -d timendus/meshbot

export-image: build
	@echo "Exporting docker image..."
	@docker save timendus/meshbot:latest | gzip > meshbot-docker-image.tar.gz

dependencies:
	@python3 -m venv .venv
	@.venv/bin/pip3 install -r requirements.txt

run:
	@.venv/bin/python3 -m meshbot
