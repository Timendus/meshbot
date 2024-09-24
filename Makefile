SHELL := /bin/bash

all: run

create-virtual-env:
	@python3 -m venv ./.venv

use-virtual-env:
	@source .venv/bin/activate

build:
	@echo "Building docker image..."
	@docker build -t timendus/meshbot .

run-image:
	@echo "Running as a docker image..."
	@docker run --name=meshbot -d timendus/meshbot

export-image: build
	@echo "Exporting docker image..."
	@docker save timendus/meshbot:latest | gzip > meshbot-docker-image.tar.gz

dependencies: create-virtual-env use-virtual-env
	@pip3 install -r requirements.txt

run: use-virtual-env
	@python3 -m meshbot
