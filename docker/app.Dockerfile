FROM golang:1.18-alpine AS builder

RUN apk update && apk --no-cache add tzdata git

WORKDIR /build

COPY . .

ARG BUILD_VERSION=v0.0.0
ARG BUILD_TIMESTAMP=1970-01-01T00:00:00Z

RUN go build -ldflags "-X 'main.buildVersion=${BUILD_VERSION}' -X 'main.buildTimestamp=${BUILD_TIMESTAMP}'" -mod=mod -o /prediction-league ./service/cmd/api

FROM alpine

COPY --from=builder /prediction-league /app/prediction-league

# copy additional source files that are not included in binary
COPY --from=builder /usr/local/go/lib/time /usr/local/go/lib/time
COPY ./resources/dist /app/resources/dist
COPY ./service/database/migrations /app/service/database/migrations
COPY ./service/views /app/service/views

WORKDIR /app

CMD ["./prediction-league"]
