
all: format install

.PHONY: format
format:
	GO111MODULE=on go mod tidy
	goimports -l -w -local github.com/howardjohn/kubectl-resources *.go cmd/*.go client/*.go

.PHONY: install
install:
	GO111MODULE=on go install -v

