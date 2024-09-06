.DEFAULT_GOAL := build

.PHONY: fmt vet build
fmt:
	go fmt ./...
vet: fmt
	go mod tidy
	go vet ./...
run: vet
	go run main.go $(arg)
docs: vet
	go run docs/main.go
build: docs
	go build -o bin/pocdeploy
