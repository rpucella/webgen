build: fmt etags
	go build -o bin/webgen ./cmd/webgen

fmt:
	go fmt ./cmd/webgen
	go fmt ./internal/*

test:
	go test ./cmd/*
	go test ./internal/*

etags:
	rm -f TAGS
	find . -name "*.go" -print | etags -
