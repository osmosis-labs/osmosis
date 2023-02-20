package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icqkeeper "github.com/strangelove-ventures/async-icq/v4/keeper"
	icqtypes "github.com/strangelove-ventures/async-icq/v4/types"
	packetforwardtypes "github.com/strangelove-ventures/packet-forward-middleware/v4/router/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/v14/app/keepers"
	appParams "github.com/osmosis-labs/osmosis/v14/app/params"
	"github.com/osmosis-labs/osmosis/v14/app/upgrades"
	gammkeeper "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

// an array of 10 strings containing the numbers from 1 to 10:

// Using the same queries as stargate_whitelist.go
var whitelistedQueries = []string{
	// cosmos-sdk queries

	// auth
	"/cosmos.auth.v1beta1.Query/Account",
	"/cosmos.auth.v1beta1.Query/Params",

	// bank
	"/cosmos.bank.v1beta1.Query/Balance",
	"/cosmos.bank.v1beta1.Query/DenomMetadata",
	"/cosmos.bank.v1beta1.Query/Params",
	"/cosmos.bank.v1beta1.Query/SupplyOf",

	// distribution
	"/cosmos.distribution.v1beta1.Query/Params",
	"/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress",
	"/cosmos.distribution.v1beta1.Query/ValidatorCommission",

	// gov
	"/cosmos.gov.v1beta1.Query/Deposit",
	"/cosmos.gov.v1beta1.Query/Params",
	"/cosmos.gov.v1beta1.Query/Vote",

	//slashing
	"/cosmos.slashing.v1beta1.Query/Params",
	"/cosmos.slashing.v1beta1.Query/SigningInfo",

	//staking
	"/cosmos.staking.v1beta1.Query/Delegation",
	"/cosmos.staking.v1beta1.Query/Params",
	"/cosmos.staking.v1beta1.Query/Validator",

	//osmosis queries

	//epochs
	"/osmosis.epochs.v1beta1.Query/EpochInfos",
	"/osmosis.epochs.v1beta1.Query/CurrentEpoch",

	//gamm
	"/osmosis.gamm.v1beta1.Query/NumPools",
	"/osmosis.gamm.v1beta1.Query/TotalLiquidity",
	"/osmosis.gamm.v1beta1.Query/Pool",
	"/osmosis.gamm.v1beta1.Query/PoolParams",
	"/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity",
	"/osmosis.gamm.v1beta1.Query/TotalShares",
	"/osmosis.gamm.v1beta1.Query/CalcJoinPoolShares",
	"/osmosis.gamm.v1beta1.Query/CalcExitPoolCoinsFromShares",
	"/osmosis.gamm.v1beta1.Query/CalcJoinPoolNoSwapShares",
	"/osmosis.gamm.v1beta1.Query/PoolType",
	"/osmosis.gamm.v2.Query/SpotPrice",
	"/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn",
	"/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut",

	//incentives
	"/osmosis.incentives.Query/ModuleToDistributeCoins",
	"/osmosis.incentives.Query/LockableDurations",

	//lockup
	"/osmosis.lockup.Query/ModuleBalance",
	"/osmosis.lockup.Query/ModuleLockedAmount",
	"/osmosis.lockup.Query/AccountUnlockableCoins",
	"/osmosis.lockup.Query/AccountUnlockingCoins",
	"/osmosis.lockup.Query/LockedDenom",
	"/osmosis.lockup.Query/LockedByID",

	//mint
	"/osmosis.mint.v1beta1.Query/EpochProvisions",
	"/osmosis.mint.v1beta1.Query/Params",

	//pool-incentives
	"/osmosis.poolincentives.v1beta1.Query/GaugeIds",

	//superfluid
	"/osmosis.superfluid.Query/Params",
	"/osmosis.superfluid.Query/AssetType",
	"/osmosis.superfluid.Query/AllAssets",
	"/osmosis.superfluid.Query/AssetMultiplier",

	//poolmanager
	"/osmosis.poolmanager.v1beta1.Query/NumPools",
	"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn",
	"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut",

	//txfees
	"/osmosis.txfees.v1beta1.Query/FeeTokens",
	"/osmosis.txfees.v1beta1.Query/DenomSpotPrice",
	"/osmosis.txfees.v1beta1.Query/DenomPoolId",
	"/osmosis.txfees.v1beta1.Query/BaseDenom",

	//tokenfactory
	"/osmosis.tokenfactory.v1beta1.Query/params",
	"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata",

	//twap
	"/osmosis.twap.v1beta1.Query/ArithmeticTwap",
	"/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow",
	"/osmosis.twap.v1beta1.Query/GeometricTwap",
	"/osmosis.twap.v1beta1.Query/GeometricTwapToNow",
	"/osmosis.twap.v1beta1.Query/Params",

	//downtime-detector
	"/osmosis.downtimedetector.v1beta1.Query/RecoveredSinceDowntimeOfLength",
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		poolmanagerParams := poolmanagertypes.NewParams(keepers.GAMMKeeper.GetParams(ctx).PoolCreationFee)

		keepers.PoolManagerKeeper.SetParams(ctx, poolmanagerParams)
		keepers.PacketForwardKeeper.SetParams(ctx, packetforwardtypes.DefaultParams())
		setICQParams(ctx, keepers.ICQKeeper)

		// N.B: pool id in gamm is to be deprecated in the future
		// Instead,it is moved to poolmanager.
		migrateNextPoolId(ctx, keepers.GAMMKeeper, keepers.PoolManagerKeeper)

		//  N.B.: this is done to avoid initializing genesis for poolmanager module.
		// Otherwise, it would overwrite migrations with InitGenesis().
		// See RunMigrations() for details.
		fromVM[poolmanagertypes.ModuleName] = 0

		// Metadata for uosmo and uion were missing prior to this upgrade.
		// They are added in this upgrade.
		registerOsmoIonMetadata(ctx, keepers.BankKeeper)

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func setICQParams(ctx sdk.Context, icqKeeper *icqkeeper.Keeper) {
	icqparams := icqtypes.DefaultParams()
	icqparams.AllowQueries = whitelistedQueries
	icqKeeper.SetParams(ctx, icqparams)
}

func migrateNextPoolId(ctx sdk.Context, gammKeeper *gammkeeper.Keeper, poolmanagerKeeper *poolmanager.Keeper) {
	// N.B: pool id in gamm is to be deprecated in the future
	// Instead,it is moved to poolmanager.
	// nolint: staticcheck
	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	poolmanagerKeeper.SetNextPoolId(ctx, nextPoolId)

	for poolId := uint64(1); poolId < nextPoolId; poolId++ {
		poolType, err := gammKeeper.GetPoolType(ctx, poolId)
		if err != nil {
			panic(err)
		}

		poolmanagerKeeper.SetPoolRoute(ctx, poolId, poolType)
	}
}

func registerOsmoIonMetadata(ctx sdk.Context, bankKeeper bankkeeper.Keeper) {
	uosmoMetadata := banktypes.Metadata{
		Description: "The native token of Osmosis",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    appParams.BaseCoinUnit,
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    appParams.HumanCoinUnit,
				Exponent: appParams.OsmoExponent,
				Aliases:  nil,
			},
		},
		Base:    appParams.BaseCoinUnit,
		Display: appParams.HumanCoinUnit,
	}

	uionMetadata := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uion",
				Exponent: 0,
				Aliases:  nil,
			},
			{
				Denom:    "ion",
				Exponent: 6,
				Aliases:  nil,
			},
		},
		Base:    "uion",
		Display: "ion",
	}

	bankKeeper.SetDenomMetaData(ctx, uosmoMetadata)
	bankKeeper.SetDenomMetaData(ctx, uionMetadata)
}
