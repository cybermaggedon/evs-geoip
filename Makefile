
# Create version tag from git tag
VERSION=$(shell git describe | sed 's/^v//')
REPO=cybermaggedon/evs-geoip
DOCKER=docker
GO=GOPATH=$$(pwd)/go go

all: evs-geoip build

evs-geoip: geoip.go go.mod go.sum
	${GO} build -o $@ geoip.go

build: evs-geoip
	${DOCKER} build -t ${REPO}:${VERSION} -f Dockerfile .

push:
	${DOCKER} push ${REPO}:${VERSION}

