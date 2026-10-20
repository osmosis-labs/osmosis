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
	osmoapp "github.com/osmosis-labs/osmosis/v32/app"
	"github.com/osmosis-labs/osmosis/v32/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v32/x/mint/types"

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
		DeveloperRewards: osmomath.ZeroDec(),
		CommunityPool:    osmomath.NewDecWithPrec(30, 2),
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
		emptyDeveloperVesting   bool

		// Expected results.
		expectedLastReductionEpochNum int64
		expectedDistribution          osmomath.Dec
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
				DeveloperRewards: osmomath.ZeroDec(),
				CommunityPool:    osmomath.NewDecWithPrec(67, 2),
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
		"empty developer vesting module account - distributes": {
			hookArgEpochNum: defaultMintingRewardsDistributionStartEpoch,

			mintDenom:               sdk.DefaultBondDenom,
			genesisEpochProvisions:  defaultGenesisEpochProvisionsDec,
			epochIdentifier:         defaultEpochIdentifier,
			reductionPeriodInEpochs: defaultReductionPeriodInEpochs,
			reductionFactor:         defaultReductionFactor,
			distributionProportions: defaultDistributionProportions,
			weightedAddresses:       testWeightedAddresses,
			mintStartEpoch:          defaultMintingRewardsDistributionStartEpoch,
			emptyDeveloperVesting:   true,

			expectedDistribution:          defaultGenesisEpochProvisionsDec,
			expectedLastReductionEpochNum: defaultMintingRewardsDistributionStartEpoch,
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

			// We require a validator be setup in the app setup logic, or else tests won't run.
			// This validator has a delegation equal to sdk.DefaultPowerReduction, so we add this
			// to the expected supply.
			s.Require().Equal(osmomath.NewInt(keeper.DeveloperVestingAmount).Add(sdk.DefaultPowerReduction), bankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount)

			developerVestingAddr := accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
			if tc.emptyDeveloperVesting {
				// Minting must not depend on the developer vesting module account balance.
				developerVestingBalance := bankKeeper.GetBalance(ctx, developerVestingAddr, sdk.DefaultBondDenom)
				s.Require().NoError(distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(developerVestingBalance), developerVestingAddr))
			}

			developerAccountBalanceBeforeHook := bankKeeper.GetBalance(ctx, developerVestingAddr, sdk.DefaultBondDenom)
			supplyBeforeHook := bankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount
			supplyWithOffsetBeforeHook := bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount

			s.Require().NoError(mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, tc.hookArgEpochNum))

			// The developer vesting module account is not part of mint distribution.
			s.Require().Equal(developerAccountBalanceBeforeHook, bankKeeper.GetBalance(ctx, developerVestingAddr, sdk.DefaultBondDenom))

			// Everything minted is new supply: nothing is burned and the supply offset does not move.
			expectedMinted := tc.expectedDistribution.TruncateInt()
			s.Require().Equal(supplyBeforeHook.Add(expectedMinted), bankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount)
			s.Require().Equal(supplyWithOffsetBeforeHook.Add(expectedMinted), bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)

			// Deprecated developer rewards receivers get nothing.
			for _, w := range tc.weightedAddresses {
				s.Require().True(bankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(w.Address), sdk.DefaultBondDenom).IsZero())
			}

			// Validate epoch provisions.
			s.Require().Equal(tc.expectedLastReductionEpochNum, mintKeeper.GetLastReductionEpochNum(ctx))

			if !tc.expectedDistribution.IsZero() {
				// Validate distribution.
				osmoassert.DecApproxEq(s.T(), tc.expectedDistribution, mintKeeper.GetMinter(ctx).EpochProvisions, osmomath.NewDecWithPrec(1, 6))
			}
		})
	}
}

func (s *KeeperTestSuite) TestAfterEpochEnd_FirstYearThirdening_RealParameters() {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	mintKeeper := app.MintKeeper
	accountKeeper := app.AccountKeeper

	genesisEpochProvisionsDec, err := osmomath.NewDecFromStr(defaultGenesisEpochProvisions)
	s.Require().NoError(err)

	mintParams := types.Params{
		MintDenom:                            sdk.DefaultBondDenom,
		GenesisEpochProvisions:               genesisEpochProvisionsDec,
		EpochIdentifier:                      defaultEpochIdentifier,
		ReductionPeriodInEpochs:              defaultReductionPeriodInEpochs,
		ReductionFactor:                      defaultReductionFactor,
		DistributionProportions:              defaultDistributionProportions,
		WeightedDeveloperRewardsReceivers:    []types.WeightedAddress{},
		MintingRewardsDistributionStartEpoch: defaultMintingRewardsDistributionStartEpoch,
	}

	// Test setup parameters are not identical with mainnet.
	// Therefore, we set them here to the desired mainnet values.
	mintKeeper.SetParams(ctx, mintParams)
	mintKeeper.SetLastReductionEpochNum(ctx, 0)
	mintKeeper.SetMinter(ctx, types.Minter{
		EpochProvisions: genesisEpochProvisionsDec,
	})

	// In test setup, we set a validator with a delegation equal to sdk.DefaultPowerReduction.
	expectedSupplyWithOffset := sdk.DefaultPowerReduction
	expectedSupply := osmomath.NewInt(keeper.DeveloperVestingAmount).Add(sdk.DefaultPowerReduction)

	s.Require().Equal(expectedSupplyWithOffset, app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)
	s.Require().Equal(expectedSupply, app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount)

	developerVestingAddr := accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	developerAccountBalance := app.BankKeeper.GetBalance(ctx, developerVestingAddr, sdk.DefaultBondDenom)

	// System truncates EpochProvisions because bank takes an Int, so each epoch mints the truncated amount.
	truncatedEpochProvisions := genesisEpochProvisionsDec.TruncateInt()

	// Actual test for running AfterEpochEnd hook thirdeningEpoch times.
	for i := int64(1); i <= defaultReductionPeriodInEpochs; i++ {
		// System under test.
		err = mintKeeper.AfterEpochEnd(ctx, defaultEpochIdentifier, i)
		s.Require().NoError(err)

		// Everything minted is new supply and the developer vesting balance does not move.
		expectedSupply = expectedSupply.Add(truncatedEpochProvisions)
		expectedSupplyWithOffset = expectedSupplyWithOffset.Add(truncatedEpochProvisions)
		s.Require().Equal(expectedSupply, app.BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount)
		s.Require().Equal(expectedSupplyWithOffset, app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)
		s.Require().Equal(developerAccountBalance, app.BankKeeper.GetBalance(ctx, developerVestingAddr, sdk.DefaultBondDenom))

		// Validate that the epoch provisions have not been reduced.
		s.Require().Equal(defaultMintingRewardsDistributionStartEpoch, mintKeeper.GetLastReductionEpochNum(ctx))
		s.Require().Equal(defaultGenesisEpochProvisions, mintKeeper.GetMinter(ctx).EpochProvisions.String())
	}

	// Validate total supply. The only rounding is the per-epoch truncation of the provisions.
	// expectedTotalProvisionedSupply = 365 * 821917808219.178082191780821917 = 299_999_999_999_999.999999999999999705
	// epochProvisionsDelta = 0.178082191780821917 * 365 = 64.999999999999999705
	expectedTotalProvisionedSupply := osmomath.NewDec(defaultReductionPeriodInEpochs).Mul(genesisEpochProvisionsDec).Add(sdk.DefaultPowerReduction.ToLegacyDec())
	epochProvisionsDelta := genesisEpochProvisionsDec.Sub(truncatedEpochProvisions.ToLegacyDec()).Mul(osmomath.NewDec(defaultReductionPeriodInEpochs))
	actualTotalProvisionedSupply := app.BankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount.ToLegacyDec()
	s.Require().Equal(expectedTotalProvisionedSupply, actualTotalProvisionedSupply.Add(epochProvisionsDelta))

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
