all: format test lint install

.PHONY: format
format:
	GO111MODULE=on go mod tidy
	goimports -l -w -local github.com/howardjohn/kubectl-resources .

.PHONY: lint
lint:
	GO111MODULE=on golangci-lint run --fix

.PHONY: install
install:
	GO111MODULE=on go install -v

.PHONY: test
test:
	GO111MODULE=on go test ./...

.PHONY: run
run:
	GO111MODULE=on go run main.go
