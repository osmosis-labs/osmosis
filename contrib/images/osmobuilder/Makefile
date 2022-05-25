# Script to automatically build binaries 
# for linux/amd64, linux/arm64

# It builds a osmosisd binary while build the docker image
# in order to create a single binary statically linked to cosmwasm

# Usage:
# 
# Create a new release by running from the root of the repository:
# make -f contrib/images/osmobuilder/Makefile release
#
# A release/ folder will be created with the appropriate files 

TAG:=9.0.0
IMAGE:=osmobuilder:$(TAG)

.PHONY: create-dockerx-builder build-binary-amd64 get-binary-amd64

release: get-binary-arm64 get-binary-amd64 git sha

# Run `create-dockerx-builder` to create the builder first
create-dockerx-builder:
	docker buildx create --name osmobuilder || true

# Build image for amd64 architecture
build-binary-amd64: create-dockerx-builder
	docker buildx use osmobuilder
	docker buildx build --platform linux/amd64 -t $(IMAGE)-amd64 --load -f contrib/images/osmobuilder/Dockerfile .

# Get binary from image for amd64 architecture
get-binary-amd64: build-binary-amd64
	mkdir -p release/
	docker rm -f osmobinary || true
	docker create -ti --name osmobinary $(IMAGE)-amd64
	docker cp osmobinary:/artifacts/osmosisd release/osmosis-$(TAG)-linux-amd64
	tar -zcvf release/osmosis-$(TAG)-linux-amd64.tar.gz release/osmosis-$(TAG)-linux-amd64
	docker rm -f osmobinary

# Build image for arm64 architecture
build-binary-arm64: create-dockerx-builder
	docker buildx use osmobuilder
	docker buildx build --platform linux/arm64 -t $(IMAGE)-arm64 --load -f contrib/images/osmobuilder/Dockerfile .

# Get binary from image for arm64 architecture
get-binary-arm64: build-binary-arm64
	mkdir -p release/
	docker rm -f osmobinary || true
	docker create -ti --name osmobinary $(IMAGE)-arm64
	docker cp osmobinary:/artifacts/osmosisd release/osmosis-$(TAG)-linux-arm64
	tar -zcvf release/osmosis-$(TAG)-linux-arm64.tar.gz release/osmosis-$(TAG)-linux-arm64
	docker rm -f osmobinary

# Calculate sha
sha:
	mkdir -p release/
	rm -f release/sha256sum.txt
	sha256sum release/* | sed 's#release/##g' > release/sha256sum.txt

# Create git archive
git:
	mkdir -p release/
	git archive \
		--format zip \
		--prefix "osmosis-$(TAG)/" \
		-o "release/Source code.zip" \
		HEAD 

	git archive \
		--format tar.gz \
		--prefix "osmosis-$(TAG)/" \
		-o "release/Source code.tar.gz" \
		HEAD 

