FROM golang:1.14-alpine

RUN apk update && apk --no-cache add tzdata git

WORKDIR /app

COPY . .

RUN go mod vendor

RUN go build -mod vendor -o ./bin/prediction-league ./service

CMD ["./bin/prediction-league"]
