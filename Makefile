build:
	go build ./...
	cd ./test/allocator && go build . && cd -

test:
	go test -v -count=1 ./...

lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: test