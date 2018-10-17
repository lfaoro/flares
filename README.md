# Flares 🔥

A collection of utilities that help you interact with the CloudFlare service.

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfaoro/flares)](https://goreportcard.com/report/github.com/lfaoro/flares)

## Installation
Golang must be installed: https://golang.org/dl/
```bash
# flaredns backups your domain DNS table into a git repo.
$ go get -u github.com/lfaoro/flares/cmd/flaredns
$ cd $GOPATH/src/github.com/lfaoro/flares/
# flarelogs (coming soon)
# $ go get -u github.com/lfaoro/flares/cmd/flarelogs
```

## Quick start (painless)
```bash
# CloudFlare auth key is here: https://dash.cloudflare.com/profile ->
# Global API Key -> View
$ docker run -it --rm \
-e CF_AUTH_KEY="" \
-e CF_AUTH_KEY="" \
-e CF_AUTH_EMAIL="" \
-e GIT_REPO="" \
-e GIT_USERNAME="" \
-e GIT_PASSWORD="" \
lfaoro/flares flaredns -domains awesome.tld,awesome2.tld
```

## Quick start (I want full control)
### Fill the .env with your CloudFlare and Git credentials
```bash
# Provide a .env file in your project with the following variables or export them.
# Check .env.example
$ cat > .env << EOF
CF_AUTH_KEY=""
CF_AUTH_EMAIL=""
GIT_REPO=""
GIT_USERNAME=""
GIT_PASSWORD=""
EOF
```
### Create your docker container
```bash
docker build -t flares .
```
### Run the app
```bash
docker run -it --rm flares flaredns -domains awesome.tld,awesome2.tld
```

# Contibuting
> Any help and suggestions are very welcome and appreciated. Start by opening an [issue](https://github.com/lfaoro/flares/issues/new).

- Fork the project
- Create your feature branch `git checkout -b my-new-feature`
- Commit your changes `git commit -am 'Add my feature'`
- Push to the branch `git push origin my-new-feature`
- Create a new pull request against the master branch

## TODO
- [ ] use https://github.com/spf13/cobra for the CLI interface
- [ ] add the flarelogs command
- [ ] use a git library instead of our own hacky interface
