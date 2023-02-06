export GOVCS=*:git

TARGET   := protohack
WORKDIR  := ./proto/${TASK}
MAIN     := ${WORKDIR}/cmd/main.go
BIN      := ./bin
TIMEOUT  := 15
COVEROUT := cover.out
BRANCH   := ${shell git rev-parse --abbrev-ref HEAD}
REVCNT   := ${shell git rev-list --count ${BRANCH}}
REVHASH  := ${shell git log -1 --format="%h"}
LDFLAGS  := -s -X main.version=${BRANCH}.${REVCNT}.${REVHASH}
CFLAGS   := --ldflags '${LDFLAGS}' -o ${BIN}/${TARGET}
FLY_APP  := protohacker-go

all: lint test build

lint:
	ls ./proto | xargs -I@ golangci-lint run --modules-download-mode= ./proto/@/...

run:
	go run --ldflags '${LDFLAGS}' ${MAIN}

dev-run:
	go run -race --ldflags '${LDFLAGS}' ${MAIN}


cover:
	go tool cover -html=${COVEROUT}
	-rm -f ${COVEROUT}

test:
	CGO_ENABLED=0 go list -f '{{.Dir}}' proto/... | xargs \
		go test -timeout ${TIMEOUT}s -cover -coverprofile=${COVEROUT}

${WORKDIR}:
	mkdir -p ${WORKDIR}/cmd && \
		cd ${WORKDIR} && \
		go mod init proto/${TASK} && \
		echo "package main\n\nfunc main{} {\n}" > cmd/main.go && \
		cd - && \
	go work use ${WORKDIR}

new: ${WORKDIR}

build:
	go build ${CFLAGS} ${MAIN}

build-linux: clean
	CGO_ENABLED=0 GOOS=linux go build ${CFLAGS} -a -installsuffix cgo ${MAIN}

build-rpi:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build ${CFLAGS} ${MAIN}

launch:
	@cp ${WORKDIR}/fly.toml . 2>/dev/null && \
	sed "s/##TASK##/${TASK}/" Dockerfile.tmpl > Dockerfile && \
	fly launch --copy-config --local-only --name ${FLY_APP} \
		--no-deploy -r lhr && \
	fly ips allocate-v6 -a ${FLY_APP} || echo 'error: export TASK variable'
	@rm -f ./fly.toml
	@rm -f ./Dockerfile

deploy:
	@cp ${WORKDIR}/fly.toml . 2>/dev/null && \
	sed "s/##TASK##/${TASK}/" Dockerfile.tmpl > Dockerfile && \
	fly deploy --local-only || echo 'error: export TASK variable'
	@rm -f ./Dockerfile
	@rm -f ./fly.toml

destroy:
	fly destroy protohacker-go

clean:
	-rm -rf ${BIN}
	-rm -rf ./dist
	-rm -f ${COVEROUT}
	-rm -f ./Dockerfile
	-rm -f ./fly.toml

.PHONY: build test
