# STAGE 1: Mengompilasi Kode Binary Go
FROM golang:1.26.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main main.go

# STAGE 2: Runtime Image Ringan
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 36665
ENV PG_HOST=localhost
ENV PG_PORT=5432
ENV PG_USER=dbe
ENV PG_PASS=rahasia
ENV PG_DATA=erp
ENV IS_POOL=false
CMD ["./main"]
