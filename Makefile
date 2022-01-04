VGO=go # Set to vgo if building in Go 1.10
BINARY_NAME=chainwall
GOFILES := $(shell find . -name '*.go' -print)
GOBIN := $(shell $(VGO) env GOPATH)/bin
LINT := $(GOBIN)/golangci-lint
.DELETE_ON_ERROR:

all: ethbinding.so
test: deps lint
		$(VGO) test  ./... -cover -coverprofile=coverage.txt -covermode=atomic
coverage.html:
		$(VGO) tool cover -html=coverage.txt
coverage: test coverage.html
ethbinding.so: ${GOFILES} test
		go build -trimpath -o ethbinding.so -buildmode=plugin -tags=prod -v
lint: ${LINT}
		GOGC=20 $(LINT) run -v --timeout 5m
build: ethbinding.so
clean: 
		$(VGO) clean
		rm -f *.so
deps:
		$(VGO) get
${LINT}:
		$(VGO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go-mod-tidy: .ALWAYS
		$(VGO) mod tidy