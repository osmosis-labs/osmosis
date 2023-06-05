package v16

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/osmosis-labs/osmosis/v16/app/keepers"
	"github.com/osmosis-labs/osmosis/v16/app/upgrades"

	cosmwasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	cltypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v16/x/superfluid/types"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v16/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v16/x/tokenfactory/types"
)

const (
	// DAI/OSMO pool ID
	// https://app.osmosis.zone/pool/674
	// Note, new concentrated liquidity pool
	// spread factor is initialized to be the same as the balancers pool spread factor of 0.2%.
	DaiOsmoPoolId = uint64(674)
	// Denom0 translates to a base asset while denom1 to a quote asset
	// We want quote asset to be DAI so that when the limit orders on ticks
	// are implemented, we have tick spacing in terms of DAI as the quote.
	DesiredDenom0 = "uosmo"
	// TODO: confirm pre-launch.
	TickSpacing = 1

	// isPermissionlessPoolCreationEnabledCL is a boolean that determines if
	// concentrated liquidity pools can be created via message. At launch,
	// we consider allowing only governance to create pools, and then later
	// allowing permissionless pool creation by switching this flag to true
	// with a governance proposal.
	IsPermissionlessPoolCreationEnabledCL = false
)

var (
	ATOMIBCDenom = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	DAIIBCDenom  = "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"
	USDCIBCDenom = "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"

	// authorized_quote_denoms quote assets that can be used as token1
	// when creating a pool. We limit the quote assets to a small set
	// for the purposes of having convenient price increments stemming
	// from tick to price conversion. These increments are in a human
	// understandeable magnitude only for token1 as a quote.
	authorizedQuoteDenoms []string = []string{
		"uosmo",
		ATOMIBCDenom,
		DAIIBCDenom,
		USDCIBCDenom,
	}

	// authorizedUptimes is the list of uptimes that are allowed to be
	// incentivized. It is a subset of SupportedUptimes (which can be
	// found under CL types) and is set initially to be 1ns, which is
	// equivalent to a 1 block required uptime to qualify for claiming incentives.
	authorizedUptimes []time.Duration = []time.Duration{time.Nanosecond}
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

		// Update expedited governance param
		// In particular, set expedited quorum to 2/3.
		params := keepers.GovKeeper.GetTallyParams(ctx)
		params.ExpeditedQuorum = sdk.NewDec(2).Quo(sdk.NewDec(3))
		keepers.GovKeeper.SetTallyParams(ctx, params)

		// Add both MsgExecuteContract and MsgInstantiateContract to the list of allowed messages.
		hostParams := keepers.ICAHostKeeper.GetParams(ctx)
		msgExecuteContractExists := false
		msgInstantiateContractExists := false
		for _, msg := range hostParams.AllowMessages {
			if msg == sdk.MsgTypeURL(&cosmwasmtypes.MsgExecuteContract{}) {
				msgExecuteContractExists = true
			}
			if msg == sdk.MsgTypeURL(&cosmwasmtypes.MsgInstantiateContract{}) {
				msgInstantiateContractExists = true
			}
		}
		if !msgExecuteContractExists {
			hostParams.AllowMessages = append(hostParams.AllowMessages, sdk.MsgTypeURL(&cosmwasmtypes.MsgExecuteContract{}))
		}

		if !msgInstantiateContractExists {
			hostParams.AllowMessages = append(hostParams.AllowMessages, sdk.MsgTypeURL(&cosmwasmtypes.MsgInstantiateContract{}))
		}
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)

		// Although parameters are set on InitGenesis() in RunMigrations(), we reset them here
		// for visibility of the final configuration.
		defaultConcentratedLiquidityParams := keepers.ConcentratedLiquidityKeeper.GetParams(ctx)
		defaultConcentratedLiquidityParams.AuthorizedQuoteDenoms = authorizedQuoteDenoms
		defaultConcentratedLiquidityParams.AuthorizedQuoteDenoms = authorizedQuoteDenoms
		defaultConcentratedLiquidityParams.AuthorizedUptimes = authorizedUptimes
		defaultConcentratedLiquidityParams.IsPermissionlessPoolCreationEnabled = IsPermissionlessPoolCreationEnabledCL
		keepers.ConcentratedLiquidityKeeper.SetParams(ctx, defaultConcentratedLiquidityParams)

		// Create a concentrated liquidity pool for DAI/OSMO.
		// Link the DAI/OSMO balancer pool to the cl pool.
		clPool, err := createCanonicalConcentratedLiquidityPoolAndMigrationLink(ctx, DaiOsmoPoolId, DesiredDenom0, keepers)
		if err != nil {
			return nil, err
		}
		clPoolId := clPool.GetId()
		clPoolDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPoolId)

		// Create a position to initialize the balancerPool.

		// Get community pool and DAI/OSMO pool address.
		communityPoolAddress := keepers.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
		daiOsmoPool, err := keepers.PoolManagerKeeper.GetPool(ctx, DaiOsmoPoolId)
		if err != nil {
			return nil, err
		}

		// Swap one DAI for OSMO from the community pool.
		oneDai := sdk.NewCoin(DAIIBCDenom, sdk.NewInt(1000000))
		tokenInAmt, err := keepers.GAMMKeeper.SwapExactAmountOut(ctx, communityPoolAddress, daiOsmoPool, DesiredDenom0, sdk.NewInt(10000000), oneDai, sdk.ZeroDec())
		if err != nil {
			return nil, err
		}

		// Create a full range position via the community pool with the funds that were swapped.
		fullRangeOsmoDaiCoins := sdk.NewCoins(sdk.NewCoin(DesiredDenom0, tokenInAmt), oneDai)
		_, _, _, _, _, err = keepers.ConcentratedLiquidityKeeper.CreateFullRangePosition(ctx, clPoolId, communityPoolAddress, fullRangeOsmoDaiCoins)
		if err != nil {
			return nil, err
		}

		// Add the cl pool's full range denom as an authorized superfluid asset.
		superfluidAsset := superfluidtypes.SuperfluidAsset{
			Denom:     clPoolDenom,
			AssetType: superfluidtypes.SuperfluidAssetTypeConcentratedShare,
		}
		err = keepers.SuperfluidKeeper.AddNewSuperfluidAsset(ctx, superfluidAsset)
		if err != nil {
			return nil, err
		}

		clPoolTwapRecords, err := keepers.TwapKeeper.GetAllMostRecentRecordsForPool(ctx, clPoolId)
		if err != nil {
			return nil, err
		}

		for _, twapRecord := range clPoolTwapRecords {
			twapRecord.LastErrorTime = time.Time{}
			keepers.TwapKeeper.StoreNewRecord(ctx, twapRecord)
		}

		updateTokenFactoryParams(ctx, keepers.TokenFactoryKeeper)

		// Transfers out all the dev fees in kvstore to dev account during upgrade
		if err := keepers.ProtoRevKeeper.SendDeveloperFeesToDeveloperAccount(ctx); err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

func updateTokenFactoryParams(ctx sdk.Context, tokenFactoryKeeper *tokenfactorykeeper.Keeper) {
	tokenFactoryKeeper.SetParams(ctx, tokenfactorytypes.NewParams(nil, NewDenomCreationGasConsume))
}
