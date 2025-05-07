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
	@echo "  all                   Run all linters"
	@echo "  fix-typo              Run codespell to fix typos"
	@echo "  format                Run linters with auto-fix"
	@echo "  markdown              Run markdown linter with auto-fix"
	@echo "  mdlint                Run markdown linter"
	@echo "  setup-pre-commit      Set pre-commit git hook"
	@echo "  typo                  Run codespell to check typos"
lint: lint-help

lint-all:
	@echo "--> Running linter"
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run --timeout=10m
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md"

lint-format:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run ./... --fix
	@go run mvdan.cc/gofumpt -l -w x/ app/ ante/ tests/
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md" --fix

lint-mdlint:
	@echo "--> Running markdown linter"
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md"

lint-markdown:
	@docker run -v $(PWD):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest "**/*.md" --fix

lint-typo:
	@codespell

lint-fix-typo:
	@codespell -w

lint-setup-pre-commit:
	@cp .git/hooks/pre-commit .git/hooks/pre-commit.bak 2>/dev/null || true
	@echo "Installing pre-commit hook..."
	@ln -sf ../../scripts/hooks/pre-commit.sh .git/hooks/pre-commit
	@echo "Pre-commit hook installed successfully"
	
