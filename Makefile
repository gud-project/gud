LIB_DIR=./gud/
CLI_DIR=./client/
SERVER_DIR=./server/

go_src=$(shell find $(1) -not -path '$(1)/vendor/**' -not -name *_test.go \( -name *.go -o -name go.* \))

LIB_SRC=$(call go_src,$(LIB_DIR))
CLI_SRC=$(call go_src,$(CLI_DIR))
SERVER_SRC=$(call go_src,$(SERVER_DIR))

.PHONY: client server
.ONESHELL: client server

define build_with_lib
	cd $(1)
	sed -i 's/"gitlab.com\/magsh-2019\/2\/gud\/gud"/\/\/ \0/g' **/*.go
	go mod vendor
	sed -i 's/\/\/ \("gitlab.com\/magsh-2019\/2\/gud\/gud"\)/\1/g' **/*.go
	GO111MODULE=off go build
endef

all: client server
client: client/client
server: server/server

gud/gud.a: $(LIB_SRC)
	cd gud
	go mod vendor
	go build -o gud.a

client/client: gud/gud.a $(CLI_SRC)
	$(call build_with_lib,$(CLI_DIR))

server/server: gud/gud.a $(SERVER_SRC)
	$(call build_with_lib,$(SERVER_DIR))
