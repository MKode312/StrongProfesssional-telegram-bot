FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/str-prof-bot ./cmd
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrator ./cmd/migrator

FROM alpine:3.22 AS migrator

RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -u 65532 migrator

WORKDIR /app

COPY --from=builder /out/migrator /app/migrator
COPY config /app/config
COPY migrations /app/migrations

USER migrator

ENTRYPOINT ["/app/migrator", "-config", "/app/config/local.yaml"]

FROM alpine:3.22 AS bot

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 65532 bot

WORKDIR /app

COPY --from=builder /out/str-prof-bot /app/str-prof-bot
COPY config /app/config

USER bot

ENTRYPOINT ["/app/str-prof-bot", "-config", "/app/config/local.yaml"]
