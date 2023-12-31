package v22

import (
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v21/app/keepers"
	"github.com/osmosis-labs/osmosis/v21/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Run migrations before applying any other state changes.
		// NOTE: DO NOT PUT ANY STATE CHANGES BEFORE RunMigrations().
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Properly register consensus params. In the process, change params as per:
		// https://forum.osmosis.zone/t/raise-maximum-gas-to-300m-and-lower-max-bytes-to-5mb/1116
		defaultConsensusParams := tmtypes.DefaultConsensusParams().ToProto()
		defaultConsensusParams.Block.MaxBytes = 5000000
		defaultConsensusParams.Block.MaxGas = 300000000
		keepers.ConsensusParamsKeeper.Set(ctx, &defaultConsensusParams)

		return migrations, nil
	}
}
