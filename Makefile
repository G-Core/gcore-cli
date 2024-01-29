all: build

.PHONY: build

build: cmd/gcore/gcore

cmd/gcore/gcore: cmd/gcore/client.go pkg/sdk/sdk.gen.go
	cd cmd/gcore && CGO_ENABLED=0 go build -ldflags="-extldflags=-static"
	cp cmd/gcore/gcore .

generate: pkg/sdk/sdk.gen.go

pkg/sdk/sdk.gen.go: pkg/sdk/api.yml
	oapi-codegen -config oapi-gen.yml pkg/sdk/api.yml