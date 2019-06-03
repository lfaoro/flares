FROM golang:alpine as builder
WORKDIR /build
COPY . .
RUN apk add --update git gcc
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go install -mod vendor -gcflags "-N -l" ./cmd/flares

FROM alpine:latest
RUN apk add --update --no-cache \
    ca-certificates && \
    update-ca-certificates
COPY --from=builder /go/bin/ /usr/local/bin/
WORKDIR /usr/local/bin/
ENTRYPOINT ["flares"]
