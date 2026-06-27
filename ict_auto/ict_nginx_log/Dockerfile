# Stage 1: Mengompilasi Kode Binary Go
FROM golang:1.26.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o auto_ict_nginx_log .

# Stage 2: Runtime Image Ringan
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/auto_ict_nginx_log .
ENV PG_HOST=erp_pool
ENV PG_PORT=5432
ENV PG_USER=dbe
ENV PG_PASS=rahasia
ENV PG_DATA=erp
ENV IS_POOL=true
ENV ES_LINK=http://elasticsearch:9200
ENV FT_HTTP=150
CMD ["./auto_ict_nginx_log"]
