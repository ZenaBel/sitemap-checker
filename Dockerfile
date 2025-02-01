FROM golang:1.23-alpine AS builder
LABEL authors="nitro"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/cmd/sitemap-checker/sitemap-checker /app/cmd/sitemap-checker/main.go

FROM alpine:3.20
LABEL authors="nitro"

COPY --from=builder /app/cmd/sitemap-checker/sitemap-checker /sitemap-checker

CMD ["/sitemap-checker"]