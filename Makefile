.PHONY: build get-deps
.DEFAULT_GOAL := install

GITHUB_ORG=pirogoeth
GITHUB_REPO=stratoid

UPSTREAM=$(GITHUB_ORG)/$(GITHUB_REPO)

VERSION=0.1.0
SHA=$(shell git rev-parse HEAD | cut -b1-9)

LDFLAGS="-X main.Version=$(VERSION) -X main.BuildHash=$(SHA)"

clean:
	rm -rf release/

get-deps:
	go get -u github.com/Masterminds/glide
	glide install

build:
	go build -v -ldflags $(LDFLAGS) ./cmd/stratoid
	go build -v -ldflags $(LDFLAGS) ./cmd/stratctl

install:
	go install -v -ldflags $(LDFLAGS) ./cmd/stratoid
	go install -v -ldflags $(LDFLAGS) ./cmd/stratctl

build-release: build-release-stratoid build-release-stratctl

build-release-stratoid: release/stratoid-linux-amd64 release/stratoid-darwin-amd64

build-release-stratcmd: release/stratcmd-linux-amd64 release/stratctl-darwin-amd64

release/stratoid-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) \
		 -o release/stratoid-linux-amd64 cmd/stratoid/main.go

release/stratoid-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) \
		 -o release/stratoid-darwin-amd64 cmd/stratoid/main.go

release/stratctl-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) \
		 -o release/stratctl-linux-amd64 cmd/stratctl/main.go

release/stratctl-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) \
		 -o release/stratctl-darwin-amd64 cmd/stratctl/main.go

release: build-release
	@if [ "$(VERSION)" = "" ]; then \
		echo " # 'VERSION' variable not set! To preform a release do the following"; \
		echo "  git tag v1.0.0"; \
		echo "  git push --tags"; \
		echo "  make release VERSION=v1.0.0"; \
		echo ""; \
		exit 1; \
	fi
	@if ! which github-release 2>&1 >> /dev/null; then \
		echo " # github-release not found in path; install and create a github token with 'repo' access"; \
		echo " # See (https://help.github.com/articles/creating-an-access-token-for-command-line-use)"; \
		echo " go get github.com/aktau/github-release"; \
		echo " export GITHUB_TOKEN=<your-token>";\
		echo ""; \
		exit 1; \
	fi
	@github-release release \
		--user $(GITHUB_ORG) \
		--repo $(GITHUB_REPO) \
		--tag $(VERSION)
	@github-release upload \
		--user $(GITHUB_ORG) \
		--repo $(GITHUB_REPO) \
		--tag $(VERSION) \
		--name "stratoid-linux-amd64" \
		--file release/stratoid-linux-amd64
	@github-release upload \
		--user $(GITHUB_ORG) \
		--repo $(GITHUB_REPO) \
		--tag $(VERSION) \
		--name "stratoid-darwin-amd64" \
		--file release/stratoid-darwin-amd64
	@github-release upload \
		--user $(GITHUB_ORG) \
		--repo $(GITHUB_REPO) \
		--tag $(VERSION) \
		--name "stratctl-linux-amd64" \
		--file release/stratctl-linux-amd64
	@github-release upload \
		--user $(GITHUB_ORG) \
		--repo $(GITHUB_REPO) \
		--tag $(VERSION) \
		--name "stratctl-darwin-amd64" \
		--file release/stratctl-darwin-amd64