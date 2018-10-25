FROM golang:alpine as builder
WORKDIR /build
COPY . .
RUN apk update && apk upgrade && \
    apk add git gcc
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go install -mod vendor -gcflags "-N -l" ./cmd/...

FROM alpine:latest
RUN apk update && apk add ca-certificates git && \
    rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/ /usr/local/bin/
WORKDIR /usr/local/bin/
ENTRYPOINT ["flaredns"]
