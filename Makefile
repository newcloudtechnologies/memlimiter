build:
	go build ./...
	pushd ./test/allocator && go build . && popd

test:
	go test -v -count=1 ./...

lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: test