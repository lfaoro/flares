# Flares 🔥

Flares is a CloudFlare DNS backup tool: every time it runs, dumps your DNS table to the screen || optionally exports it into nicely formatted files.

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/lfaoro/flares)](https://goreportcard.com/report/github.com/lfaoro/flares)

![flaredns_demo](static/flaredns_demo.gif)

## Docker Quick start (painless)
```bash
# CloudFlare auth key is here: https://dash.cloudflare.com/profile ->
# Global API Key -> View
$ docker run -it --rm \
-e CF_AUTH_KEY="" \
-e CF_AUTH_EMAIL="" \
lfaoro/flares domain1.tld domain2.tld
```

## Quick start (full control)
Golang must be installed: https://golang.org/dl/
```bash
# flaredns
$ go get -u github.com/lfaoro/flares/cmd/flaredns
$ cd $GOPATH/src/github.com/lfaoro/flares/
# flarelogs (coming soon)
# $ go get -u github.com/lfaoro/flares/cmd/flarelogs
```
### Fill the .env with your CloudFlare and Git credentials
```bash
# Provide a .env file in your project with the following variables or export them.
# Check .env.example
$ cat > .env << EOF
CF_AUTH_KEY=""
CF_AUTH_EMAIL=""
EOF
```
### Run the app
```bash
$ make install
$ flaredns -h
$ flaredns domain.tld
$ flaredns export -d /tmp/tables/ domain.tld
```

# Contributing
> Any help and suggestions are very welcome and appreciated. Start by opening an [issue](https://github.com/lfaoro/flares/issues/new).

- Fork the project
- Create your feature branch `git checkout -b my-new-feature`
- Commit your changes `git commit -am 'Add my feature'`
- Push to the branch `git push origin my-new-feature`
- Create a new pull request against the master branch

## TODO
- [ ] use https://github.com/spf13/cobra for the CLI interface
- [ ] add the flarelogs command
- [ ] add `all` keyword to export all the domains available in the account
