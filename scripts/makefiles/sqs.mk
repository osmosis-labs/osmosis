###############################################################################
###                                SQS                                      ###
###############################################################################

sqs-help:
	@echo "sqs subcommands"
	@echo ""
	@echo "Usage:"
	@echo "  make sqs-[command] or make redis-start, make redis-stop"
	@echo ""
	@echo "Available Commands:"
	@echo "  load-test-ui                       Start SQS load ui service"
	@echo "  profile                            Profiling SQS service"
	@echo "  quote-compare                      Compares the quotes between SQS and chain over pool 1136 which is concentrated"
	@echo "  quote-compare-stage                Compares the quotes between SQS and chain over pool 1136 which is concentrated"
	@echo "  start                              Start SQS service"
	@echo "  update-mainnet-state               Updates go tests with the latest mainnet state and make sure that the node is running locally"
	@echo "  validate-cl-state                  Validates that SQS concentrated liquidity pool state is consistent with the state of the chain"

sqs: sqs-help

redis-start:
	docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 -v ./redis-cache/:/data redis/redis-stack:7.2.0-v3

redis-stop:
	docker container rm -f redis-stack

sqs-start:
	./scripts/debug_builder.sh
	build/osmosisd start

sqs-load-test-ui:
	docker compose -f ingest/sqs/locust/docker-compose.yml up --scale worker=4

sqs-profile:
	go tool pprof -http=:8080 http://localhost:9092/debug/pprof/profile?seconds=15

# Validates that SQS concentrated liquidity pool state is
# consistent with the state of the chain.
sqs-validate-cl-state:
	ingest/sqs/scripts/validate-cl-state.sh "http://localhost:9092"

# Compares the quotes between SQS and chain over pool 1136
# which is concentrated.
sqs-quote-compare:
	ingest/sqs/scripts/quote.sh "http://localhost:9092"

sqs-quote-compare-stage:
	ingest/sqs/scripts/quote.sh "http://165.227.168.61"

# Updates go tests with the latest mainnet state
# Make sure that the node is running locally
sqs-update-mainnet-state:
	curl -X POST "http:/localhost:9092/router/store-state"
	mv pools.json ingest/sqs/router/usecase/routertesting/parsing/pools.json
	mv taker_fees.json ingest/sqs/router/usecase/routertesting/parsing/taker_fees.json