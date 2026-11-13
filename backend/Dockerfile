# syntax=docker/dockerfile:1

FROM golang:1.27.4-alpine AS builder

WORKDIR /src

# Cache module downloads before copying the rest of the source.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Fully static binaries so they run on a distroless base without a libc.
ENV CGO_ENABLED=0

RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# The goose CLI is used by the compose `migrate` job.
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.30.0

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /src/db/migrations /migrations

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/api"]
