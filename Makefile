all: gcore

.PHONY: gcore

build: gcore

gcore: 
	CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o gcore
