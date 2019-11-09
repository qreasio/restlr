SERVICE = restlr
BUILD_TIME=`date -u '+%Y-%m-%d_%I:%M:%S%p'`
COMMIT=`git rev-parse HEAD`
LDFLAGS=-ldflags "-s -w -extldflags '-static' -X main.BuildStamp=$(BUILD_TIME) -X main.GitHash=$(COMMIT)"
LINT_TOOL=$(shell go env GOPATH)/bin/golangci-lint
GO_PKGS=$(foreach pkg, $(shell go list ./...), $(if $(findstring /vendor/, $(pkg)), , $(pkg)))
GO_FILES=$(shell find . -type f -name '*.go' -not -path './vendor/*')

run:
	go run main.go

clean:
	rm -rf ./bin

fmt:
	@go fmt $(GO_PKGS)
	@goimports -w -l $(GO_FILES)

build:
	env GOOS=linux GOARCH=amd64 go build -o bin/$(SERVICE) -a $(LDFLAGS) .
	chmod +x bin/$(SERVICE)

test:
	@go test -v

$(LINT_TOOL):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.16.0

qc: $(LINT_TOOL)
	$(LINT_TOOL) run --config=.golangci.yaml ./...