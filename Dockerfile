FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY flares /usr/local/bin/flares
ENTRYPOINT ["flares"]