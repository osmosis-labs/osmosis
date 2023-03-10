localosmosis-init: localosmosis-clean localosmosis-build

localosmosis-build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f $(PWD)/tests/localosmosis/docker-compose.yml --profile empty-state build

localosmosis-start:
	@docker-compose -f $(PWD)/tests/localosmosis/docker-compose.yml --profile empty-state up 

localosmosis-startd:
	@docker-compose -f $(PWD)/tests/localosmosis/docker-compose.yml --profile empty-state up -d

localosmosis-stop:
	@docker-compose -f $(PWD)/tests/localosmosis/docker-compose.yml --profile empty-state down 

localosmosis-clean:
	@rm -rfI $(HOME)/.osmosisd-local/

localosmosis-mainnet-init: localosmosis-clean localosmosis-build 

localosmosis-mainnet-start:
	@docker-compose -f $(PWD)/tests/localosmosis/docker-compose.yml --profile mainnet-state up 

localosmosis-mainnet-startd:
	@docker-compose -f $(PWD)/tests/localosmosis/docker-compose.yml --profile mainnet-state up -d

localosmosis-mainnet-stop:
	@docker-compose $(PWD)/tests/localosmosis/docker-compose.yml --profile mainnet-state down

localosmosis-keys:
	$(PWD)/tests/localosmosis/utils/add_keys.sh

localosmosis-upgrade:
	$(PWD)/tests/localosmosis/utils/submit_upgrade_proposal.sh
