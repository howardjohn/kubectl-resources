
.PHONY: format
format:
	goimports -l -w -local github.com/howardjohn/kubectl-resources *.go cmd/*.go client/*.go

.PHONY: install
install:
	GO111MODULE=on go install -v

all: format install
