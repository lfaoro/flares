FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /flares ./cmd/flares

FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=builder /flares /usr/local/bin/flares
ENTRYPOINT ["flares"]