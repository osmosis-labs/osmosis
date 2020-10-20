
.PHONY: osmosis message

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

message:
	@echo
	@echo "=========== Generate Protobuf ============"
	@echo
	./scripts/protocgen.sh
	@echo
	@echo "=========== Generate Complete ============"
	@echo
