test:
	go test -v -count=1 ./...

build:
	go build ./...

lint:
	golangci-lint run -c .golangci.yml ./...


.PHONY: test