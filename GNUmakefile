default: build

build:
	go build -o terraform-provider-turingpi .

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/davidroman0O/turingpi/0.0.1/darwin_arm64
	cp terraform-provider-turingpi ~/.terraform.d/plugins/registry.terraform.io/davidroman0O/turingpi/0.0.1/darwin_arm64/

test:
	go test -v ./...

testacc:
	TF_ACC=1 go test -v ./... -timeout 3h

testacc-quick:
	TF_ACC=1 go test -v ./... -timeout 30m -skip TestAccNodeFlash

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

generate:
	go generate ./...

clean:
	rm -f terraform-provider-turingpi
	go clean -cache

.PHONY: build install test testacc testacc-quick lint fmt generate clean
