FROM golang:1.25 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o /out/dos ./cmd/client/cli

FROM ubuntu:24.04
WORKDIR /work

COPY --from=builder /out/dos /usr/local/bin/dos
COPY cmd/client/cli/config.yml /work/config.yml

ENTRYPOINT ["/bin/bash", "-c"]
