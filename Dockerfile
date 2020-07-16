FROM golang:1.14

RUN apt-get update && apt-get -y install xvfb chromium

WORKDIR $GOPATH/src/github.com/pmurley/mida

COPY . .

RUN go get -d -v ./...

RUN go build

CMD ["./mida", "file"]