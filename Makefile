# Makefile
# help:
#    all)
#    make --always-make
#.DEFAULT_GOAL := help
MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

.PHONY: help
help:
	@echo "\n>> help [ command list ]"
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@grep -E '^.PHONY: [a-zA-Z_-]+.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = " "}; {printf "\033[35m%-30s\033[32m %s\n", $$2, $$4}'
	@echo ""

.PHONY: setup ## [category]`description`.
setup:
	@echo 'hello world'

.PHONY: protoc_model_only ## [category]`description`.
protoc_model_only:
	protoc --go_out=model/ proto/*.proto

.PHONY: build_proto ## [category]`description`.
build_proto:
	protoc --go_out=plugins=grpc:./model proto/*.proto
	
