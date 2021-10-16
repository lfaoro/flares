APP ?= "./cmd/flares"
VERSION ?= 1.0.0
EPOCH ?= 1
MAINTAINER ?= "The Flares Developers"
LDFLAGS += -X "main.date=$(shell date '+%Y-%m-%d %I:%M:%S %Z')"

install:
	@go install -ldflags='$(LDFLAGS)' "$(APP)"

build:
	@go build -o flares "$(APP)"

tag?=""
tag:
	git tag -f -a $(tag) -m "$(tag)"
	git push -f origin $(tag)

release:
	cd cmd/flares && \
	goreleaser release --rm-dist --config=../../.goreleaser.yml

reltest:
	cd cmd/flares && \
	goreleaser release --snapshot --rm-dist --skip-publish --config=../../.goreleaser.yml

installer:
	godownloader --repo=lfaoro/flares > ./install.sh

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
