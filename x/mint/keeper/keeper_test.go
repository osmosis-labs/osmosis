package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

// setupDeveloperVestingModuleAccountTest sets up test cases that utilize developer vesting
// module account logic. It reverts some default logic added by suite.Setup()
// Specifically, it removes the developer vesting module account
// from account keeper if isDeveloperModuleAccountCreated is true.
// Additionally, it initializes suite's Ctx with blockHeight
func (suite *KeeperTestSuite) setupDeveloperVestingModuleAccountTest(blockHeight int64, isDeveloperModuleAccountCreated bool) {
	suite.Setup()
	// Reset height to the desired value since test suite setup initialized
	// it to 1.
	bankKeeper := suite.App.BankKeeper
	accountKeeper := suite.App.AccountKeeper

	suite.Ctx = suite.Ctx.WithBlockHeader(tmproto.Header{Height: blockHeight})

	if !isDeveloperModuleAccountCreated {
		// Remove the developer vesting account since suite setup creates and initializes it.
		// This environment w/o the developer vesting account configured is necessary for
		// testing edge cases of multiple tests.
		developerVestingAccount := accountKeeper.GetAccount(suite.Ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
		accountKeeper.RemoveAccount(suite.Ctx, developerVestingAccount)
		bankKeeper.BurnCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(keeper.DeveloperVestingAmount))))

		// If developer module account is created, the suite.Setup() also sets the offset,
		// therefore, we should reset it to 0 to set up the environment truly w/o the module account.
		supplyOffset := bankKeeper.GetSupplyOffset(suite.Ctx, sdk.DefaultBondDenom)
		bankKeeper.AddSupplyOffset(suite.Ctx, sdk.DefaultBondDenom, supplyOffset.Mul(sdk.NewInt(-1)))
		suite.Equal(sdk.ZeroInt(), bankKeeper.GetSupplyOffset(suite.Ctx, sdk.DefaultBondDenom))
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestMintCoinsToFeeCollectorAndGetProportions() {
	tests := []struct {
		name                 string
		ratio                sdk.Dec
		hasPreExistingSupply bool
		expectedCoin         sdk.Coin
		fee                  sdk.Coin
	}{
		{
			name:                 "coin is minted to the fee collector",
			fee:                  sdk.NewCoin("stake", sdk.NewInt(0)),
			ratio:                sdk.NewDecWithPrec(2, 1),
			hasPreExistingSupply: false,
			expectedCoin:         sdk.NewCoin("stake", sdk.NewInt(0)),
		}, {
			name:                 "mint the 100K stake coin to the fee collector",
			fee:                  sdk.NewCoin("stake", sdk.NewInt(100000)),
			ratio:                sdk.NewDecWithPrec(2, 1),
			hasPreExistingSupply: true,
			expectedCoin:         sdk.NewCoin("stake", sdk.NewInt(100000).Quo(sdk.NewInt(5))),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			bankKeeper := suite.App.BankKeeper
			mintKeeper := suite.App.MintKeeper

			if tc.hasPreExistingSupply {
				fee := sdk.NewCoin("stake", sdk.NewInt(100000))
				fees := sdk.NewCoins(fee)
				err := simapp.FundModuleAccount(bankKeeper,
					suite.Ctx,
					authtypes.FeeCollectorName,
					fees)
				suite.NoError(err)
			}

			coin := mintKeeper.GetProportions(suite.Ctx, tc.fee, tc.ratio)
			suite.Equal(tc.expectedCoin, coin)
		})
	}
}

func (suite *KeeperTestSuite) TestDistrAssetToDeveloperRewardsAddr() {
	var (
		distrTo = lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		}
		params       = suite.App.MintKeeper.GetParams(suite.Ctx)
		gaugeCoins   = sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
		gaugeCreator = sdk.AccAddress([]byte("addr2---------------"))
		mintLPtokens = sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	)

	tests := []struct {
		name              string
		weightedAddresses []types.WeightedAddress
		mintCoin          sdk.Coin
	}{
		{
			name: "one dev reward address",
			weightedAddresses: []types.WeightedAddress{
				{
					Address: sdk.AccAddress([]byte("addr1---------------")).String(),
					Weight:  sdk.NewDec(1),
				}},
			mintCoin: sdk.NewCoin("stake", sdk.NewInt(10000)),
		},
		{
			name: "multiple dev reward addresses",
			weightedAddresses: []types.WeightedAddress{
				{
					Address: sdk.AccAddress([]byte("addr3---------------")).String(),
					Weight:  sdk.NewDecWithPrec(6, 1),
				},
				{
					Address: sdk.AccAddress([]byte("addr4---------------")).String(),
					Weight:  sdk.NewDecWithPrec(4, 1),
				}},
			mintCoin: sdk.NewCoin("stake", sdk.NewInt(100000)),
		},
		{
			name:              "nil dev reward address",
			weightedAddresses: nil,
			mintCoin:          sdk.NewCoin("stake", sdk.NewInt(100000)),
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Setup()

			mintKeeper := suite.App.MintKeeper
			bankKeeper := suite.App.BankKeeper
			intencentivesKeeper := suite.App.IncentivesKeeper
			poolincentivesKeeper := suite.App.PoolIncentivesKeeper
			distrKeeper := suite.App.DistrKeeper
			accountKeeper := suite.App.AccountKeeper

			// set WeightedDeveloperRewardsReceivers
			params.WeightedDeveloperRewardsReceivers = tc.weightedAddresses
			mintKeeper.SetParams(suite.Ctx, params)

			// mints coins so supply exists on chain
			suite.FundAcc(gaugeCreator, gaugeCoins)
			suite.FundAcc(gaugeCreator, mintLPtokens)

			gaugeId, err := intencentivesKeeper.CreateGauge(suite.Ctx, true, gaugeCreator, gaugeCoins, distrTo, time.Now(), 1)
			suite.NoError(err)
			err = poolincentivesKeeper.UpdateDistrRecords(suite.Ctx, poolincentivestypes.DistrRecord{
				GaugeId: gaugeId,
				Weight:  sdk.NewInt(100),
			})
			suite.NoError(err)

			err = mintKeeper.MintCoins(suite.Ctx, sdk.NewCoins(tc.mintCoin))
			suite.NoError(err)

			err = mintKeeper.DistributeMintedCoin(suite.Ctx, tc.mintCoin)
			suite.NoError(err)

			// check feePool
			feePool := distrKeeper.GetFeePool(suite.Ctx)
			feeCollector := accountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			suite.Equal(
				tc.mintCoin.Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt(),
				bankKeeper.GetAllBalances(suite.Ctx, feeCollector).AmountOf("stake"))

			if tc.weightedAddresses != nil {
				suite.Equal(
					tc.mintCoin.Amount.ToDec().Mul(params.DistributionProportions.CommunityPool),
					feePool.CommunityPool.AmountOf("stake"))
			} else {
				suite.Equal(
					//distribution go to community pool because nil dev reward addresses.
					tc.mintCoin.Amount.ToDec().Mul((params.DistributionProportions.DeveloperRewards).Add(params.DistributionProportions.CommunityPool)),
					feePool.CommunityPool.AmountOf("stake"))
			}

			// check devAddress balances
			for i, weightedAddress := range tc.weightedAddresses {
				devRewardsReceiver, _ := sdk.AccAddressFromBech32(weightedAddress.GetAddress())
				suite.Equal(
					tc.mintCoin.Amount.ToDec().Mul(params.DistributionProportions.DeveloperRewards).Mul(params.WeightedDeveloperRewardsReceivers[i].Weight).TruncateInt(),
					bankKeeper.GetBalance(suite.Ctx, devRewardsReceiver, "stake").Amount)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDistrAssetToCommunityPoolWhenNoDeveloperRewardsAddr() {
	mintKeeper := suite.App.MintKeeper
	bankKeeper := suite.App.BankKeeper
	distrKeeper := suite.App.DistrKeeper
	accountKeeper := suite.App.AccountKeeper

	params := suite.App.MintKeeper.GetParams(suite.Ctx)
	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoin := sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins := sdk.Coins{mintCoin}
	err := mintKeeper.MintCoins(suite.Ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.Ctx, mintCoin)
	suite.NoError(err)

	distribution.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{}, *distrKeeper)

	feePool := distrKeeper.GetFeePool(suite.Ctx)
	feeCollector := accountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	// PoolIncentives + DeveloperRewards + CommunityPool => CommunityPool
	proportionToCommunity := params.DistributionProportions.PoolIncentives.
		Add(params.DistributionProportions.DeveloperRewards).
		Add(params.DistributionProportions.CommunityPool)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt(),
		bankKeeper.GetBalance(suite.Ctx, feeCollector, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(proportionToCommunity),
		feePool.CommunityPool.AmountOf("stake"))

	// Mint more and community pool should be increased
	err = mintKeeper.MintCoins(suite.Ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.Ctx, mintCoin)
	suite.NoError(err)

	distribution.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{}, *distrKeeper)

	feePool = distrKeeper.GetFeePool(suite.Ctx)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt().Mul(sdk.NewInt(2)),
		bankKeeper.GetBalance(suite.Ctx, feeCollector, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(proportionToCommunity).Mul(sdk.NewDec(2)),
		feePool.CommunityPool.AmountOf("stake"))
}

func (suite *KeeperTestSuite) TestCreateDeveloperVestingModuleAccount() {
	testcases := map[string]struct {
		blockHeight                     int64
		amount                          sdk.Coin
		isDeveloperModuleAccountCreated bool

		expectedError error
	}{
		"valid call": {
			blockHeight: 0,
			amount:      sdk.NewCoin("stake", sdk.NewInt(keeper.DeveloperVestingAmount)),
		},
		"nil amount": {
			blockHeight:   0,
			expectedError: keeper.ErrAmountCannotBeNilOrZero,
		},
		"zero amount": {
			blockHeight:   0,
			amount:        sdk.NewCoin("stake", sdk.NewInt(0)),
			expectedError: keeper.ErrAmountCannotBeNilOrZero,
		},
		"module account is already created": {
			blockHeight:                     0,
			amount:                          sdk.NewCoin("stake", sdk.NewInt(keeper.DeveloperVestingAmount)),
			isDeveloperModuleAccountCreated: true,
			expectedError:                   keeper.ErrDevVestingModuleAccountAlreadyCreated,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.setupDeveloperVestingModuleAccountTest(tc.blockHeight, tc.isDeveloperModuleAccountCreated)
			mintKeeper := suite.App.MintKeeper

			// Test
			actualError := mintKeeper.CreateDeveloperVestingModuleAccount(suite.Ctx, tc.amount)

			if tc.expectedError != nil {
				suite.Error(actualError)
				suite.Equal(actualError, tc.expectedError)
				return
			}
			suite.NoError(actualError)
		})
	}
}

func (suite *KeeperTestSuite) TestSetInitialSupplyOffsetDuringMigration() {
	testcases := map[string]struct {
		blockHeight                     int64
		isDeveloperModuleAccountCreated bool

		expectedError error
	}{
		"valid call": {
			blockHeight:                     1,
			isDeveloperModuleAccountCreated: true,
		},
		"dev vesting module account does not exist": {
			blockHeight:   1,
			expectedError: keeper.ErrDevVestingModuleAccountNotCreated,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.setupDeveloperVestingModuleAccountTest(tc.blockHeight, tc.isDeveloperModuleAccountCreated)
			ctx := suite.Ctx
			bankKeeper := suite.App.BankKeeper
			mintKeeper := suite.App.MintKeeper

			supplyWithOffsetBefore := bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom)
			supplyOffsetBefore := bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom)

			// Test
			actualError := mintKeeper.SetInitialSupplyOffsetDuringMigration(ctx)

			if tc.expectedError != nil {
				suite.Error(actualError)
				suite.Equal(actualError, tc.expectedError)

				suite.Equal(supplyWithOffsetBefore.Amount, bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)
				suite.Equal(supplyOffsetBefore, bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom))
				return
			}
			suite.NoError(actualError)
			suite.Equal(supplyWithOffsetBefore.Amount.Sub(sdk.NewInt(keeper.DeveloperVestingAmount)), bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)
			suite.Equal(supplyOffsetBefore.Sub(sdk.NewInt(keeper.DeveloperVestingAmount)), bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom))
		})
	}
}
