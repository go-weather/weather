GOPATH?=$$HOME/go

all: build

build:
	go build

fmt:
	for f in *.go; do go fmt $$f && sed -i -e 's/	/  /g' $$f; done
