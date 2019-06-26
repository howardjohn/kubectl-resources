
.PHONY: format
format:
	goimports -l -w -local github.com/howardjohn/kubectl-resources *.go cmd/*.go

.PHONY: vendor
vendor:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

.PHONY: install
install:
	go install -v

all: install vendor
