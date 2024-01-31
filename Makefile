all: build

.PHONY: build

build: cmd/gcore/gcore

cmd/gcore/gcore: cmd/gcore/client.go cmd/gcore/output.go cmd/gcore/fastedge.go pkg/sdk/sdk.gen.go
	cd cmd/gcore && CGO_ENABLED=0 go build -ldflags="-extldflags=-static"
	cp cmd/gcore/gcore .

generate: pkg/sdk/sdk.gen.go

pkg/sdk/sdk.gen.go: pkg/sdk/api.yml
	sed -i 's/ \/v1/ \/fastedge\/v1/g' pkg/sdk/api.yml # add /fastedge prefix to endpoints
	oapi-codegen -config oapi-gen.yml pkg/sdk/api.yml
