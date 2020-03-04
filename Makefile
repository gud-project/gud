LIB_DIR=./gud/
CLI_DIR=.
SERVER_DIR=./server/

go_src=$(shell find $(1) -not -path '**/vendor/**' -not -name *_test.go \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \))

LIB_SRC=$(call go_src,$(LIB_DIR))
CLI_SRC=$(call go_src,$(CLI_DIR))
SERVER_SRC=$(call go_src,$(SERVER_DIR))

.PHONY: all cli server lib
.ONESHELL: cli server lib

define vendor
	sed -i 's/"gitlab.com\/magsh-2019\/2\/gud\/gud"/\/\/ \0/g' $(call go_src,$(1))
	cd $(1)
	go mod vendor
	cd - >/dev/null
	sed -i 's/\/\/ \("gitlab.com\/magsh-2019\/2\/gud\/gud"\)/\1/g' $(call go_src,$(1))
endef

all: cli server

lib: gud/gud.a
gud/gud.a: $(LIB_SRC)
	cd gud
	go mod vendor
	go build -o gud.a

cli: gud/gud.a $(CLI_SRC)
	$(call vendor,$(CLI_DIR))
	cd $(CLI_DIR)
	GO111MODULE=off go install

server: server/server
server/server: gud/gud.a $(SERVER_SRC)
	$(call vendor,$(SERVER_DIR))
	cd $(SERVER_DIR)
	GO111MODULE=off go build
