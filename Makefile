STATICCHECK := $(shell command -v staticcheck)
RELEASE := $(shell command -v github-release)

vet:
ifndef STATICCHECK
	go get -u honnef.co/go/tools/cmd/staticcheck
endif
	go vet ./...
	staticcheck ./...

test: vet
	go test ./...

release:
ifndef version
	@echo "Please provide a version"
	exit 1
endif
ifndef GITHUB_TOKEN
	@echo "Please set GITHUB_TOKEN in the environment"
	exit 1
endif
	git tag $(version)
	git push origin --tags
	mkdir -p releases/$(version)
	GOOS=linux GOARCH=amd64 go build -o releases/$(version)/write_mailmap-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o releases/$(version)/write_mailmap-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o releases/$(version)/write_mailmap-windows-amd64 .
ifndef RELEASE
	go get -u github.com/aktau/github-release
endif
	# these commands are not idempotent so ignore failures if an upload repeats
	github-release release --user kevinburke --repo write_mailmap --tag $(version) || true
	github-release upload --user kevinburke --repo write_mailmap --tag $(version) --name write_mailmap-linux-amd64 --file releases/$(version)/write_mailmap-linux-amd64 || true
	github-release upload --user kevinburke --repo write_mailmap --tag $(version) --name write_mailmap-darwin-amd64 --file releases/$(version)/write_mailmap-darwin-amd64 || true
	github-release upload --user kevinburke --repo write_mailmap --tag $(version) --name write_mailmap-windows-amd64 --file releases/$(version)/write_mailmap-windows-amd64 || true
