LIB_DIR=./gud
CMD_DIR=./cmd
SERVER_DIR=./server
FRONT_DIR=./server/front
RELEASE_DIR=./server/front/dist/release

go_src=$(shell find $(1) -not -name *_test.go \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \))

LIB_SRC=$(call go_src,$(LIB_DIR))
CMD_SRC=$(call go_src,$(CMD_DIR))
SERVER_SRC=$(call go_src,$(SERVER_DIR))
FRONT_SRC=$(shell find $(FRONT_DIR)/src/ \( -name *.js -o -name *.vue \))

.PHONY: all cli server back front bin
.ONESHELL: back

all: cli server

cli: main.go $(CMD_SRC) $(LIB_SRC)
	go install

server: back front bin

back: $(SERVER_DIR)/server
$(SERVER_DIR)/server: $(SERVER_SRC) $(LIB_SRC)
	cd $(SERVER_DIR)
	go build

front: $(FRONT_DIR)/dist/index.html
$(FRONT_DIR)/dist/index.html: $(FRONT_SRC)
	npm install --prefix $(FRONT_DIR)
	npm run --prefix $(FRONT_DIR) build -- --mode $(if $(PROD),production,development)

bin: main.go $(CMD_SRC) $(LIB_SRC)
	for target in 'windows' 'darwin' 'linux'
	do
		[ $$target = 'windows' ] && ext='.exe' || ext=''
		GOOS="$$target" GOARCH=amd64 go build -o "$(RELEASE_DIR)/$$target/gud$$ext"
	done
