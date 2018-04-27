BUMP_VERSION := $(GOPATH)/bin/bump_version
DIFFER := $(GOPATH)/bin/differ
MEGACHECK := $(GOPATH)/bin/megacheck
RELEASE := $(GOPATH)/bin/github-release
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap

UNAME := $(shell uname)

$(MEGACHECK):
ifeq ($(UNAME), Darwin)
	curl --silent --location --output $(MEGACHECK) https://github.com/kevinburke/go-tools/releases/download/2017-10-04/megacheck-darwin-amd64
endif
ifeq ($(UNAME), Linux)
	curl --silent --location --output $(MEGACHECK) https://github.com/kevinburke/go-tools/releases/download/2017-10-04/megacheck-linux-amd64
endif
	chmod 755 $(MEGACHECK)

vet: $(MEGACHECK)
	go vet ./...
	$(MEGACHECK) ./...

test: vet
	go test ./...

$(BUMP_VERSION):
	go get github.com/kevinburke/bump_version

$(DIFFER):
	go get github.com/kevinburke/differ

$(RELEASE):
	go get -u github.com/aktau/github-release

$(WRITE_MAILMAP):
	go install .

force: ;

AUTHORS.txt: force | $(WRITE_MAILMAP)
	$(WRITE_MAILMAP) > AUTHORS.txt

authors: AUTHORS.txt

release: | $(BUMP_VERSION) $(DIFFER) $(RELEASE)
ifndef version
	@echo "Please provide a version"
	exit 1
endif
ifndef GITHUB_TOKEN
	@echo "Please set GITHUB_TOKEN in the environment"
	exit 1
endif
	$(DIFFER) $(MAKE) authors
	$(BUMP_VERSION) minor main.go
	git push origin --tags
	mkdir -p releases/$(version)
	GOOS=linux GOARCH=amd64 go build -o releases/$(version)/write_mailmap-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o releases/$(version)/write_mailmap-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o releases/$(version)/write_mailmap-windows-amd64 .
	# these commands are not idempotent so ignore failures if an upload repeats
	$(RELEASE) release --user kevinburke --repo write_mailmap --tag $(version) || true
	$(RELEASE) upload --user kevinburke --repo write_mailmap --tag $(version) --name write_mailmap-linux-amd64 --file releases/$(version)/write_mailmap-linux-amd64 || true
	$(RELEASE) upload --user kevinburke --repo write_mailmap --tag $(version) --name write_mailmap-darwin-amd64 --file releases/$(version)/write_mailmap-darwin-amd64 || true
	$(RELEASE) upload --user kevinburke --repo write_mailmap --tag $(version) --name write_mailmap-windows-amd64 --file releases/$(version)/write_mailmap-windows-amd64 || true
