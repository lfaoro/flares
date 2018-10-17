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
RUN git config --global user.email "flares@github.com" && \
    git config --global user.name "Flares"
COPY --from=builder /go/bin/ /usr/local/bin/
COPY --from=builder /build/.env /usr/local/bin/
WORKDIR /usr/local/bin/
# CMD ["server", "-hostPort", ":8080", "-devel=false"]
