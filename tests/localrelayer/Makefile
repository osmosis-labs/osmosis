init: clean build

build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f docker-compose.yml build

start:
	@docker-compose -f docker-compose.yml up

stop:
	@docker-compose -f docker-compose.yml down

clean:
	@rm -rfI $(HOME)/.osmosisd-local-a/
	@rm -rfI $(HOME)/.osmosisd-local-b/
