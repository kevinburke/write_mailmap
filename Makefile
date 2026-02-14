SHELL = /bin/bash -o pipefail

BUMP_VERSION := $(GOPATH)/bin/bump_version
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap

vet:
	go vet ./...
	staticcheck ./...

test: vet
	go test ./...

race-test: vet
	go test -race ./...

$(WRITE_MAILMAP):
	go install .

force: ;

AUTHORS.txt: force | $(WRITE_MAILMAP)
	$(WRITE_MAILMAP) > AUTHORS.txt

authors: AUTHORS.txt

$(BUMP_VERSION):
	go install github.com/kevinburke/bump_version@latest

release: race-test | $(BUMP_VERSION)
	$(BUMP_VERSION) --tag-prefix=v minor main.go
