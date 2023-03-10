localosmosis-keys:
	./utils/add_keys.sh

localosmosis-init: localosmosis-clean localosmosis-build

localosmosis-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f docker-compose.yml build

localosmosis-start:
	@docker compose -f docker-compose.yml --profile empty-state up 

localosmosis-startd:
	@docker-compose -f docker-compose.yml --profile empty-state up -d

localosmosis-stop:
	@docker-compose -f docker-compose.yml down

localosmosis-clean:
	@rm -rfI $(HOME)/.osmosisd-local/

localosmosis-state-export-init: localosmosis-state-export-clean localosmosis-state-export-build 

localosmosis-state-export-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f state_export/docker-compose.yml build

localosmosis-state-export-start:
	@docker-compose -f state_export/docker-compose.yml up

localosmosis-state-export-startd:
	@docker-compose -f state_export/docker-compose.yml up -d

localosmosis-state-export-stop:
	@docker-compose -f docker-compose.yml down

localosmosis-state-export-clean: localosmosis-clean

.PHONY: all 