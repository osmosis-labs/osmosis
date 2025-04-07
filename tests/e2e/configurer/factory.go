package configurer

import (
	"errors"
	"testing"

	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/initialization"
)

type Configurer interface {
	ConfigureChains() error

	ClearResources() error

	GetChainConfig(chainIndex int) *chain.Config

	RunSetup() error

	RunValidators() error

	RunIBC() error
}

var (
	// each started validator containers corresponds to one of
	// the configurations below.
	validatorConfigsChainA = []*initialization.NodeConfig{
		{
			// this is a node that is used to state-sync from so its snapshot-interval
			// is frequent.
			Name:               "prune-default-snapshot-state-sync-from",
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   25,
			SnapshotKeepRecent: 10,
			IsValidator:        true,
		},
		{
			Name:               "prune-nothing-snapshot",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "prune-custom-10000-13-snapshot",
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "prune-everything-no-snapshot",
			Pruning:            "everything",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   0,
			SnapshotKeepRecent: 0,
			IsValidator:        true,
		},
	}
	validatorConfigsChainB = []*initialization.NodeConfig{
		{
			Name:               "prune-default-snapshot",
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "prune-nothing-snapshot",
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
		{
			Name:               "prune-custom-snapshot",
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
			IsValidator:        true,
		},
	}
)

// New returns a new Configurer depending on the values of its parameters.
// - If only isIBCEnabled, we want to have 2 chains initialized at the current
// Git branch version of Symphony codebase.
// - If only isUpgradeEnabled, that is invalid and an error is returned.
// - If both isIBCEnabled and isUpgradeEnabled, we want 2 chains with IBC initialized
// at the previous Symphony version.
// - If !isIBCEnabled and !isUpgradeEnabled, we only need one chain at the current
// Git branch version of the Symphony code.
func New(t *testing.T, isIBCEnabled, isDebugLogEnabled bool, upgradeSettings UpgradeSettings) (Configurer, error) {
	t.Helper()
	containerManager, err := containers.NewManager(upgradeSettings.IsEnabled, upgradeSettings.ForkHeight > 0, isDebugLogEnabled)
	if err != nil {
		return nil, err
	}

	if isIBCEnabled && upgradeSettings.IsEnabled {
		// skip none - configure two chains via Docker
		// to utilize the older version of symphony to upgrade from
		return NewUpgradeConfigurer(t,
			[]*chain.Config{
				chain.New(t, containerManager, initialization.ChainAID, validatorConfigsChainA),
				chain.New(t, containerManager, initialization.ChainBID, validatorConfigsChainB),
			},
			withUpgrade(withIBC(baseSetup)), // base set up with IBC and upgrade
			containerManager,
			upgradeSettings.Version,
			upgradeSettings.ForkHeight,
		), nil
	} else if isIBCEnabled {
		// configure two chains from current Git branch
		return NewCurrentBranchConfigurer(t,
			[]*chain.Config{
				chain.New(t, containerManager, initialization.ChainAID, validatorConfigsChainA),
				chain.New(t, containerManager, initialization.ChainBID, validatorConfigsChainB),
			},
			withIBC(baseSetup), // base set up with IBC
			containerManager,
		), nil
	} else if upgradeSettings.IsEnabled {
		// invalid - IBC tests must be enabled for upgrade
		// to function
		return nil, errors.New("IBC tests must be enabled for upgrade to work")
	} else {
		// configure one chain from current Git branch
		return NewCurrentBranchConfigurer(t,
			[]*chain.Config{
				chain.New(t, containerManager, initialization.ChainAID, validatorConfigsChainA),
			},
			baseSetup, // base set up only
			containerManager,
		), nil
	}
}
