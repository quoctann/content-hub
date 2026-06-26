FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/app ./cmd/app

FROM alpine:3.23
WORKDIR /app

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app

COPY --from=builder /out/app /app/app
COPY backend/migrations /app/migrations

ENV APP_ENV=prod

USER app

EXPOSE 8080

CMD ["./app"]
