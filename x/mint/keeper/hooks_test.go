package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmoapp "github.com/osmosis-labs/osmosis/v7/app"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestEndOfEpochMintedCoinDistribution(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	setupGaugeForLPIncentives(t, app, ctx)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	// set developer rewards address
	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.WeightedDeveloperRewardsReceivers = []types.WeightedAddress{
		{
			Address: sdk.AccAddress([]byte("addr1---------------")).String(),
			Weight:  sdk.NewDec(1),
		},
	}
	app.MintKeeper.SetParams(ctx, mintParams)

	// setup developer rewards account
	app.MintKeeper.CreateDeveloperVestingModuleAccount(
		ctx, sdk.NewCoin("stake", sdk.NewInt(156*500000*2)))

	height := int64(1)
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	// correct rewards
	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)

		// get pre-epoch osmo supply and supplyWithOffset
		presupply := app.BankKeeper.GetSupply(ctx, mintParams.MintDenom)
		presupplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, mintParams.MintDenom)

		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams = app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// ensure post-epoch supply with offset changed by exactly the minted coins amount
		// ensure post-epoch supply with offset changed by less than the minted coins amount (because of developer vesting account)
		postsupply := app.BankKeeper.GetSupply(ctx, mintParams.MintDenom)
		postsupplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, mintParams.MintDenom)
		require.False(t, postsupply.IsEqual(presupply.Add(mintedCoin)))
		require.True(t, postsupplyWithOffset.IsEqual(presupplyWithOffset.Add(mintedCoin)))

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, height)

		// test that the dev rewards module account balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}

	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

	lastHalvenPeriod = app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs)

	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)

		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams = app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, height)

		// test that the balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}
}

func TestMintedCoinDistributionWhenDevRewardsAddressEmpty(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	setupGaugeForLPIncentives(t, app, ctx)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	// setup developer rewards account
	app.MintKeeper.CreateDeveloperVestingModuleAccount(
		ctx, sdk.NewCoin("stake", sdk.NewInt(156*500000*2)))

	height := int64(1)
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	// correct rewards
	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)
		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams := app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking.Add(mintParams.DistributionProportions.DeveloperRewards)).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, height)

		// test that the dev rewards module account balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}

	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

	lastHalvenPeriod = app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs)

	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)

		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams := app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking.Add(mintParams.DistributionProportions.DeveloperRewards)).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, expectedRewards.String())

		// test that the dev rewards module account balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}
}

func TestEndOfEpochNoDistributionWhenIsNotYetStartTime(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.MintingRewardsDistributionStartEpoch = 4
	app.MintKeeper.SetParams(ctx, mintParams)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	setupGaugeForLPIncentives(t, app, ctx)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	height := int64(1)
	// Run through epochs 0 through mintParams.MintingRewardsDistributionStartEpoch - 1
	// ensure no rewards sent out
	for ; height < mintParams.MintingRewardsDistributionStartEpoch; height++ {
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)
		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		// check community pool balance not increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool, feePoolNew.CommunityPool, "height = %v", height)
	}
	// Run through epochs mintParams.MintingRewardsDistributionStartEpoch
	// ensure tokens distributed
	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)
	require.NotEqual(t, sdk.DecCoins{}, app.DistrKeeper.GetFeePool(ctx).CommunityPool,
		"Tokens to community pool at start distribution epoch")

	// halven period should be set to mintParams.MintingRewardsDistributionStartEpoch
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, mintParams.MintingRewardsDistributionStartEpoch)
}

func TestAfterEpochEnd_FirstYearThirdening_RealParameters(t *testing.T) {
	// Most values in this test are taken from mainnet genesis to mimic real-world behavior:
	// https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
	const (
		reductionPeriodInEpochs                    = 365
		mintingRewardsDistributionStartEpoch       = 1
		thirdeningEpoch                      int64 = 365 + 1

		// different from mainnet since the difference is insignificant for testnig purposes.
		mintDenom              = "stake"
		genesisEpochProvisions = "821917808219.178082191780821917"
		epochIdentifier        = "day"

		// Actual value taken from mainnet for sanity checking calculations.
		mainnetThirdenedProvisions = "547945205479.452055068493150684"
	)

	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	genesisEpochProvisionsDec, err := sdk.NewDecFromStr(genesisEpochProvisions)
	require.NoError(t, err)

	mintParams := types.Params{
		MintDenom:               mintDenom,
		GenesisEpochProvisions:  genesisEpochProvisionsDec,
		EpochIdentifier:         epochIdentifier,
		ReductionPeriodInEpochs: reductionPeriodInEpochs,
		ReductionFactor:         sdk.NewDec(2).Quo(sdk.NewDec(3)),
		DistributionProportions: types.DistributionProportions{
			Staking:          sdk.NewDecWithPrec(25, 2),
			PoolIncentives:   sdk.NewDecWithPrec(45, 2),
			DeveloperRewards: sdk.NewDecWithPrec(25, 2),
			CommunityPool:    sdk.NewDecWithPrec(05, 2),
		},
		WeightedDeveloperRewardsReceivers: []types.WeightedAddress{
			{
				Address: "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
				Weight:  sdk.NewDecWithPrec(2887, 4),
			},
			{
				Address: "osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
				Weight:  sdk.NewDecWithPrec(229, 3),
			},
			{
				Address: "osmo13lt0hzc6u3htsk7z5rs6vuurmgg4hh2ecgxqkf",
				Weight:  sdk.NewDecWithPrec(1625, 4),
			},
			{
				Address: "osmo1kvc3he93ygc0us3ycslwlv2gdqry4ta73vk9hu",
				Weight:  sdk.NewDecWithPrec(109, 3),
			},
			{
				Address: "osmo19qgldlsk7hdv3ddtwwpvzff30pxqe9phq9evxf",
				Weight:  sdk.NewDecWithPrec(995, 3).Quo(sdk.NewDec(10)), // 0.0995
			},
			{
				Address: "osmo19fs55cx4594een7qr8tglrjtt5h9jrxg458htd",
				Weight:  sdk.NewDecWithPrec(6, 1).Quo(sdk.NewDec(10)), // 0.06
			},
			{
				Address: "osmo1ssp6px3fs3kwreles3ft6c07mfvj89a544yj9k",
				Weight:  sdk.NewDecWithPrec(15, 2).Quo(sdk.NewDec(10)), // 0.015
			},
			{
				Address: "osmo1c5yu8498yzqte9cmfv5zcgtl07lhpjrj0skqdx",
				Weight:  sdk.NewDecWithPrec(1, 1).Quo(sdk.NewDec(10)), // 0.01
			},
			{
				Address: "osmo1yhj3r9t9vw7qgeg22cehfzj7enwgklw5k5v7lj",
				Weight:  sdk.NewDecWithPrec(75, 2).Quo(sdk.NewDec(100)), // 0.0075
			},
			{
				Address: "osmo18nzmtyn5vy5y45dmcdnta8askldyvehx66lqgm",
				Weight:  sdk.NewDecWithPrec(7, 1).Quo(sdk.NewDec(100)), // 0.007
			},
			{
				Address: "osmo1z2x9z58cg96ujvhvu6ga07yv9edq2mvkxpgwmc",
				Weight:  sdk.NewDecWithPrec(5, 1).Quo(sdk.NewDec(100)), // 0.005
			},
			{
				Address: "osmo1tvf3373skua8e6480eyy38avv8mw3hnt8jcxg9",
				Weight:  sdk.NewDecWithPrec(25, 2).Quo(sdk.NewDec(100)), // 0.0025
			},
			{
				Address: "osmo1zs0txy03pv5crj2rvty8wemd3zhrka2ne8u05n",
				Weight:  sdk.NewDecWithPrec(25, 2).Quo(sdk.NewDec(100)), // 0.0025
			},
			{
				Address: "osmo1djgf9p53n7m5a55hcn6gg0cm5mue4r5g3fadee",
				Weight:  sdk.NewDecWithPrec(1, 1).Quo(sdk.NewDec(100)), // 0.001
			},
			{
				Address: "osmo1488zldkrn8xcjh3z40v2mexq7d088qkna8ceze",
				Weight:  sdk.NewDecWithPrec(8, 1).Quo(sdk.NewDec(1000)), // 0.0008
			},
		},
		MintingRewardsDistributionStartEpoch: mintingRewardsDistributionStartEpoch,
	}

	sumOfWeights := sdk.ZeroDec()
	// As a sanity check, ensure developer reward receivers add up to 1.
	for _, w := range mintParams.WeightedDeveloperRewardsReceivers {
		sumOfWeights = sumOfWeights.Add(w.Weight)
	}
	require.Equal(t, sdk.OneDec(), sumOfWeights)

	// Test setup parameters are not identical with mainnet.
	// Therfore, we set them here to the desired mainnet values.
	app.MintKeeper.SetParams(ctx, mintParams)
	app.MintKeeper.SetLastHalvenEpochNum(ctx, 0)
	app.MintKeeper.SetMinter(ctx, types.Minter{
		EpochProvisions: genesisEpochProvisionsDec,
	})

	expectedSupplyWithOffset := sdk.NewDec(0)

	supplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, mintDenom)
	require.Equal(t, expectedSupplyWithOffset.TruncateInt64(), supplyWithOffset.Amount.Int64())

	// Actual test for running AfterEpochEnd hook thirdeningEpoch times.
	for i := int64(0); i < thirdeningEpoch; i++ {
		app.MintKeeper.AfterEpochEnd(ctx, epochIdentifier, i)
		require.Equal(t, genesisEpochProvisions, app.MintKeeper.GetMinter(ctx).EpochProvisions.String())

		if i >= mintingRewardsDistributionStartEpoch {
			expectedSupplyWithOffset = expectedSupplyWithOffset.Add(genesisEpochProvisionsDec)
			require.Equal(t, expectedSupplyWithOffset.String(), app.BankKeeper.GetSupplyWithOffset(ctx, mintDenom).Amount.String())
		}
	}

	// This end of epoch should trigger thirderning.
	app.MintKeeper.AfterEpochEnd(ctx, epochIdentifier, thirdeningEpoch)

	require.Equal(t, thirdeningEpoch, app.MintKeeper.GetLastHalvenEpochNum(ctx))

	expectedThirdenedProvisions := mintParams.ReductionFactor.Mul(genesisEpochProvisionsDec)
	// Sanity check with the actual value on mainnet.
	require.Equal(t, mainnetThirdenedProvisions, expectedThirdenedProvisions.String())
	require.Equal(t, expectedThirdenedProvisions, app.MintKeeper.GetMinter(ctx).EpochProvisions)
}

func setupGaugeForLPIncentives(t *testing.T, app *osmoapp.OsmosisApp, ctx sdk.Context) {
	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, coins)
	require.NoError(t, err)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}

	// mints coins so supply exists on chain
	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	err = simapp.FundAccount(app.BankKeeper, ctx, addr, mintLPtokens)
	require.NoError(t, err)

	_, err = app.IncentivesKeeper.CreateGauge(ctx, true, addr, coins, distrTo, time.Now(), 1)
	require.NoError(t, err)
}
