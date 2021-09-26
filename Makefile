EXECUTABLE := buyte
VERSION := 1.0.0
PACKAGENAME := $(shell go list -f '{{ .ImportPath }}')

.PHONY: default
default: ${EXECUTABLE}

.PHONY: init
init:
	curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s

.PHONY: clean
clean:
	rm -rf /usr/local/bin/${EXECUTABLE} ./bin ./vendor

.PHONY: build
build: dep
	go build -o bin/${EXECUTABLE} *.go

.PHONY: container-build
container-build:
	GOOS=linux GOARCH=amd64 go build -ldflags '-s -w -X "${PACKAGENAME}/conf.Executable=${EXECUTABLE}" -X "${PACKAGENAME}/conf.Version=${VERSION}"' -o ./bin/${EXECUTABLE}

.PHONY: ${EXECUTABLE}
${EXECUTABLE}:
	# Compiling...
	go build -ldflags '-X "${PACKAGENAME}/conf.Executable=${EXECUTABLE}" -X "${PACKAGENAME}/conf.Version=${VERSION}"' -o bin/${EXECUTABLE}
	chmod +x bin/${EXECUTABLE}
	rm -rf /usr/local/bin/${EXECUTABLE}
	ln -s $(shell echo $(CURDIR))/bin/${EXECUTABLE} /usr/local/bin/${EXECUTABLE}

.PHONY: watch
watch:
	./bin/air -c ./.air.conf

.PHONY: sls-build
sls-build:
	sls build
	sls create_domain

.PHONY: sls-deploy
sls-deploy: sls-build
	sls deploy --verbose
