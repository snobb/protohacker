export GO111MODULE=on
export GOVCS=*:git

TARGET   := protohack
MAIN     := ./cmd/main.go
BIN      := ./bin
TIMEOUT  := 15
COVEROUT := cover.out
BRANCH   := ${shell git rev-parse --abbrev-ref HEAD}
REVCNT   := ${shell git rev-list --count $(BRANCH)}
REVHASH  := ${shell git log -1 --format="%h"}
LDFLAGS  := -X main.version=${BRANCH}.${REVCNT}.${REVHASH}
CFLAGS   := --ldflags '${LDFLAGS}' -o $(BIN)/$(TARGET)

all: lint test build

lint:
	golangci-lint run

run:
	go run --ldflags '${LDFLAGS}' $(MAIN)

dev-run:
	go run -race --ldflags '${LDFLAGS}' $(MAIN)

cover:
	go tool cover -html=$(COVEROUT)
	-rm -f $(COVEROUT)

test:
	CGO_ENABLED=0 go test -timeout $(TIMEOUT)s -cover -coverprofile=$(COVEROUT) ./pkg/...

build:
	go build ${CFLAGS} $(MAIN)

build-linux: clean
	CGO_ENABLED=0 GOOS=linux go build ${CFLAGS} -a -installsuffix cgo $(MAIN)

build-rpi:
	GOOS=linux GOARCH=arm GOARM=5 go build ${CFLAGS} -o $(BIN)/$(TARGET) $(MAIN)

clean:
	-rm -rf $(BIN)
	-rm -rf ./dist
	-rm -f $(COVEROUT)

.PHONY: build test
