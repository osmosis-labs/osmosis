package v12

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	gammtypes "github.com/osmosis-labs/osmosis/v11/x/gamm/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v11/x/superfluid/types"

	"github.com/osmosis-labs/osmosis/v11/app/keepers"
	"github.com/osmosis-labs/osmosis/v11/app/upgrades"
)

// We set the app version to pre-upgrade because it will be incremented by one
// after the upgrade is applied by the handler.
const preUpgradeAppVersion = 11

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Although the app version was already set during the v9 upgrade, our v10 was a fork and
		// v11 was decided to be limited to the "gauge creation minimum fee" change only:
		// https://github.com/osmosis-labs/osmosis/pull/2202
		//
		// As a result, the upgrade handler was not executed to increment the app version.
		// This change helps to correctly set the app version for v12.
		if err := keepers.UpgradeKeeper.SetAppVersion(ctx, preUpgradeAppVersion); err != nil {
			return nil, err
		}

		// Set the max_age_num_blocks in the evidence params to reflect the 14 day
		// unbonding period.
		//
		// Ref: https://github.com/osmosis-labs/osmosis/issues/1160
		cp := bpm.GetConsensusParams(ctx)
		if cp != nil && cp.Evidence != nil {
			evParams := cp.Evidence
			evParams.MaxAgeNumBlocks = 186_092

			bpm.StoreConsensusParams(ctx, cp)
		}

		// Update ICA allowedMessages param
		params := keepers.ICAHostKeeper.GetParams(ctx)
		// Specifying the whole list instead of adding and removing. Less fragile.
		params.AllowMessages = []string{
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
			sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
			sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
			sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
			sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
			sdk.MsgTypeURL(&govtypes.MsgVote{}),
			// Change: Removed authz messages
			sdk.MsgTypeURL(&gammtypes.MsgJoinPool{}),
			sdk.MsgTypeURL(&gammtypes.MsgExitPool{}),
			sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountIn{}),
			sdk.MsgTypeURL(&gammtypes.MsgSwapExactAmountOut{}),
			sdk.MsgTypeURL(&gammtypes.MsgJoinSwapExternAmountIn{}),
			sdk.MsgTypeURL(&gammtypes.MsgJoinSwapShareAmountOut{}),
			sdk.MsgTypeURL(&gammtypes.MsgExitSwapExternAmountOut{}),
			sdk.MsgTypeURL(&gammtypes.MsgExitSwapShareAmountIn{}),
			// Change: Added superfluid unbound
			sdk.MsgTypeURL(&superfluidtypes.MsgSuperfluidUnbondLock{}),
		}
		keepers.ICAHostKeeper.SetParams(ctx, params)

		// Initialize TWAP state
		// TODO: Get allPoolIds from gamm keeper, and write test for migration.
		allPoolIds := []uint64{}
		err := keepers.TwapKeeper.MigrateExistingPools(ctx, allPoolIds)
		if err != nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
