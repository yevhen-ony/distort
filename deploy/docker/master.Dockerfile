FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP=./cmd/master
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/master ${APP}

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/master /app/master
COPY cmd/master/config.yml /app/config.yml

ENTRYPOINT ["/app/master"]
