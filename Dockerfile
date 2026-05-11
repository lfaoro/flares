FROM alpine:3.23
ARG TARGETPLATFORM
RUN apk add --no-cache ca-certificates
COPY $TARGETPLATFORM/flares /usr/local/bin/flares
ENTRYPOINT ["flares"]
