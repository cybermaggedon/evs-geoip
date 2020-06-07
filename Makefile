
VERSION=$(shell git describe | sed 's/^v//')
REPO=cybermaggedon/evs-geoip
DOCKER=docker

all: build

build:
	${DOCKER} build -t ${REPO}:${VERSION} -f Dockerfile .

push:
	${DOCKER} push ${REPO}:${VERSION}

