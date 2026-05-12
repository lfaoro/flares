APP     ?= ./cmd/flares
VERSION ?= 1.0.0
LDFLAGS += -X "main.date=$(shell date '+%Y-%m-%d %I:%M:%S %Z')"

.PHONY: install build build-all tag release reltest dep clean docker nix fmt vet test lint tidy mise tidy-check check hooks dev

install:
	@go install -ldflags='$(LDFLAGS)' "$(APP)"

build:
	@go build -ldflags='$(LDFLAGS)' -o flares "$(APP)"

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-freebsd-amd64 build-freebsd-arm64

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o dist/flares-linux-amd64 "$(APP)"

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o dist/flares-linux-arm64 "$(APP)"

build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o dist/flares-darwin-amd64 "$(APP)"

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o dist/flares-darwin-arm64 "$(APP)"

build-freebsd-amd64:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o dist/flares-freebsd-amd64 "$(APP)"

build-freebsd-arm64:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o dist/flares-freebsd-arm64 "$(APP)"

tag:
	git tag -f -a $(tag) -m "$(tag)"
	git push -f origin $(tag)

release:
	goreleaser release --clean --config=.goreleaser.yml

reltest:
	goreleaser release --snapshot --clean --skip=validate,publish,docker --config=.goreleaser.yml

dep:
	go mod tidy
	go mod verify
	go mod download

clean:
	rm -f flares
	rm -rf dist/

docker:
	docker build -t ghcr.io/lfaoro/flares .
	docker push ghcr.io/lfaoro/flares

nix:
	nix-shell --run "echo 'Nix dev shell ready'"

fmt:
	gofumpt -w .

vet:
	go vet ./...

test:
	go test -v -race -shuffle=on -count=1 ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

mise:
	mise install

# Run the same checks as CI before pushing.
check: tidy-check build vet lint test

tidy-check:
	go mod tidy
	git diff --exit-code go.mod go.sum

hooks: .githooks/pre-push
	git config core.hooksPath .githooks

dev:
	@echo "Choose your dev environment:"
	@echo "  make nix   — nix-shell"
	@echo "  make mise  — mise install"
	@echo "Then: make check"

