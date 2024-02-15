all: gcore-cli

.PHONY: gcore-cli

build: gcore-cli

gcore-cli: 
	CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o gcore-cli cmd/gcore-cli/main.go
