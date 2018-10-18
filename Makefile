export GO111MODULE=on

run:
	go run cmd/flaredns/*.go -domains vlct.io

install:
	go install ./cmd/flaredns/.

build:
	go build -o flaredns cmd/flaredns/*.go

dep:
	go mod init || :
	go mod verify
	go mod tidy
	go mod download
	go mod vendor

clean:
	rm -rf vendor/ go.mod go.sum

docker:
	rm .dockerignore || :
	echo ".env" > .dockerignore
	docker build -t lfaoro/flares .
	docker push lfaoro/flares
	rm .dockerignore || :

.PHONY: install
