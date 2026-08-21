default: fmt lint install generate

build:
	go build -v ./...

fmt:
	gofmt -s -w -e .

generate:
	cd tools; go generate ./...

# Regenerates the predefined machine type table from the Compute Engine API.
# Requires GCE_METADATA_GEN_PROJECT and working gcloud credentials. Not part of
# `make generate`, because it needs GCP access that CI does not have.
generate-machine-types:
	tools/generate-machine-types.sh

install: build
	go install -v ./...

lint:
	golangci-lint run

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: build fmt generate generate-machine-types install lint test testacc
