# Copyright (C) 2022 Jingli Chen (Wine93), NetEase Inc.
.PHONY: build debug init proto clean install_grpc_protobuf

version=4.2
GITHUB_PROXY="https://ghproxy.com/"
PROTOC_VERSION= 21.8
PROTOC_GEN_GO_VERSION= "v1.28"
PROTOC_GEN_GO_GRPC_VERSION= "v1.2"

ifndef THIRD_PARTY_INSTALL_PATH
	USER_HOME := $(shell echo $$HOME)
	THIRD_PARTY_INSTALL_PATH := $(USER_HOME)/.local/dingo-eureka
endif

# go env
# GOPROXY     :=https://goproxy.cn,direct
GOPROXY     := "https://proxy.golang.org,direct"
GOOS        := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH      := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
CC          := gcc

GOENV := GO111MODULE=on
GOENV += GOPROXY=$(GOPROXY)
GOENV += CC=$(CC)
GOENV += CGO_ENABLED=1 CGO_LDFLAGS=$(CGO_LDFLAGS)
GOENV += GOOS=$(GOOS) GOARCH=$(GOARCH)
GOENV += CGO_CFLAGS="-I$(THIRD_PARTY_INSTALL_PATH)/include"

# go
GO := go

# output
OUTPUT := sbin/dingo

# version
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status -s | grep -v third-party`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%dT%H:%M:%SZ')
VERSION_FLAG := -X github.com/dingodb/dingofs-tools/pkg/cli/command/common/version.Version=$(version)
VERSION_FLAG += -X github.com/dingodb/dingofs-tools/pkg/cli/command/common/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY}
VERSION_FLAG += -X github.com/dingodb/dingofs-tools/pkg/cli/command/common/version.BuildDate=${BUILD_DATE}

# for rados
BUILD_LDFLAGS	 :=-L$(THIRD_PARTY_INSTALL_PATH)/lib
BUILD_LDFLAGS +=  -l:librados.a  #rados
BUILD_LDFLAGS +=  -l:librdmacm.a -l:libibverbs.a #rdma
BUILD_LDFLAGS +=  -l:libcrypto.a -l:libssl.a #openssl
BUILD_LDFLAGS +=  -l:libboost_thread.a -l:libboost_iostreams.a #boost
BUILD_LDFLAGS +=  -l:libz.a  #libz
BUILD_LDFLAGS +=  -lstdc++ -lm -ldl -lpthread -Wl,--allow-multiple-definition

# build flags
CGO_BUILD_LDFLAGS := -s -w -linkmode external
CGO_BUILD_LDFLAGS += -extldflags "$(BUILD_LDFLAGS)"
CGO_BUILD_FLAG += -ldflags '$(CGO_BUILD_LDFLAGS) $(VERSION_FLAG)'

BUILD_FLAGS := -a
BUILD_FLAGS += -trimpath
BUILD_FLAGS += $(CGO_BUILD_FLAG)
BUILD_FLAGS += $(EXTRA_FLAGS)

# debug flags
GCFLAGS := "all=-N -l"

CGO_DEBUG_LDFLAGS := -linkmode external
CGO_DEBUG_FLAG += -ldflags '$(CGO_DEBUG_LDFLAGS) $(VERSION_FLAG)'

DEBUG_FLAGS := -gcflags=$(GCFLAGS)
DEBUG_FLAGS += $(CGO_DEBUG_FLAG)

# packages
PACKAGES := $(PWD)/cmd/dingo/main.go

build: proto
	$(GOENV) $(GO) build -o $(OUTPUT) $(BUILD_FLAGS) $(PACKAGES)
debug: proto
	$(GOENV) $(GO) build -o $(OUTPUT) $(DEBUG_FLAGS) $(PACKAGES)
init: proto
	go mod init github.com/dingodb/dingofs-tools
	go mod tidy

proto: install_grpc_protobuf
	@bash mk-proto.sh

install_grpc_protobuf:
	# wget ${GITHUB_PROXY}https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip \
    # && unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip "bin/protoc" -d /usr/ \
    # && rm protoc-${PROTOC_VERSION}-linux-x86_64.zip
	go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}

clean:
	rm -rf sbin
	rm -rf proto/*
