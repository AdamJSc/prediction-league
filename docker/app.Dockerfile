FROM golang:1.16-alpine AS builder

RUN apk update && apk --no-cache add tzdata git

WORKDIR /build

COPY . .

RUN go build -mod=mod -o /prediction-league ./service

FROM alpine

COPY --from=builder /prediction-league /app/prediction-league

# copy additional source files that are not included in binary
COPY --from=builder /usr/local/go/lib/time /usr/local/go/lib/time
COPY ./resources/dist /app/resources/dist
COPY ./service/database/migrations /app/service/database/migrations
COPY ./service/views /app/service/views

WORKDIR /app

CMD ["./prediction-league"]
