FROM golang:1.23-alpine AS builder
LABEL authors="nitro"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/cmd/sitemap-checker/sitemap-checker /app/cmd/sitemap-checker/main.go

CMD ["/app/cmd/sitemap-checker/sitemap-checker"]