###############################################################################
###                           Dependency Updates                            ###
###############################################################################
deps-help:
	@echo "Dependency Update subcommands"
	@echo ""
	@echo "Usage:"
	@echo "  make deps-[command]"
	@echo ""
	@echo "Available Commands:"
	@echo "  go-mod-cache             Download go modules to local cache"
	@echo "  go.sum                   Ensure dependencies have not been modified"
	@echo "  draw                     Create a dependency graph"
	@echo "  clean                    Remove artifacts"
	@echo "  distclean                Remove vendor directory"
	@echo "  update-sdk-version       Update SDK version"
	@echo "  tidy-workspace           Tidy workspace"
deps: deps-help

deps-go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

deps-go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@GOWORK=off go mod verify

deps-draw:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/osmosisd -d 2 | dot -Tpng -o dependency-graph.png

dpes-clean:
	rm -rf $(CURDIR)/artifacts/

deps-distclean: clean
	rm -rf vendor/

VERSION := 
MODFILES := ./go.mod ./osmoutils/go.mod ./osmomath/go.mod ./x/epochs/go.mod ./x/ibc-hooks/go.mod ./tests/cl-genesis-positions/go.mod ./tests/cl-go-client/go.mod
# run with VERSION argument specified
# e.g) make update-sdk-version VERSION=v0.45.1-0.20230523200430-193959b898ec
# This will change sdk dependencyu version for go.mod in root directory + all sub-modules in this repo.
deps-update-sdk-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION not set"; \
		exit 1; \
	fi
	@echo "Updating version to $(VERSION)"
	@for modfile in $(MODFILES); do \
		if [ -e "$$modfile" ]; then \
			sed -i '' 's|github.com/osmosis-labs/cosmos-sdk v[0-9a-z.\-]*|github.com/osmosis-labs/cosmos-sdk $(VERSION)|g' $$modfile; \
			cd `dirname $$modfile`; \
			go mod tidy; \
			cd - > /dev/null; \
		else \
			echo "File $$modfile does not exist"; \
		fi; \
	done

deps-tidy-workspace:
	@./scripts/tidy_workspace.sh
