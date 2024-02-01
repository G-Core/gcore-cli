all: gcore

.PHONY: gcore

build: gcore

gcore: 
	CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o gcore
	
generate: pkg/sdk/sdk.gen.go

pkg/sdk/sdk.gen.go: pkg/sdk/api.yml
	oapi-codegen -config oapi-gen.yml pkg/sdk/api.yml
