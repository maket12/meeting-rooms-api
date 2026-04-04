FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o meeting-rooms-api ./cmd/app/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/meeting-rooms-api .

EXPOSE 8080
CMD ["/app/meeting-rooms-api"]
