FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY ["github.com/maket12/meeting-rooms-api/go.mod", "github.com/maket12/meeting-rooms-api/go.sum", "./"]

ENV GOPROXY=https://proxy.golang.org,https://goproxy.io,https://mirror.yandex.ru/mirrors/go-proxy/,direct

RUN go mod download

COPY backend/pkg ./pkg
COPY "github.com/maket12/meeting-rooms-api/cmd" "./cmd"
COPY backend/internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o meeting-rooms-api ./cmd/app/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/meeting-rooms-api .

COPY docs/api ./docs/api

EXPOSE 8080
CMD ["/app/meeting-rooms-api"]
