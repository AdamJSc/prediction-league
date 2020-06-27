FROM golang:1.14-alpine

RUN apk update && apk --no-cache add git

RUN go get github.com/markbates/pkger/cmd/pkger

WORKDIR /app
