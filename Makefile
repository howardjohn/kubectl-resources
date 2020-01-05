.ONESHELL:
GOBIN ?= $(GOPATH)/bin
MODULE = github.com/howardjohn/kubectl-resources
export GO111MODULE ?= on

all: format lint install

$(GOBIN)/goimports:
	(cd /tmp; go get golang.org/x/tools/cmd/goimports@v0.0.0-20200103221440-774c71fcf114)

$(GOBIN)/golangci-lint:
	(cd /tmp; go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.22.2)

.PHONY: deps
deps: $(GOBIN)/goimports $(GOBIN)/golangci-lint

.PHONY: check-git
check-git:
	@
	if [[ -n $$(git status --porcelain) ]]; then
		echo "Error: git is not clean"
		git status
		git diff
		exit 1
	fi

.PHONY: gen-check
gen-check: check-git format tidy

.PHONY: format
format: $(GOBIN)/goimports
	@go mod tidy
	@goimports -l -w -local $(MODULE) .

.PHONY: lint
lint: $(GOBIN)/golangci-lint
	@golangci-lint run --fix

.PHONY: install
install:
	@go install
