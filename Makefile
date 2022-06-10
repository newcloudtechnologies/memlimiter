test:
	go test -v -count=1 ./...

build:
	go build ./...
	pushd ./test/allocator && go build . && popd

lint:
	golangci-lint run -c .golangci.yml ./...


.PHONY: test