build:
	go build ./...
	cd ./test/allocator && go build . && cd -

UNIT_TEST_PACKAGES=$(shell go list ./... | grep -v test)
unit_test:
	go test -v -count=1 -cover $(UNIT_TEST_PACKAGES) -coverprofile=coverage.unit.out -coverpkg ./...

integration_test:
	go test -c ./test/ -o ./test/allocator/allocator-test -coverpkg ./...
	./test/allocator/allocator-test -test.v -test.coverprofile=coverage.integration.out

coverage:
	go tool cover -func=coverage.integration.out -o=coverage.out

lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: test