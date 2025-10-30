# syntax=docker/dockerfile:1

#######################
### Building container

FROM golang:latest AS build
WORKDIR /app

# Install dependencies
COPY go.mod go.sum .
RUN go mod download

# Copy source
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o output/meshbot *.go

######################
### Running container

FROM alpine:latest AS run
WORKDIR /app

# Copy the application executable from the build image
COPY --from=build /app/output /app

# For when we have a web interface:
# EXPOSE 8080

# Have a little runner script that copies the default config and plugins to the
# host directory if not yet present
COPY ./config.json /app/default-config/config.json
RUN cat >./run-meshbot.sh <<EOF
#!/bin/sh
if [ ! -f "/app/config/config.json" ]; then
    echo "Copying default config"
    cp /app/default-config/config.json /app/config/config.json
fi
ln -sfn /app/config/config.json /app/config.json
./meshbot
EOF
RUN chmod +x ./run-meshbot.sh

# Run the application
CMD ["./run-meshbot.sh"]
