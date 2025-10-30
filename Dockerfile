FROM golang:1.24.9 AS builder

WORKDIR /app

COPY . .

RUN go mod download && go mod tidy

ARG VERSION="unknown"
ARG GIT_COMMIT="unknown"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.Version=$VERSION -X main.GitCommit=$GIT_COMMIT" \
    -o main ./cmd/bodybalance/main.go


FROM alpine AS app

WORKDIR /app

RUN apk add --no-cache tzdata

COPY --from=builder /app/main ./
COPY --from=builder /app/config/* ./config/
COPY --from=builder /app/docs/* ./docs/
COPY --from=builder /app/web/ ./web/
RUN mkdir -p /app/data/video
RUN mkdir -p /app/data/img

ENV TZ=Europe/Moscow

CMD ["/app/main"]