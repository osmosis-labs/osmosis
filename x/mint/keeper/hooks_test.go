package keeper_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	osmoapp "github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/mint/types"

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
	defaultReductionFactor         = osmomath.NewDec(2).Quo(osmomath.NewDec(3))
	defaultDistributionProportions = types.DistributionProportions{
		Staking:          osmomath.NewDecWithPrec(25, 2),
		PoolIncentives:   osmomath.NewDecWithPrec(45, 2),
		DeveloperRewards: osmomath.NewDecWithPrec(25, 2),
		CommunityPool:    osmomath.NewDecWithPrec(0o5, 2),
	}
)

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TestAfterEpochEnd tests that the after epoch end hook correctly
// distributes the rewards depending on what epoch it is in.
func (s *KeeperTestSuite) TestAfterEpochEnd() {
	var (
		testWeightedAddresses = []types.WeightedAddress{
			{
				Address: testAddressOne.String(),
				Weight:  osmomath.NewDecWithPrec(233, 3),
			},
			{
				Address: testAddressTwo.String(),
				Weight:  osmomath.NewDecWithPrec(5, 1),
			},
			{
				Address: testAddressThree.String(),
				Weight:  osmomath.NewDecWithPrec(50, 3),
			},
			{
				Address: testAddressFour.String(),
				Weight:  osmomath.NewDecWithPrec(217, 3),
			},
		}
		maxArithmeticTolerance = osmomath.NewDec(5)
		// In test setup, we set a validator with a delegation equal to sdk.DefaultPowerReduction.
		expectedSupplyWithOffset = sdk.DefaultPowerReduction.ToLegacyDec()
		expectedSupply           = osmomath.NewDec(keeper.DeveloperVestingAmount).Add(sdk.DefaultPowerReduction.ToLegacyDec())
	)

	s.assertAddressWeightsAddUpToOne(testWeightedAddresses)

	defaultGenesisEpochProvisionsDec, err := osmomath.NewDecFromStr(defaultGenesisEpochProvisions)
	s.Require().NoError(err)

	defaultMainnetThirdenedProvisionsDec, err := osmomath.NewDecFromStr(defaultMainnetThirdenedProvisions)
	s.Require().NoError(err)

	testcases := map[string]struct {
		// Args.
		hookArgEpochNum int64

		// Presets.
		preExistingEpochNum     int64
		mintDenom               string
		epochIdentifier         string
		genesisEpochProvisions  osmomath.Dec
		reductionPeriodInEpochs int64
		reductionFactor         osmomath.Dec
		distributionProportions types.DistributionProportions
		weightedAddresses       []types.WeightedAddress
		mintStartEpoch          int64

		// Expected results.
		expectedLastReductionEpochNum int64
		expectedDistribution          osmomath.Dec
		expectedError                 bool
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

			expectedDistribution: osmomath.ZeroDec(),
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

			expectedDistribution: osmomath.ZeroDec(),
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

			expectedDistribution: osmomath.ZeroDec(),
		},
		"custom genesisEpochProvisions, at start epoch": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  osmomath.NewDec(1_000_000_000),
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
			reductionFactor:         osmomath.NewDec(43).Quo(osmomath.NewDec(55)),
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,

			expectedDistribution:          defaultGenesisEpochProvisionsDec.Mul(osmomath.NewDec(43)).Quo(osmomath.NewDec(55)),
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
				Staking:          osmomath.NewDecWithPrec(11, 2),
				PoolIncentives:   osmomath.NewDecWithPrec(22, 2),
				DeveloperRewards: osmomath.NewDecWithPrec(33, 2),
				CommunityPool:    osmomath.NewDecWithPrec(34, 2),
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
					Weight:  osmomath.NewDecWithPrec(11, 2),
				},
				{
					Address: testAddressTwo.String(),
					Weight:  osmomath.NewDecWithPrec(22, 2),
				},
				{
					Address: testAddressThree.String(),
					Weight:  osmomath.NewDecWithPrec(33, 2),
				},
				{
					Address: testAddressFour.String(),
					Weight:  osmomath.NewDecWithPrec(34, 2),
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

			expectedDistribution:          osmomath.ZeroDec(),
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
			expectedError:                 true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
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

			dirName := fmt.Sprintf("%d", rand.Int())
			app := osmoapp.SetupWithCustomHome(false, dirName)
			defer os.RemoveAll(dirName)

			ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})

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

			if tc.expectedError {
				// If panic is expected, burn developer module account balance so that it causes an error that leads to a
				// panic in the hook.
				s.Require().NoError(distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(developerAccountBalanceBeforeHook), accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)))
				developerAccountBalanceBeforeHook.Amount = osmomath.ZeroInt()
			}

			// Old supply
			oldSupply := app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount
			// We require a validator be setup in the app setup logic, or else tests won't run.
			// This validator has a delegation equal to sdk.DefaultPowerReduction, so we add this
			// to the expected supply.
			s.Require().Equal(osmomath.NewInt(keeper.DeveloperVestingAmount).Add(sdk.DefaultPowerReduction), oldSupply)

			if tc.expectedError {
				s.Require().Error(mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, tc.hookArgEpochNum))
			} else {
				s.Require().NoError(mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, tc.hookArgEpochNum))
			}

			// If panics, the behavior is undefined.
			if tc.expectedError {
				return
			}

			// Validate developer account balance.
			developerAccountBalanceAfterHook := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)
			osmoassert.DecApproxEq(s.T(), developerAccountBalanceBeforeHook.Amount.Sub(expectedDevRewards.TruncateInt()).ToLegacyDec(), developerAccountBalanceAfterHook.Amount.ToLegacyDec(), maxArithmeticTolerance)

			// Validate supply.
			osmoassert.DecApproxEq(s.T(), expectedSupply.Add(tc.expectedDistribution).Sub(expectedDevRewards), app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount.ToLegacyDec(), maxArithmeticTolerance)

			// Validate supply with offset.
			osmoassert.DecApproxEq(s.T(), expectedSupplyWithOffset.Add(tc.expectedDistribution), app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.ToLegacyDec(), maxArithmeticTolerance)

			// Validate epoch provisions.
			s.Require().Equal(tc.expectedLastReductionEpochNum, mintKeeper.GetLastReductionEpochNum(ctx))

			if !tc.expectedDistribution.IsZero() {
				// Validate distribution.
				osmoassert.DecApproxEq(s.T(), tc.expectedDistribution, mintKeeper.GetMinter(ctx).EpochProvisions, osmomath.NewDecWithPrec(1, 6))
			}
		})
	}
}

// TODO: Remove after rounding errors are addressed and resolved.
// Make sure that more specific test specs are added to validate the expected
// supply for correctness.
//
// Ref: https://github.com/osmosis-labs/osmosis/issues/1917
func (s *KeeperTestSuite) TestAfterEpochEnd_FirstYearThirdening_RealParameters() {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	mintKeeper := app.MintKeeper
	accountKeeper := app.AccountKeeper

	genesisEpochProvisionsDec, err := osmomath.NewDecFromStr(defaultGenesisEpochProvisions)
	s.Require().NoError(err)

	mintParams := types.Params{
		MintDenom:               sdk.DefaultBondDenom,
		GenesisEpochProvisions:  genesisEpochProvisionsDec,
		EpochIdentifier:         defaultEpochIdentifier,
		ReductionPeriodInEpochs: defaultReductionPeriodInEpochs,
		ReductionFactor:         defaultReductionFactor,
		DistributionProportions: types.DistributionProportions{
			Staking:          osmomath.NewDecWithPrec(25, 2),
			PoolIncentives:   osmomath.NewDecWithPrec(45, 2),
			DeveloperRewards: osmomath.NewDecWithPrec(25, 2),
			CommunityPool:    osmomath.NewDecWithPrec(0o5, 2),
		},
		WeightedDeveloperRewardsReceivers: []types.WeightedAddress{
			{
				Address: "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
				Weight:  osmomath.NewDecWithPrec(2887, 4),
			},
			{
				Address: "osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
				Weight:  osmomath.NewDecWithPrec(229, 3),
			},
			{
				Address: "osmo13lt0hzc6u3htsk7z5rs6vuurmgg4hh2ecgxqkf",
				Weight:  osmomath.NewDecWithPrec(1625, 4),
			},
			{
				Address: "osmo1kvc3he93ygc0us3ycslwlv2gdqry4ta73vk9hu",
				Weight:  osmomath.NewDecWithPrec(109, 3),
			},
			{
				Address: "osmo19qgldlsk7hdv3ddtwwpvzff30pxqe9phq9evxf",
				Weight:  osmomath.NewDecWithPrec(995, 3).Quo(osmomath.NewDec(10)), // 0.0995
			},
			{
				Address: "osmo19fs55cx4594een7qr8tglrjtt5h9jrxg458htd",
				Weight:  osmomath.NewDecWithPrec(6, 1).Quo(osmomath.NewDec(10)), // 0.06
			},
			{
				Address: "osmo1ssp6px3fs3kwreles3ft6c07mfvj89a544yj9k",
				Weight:  osmomath.NewDecWithPrec(15, 2).Quo(osmomath.NewDec(10)), // 0.015
			},
			{
				Address: "osmo1c5yu8498yzqte9cmfv5zcgtl07lhpjrj0skqdx",
				Weight:  osmomath.NewDecWithPrec(1, 1).Quo(osmomath.NewDec(10)), // 0.01
			},
			{
				Address: "osmo1yhj3r9t9vw7qgeg22cehfzj7enwgklw5k5v7lj",
				Weight:  osmomath.NewDecWithPrec(75, 2).Quo(osmomath.NewDec(100)), // 0.0075
			},
			{
				Address: "osmo18nzmtyn5vy5y45dmcdnta8askldyvehx66lqgm",
				Weight:  osmomath.NewDecWithPrec(7, 1).Quo(osmomath.NewDec(100)), // 0.007
			},
			{
				Address: "osmo1z2x9z58cg96ujvhvu6ga07yv9edq2mvkxpgwmc",
				Weight:  osmomath.NewDecWithPrec(5, 1).Quo(osmomath.NewDec(100)), // 0.005
			},
			{
				Address: "osmo1tvf3373skua8e6480eyy38avv8mw3hnt8jcxg9",
				Weight:  osmomath.NewDecWithPrec(25, 2).Quo(osmomath.NewDec(100)), // 0.0025
			},
			{
				Address: "osmo1zs0txy03pv5crj2rvty8wemd3zhrka2ne8u05n",
				Weight:  osmomath.NewDecWithPrec(25, 2).Quo(osmomath.NewDec(100)), // 0.0025
			},
			{
				Address: "osmo1djgf9p53n7m5a55hcn6gg0cm5mue4r5g3fadee",
				Weight:  osmomath.NewDecWithPrec(1, 1).Quo(osmomath.NewDec(100)), // 0.001
			},
			{
				Address: "osmo1488zldkrn8xcjh3z40v2mexq7d088qkna8ceze",
				Weight:  osmomath.NewDecWithPrec(8, 1).Quo(osmomath.NewDec(1000)), // 0.0008
			},
		},
		MintingRewardsDistributionStartEpoch: defaultMintingRewardsDistributionStartEpoch,
	}

	s.assertAddressWeightsAddUpToOne(mintParams.WeightedDeveloperRewardsReceivers)

	// Test setup parameters are not identical with mainnet.
	// Therefore, we set them here to the desired mainnet values.
	mintKeeper.SetParams(ctx, mintParams)
	mintKeeper.SetLastReductionEpochNum(ctx, 0)
	mintKeeper.SetMinter(ctx, types.Minter{
		EpochProvisions: genesisEpochProvisionsDec,
	})

	// In test setup, we set a validator with a delegation equal to sdk.DefaultPowerReduction.
	expectedSupplyWithOffset := sdk.DefaultPowerReduction.ToLegacyDec()
	expectedSupply := osmomath.NewDec(keeper.DeveloperVestingAmount).Add(sdk.DefaultPowerReduction.ToLegacyDec())

	supplyWithOffset := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom)
	s.Require().Equal(expectedSupplyWithOffset.TruncateInt64(), supplyWithOffset.Amount.Int64())

	supply := app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom)
	s.Require().Equal(expectedSupply.TruncateInt64(), supply.Amount.Int64())

	devRewardsDelta := osmomath.ZeroDec()
	epochProvisionsDelta := genesisEpochProvisionsDec.Sub(genesisEpochProvisionsDec.TruncateInt().ToLegacyDec()).Mul(osmomath.NewDec(defaultReductionPeriodInEpochs))

	// Actual test for running AfterEpochEnd hook thirdeningEpoch times.
	for i := int64(1); i <= defaultReductionPeriodInEpochs; i++ {
		developerAccountBalanceBeforeHook := app.BankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)

		// System under test.
		err = mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, i)
		s.Require().NoError(err)

		// System truncates EpochProvisions because bank takes an Int.
		// This causes rounding errors. Let's refer to this source as #1.
		//
		// Since this is truncated, our total supply calculation at the end will
		// be off by reductionPeriodInEpochs * (genesisEpochProvisionsDec - truncatedEpochProvisions)
		// Therefore, we store this delta in epochProvisionsDelta to add to the actual supply to compare
		// to expected at the end.
		truncatedEpochProvisions := genesisEpochProvisionsDec.TruncateInt().ToLegacyDec()

		// We want supply with offset to exclude unvested developer rewards
		// Truncation also happens when subtracting dev rewards.
		// Potential source of minor rounding errors #2.
		devRewards := truncatedEpochProvisions.Mul(mintParams.DistributionProportions.DeveloperRewards).TruncateInt().ToLegacyDec()

		// We aim to exclude developer account balance from the supply with offset calculation.
		developerAccountBalance := app.BankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)

		// Make sure developer account balance has decreased by devRewards.
		// This check is now failing because of rounding errors.
		// To prove that this is the source of errors, we keep accumulating
		// the delta and add it to the expected supply validation after the loop.
		if !developerAccountBalanceBeforeHook.Amount.ToLegacyDec().Sub(devRewards).Equal(developerAccountBalance.Amount.ToLegacyDec()) {
			expectedDeveloperAccountBalanceAfterHook := developerAccountBalanceBeforeHook.Amount.ToLegacyDec().Sub(devRewards)
			actualDeveloperAccountBalanceAfterHook := developerAccountBalance.Amount.ToLegacyDec()

			// This is supposed to be equal but is failing due to the rounding errors from devRewards.
			s.Require().NotEqual(expectedDeveloperAccountBalanceAfterHook, actualDeveloperAccountBalanceAfterHook)

			devRewardsDelta = devRewardsDelta.Add(actualDeveloperAccountBalanceAfterHook.Sub(expectedDeveloperAccountBalanceAfterHook))
		}

		expectedSupply = expectedSupply.Add(truncatedEpochProvisions).Sub(devRewards)
		s.Require().Equal(expectedSupply.RoundInt(), app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount)

		expectedSupplyWithOffset = expectedSupply.Sub(developerAccountBalance.Amount.ToLegacyDec())
		s.Require().Equal(expectedSupplyWithOffset.RoundInt(), app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)

		// Validate that the epoch provisions have not been reduced.
		s.Require().Equal(defaultMintingRewardsDistributionStartEpoch, mintKeeper.GetLastReductionEpochNum(ctx))
		s.Require().Equal(defaultGenesisEpochProvisions, mintKeeper.GetMinter(ctx).EpochProvisions.String())
	}

	// Validate total supply.
	// This test check is now failing due to rounding errors.
	// Every epoch, we accumulate the rounding delta from every problematic component
	// Here, we add the deltas to the actual supply and compare against expected.
	//
	// expectedTotalProvisionedSupply = 365 * 821917808219.178082191780821917 = 299_999_999_999_999.999999999999999705
	// In test setup, we set a validator with a delegation equal to sdk.DefaultPowerReduction.
	expectedTotalProvisionedSupply := osmomath.NewDec(defaultReductionPeriodInEpochs).Mul(genesisEpochProvisionsDec).Add(sdk.DefaultPowerReduction.ToLegacyDec())
	// actualTotalProvisionedSupply = 299_999_999_997_380 (off by 2619.999999999999999705)
	// devRewardsDelta = 2555 (hard to estimate but the source is from truncating dev rewards )
	// epochProvisionsDelta = 0.178082191780821917 * 365 = 64.999999999999999705
	actualTotalProvisionedSupply := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.ToLegacyDec()

	// 299_999_999_999_999.999999999999999705 == 299_999_999_997_380 + 2555 + 64.999999999999999705
	s.Require().Equal(expectedTotalProvisionedSupply, actualTotalProvisionedSupply.Add(devRewardsDelta).Add(epochProvisionsDelta))

	// This end of epoch should trigger thirdening. It will utilize the updated
	// (reduced) provisions.
	err = mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, defaultThirdeningEpochNum)
	s.Require().NoError(err)

	s.Require().Equal(defaultThirdeningEpochNum, mintKeeper.GetLastReductionEpochNum(ctx))

	expectedThirdenedProvisions := mintParams.ReductionFactor.Mul(genesisEpochProvisionsDec)
	// Sanity check with the actual value on mainnet.
	s.Require().Equal(defaultMainnetThirdenedProvisions, expectedThirdenedProvisions.String())
	s.Require().Equal(expectedThirdenedProvisions, mintKeeper.GetMinter(ctx).EpochProvisions)
}

func (s *KeeperTestSuite) assertAddressWeightsAddUpToOne(receivers []types.WeightedAddress) { //nolint:govet // this is a test and we can copy locks here.
	sumOfWeights := osmomath.ZeroDec()
	// As a sanity check, ensure developer reward receivers add up to 1.
	for _, w := range receivers {
		sumOfWeights = sumOfWeights.Add(w.Weight)
	}
	s.Require().Equal(osmomath.OneDec(), sumOfWeights)
}
