build: fmt etags
	go build -o bin/webgen ./cmd/webgen
	go build -o bin/weblog ./cmd/weblog

fmt:
	go fmt ./cmd/*
	go fmt ./internal/*

test:
	go test ./cmd/*
	go test ./internal/*

etags:
	rm -f TAGS
	find . -name "*.go" -print | etags -
