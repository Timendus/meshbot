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

run:
	@python3 ./main.py
