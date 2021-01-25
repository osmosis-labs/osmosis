
.PHONY: clean buildpath osmosis proto docs

all: osmosis proto

clean:
	rm -rf ./build

buildpath:
	mkdir -p build

osmosis: clean buildpath
	@echo
	@echo "=========== Build Osmosis ================"
	@echo
	go build -o ./build/osmosisd ./cmd/osmosisd
	@echo
	@echo "=========== Build Complete ==============="
	@echo

proto:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/generate-proto.sh
	@echo
	@echo "=========== Generate Complete ============"
	@echo

test:
	@go test -v ./x/...
docs:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/generate-docs.sh

	statik -src=client/docs/static -dest=client/docs -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
        exit 1;\
    else \
    	echo "\033[92mSwagger docs are in sync\033[0m";\
    fi
	@echo
	@echo "=========== Generate Complete ============"
	@echo
