FROM golang:1.10-alpine3.7

COPY . /go/src/go_gomoku

WORKDIR /go/src/go_gomoku

RUN apk add --no-cache git mercurial \
    && go get -d -v github.com/google/uuid \
    && apk del git mercurial

RUN go install -v github.com/google/uuid

RUN go build ./main.go

CMD ./main -connect server