all: simd-env

simd-env:
	docker build --build-arg UID=$(shell id -u) --build-arg GID=$(shell id -g) --tag cosmossdk/simd-env simd-env

rbuilder:
	docker build --tag osmolabs/rbuilder rbuilder

.PHONY: all simd-env rbuilder
