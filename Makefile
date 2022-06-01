SHELL:=/bin/sh
.PHONY: all build test clean

export GO111MODULE=on

# Path Related
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
RELEASE_DIR := ${MKFILE_DIR}/build/bin

# Version
RELEASE?=v0.1.0

# Git Related
GIT_REPO_INFO=$(shell cd ${MKFILE_DIR} && git config --get remote.origin.url)
ifndef GIT_COMMIT
  GIT_COMMIT := git-$(shell git rev-parse --short HEAD)
endif


# go source files, ignore vendor directory
SOURCE = $(shell find ${MKFILE_DIR} -type f -name "*.go")
TARGET = ${RELEASE_DIR}/easeprobe

all: ${TARGET}

${TARGET}: ${SOURCE}
	mkdir -p ${RELEASE_DIR}
	go mod tidy
	go build -a -ldflags '-s -w -extldflags "-static"' -gcflags=-G=3 -o ${TARGET} ${MKFILE_DIR}cmd/easeprobe

build: all

test:
	go test -gcflags=-l  -cover -race -count=1 ./...

docker:
	sudo DOCKER_BUILDKIT=1 docker build -t megaease/easeprobe -f ${MKFILE_DIR}/resources/Dockerfile ${MKFILE_DIR}

clean:
	@rm -rf ${MKFILE_DIR}/build
