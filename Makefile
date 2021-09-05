# go env
GO_CMD = go
GO_CLEAN = go clean
GO_BUILD = go build
GO_TEST = go test

all: clean build test

clean:
	$(GO_CLEAN) -v

build:
	$(GO_BUILD) -v

test:
	$(GO_TEST)