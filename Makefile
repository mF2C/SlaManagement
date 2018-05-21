# Executable name
OUT := SLALite
# Package name
PKG := SLALite
# e.g. mf2c/sla-management
IMAGE := slalite
# Version identifier for HEAD
VERSION := $(shell git describe --always --dirty)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%S)

all: run

build:
	go build -i -v -o ${OUT} -ldflags="-X main.version=${VERSION} -X main.date=${DATE}" ${PKG}

test:
	go test -v ./...

run: build
	./${OUT}

docker: 
	docker build --build-arg VERSION=${VERSION} --build-arg DATE=${DATE} -t ${IMAGE}:${VERSION} .

release_patch:
	resources/bin/release.sh patch

release_minor:
	resources/bin/release.sh minor

release_major:
	resources/bin/release.sh major

clean:
	go clean
	-@rm ${OUT} ${OUT}-v*

.PHONY: build run docker release_patch release_minor release_major clean
