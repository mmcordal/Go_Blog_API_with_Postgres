# 1. Build aşaması
FROM golang:1.24 AS builder

WORKDIR /app

# Modülleri indir
COPY go.mod go.sum ./
RUN go mod download

# Projeyi kopyala
COPY . .

# Binary oluştur
RUN go build -o cleanarch_with_postgres ./cmd/main.go

# 2. Run aşaması
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/cleanarch_with_postgres .

EXPOSE 3000

CMD ["./cleanarch_with_postgres"]