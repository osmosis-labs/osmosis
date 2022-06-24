package containers

// ImageConfig contains all images and their respective tags
// needed for running e2e tests.
type ImageConfig struct {
	InitRepository string
	InitTag        string

	OsmosisRepository string
	OsmosisTag        string

	RelayerRepository string
	RelayerTag        string
}

const (
	// Current Git branch osmosis repo/version. It is meant to be built locally.
	// It is used when skipping upgrade by setting OSMOSIS_E2E_SKIP_UPGRADE to true).
	// This image should be pre-built with `make docker-build-debug` either in CI or locally.
	CurrentBranchOsmoRepository = "osmosis"
	CurrentBranchOsmoTag        = "debug"
	/// Current Git branch repo/version for osmosis initialization. It is meant to be built locally.
	// It is used when skipping upgrade by setting OSMOSIS_E2E_SKIP_UPGRADE to true).
	// This image should be pre-built with `make docker-build-e2e-chain-init` either in CI or locally.
	currentBranchInitRepository = "osmosis-e2e-chain-init"
	currentBranchInitTag        = "debug"
	// Pre-upgrade osmosis repo/tag to pull.
	// It should be uploaded to Docker Hub. OSMOSIS_E2E_SKIP_UPGRADE should be unset
	// for this functionality to be used.
	previousVersionOsmoRepository = "osmolabs/osmosis-dev"
	previousVersionOsmoTag        = "v8.0.0-2-debug"
	// Pre-upgrade repo/tag for osmosis initialization (this should be one version below upgradeVersion)
	previousVersionInitRepository = "osmolabs/osmosis-e2e-init-chain"
	previousVersionInitTag        = "v8.0.0-rc0"
	// Hermes repo/version for relayer
	relayerRepository = "osmolabs/hermes"
	relayerTag        = "0.13.0"
)

// Returns ImageConfig needed for running e2e test.
// If isUpgrade is true, returns images for running the upgrade
// Otherwise, returns images for running non-upgrade e2e tests.
func NewImageConfig(isUpgrade bool) ImageConfig {
	config := ImageConfig{
		RelayerRepository: relayerRepository,
		RelayerTag:        relayerTag,
	}

	if isUpgrade {
		config.InitRepository = previousVersionInitRepository
		config.InitTag = previousVersionInitTag

		config.OsmosisRepository = previousVersionOsmoRepository
		config.OsmosisTag = previousVersionOsmoTag
	} else {
		config.InitRepository = currentBranchInitRepository
		config.InitTag = currentBranchInitTag

		config.OsmosisRepository = CurrentBranchOsmoRepository
		config.OsmosisTag = CurrentBranchOsmoTag
	}

	return config
}
