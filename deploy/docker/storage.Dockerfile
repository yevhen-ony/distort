FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP=./cmd/storage
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/storage ${APP}

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/storage /app/storage
COPY cmd/storage/config.yml /app/config.yml

ENTRYPOINT ["/app/storage"]
