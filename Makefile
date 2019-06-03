APP ?= "./cmd/flares"

VERSION ?= 1.0.0
EPOCH ?= 1
MAINTAINER ?= "Community"

LDFLAGS += -X "main.date=$(shell date '+%Y-%m-%d %I:%M:%S %Z')"

install:
	@go install -ldflags='$(LDFLAGS)' "$(APP)"

build:
	@go build -o flares "$(APP)"

dep:
	go mod init || :
	go mod tidy
	go mod verify
	go mod download
	go mod vendor

clean:
	rm -rf vendor/ go.mod go.sum

docker:
	docker build -t lfaoro/flares .
	#docker push lfaoro/flares

.PHONY: install
