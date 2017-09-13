FROM golang:latest
WORKDIR /go/src/github.com/gempir/relaybroker
RUN go get github.com/op/go-logging \
	&& go get github.com/stretchr/testify/assert
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/gempir/relaybroker/app .
CMD ["./app"]  
EXPOSE 3333
