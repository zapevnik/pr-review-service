FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o pr-review-service ./cmd/pr-review-service

FROM alpine:3.18
WORKDIR /app

COPY --from=build /app/pr-review-service .
COPY config.yaml .
COPY migrations ./migrations

EXPOSE 8080

CMD ["./pr-review-service"]
