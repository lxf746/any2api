FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/any2api-go ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/any2api-go ./any2api-go

ENV NEWPLATFORM2API_PORT=8099 \
    NEWPLATFORM2API_DATA_DIR=/app/data

EXPOSE 8099

CMD ["./any2api-go"]