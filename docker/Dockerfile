# Build
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /poker ./cmd/poker

# Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates
COPY --from=builder /poker /usr/local/bin/poker

ENTRYPOINT ["poker"]
CMD ["--help"]