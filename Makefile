.PHONY: start build

NOW = $(shell date -u '+%Y%m%d%I%M%S')


VERSION         = v1.0.0      #docker build tag
BIN_ROOT        = ./bin/
SERVER_BIN  	= ./bin/$(WHAT)
RELEASE_ROOT 	= release
RELEASE_SERVER 	= release/$(WHAT)

EXPOSE_PORT = 8080

ifeq ($(WHAT), chatserver)
	EXPOSE_PORT = 6387
endif

all: start

# eg.  make build WHAT=chatserver
build:
	@go build -ldflags "-w -s" -o $(SERVER_BIN) ./cmd/$(WHAT)/main.go

# eg. make image WHAT=chatserver
image:
	docker build -t $(WHAT):$(VERSION) --build-arg SERVER_NAME=$(WHAT) --build-arg SERVER_PORT=$(EXPOSE_PORT) .

run:
	@go run ./cmd/$(WHAT)/main.go

test:
	cd ./test && go test -v

clean:
	rm -rf $(BIN_ROOT)/*

proto:
	protoc --go_out=proto/ proto/*.proto

tidy:
	go mod tidy