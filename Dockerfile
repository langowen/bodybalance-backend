FROM golang:1.24.3 AS builder

WORKDIR /app

COPY . .

RUN go mod download

ARG VERSION="unknown"
ARG BUILD_TIME="unknown"
ARG GIT_COMMIT="unknown"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
    -o main ./cmd/bodybalance/main.go


FROM alpine AS app

WORKDIR /app

COPY --from=builder /app/main ./
COPY --from=builder /app/config/* ./config/
COPY --from=builder /app/docs/* ./docs/
RUN mkdir -p /app/data/video

CMD ["/app/main"]