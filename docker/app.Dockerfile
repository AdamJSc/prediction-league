FROM golang:1.16-alpine AS builder

RUN apk update && apk --no-cache add tzdata git

WORKDIR /build

COPY . .

RUN go build -mod=mod -o /prediction-league ./service

FROM alpine

COPY --from=builder /prediction-league /app/prediction-league

WORKDIR /app

CMD ["./prediction-league"]
