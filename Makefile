build:
	go build ./...
	cd ./test/allocator && go build . && cd -

#UNIT_TEST_PACKAGES=$(shell go list ./... | grep -v test/integration test/allocator/app)
UNIT_TEST_PACKAGES=$(shell go list ./...)
unit_test:
	go test -v -count=1 -cover $(UNIT_TEST_PACKAGES) -coverprofile=coverage.unit.out -coverpkg ./...

integration_test:
	go test -c ./test/integration/main_test.go -o ./test/integration/integration-test -coverpkg ./...
	./test/integration/integration-test -test.v -test.coverprofile=coverage.integration.out

test_coverage: unit_test integration_test
	# merge outputs from unit and integration testing
	cp coverage.unit.out coverage.overall.out
	tail --lines=+2 coverage.integration.out >> coverage.overall.out
	# cannot cover main function and CLI package
	sed -i '/test\/allocator\/app/d' ./coverage.overall.out
	sed -i '/test\/allocator\/main.go/d' ./coverage.overall.out
	# final report
	go tool cover -func=coverage.overall.out -o=coverage.out

fix:
	go fmt .
	go mod tidy

lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: test