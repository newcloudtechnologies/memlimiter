build:
	go build ./...
	cd ./test/allocator && go build . && cd -

UNIT_TEST_PACKAGES=$(shell go list ./... | grep -v test)
unit_test:
	go test -v -count=1 -cover $(UNIT_TEST_PACKAGES) -coverprofile=coverage.unit.out -coverpkg ./...

integration_test:
	go test -c ./test/ -o ./test/allocator/allocator-test -coverpkg ./...
	./test/allocator/allocator-test -test.v -test.coverprofile=coverage.integration.out

test_coverage: unit_test integration_test
	cp coverage.unit.out coverage.out
	tail --lines=+2 coverage.integration.out >> coverage.out
	go tool cover -func=coverage.out -o=coverage.out

lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: test