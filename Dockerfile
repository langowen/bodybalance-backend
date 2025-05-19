FROM golang:1.24.3 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/bodybalance/main.go

FROM alpine AS app

WORKDIR /app

COPY --from=builder /app/main ./
COPY --from=builder /app/config/* ./config/
RUN mkdir -p /app/data/video

CMD ["/app/main"]