FROM golang:1.14-alpine AS builder

RUN apk update && apk --no-cache add tzdata git

WORKDIR /app

COPY . .

RUN go build -o /prediction-league ./service

FROM alpine

COPY --from=builder /prediction-league /prediction-league

CMD ["/prediction-league"]
