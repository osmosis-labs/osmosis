package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmoapp "github.com/osmosis-labs/osmosis/v11/app"
	"github.com/osmosis-labs/osmosis/v11/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v11/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Most values here are taken from mainnet genesis to mimic real-world behavior:
	// https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
	defaultGenesisEpochProvisions = "821917808219.178082191780821917"
	defaultEpochIdentifier        = "day"
	// actual value taken from mainnet for sanity checking calculations.
	defaultMainnetThirdenedProvisions                 = "547945205479.452055068493150684"
	defaultReductionPeriodInEpochs                    = 365
	defaultMintingRewardsDistributionStartEpoch int64 = 1
	defaultThirdeningEpochNum                   int64 = defaultReductionPeriodInEpochs + defaultMintingRewardsDistributionStartEpoch
)

var (
	defaultReductionFactor         = sdk.NewDec(2).Quo(sdk.NewDec(3))
	defaultDistributionProportions = types.DistributionProportions{
		Staking:          sdk.NewDecWithPrec(25, 2),
		PoolIncentives:   sdk.NewDecWithPrec(45, 2),
		DeveloperRewards: sdk.NewDecWithPrec(25, 2),
		CommunityPool:    sdk.NewDecWithPrec(0o5, 2),
	}
)

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TestAfterEpochEnd tests that the after epoch end hook correctly
// distributes the rewards depending on what epoch it is in.
func (suite *KeeperTestSuite) TestAfterEpochEnd() {
	var (
		testWeightedAddresses = []types.WeightedAddress{
			{
				Address: testAddressOne.String(),
				Weight:  sdk.NewDecWithPrec(233, 3),
			},
			{
				Address: testAddressTwo.String(),
				Weight:  sdk.NewDecWithPrec(5, 1),
			},
			{
				Address: testAddressThree.String(),
				Weight:  sdk.NewDecWithPrec(50, 3),
			},
			{
				Address: testAddressFour.String(),
				Weight:  sdk.NewDecWithPrec(217, 3),
			},
		}
		maxArithmeticTolerance   = sdk.NewDec(5)
		expectedSupplyWithOffset = sdk.NewDec(0)
		expectedSupply           = sdk.NewDec(keeper.DeveloperVestingAmount)
	)

	suite.assertAddressWeightsAddUpToOne(testWeightedAddresses)

	defaultGenesisEpochProvisionsDec, err := sdk.NewDecFromStr(defaultGenesisEpochProvisions)
	suite.Require().NoError(err)

	defaultMainnetThirdenedProvisionsDec, err := sdk.NewDecFromStr(defaultMainnetThirdenedProvisions)
	suite.Require().NoError(err)

	testcases := map[string]struct {
		// Args.
		hookArgEpochNum int64

		// Presets.
		preExistingEpochNum     int64
		mintDenom               string
		epochIdentifier         string
		genesisEpochProvisions  sdk.Dec
		reductionPeriodInEpochs int64
		reductionFactor         sdk.Dec
		distributionProportions types.DistributionProportions
		weightedAddresses       []types.WeightedAddress
		mintStartEpoch          int64

		// Expected results.
		expectedLastReductionEpochNum int64
		expectedDistribution          sdk.Dec
		expectedPanic                 bool
	}{
		"before start epoch - no distributions": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch - 1,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution: sdk.ZeroDec(),
		},
		"at start epoch - distributes": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"after start epoch - distributes": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + 5,

			preExistingEpochNum:     defaultMintingRewardsDistributionStartEpoch,
			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"before reduction epoch - distributes, no reduction": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			preExistingEpochNum:     defaultMintingRewardsDistributionStartEpoch,
			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"at reduction epoch - distributes, reduction occurs": {
			preExistingEpochNum: defaultMintingRewardsDistributionStartEpoch,

			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultMainnetThirdenedProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,
		},
		"after reduction epoch - distributes, with reduced amounts": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultMainnetThirdenedProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,
		},
		"start epoch == reduction epoch = curEpoch": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultReductionPeriodInEpochs,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultReductionPeriodInEpochs,
		},
		"start epoch == curEpoch + 1 && reduction epoch == curEpoch": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultReductionPeriodInEpochs - 1,

			expectedDistribution:          defaultMainnetThirdenedProvisionsDec,
			expectedLastReductionEpochNum: defaultReductionPeriodInEpochs,
		},
		"start epoch > reduction epoch": {
			hookArgEpochNum: defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultReductionPeriodInEpochs + 1,

			expectedDistribution: sdk.ZeroDec(),
		},
		// N.B.: This test case would not work since it would require changing default genesis denom.
		// Leaving it to potentially revisit in the future.
		// "custom mint denom, at start epoch": {},
		"custom epochIdentifier, at start epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         "week",
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution: sdk.ZeroDec(),
		},
		"custom genesisEpochProvisions, at start epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  sdk.NewDec(1_000_000_000),
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"custom reduction factor, reduction epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         sdk.NewDec(43).Quo(sdk.NewDec(55)),
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec.Mul(sdk.NewDec(43)).Quo(sdk.NewDec(55)),
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch + defaultReductionPeriodInEpochs,
		},
		"custom distribution proportions, at start epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: types.DistributionProportions{
				Staking:          sdk.NewDecWithPrec(11, 2),
				PoolIncentives:   sdk.NewDecWithPrec(22, 2),
				DeveloperRewards: sdk.NewDecWithPrec(33, 2),
				CommunityPool:    sdk.NewDecWithPrec(34, 2),
			},
			weightedAddresses: testWeightedAddresses,
			mintStartEpoch:    defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"custom weighted addresses, at start epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch + 5,

			preExistingEpochNum:     defaultMintingRewardsDistributionStartEpoch,
			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDecWithPrec(11, 2),
				},
				{
					Address: testAddressTwo.String(),
					Weight:  sdk.NewDecWithPrec(22, 2),
				},
				{
					Address: testAddressThree.String(),
					Weight:  sdk.NewDecWithPrec(33, 2),
				},
				{
					Address: testAddressFour.String(),
					Weight:  sdk.NewDecWithPrec(34, 2),
				},
			},
			mintStartEpoch: defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
		},
		"failed to hook due to developer vesting module account not having enough balance - panic": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          sdk.ZeroDec(),
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
			expectedPanic:                 true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			mintParams := types.Params{
				MintDenom:                            tc.mintDenom,
				GenesisEpochProvisions:               tc.genesisEpochProvisions,
				EpochIdentifier:                      tc.epochIdentifier,
				ReductionPeriodInEpochs:              tc.reductionPeriodInEpochs,
				ReductionFactor:                      tc.reductionFactor,
				DistributionProportions:              tc.distributionProportions,
				WeightedDeveloperRewardsReceivers:    tc.weightedAddresses,
				MintingRewardsDistributionStartEpoch: tc.mintStartEpoch,
			}

			app := osmoapp.Setup(false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{})

			mintKeeper := app.MintKeeper
			distrKeeper := app.DistrKeeper
			accountKeeper := app.AccountKeeper
			bankKeeper := app.BankKeeper

			// Pre-set parameters and minter.
			mintKeeper.SetParams(ctx, mintParams)
			mintKeeper.SetLastReductionEpochNum(ctx, tc.preExistingEpochNum)
			mintKeeper.SetMinter(ctx, types.Minter{
				EpochProvisions: defaultGenesisEpochProvisionsDec,
			})

			expectedDevRewards := tc.expectedDistribution.Mul(mintParams.DistributionProportions.DeveloperRewards)

			developerAccountBalanceBeforeHook := app.BankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)

			if tc.expectedPanic {
				// If panic is expected, burn developer module account balance so that it causes an error that leads to a
				// panic in the hook.
				suite.Require().NoError(distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(developerAccountBalanceBeforeHook), accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)))
				developerAccountBalanceBeforeHook.Amount = sdk.ZeroInt()
			}

			// Old supply
			oldSupply := app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount
			suite.Require().Equal(sdk.NewInt(keeper.DeveloperVestingAmount), oldSupply)

			osmoassert.ConditionalPanic(suite.T(), tc.expectedPanic, func() {
				// System under test.
				mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, tc.hookArgEpochNum)
			})

			// If panics, the behavior is undefined.
			if tc.expectedPanic {
				return
			}

			// Validate developer account balance.
			developerAccountBalanceAfterHook := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)
			osmoassert.DecApproxEq(suite.T(), developerAccountBalanceBeforeHook.Amount.Sub(expectedDevRewards.TruncateInt()).ToDec(), developerAccountBalanceAfterHook.Amount.ToDec(), maxArithmeticTolerance)

			// Validate supply.
			osmoassert.DecApproxEq(suite.T(), expectedSupply.Add(tc.expectedDistribution).Sub(expectedDevRewards), app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount.ToDec(), maxArithmeticTolerance)

			// Validate supply with offset.
			osmoassert.DecApproxEq(suite.T(), expectedSupplyWithOffset.Add(tc.expectedDistribution), app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.ToDec(), maxArithmeticTolerance)

			// Validate epoch provisions.
			suite.Require().Equal(tc.expectedLastReductionEpochNum, mintKeeper.GetLastReductionEpochNum(ctx))

			if !tc.expectedDistribution.IsZero() {
				// Validate distribution.
				osmoassert.DecApproxEq(suite.T(), tc.expectedDistribution, mintKeeper.GetMinter(ctx).EpochProvisions, sdk.NewDecWithPrec(1, 6))
			}
		})
	}
}

// TODO: Remove after rounding errors are addressed and resolved.
// Make sure that more specific test specs are added to validate the expected
// supply for correctness.
//
// Ref: https://github.com/osmosis-labs/osmosis/issues/1917
func (suite *KeeperTestSuite) TestAfterEpochEnd_FirstYearThirdening_RealParameters() {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	mintKeeper := app.MintKeeper
	accountKeeper := app.AccountKeeper

	genesisEpochProvisionsDec, err := sdk.NewDecFromStr(defaultGenesisEpochProvisions)
	suite.Require().NoError(err)

	mintParams := types.Params{
		MintDenom:               sdk.DefaultBondDenom,
		GenesisEpochProvisions:  genesisEpochProvisionsDec,
		EpochIdentifier:         defaultEpochIdentifier,
		ReductionPeriodInEpochs: defaultReductionPeriodInEpochs,
		ReductionFactor:         defaultReductionFactor,
		DistributionProportions: types.DistributionProportions{
			Staking:          sdk.NewDecWithPrec(25, 2),
			PoolIncentives:   sdk.NewDecWithPrec(45, 2),
			DeveloperRewards: sdk.NewDecWithPrec(25, 2),
			CommunityPool:    sdk.NewDecWithPrec(0o5, 2),
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
		MintingRewardsDistributionStartEpoch: defaultMintingRewardsDistributionStartEpoch,
	}

	suite.assertAddressWeightsAddUpToOne(mintParams.WeightedDeveloperRewardsReceivers)

	// Test setup parameters are not identical with mainnet.
	// Therfore, we set them here to the desired mainnet values.
	mintKeeper.SetParams(ctx, mintParams)
	mintKeeper.SetLastReductionEpochNum(ctx, 0)
	mintKeeper.SetMinter(ctx, types.Minter{
		EpochProvisions: genesisEpochProvisionsDec,
	})

	expectedSupplyWithOffset := sdk.NewDec(0)
	expectedSupply := sdk.NewDec(keeper.DeveloperVestingAmount)

	supplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom)
	suite.Require().Equal(expectedSupplyWithOffset.TruncateInt64(), supplyWithOffset.Amount.Int64())

	supply := app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom)
	suite.Require().Equal(expectedSupply.TruncateInt64(), supply.Amount.Int64())

	devRewardsDelta := sdk.ZeroDec()
	epochProvisionsDelta := genesisEpochProvisionsDec.Sub(genesisEpochProvisionsDec.TruncateInt().ToDec()).Mul(sdk.NewDec(defaultReductionPeriodInEpochs))

	// Actual test for running AfterEpochEnd hook thirdeningEpoch times.
	for i := int64(1); i <= defaultReductionPeriodInEpochs; i++ {
		developerAccountBalanceBeforeHook := app.BankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)

		// System undert test.
		mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, i)

		// System truncates EpochProvisions because bank takes an Int.
		// This causes rounding errors. Let's refer to this source as #1.
		//
		// Since this is truncated, our total supply calculation at the end will
		// be off by reductionPeriodInEpochs * (genesisEpochProvisionsDec - truncatedEpochProvisions)
		// Therefore, we store this delta in epochProvisionsDelta to add to the actual supply to compare
		// to expected at the end.
		truncatedEpochProvisions := genesisEpochProvisionsDec.TruncateInt().ToDec()

		// We want supply with offset to exclude unvested developer rewards
		// Truncation also happens when subtracting dev rewards.
		// Potential source of minor rounding errors #2.
		devRewards := truncatedEpochProvisions.Mul(mintParams.DistributionProportions.DeveloperRewards).TruncateInt().ToDec()

		// We aim to exclude developer account balance from the supply with offset calculation.
		developerAccountBalance := app.BankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)

		// Make sure developer account balance has decreased by devRewards.
		// This check is now failing because of rounding errors.
		// To prove that this is the source of errors, we keep accumulating
		// the delta and add it to the expected supply validation after the loop.
		if !developerAccountBalanceBeforeHook.Amount.ToDec().Sub(devRewards).Equal(developerAccountBalance.Amount.ToDec()) {
			expectedDeveloperAccountBalanceAfterHook := developerAccountBalanceBeforeHook.Amount.ToDec().Sub(devRewards)
			actualDeveloperAccountBalanceAfterHook := developerAccountBalance.Amount.ToDec()

			// This is supposed to be equal but is failing due to the rounding errors from devRewards.
			suite.Require().NotEqual(expectedDeveloperAccountBalanceAfterHook, actualDeveloperAccountBalanceAfterHook)

			devRewardsDelta = devRewardsDelta.Add(actualDeveloperAccountBalanceAfterHook.Sub(expectedDeveloperAccountBalanceAfterHook))
		}

		expectedSupply = expectedSupply.Add(truncatedEpochProvisions).Sub(devRewards)
		suite.Require().Equal(expectedSupply.RoundInt(), app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount)

		expectedSupplyWithOffset = expectedSupply.Sub(developerAccountBalance.Amount.ToDec())
		suite.Require().Equal(expectedSupplyWithOffset.RoundInt(), app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)

		// Validate that the epoch provisions have not been reduced.
		suite.Require().Equal(defaultMintingRewardsDistributionStartEpoch, mintKeeper.GetLastReductionEpochNum(ctx))
		suite.Require().Equal(defaultGenesisEpochProvisions, mintKeeper.GetMinter(ctx).EpochProvisions.String())
	}

	// Validate total supply.
	// This test check is now failing due to rounding errors.
	// Every epoch, we accumulate the rounding delta from every problematic component
	// Here, we add the deltas to the actual supply and compare against expected.
	//
	// expectedTotalProvisionedSupply = 365 * 821917808219.178082191780821917 = 299_999_999_999_999.999999999999999705
	expectedTotalProvisionedSupply := sdk.NewDec(defaultReductionPeriodInEpochs).Mul(genesisEpochProvisionsDec)
	// actualTotalProvisionedSupply = 299_999_999_997_380 (off by 2619.999999999999999705)
	// devRewardsDelta = 2555 (hard to estimate but the source is from truncating dev rewards )
	// epochProvisionsDelta = 0.178082191780821917 * 365 = 64.999999999999999705
	actualTotalProvisionedSupply := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.ToDec()

	// 299_999_999_999_999.999999999999999705 == 299_999_999_997_380 + 2555 + 64.999999999999999705
	suite.Require().Equal(expectedTotalProvisionedSupply, actualTotalProvisionedSupply.Add(devRewardsDelta).Add(epochProvisionsDelta))

	// This end of epoch should trigger thirdening. It will utilize the updated
	// (reduced) provisions.
	mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, defaultThirdeningEpochNum)

	suite.Require().Equal(defaultThirdeningEpochNum, mintKeeper.GetLastReductionEpochNum(ctx))

	expectedThirdenedProvisions := mintParams.ReductionFactor.Mul(genesisEpochProvisionsDec)
	// Sanity check with the actual value on mainnet.
	suite.Require().Equal(defaultMainnetThirdenedProvisions, expectedThirdenedProvisions.String())
	suite.Require().Equal(expectedThirdenedProvisions, mintKeeper.GetMinter(ctx).EpochProvisions)
}

func (suite KeeperTestSuite) assertAddressWeightsAddUpToOne(receivers []types.WeightedAddress) {
	sumOfWeights := sdk.ZeroDec()
	// As a sanity check, ensure developer reward receivers add up to 1.
	for _, w := range receivers {
		sumOfWeights = sumOfWeights.Add(w.Weight)
	}
	suite.Require().Equal(sdk.OneDec(), sumOfWeights)
}
