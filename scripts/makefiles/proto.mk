###############################################################################
###                                  Proto                                  ###
###############################################################################

proto-help:
	@echo "proto subcommands"
	@echo ""
	@echo "Usage:"
	@echo "  make proto-[command]"
	@echo ""
	@echo "Available Commands:"
	@echo "  all        Run proto-format and proto-gen"
	@echo "  format     Format Protobuf files"
	@echo "  gen        Generate Protobuf files"
	@echo "  image-build  Build the protobuf Docker image"
	@echo "  image-push  Push the protobuf Docker image"

proto: proto-help
proto-all: proto-format proto-gen

PROTO_BUILDER_IMAGE=ghcr.io/cosmos/proto-builder:0.14.0
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(PROTO_BUILDER_IMAGE)

proto-all: proto-format proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(DOCKER) run --rm -u 0 -v $(CURDIR):/workspace --workdir /workspace $(PROTO_BUILDER_IMAGE) sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./proto -name "*.proto" -exec clang-format -i {} \;


SWAGGER_DIR=./swagger-proto
THIRD_PARTY_DIR=$(SWAGGER_DIR)/third_party

# Proto sources for swagger generation are pinned to the versions the chain
# builds against (keep in sync with go.mod). Pulling mutable branches breaks
# generation when upstream removes modules this chain still serves (e.g.
# x/params), documents query surfaces the deployed versions do not have, and
# lets the committed output drift without any repo dependency change.
SDK_PROTO_REPO=https://github.com/osmosis-labs/cosmos-sdk.git
SDK_PROTO_REF=v0.50.14-v30-osmo
IBC_PROTO_REF=v8.7.0
WASMD_PROTO_REF=v0.53.3
ICQ_PROTO_REF=modules/async-icq/v8.0.0
BLOCK_SDK_PROTO_REF=v2.1.9-mempool
COSMOS_PROTO_REF=v1.0.0-beta.5
GOGOPROTO_REF=v1.7.0
ICS23_REF=go/v0.11.0
GOOGLEAPIS_REF=93c3926464cdc6bd9410c3be3726e2cd22951fff

proto-download-deps:
	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	git init && \
	git remote add origin "$(SDK_PROTO_REPO)" && \
	git config core.sparseCheckout true && \
	printf "proto\nthird_party\n" > .git/info/sparse-checkout && \
	git pull --depth 1 origin "$(SDK_PROTO_REF)" && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	cd "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/ibc-go.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git pull --depth 1 origin "$(IBC_PROTO_REF)" && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/ibc_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/wasmd_tmp" && \
	cd "$(THIRD_PARTY_DIR)/wasmd_tmp" && \
	git init && \
	git remote add origin "https://github.com/CosmWasm/wasmd.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git pull --depth 1 origin "$(WASMD_PROTO_REF)" && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/wasmd_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/icq_tmp" && \
	cd "$(THIRD_PARTY_DIR)/icq_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/ibc-apps.git" && \
	git config core.sparseCheckout true && \
	printf "modules/async-icq/proto\n" > .git/info/sparse-checkout && \
	git pull --depth 1 origin "$(ICQ_PROTO_REF)" && \
	rm -f ./modules/async-icq/proto/buf.* && \
	mv ./modules/async-icq/proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/icq_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/block_sdk_tmp" && \
	cd "$(THIRD_PARTY_DIR)/block_sdk_tmp" && \
	git init && \
	git remote add origin "https://github.com/osmosis-labs/block-sdk.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git pull --depth 1 origin "$(BLOCK_SDK_PROTO_REF)" && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/block_sdk_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/cosmos-proto.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git pull --depth 1 origin "$(COSMOS_PROTO_REF)" && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_proto_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/gogoproto" && \
	curl -sSL https://raw.githubusercontent.com/cosmos/gogoproto/$(GOGOPROTO_REF)/gogoproto/gogo.proto > "$(THIRD_PARTY_DIR)/gogoproto/gogo.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/google/api" && \
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/$(GOOGLEAPIS_REF)/google/api/annotations.proto > "$(THIRD_PARTY_DIR)/google/api/annotations.proto"
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/$(GOOGLEAPIS_REF)/google/api/http.proto > "$(THIRD_PARTY_DIR)/google/api/http.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/cosmos/ics23/v1" && \
	curl -sSL https://raw.githubusercontent.com/cosmos/ics23/$(ICS23_REF)/proto/cosmos/ics23/v1/proofs.proto > "$(THIRD_PARTY_DIR)/cosmos/ics23/v1/proofs.proto"


docs:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	@make proto-download-deps
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
.PHONY: docs