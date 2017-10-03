default: get build

get:
	@go get ./...

build:
	@go build

get-test:
	@go get -t ./...

test: get-test
	@go test -v ./...