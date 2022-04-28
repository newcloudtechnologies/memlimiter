test:
	go test -v -count=1 ./...

build:
	go build ./...

lint:
	golangci-lint run -c .golangci.yml ./...

fix:
	ucs-fmt -p gitlab.stageoffice.ru/UCS-COMMON/memlimiter . -w

infra_prepare:
	./test/infra/prepare.sh

infra_run:
	pushd ./test/infra && docker-compose up --build

.PHONY: test