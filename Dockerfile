FROM golang:1.13.4

WORKDIR /go/src/crawler
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["crawler"]
