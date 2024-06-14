export GOPROXY=https://goproxy.cn

all: gen build
.PHONY: all

GOCC?=go

titan-workerd-api:
	rm -f titan-workerd-api
	$(GOCC) build $(GOFLAGS) -o titan-workerd-api .
.PHONY: titan-explorer

gen:
	sqlc generate
.PHONY: gen

build: titan-workerd-api
.PHONY: build
