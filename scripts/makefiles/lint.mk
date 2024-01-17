###############################################################################
###                                Linting                                  ###
###############################################################################
lint-help:
	@echo "lint subcommands"
	@echo ""
	@echo "Usage:"
	@echo "  make lint-[command]"
	@echo ""
	@echo "Available Commands:"
	@echo "  all          Run all linters"
	@echo "  format       Run linters with auto-fix"
	@echo "  mdlint       Run markdown linter"
	@echo "  markdown     Run markdown linter with auto-fix"
lint: lint-help

lint-all:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md"

lint-format:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --fix
	@go run mvdan.cc/gofumpt -l -w x/ app/ ante/ tests/
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md" --fix

lint-mdlint:
	@echo "--> Running markdown linter"
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md"

lint-markdown:
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md" --fix
