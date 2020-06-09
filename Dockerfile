FROM golang:1.13

RUN apt-get update && apt-get -y upgrade && apt-get -y install chromium-browser xvfb

WORKDIR $GOPATH/src/github.com/pmurley/mida

COPY . .

RUN go get -d -v ./...

RUN go build ./mida

CMD ["./mida file"]