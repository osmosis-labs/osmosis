init: clean build

build:
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker-compose -f docker-compose.yml build

start:
	@docker-compose -f docker-compose.yml up

startd:
	@docker-compose -f docker-compose.yml up -d

stop:
	@docker-compose -f docker-compose.yml down -t 3


restart: stop
	@docker-compose -f docker-compose.yml up --force-recreate 

restartd: stop
	@docker-compose -f docker-compose.yml up --force-recreate -d

clean:
	@rm -rfI $(HOME)/.osmosisd-local-a/
	@rm -rfI $(HOME)/.osmosisd-local-b/
