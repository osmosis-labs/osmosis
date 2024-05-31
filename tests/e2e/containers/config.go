package containers

// ImageConfig contains all images and their respective tags
// needed for running e2e tests.
type ImageConfig struct {
	InitRepository string
	InitTag        string

	SymphonyRepository string
	SymphonyTag        string

	RelayerRepository string
	RelayerTag        string
}

//nolint:deadcode
const (
	// Current Git branch symphony repo/version. It is meant to be built locally.
	// It is used when skipping upgrade by setting OSMOSIS_E2E_SKIP_UPGRADE to true).
	// This image should be pre-built with `make e2e-docker-build-debug` either in CI or locally.
	CurrentBranchMelodyRepository = "symphony"
	CurrentBranchMelodyTag        = "debug"
	// Pre-upgrade symphony repo/tag to pull.
	// It should be uploaded to Docker Hub. OSMOSIS_E2E_SKIP_UPGRADE should be unset
	// for this functionality to be used.
	previousVersionMelodyRepository = "osmolabs/symphony"
	previousVersionMelodyTag        = "22.0.0-alpine"
	// Pre-upgrade repo/tag for symphony initialization (this should be one version below upgradeVersion)
	previousVersionInitRepository = "osmolabs/symphony-e2e-init-chain"
	previousVersionInitTag        = "22.0.0"
	// Hermes repo/version for relayer
	relayerRepository = "informalsystems/hermes"
	relayerTag        = "1.5.1"
)

// Returns ImageConfig needed for running e2e test.
// If isUpgrade is true, returns images for running the upgrade
// If isFork is true, utilizes provided fork height to initiate fork logic
func NewImageConfig(isUpgrade, isFork bool) ImageConfig {
	config := ImageConfig{
		RelayerRepository: relayerRepository,
		RelayerTag:        relayerTag,
	}

	if !isUpgrade {
		// If upgrade is not tested, we do not need InitRepository and InitTag
		// because we directly call the initialization logic without
		// the need for Docker.
		config.SymphonyRepository = CurrentBranchMelodyRepository
		config.SymphonyTag = CurrentBranchMelodyTag
		return config
	}

	// If upgrade is tested, we need to utilize InitRepository and InitTag
	// to initialize older state with Docker
	config.InitRepository = previousVersionInitRepository
	config.InitTag = previousVersionInitTag

	if isFork {
		// Forks are state compatible with earlier versions before fork height.
		// Normally, validators switch the binaries pre-fork height
		// Then, once the fork height is reached, the state breaking-logic
		// is run.
		config.SymphonyRepository = CurrentBranchMelodyRepository
		config.SymphonyTag = CurrentBranchMelodyTag
	} else {
		// Upgrades are run at the time when upgrade height is reached
		// and are submitted via a governance proposal. Therefore, we
		// must start running the previous Symphony version. Then, the node
		// should auto-upgrade, at which point we can restart the updated
		// Symphony validator container.
		config.SymphonyRepository = previousVersionMelodyRepository
		config.SymphonyTag = previousVersionMelodyTag
	}

	return config
}
