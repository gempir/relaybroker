

default: dependencies relaybroker 

.PHONY:
dependencies:
	go get github.com/op/go-logging
	go get github.com/stretchr/testify/assert
	

.PHONY: relaybroker
relaybroker:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o relaybroker .