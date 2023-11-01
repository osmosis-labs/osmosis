###############################################################################
###                                 E2E                                     ###
###############################################################################
e2e-help:
	@echo "e2e subcommands"
	@echo ""
	@echo "Usage:"
	@echo "  make e2e-[command]"
	@echo ""
	@echo "Available Commands:"
	@echo "  build-script                          Build e2e script"
	@echo "  docker-build-debug                    Build e2e debug Docker image"
	@echo "  docker-build-e2e-init-chain           Build e2e init chain Docker image"
	@echo "  docker-build-e2e-init-node            Build e2e init node Docker image"
	@echo "  setup                                 Set up e2e environment"
	@echo "  check-image-sha                       Check e2e image SHA"
	@echo "  remove-resources                      Remove e2e resources"
e2e: e2e-help

e2e-build-script:
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./tests/e2e/initialization/$(E2E_SCRIPT_NAME)

e2e-docker-build-debug:
	@DOCKER_BUILDKIT=1 docker build -t osmosis:${COMMIT} --build-arg BASE_IMG_TAG=debug --build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_ALPINE) -f Dockerfile .
	@DOCKER_BUILDKIT=1 docker tag osmosis:${COMMIT} osmosis:debug

e2e-docker-build-e2e-init-chain:
	@DOCKER_BUILDKIT=1 docker build -t osmolabs/osmosis-e2e-init-chain:debug --build-arg E2E_SCRIPT_NAME=chain --platform=linux/x86_64 -f tests/e2e/initialization/init.Dockerfile .

e2e-docker-build-e2e-init-node:
	@DOCKER_BUILDKIT=1 docker build -t osmosis-e2e-init-node:debug --build-arg E2E_SCRIPT_NAME=node --platform=linux/x86_64 -f tests/e2e/initialization/init.Dockerfile .

e2e-setup: e2e-check-image-sha e2e-remove-resources
	@echo Finished e2e environment setup, ready to start the test

e2e-check-image-sha:
	tests/e2e/scripts/run/check_image_sha.sh

e2e-remove-resources:
	tests/e2e/scripts/run/remove_stale_resources.sh

.PHONY: test-mutation