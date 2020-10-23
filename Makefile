
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

docs:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/generate-docs.sh
	@echo
	@echo "=========== Generate Complete ============"
	@echo
