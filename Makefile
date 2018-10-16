export GO111MODULE=on

run:
	go run cmd/flaredns/*.go

build:
	go build -o flaredns cmd/flaredns/*.go

dep:
	go mod init || :
	go mod verify
	go mod tidy
	go mod download
	go mod vendor

docker:
	docker build -t flares .
