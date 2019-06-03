# Flares 🔥

Flares is a CloudFlare DNS backup tool, it dumps your DNS table to the screen or exports it as BIND formatted zone 
files.

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/lfaoro/flares)](https://goreportcard.com/report/github.com/lfaoro/flares)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flfaoro%2Fflares.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Flfaoro%2Fflares?ref=badge_shield)

## Quick Start

### [Video Tutorial](https://asciinema.org/a/NLVa6TyQzvTEhnzZDdH1q79lO)

### Docker
```bash
# Fetch your CloudFlare API key from here:
# https://dash.cloudflare.com/profile -> Global API Key -> View

$ export CF_API_KEY=abcdef1234567890
$ export CF_API_EMAIL=someone@example.com

$ docker run -it --rm \
-e CF_API_KEY="$CF_API_KEY" \
-e CF_API_EMAIL="$CF_API_EMAIL" \
lfaoro/flares domain1.tld domain2.tld
```

### macOS
```bash
brew install lfaoro/tap/flares
```

### Linux (soon)
```bash
curl apionic.com/flares.sh | bash
```

### Developers
> Go installer: https://golang.org/dl/
```bash
go get -u github.com/lfaoro/flares
make install
flares -h

make test
```

## Examples

```bash
$ make install
$ flares -h

$ flares domain1.tld
;;
;; Domain:     domain1.tld
;; Exported:   2019-06-03 06:31:29
...continued

$ flares --export domain1.tld domain2.tld 
BIND table for domain1.tld successfully exported
BIND table for domain2.tld successfully exported
$ ls
domain1.tld domain2.tld
```

## Automation

### GitLab CI/CD

- Copy [.gitlab-ci.yml](.gitlab-ci.yml) inside your repo
- Use the [pipeline schedules](https://gitlab.com/help/user/project/pipelines/schedules) feature
- Each run of the task will generate a DNS backup stored as a downloadable artifact

# Contributing

> Any help and suggestions are very welcome and appreciated. Start by opening an [issue](https://github.com/lfaoro/flares/issues/new).

- Fork the project
- Create your feature branch `git checkout -b my-new-feature`
- Commit your changes `git commit -am 'Add my feature'`
- Push to the branch `git push origin my-new-feature`
- Create a new pull request against the master branch

