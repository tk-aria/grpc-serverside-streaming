# Makefile
# help:
#    all)
#    make --always-make
.DEFAULT_GOAL := help
MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

.PHONY: help
help:
	@echo "\n>> help [ command list ]"
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'grpc playground => (http://grpcui.hinastory.com/)'
	@echo ''
	@echo 'Targets:'
	@grep -E '^.PHONY: [a-zA-Z_-]+.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = " "}; {printf "\033[35m%-30s\033[32m %s\n", $$2, $$4}'
	@echo ""

.PHONY: setup ## [category]`description`.
setup:
	go build

	apt-get install --yes htop
	wget -O- https://bit.ly/glances | /bin/bash
	wget https://tar.goaccess.io/goaccess-1.4.tar.gz
	tar -xzvf goaccess-1.4.tar.gz
	cd goaccess-1.4/ && \
		./configure --enable-utf8 --enable-geoip=legacy && \
		make && \
		make install

.PHONY: clear_proto ## [category]`description`.
clear_proto:
	rm -r model/*

.PHONY: protoc-client ## [category]`description`.
protoc-client:
	cd proto && protoc proto/*.proto \
		--go_out=model/ \
		--cpp_out=model/ \
		--csharp_out=model/ \
		--java_out=model/ \
		--js_out=model/ \
		--objc_out=model/ \
		--php_out=model/ \
		--python_out=model/ \
		--ruby_out=model/

.PHONY: protoc-server ## [category]`description`.
protoc-server:
	cd proto && protoc ./*.proto \
		--go_out=plugins=grpc:../model
	
