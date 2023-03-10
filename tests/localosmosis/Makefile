localosmosis-init: localosmosis-clean localosmosis-build

localosmosis-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f docker-compose.yml --profile empty-state build

localosmosis-start:
	@docker-compose -f docker-compose.yml --profile empty-state up 

localosmosis-startd:
	@docker-compose -f docker-compose.yml --profile empty-state up -d

localosmosis-stop:
	@docker-compose -f docker-compose.yml --profile empty-state down 

localosmosis-clean:
	@rm -rfI $(HOME)/.osmosisd-local/

localosmosis-mainnet-init: localosmosis-clean localosmosis-build 

localosmosis-mainnet-start:
	@docker-compose -f docker-compose.yml --profile mainnet-state up 

localosmosis-mainnet-startd:
	@docker-compose -f docker-compose.yml --profile mainnet-state up -d

localosmosis-mainnet-stop:
	@docker-compose docker-compose.yml --profile mainnet-state down

localosmosis-keys:
	./utils/add_keys.sh

localosmosis-upgrade:
	./utils/submit_upgrade_proposal.sh
