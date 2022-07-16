package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/btcutil/bech32"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

var (
	testAddressOne   = sdk.AccAddress([]byte("addr1---------------"))
	testAddressTwo   = sdk.AccAddress([]byte("addr2---------------"))
	testAddressThree = sdk.AccAddress([]byte("addr3---------------"))
	testAddressFour  = sdk.AccAddress([]byte("addr4---------------"))
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
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

// TestGetProportions tests that mint allocations are computed as expected.
func (suite *KeeperTestSuite) TestGetProportions() {
	complexRatioDec := sdk.NewDecWithPrec(131, 3).Quo(sdk.NewDecWithPrec(273, 3))

	tests := []struct {
		name          string
		ratio         sdk.Dec
		expectedCoin  sdk.Coin
		expectedError error
		mintedCoin    sdk.Coin
	}{
		{
			name:         "0 * 0.2 = 0",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),
			ratio:        sdk.NewDecWithPrec(2, 1),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),
		},
		{
			name:         "100000 * 0.2 = 20000",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000)),
			ratio:        sdk.NewDecWithPrec(2, 1),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000).Quo(sdk.NewInt(5))),
		},
		{
			name:         "123456 * 2/3 = 82304",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(123456)),
			ratio:        sdk.NewDecWithPrec(2, 1).Quo(sdk.NewDecWithPrec(3, 1)),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(82304)),
		},
		{
			name:       "54617981 * .131/.273 approx = 2.62",
			mintedCoin: sdk.NewCoin("uosmo", sdk.NewInt(54617981)),
			ratio:      complexRatioDec, // .131/.273
			// TODO: Should not be truncated. Remove truncation after rounding errors are addressed and resolved.
			// Ref: https://github.com/osmosis-labs/osmosis/issues/1917
			expectedCoin: sdk.NewCoin("uosmo", sdk.NewInt(54617981).ToDec().Mul(complexRatioDec).TruncateInt()),
		},
		{
			name:         "1 * 1 = 1",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)),
			ratio:        sdk.NewDec(1),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)),
		},
		{
			name:       "1 * 1.01 - error, ratio must be <= 1",
			mintedCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),
			ratio:      sdk.NewDecWithPrec(101, 2),

			expectedError: keeper.ErrInvalidRatio{ActualRatio: sdk.NewDecWithPrec(101, 2)},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			coin, err := keeper.GetProportions(suite.Ctx, tc.mintedCoin, tc.ratio)

			if tc.expectedError != nil {
				suite.Require().Equal(tc.expectedError, err)
				suite.Equal(sdk.Coin{}, coin)
				return
			}

			suite.NoError(err)
			suite.Equal(tc.expectedCoin, coin)
		})
	}
}

func (suite *KeeperTestSuite) TestDistributeMintedCoin_ToDeveloperRewardsAddr() {
	var (
		distrTo = lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		}
		params       = suite.App.MintKeeper.GetParams(suite.Ctx)
		gaugeCoins   = sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
		gaugeCreator = testAddressTwo
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
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
			mintCoin: sdk.NewCoin("stake", sdk.NewInt(10000)),
		},
		{
			name: "multiple dev reward addresses",
			weightedAddresses: []types.WeightedAddress{
				{
					Address: testAddressThree.String(),
					Weight:  sdk.NewDecWithPrec(6, 1),
				},
				{
					Address: testAddressFour.String(),
					Weight:  sdk.NewDecWithPrec(4, 1),
				},
			},
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
					// distribution go to community pool because nil dev reward addresses.
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
			expectedError: sdkerrors.Wrap(types.ErrAmountNilOrZero, "amount cannot be nil or zero"),
		},
		"zero amount": {
			blockHeight:   0,
			amount:        sdk.NewCoin("stake", sdk.NewInt(0)),
			expectedError: sdkerrors.Wrap(types.ErrAmountNilOrZero, "amount cannot be nil or zero"),
		},
		"module account is already created": {
			blockHeight:                     0,
			amount:                          sdk.NewCoin("stake", sdk.NewInt(keeper.DeveloperVestingAmount)),
			isDeveloperModuleAccountCreated: true,
			expectedError:                   sdkerrors.Wrapf(types.ErrModuleAlreadyExist, "%s vesting module account already exist", types.DeveloperVestingModuleAcctName),
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
				suite.ErrorIs(actualError, tc.expectedError)
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
			expectedError: sdkerrors.Wrapf(types.ErrModuleDoesnotExist, "%s vesting module account doesnot exist", types.DeveloperVestingModuleAcctName),
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
				suite.ErrorIs(actualError, tc.expectedError)

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

// TestDistributeToModule tests that distribution from mint module to another module helper
// function is working as expected.
func (suite *KeeperTestSuite) TestDistributeToModule() {
	const (
		denomDoesNotExist         = "denomDoesNotExist"
		moduleAccountDoesNotExist = "moduleAccountDoesNotExist"
	)

	tests := map[string]struct {
		preMintCoin sdk.Coin

		recepientModule string
		mintedCoin      sdk.Coin
		proportion      sdk.Dec

		expectedError bool
		expectPanic   bool
	}{
		"pre-mint == distribute - poolincentives module - full amount - success": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:      sdk.NewDec(1),
		},
		"pre-mint > distribute - developer vesting module - two thirds - success": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(101)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:      sdk.NewDecWithPrec(2, 1).Quo(sdk.NewDecWithPrec(3, 1)),
		},
		"pre-mint < distribute (0) - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:      sdk.NewDecWithPrec(2, 1).Quo(sdk.NewDecWithPrec(3, 1)),

			expectedError: true,
		},
		"denom does not exist - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(denomDoesNotExist, sdk.NewInt(100)),
			proportion:      sdk.NewDec(1),

			expectedError: true,
		},
		"invalid module account -panic": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),

			recepientModule: moduleAccountDoesNotExist,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:      sdk.NewDec(1),

			expectPanic: true,
		},
		"proportion greater than 1 - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(300)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:      sdk.NewDec(2),

			expectedError: true,
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			suite.Setup()
			osmoutils.ConditionalPanic(suite.T(), tc.expectPanic, func() {
				mintKeeper := suite.App.MintKeeper
				bankKeeper := suite.App.BankKeeper
				accountKeeper := suite.App.AccountKeeper
				ctx := suite.Ctx

				// Setup.
				suite.NoError(mintKeeper.MintCoins(ctx, sdk.NewCoins(tc.preMintCoin)))

				// TODO: Should not be truncated. Remove truncation after rounding errors are addressed and resolved.
				// Ref: https://github.com/osmosis-labs/osmosis/issues/1917
				expectedDistributed := tc.mintedCoin.Amount.ToDec().Mul(tc.proportion).TruncateInt()
				oldMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.mintedCoin.Denom).Amount
				oldRecepientModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(tc.recepientModule), tc.mintedCoin.Denom).Amount

				// Test.
				actualDistributed, err := mintKeeper.DistributeToModule(ctx, tc.recepientModule, tc.mintedCoin, tc.proportion)

				// Assertions.
				actualMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.mintedCoin.Denom).Amount
				actualRecepientModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(tc.recepientModule), tc.mintedCoin.Denom).Amount

				if tc.expectedError {
					suite.Error(err)
					suite.Equal(actualDistributed, sdk.Int{})
					// Old balances should not change.
					suite.Equal(oldMintModuleBalanceAmount.Int64(), actualMintModuleBalanceAmount.Int64())
					suite.Equal(oldRecepientModuleBalanceAmount.Int64(), actualRecepientModuleBalanceAmount.Int64())
					return
				}

				suite.NoError(err)
				suite.Equal(expectedDistributed, actualDistributed)

				// Updated balances.
				suite.Equal(oldMintModuleBalanceAmount.Sub(actualDistributed).Int64(), actualMintModuleBalanceAmount.Int64())
				suite.Equal(oldRecepientModuleBalanceAmount.Add(actualDistributed).Int64(), actualRecepientModuleBalanceAmount.Int64())
			})
		})
	}
}

// TestDistributeDeveloperRewards tests the following:
// - distribution from developer module account to the given weighted addressed occurs.
// - developer vesting module account balance is correctly updated.
// - all developer addressed are updated with correct proportions.
// - mint module account balance is updated - burn over allocations.
// - if recepients are empty - community pool us updated.
func (suite *KeeperTestSuite) TestDistributeDeveloperRewards() {
	const (
		invalidAddress = "invalid"
	)

	var (
		validLargePreMintAmount  = sdk.NewInt(keeper.DeveloperVestingAmount)
		validPreMintAmountAddOne = sdk.NewInt(keeper.DeveloperVestingAmount).Add(sdk.OneInt())
		validPreMintCoin         = sdk.NewCoin(sdk.DefaultBondDenom, validLargePreMintAmount)
		validPreMintCoinSubOne   = sdk.NewCoin(sdk.DefaultBondDenom, validLargePreMintAmount.Sub(sdk.OneInt()))
	)

	tests := map[string]struct {
		preMintCoin sdk.Coin

		mintedCoin         sdk.Coin
		proportion         sdk.Dec
		recepientAddresses []types.WeightedAddress

		expectedError error
		expectPanic   bool
		// See testcases with this flag set to true for details.
		allowBalanceChange bool
		// See testcases with this flag set to true for details.
		expectSameAddresses bool
	}{
		"valid case with 1 weighted address": {
			preMintCoin: validPreMintCoin,

			mintedCoin: validPreMintCoin,
			proportion: sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
		},
		"valid case with 3 weighted addresses and custom large mint amount under pre mint": {
			preMintCoin: validPreMintCoin,

			mintedCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(939_123_546_789)),
			proportion: sdk.NewDecWithPrec(31347, 5),
			recepientAddresses: []types.WeightedAddress{ // .231 + .4 + .369
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDecWithPrec(231, 3),
				},
				{
					Address: testAddressTwo.String(),
					Weight:  sdk.NewDecWithPrec(4, 1),
				},
				{
					Address: testAddressThree.String(),
					Weight:  sdk.NewDecWithPrec(369, 3),
				},
			},
		},
		"valid case with 2 addresses that are the same": {
			preMintCoin: validPreMintCoin,

			mintedCoin: validPreMintCoin,
			proportion: sdk.NewDecWithPrec(123, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDecWithPrec(5, 1),
				},
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDecWithPrec(5, 1),
				},
			},
			// Since we have double the full amount allocated
			/// to the same address, the balance assertions will
			// differ by expecting the full minted amount.
			expectSameAddresses: true,
		},
		"valid case with 0 reward receivers - goes to community pool": {
			preMintCoin: validPreMintCoin,

			mintedCoin: validPreMintCoin,
			proportion: sdk.NewDecWithPrec(153, 3),
		},
		"valid case with 0 amount of total minted coin": {
			preMintCoin: validPreMintCoin,

			mintedCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt()),
			proportion: sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
		},
		"invalid value for developer rewards proportion (> 1) - error": {
			preMintCoin: validPreMintCoin,

			mintedCoin: validPreMintCoin,
			proportion: sdk.NewDec(2),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},

			expectedError: keeper.ErrInvalidRatio{ActualRatio: sdk.NewDec(2)},
		},
		"invalid address in developer reward receivers - error": {
			preMintCoin: validPreMintCoin,

			mintedCoin: validPreMintCoin,
			proportion: sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: invalidAddress,
					Weight:  sdk.NewDec(1),
				},
			},

			expectedError: sdkerrors.Wrap(bech32.ErrInvalidLength(len(invalidAddress)), "decoding bech32 failed"),
			// This case should not happen in practice due to parameter validation.
			// The method spec also requires that all recepient addresses are valid by CONTRACT.
			// Since we still handle error returned by the converion from string to address,
			// we try to cover it explicitly. However, it changes balance so we don't test it.
			allowBalanceChange: true,
		},
		"pre-mint < distribute * proportion - error": {
			preMintCoin: validPreMintCoinSubOne,

			mintedCoin: validPreMintCoin,
			proportion: sdk.OneDec(),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
			expectedError: sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("%s is smaller than %s", validPreMintCoinSubOne, validPreMintCoin)),
		},
		"distribute * proportion < pre-mint but distribute * proportion > developer vesting amount - error": {
			preMintCoin: validPreMintCoin,

			mintedCoin: sdk.NewCoin(sdk.DefaultBondDenom, validPreMintAmountAddOne),
			proportion: sdk.OneDec(),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
			expectedError: sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("%s is smaller than %s", validPreMintCoin, sdk.NewCoin(sdk.DefaultBondDenom, validPreMintAmountAddOne))),
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			suite.Setup()

			osmoutils.ConditionalPanic(suite.T(), tc.expectPanic, func() {
				mintKeeper := suite.App.MintKeeper
				bankKeeper := suite.App.BankKeeper
				accountKeeper := suite.App.AccountKeeper
				ctx := suite.Ctx

				// Setup.
				suite.NoError(mintKeeper.MintCoins(ctx, sdk.NewCoins(tc.preMintCoin)))

				// TODO: Should not be truncated. Remove truncation after rounding errors are addressed and resolved.
				// Ref: https://github.com/osmosis-labs/osmosis/issues/1917
				expectedDistributed := tc.mintedCoin.Amount.ToDec().Mul(tc.proportion).TruncateInt()

				oldMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.mintedCoin.Denom).Amount
				oldDeveloperVestingModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), tc.mintedCoin.Denom).Amount
				oldCommunityPoolBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), tc.mintedCoin.Denom).Amount
				oldDeveloperRewardsBalanceAmounts := make([]sdk.Int, len(tc.recepientAddresses))
				for i, weightedAddress := range tc.recepientAddresses {
					// No error check to be able to test invalid addresses.
					address, _ := sdk.AccAddressFromBech32(weightedAddress.Address)
					oldDeveloperRewardsBalanceAmounts[i] = bankKeeper.GetBalance(ctx, address, tc.mintedCoin.Denom).Amount
				}

				// Test.
				actualDistributed, err := mintKeeper.DistributeDeveloperRewards(ctx, tc.mintedCoin, tc.proportion, tc.recepientAddresses)

				// Assertions.
				actualMintModuleBalance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.mintedCoin.Denom)
				actualDeveloperVestingModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), tc.mintedCoin.Denom).Amount
				actualCommunityPoolModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), tc.mintedCoin.Denom).Amount

				if tc.expectedError != nil {
					suite.Error(err)
					suite.Equal(tc.expectedError.Error(), err.Error())
					suite.Equal(actualDistributed, sdk.Int{})

					// See testcases with this flag set to true for details.
					if tc.allowBalanceChange {
						return
					}
					// Old balances should not change.
					suite.Equal(oldMintModuleBalanceAmount.Int64(), actualMintModuleBalance.Amount.Int64())
					suite.Equal(oldDeveloperVestingModuleBalanceAmount.Int64(), actualDeveloperVestingModuleBalanceAmount.Int64())
					suite.Equal(oldCommunityPoolBalanceAmount.Int64(), actualCommunityPoolModuleBalanceAmount.Int64())
					return
				}

				suite.NoError(err)
				suite.Equal(expectedDistributed, actualDistributed)

				// Updated balances.

				// Burn from mint module account. We over-allocate.
				// To be fixed: https://github.com/osmosis-labs/osmosis/issues/2025
				suite.Equal(oldMintModuleBalanceAmount.Sub(expectedDistributed).Int64(), actualMintModuleBalance.Amount.Int64())

				// Allocate to community pool when no addresses are provided.
				if len(tc.recepientAddresses) == 0 {
					suite.Equal(oldDeveloperVestingModuleBalanceAmount.Sub(expectedDistributed).Int64(), actualDeveloperVestingModuleBalanceAmount.Int64())
					suite.Equal(oldCommunityPoolBalanceAmount.Add(expectedDistributed).Int64(), actualCommunityPoolModuleBalanceAmount.Int64())
					return
				}

				// TODO: these should be equal, slightly off due to known rounding issues: https://github.com/osmosis-labs/osmosis/issues/1917
				// suite.Equal(oldDeveloperVestingModuleBalanceAmount.Sub(expectedDistributed).Int64(), actualDeveloperVestingModuleBalanceAmount.Int64())

				for i, weightedAddress := range tc.recepientAddresses {
					address, err := sdk.AccAddressFromBech32(weightedAddress.Address)
					suite.NoError(err)

					// TODO: truncation should not occur: https://github.com/osmosis-labs/osmosis/issues/1917
					expectedAllocation := expectedDistributed.ToDec().Mul(tc.recepientAddresses[i].Weight).TruncateInt()
					actualDeveloperRewardsBalanceAmounts := bankKeeper.GetBalance(ctx, address, tc.mintedCoin.Denom).Amount

					// Edge case. See testcases with this flag set to true for details.
					if tc.expectSameAddresses {
						suite.Equal(oldDeveloperRewardsBalanceAmounts[i].Add(expectedAllocation.Mul(sdk.NewInt(2))).Int64(), actualDeveloperRewardsBalanceAmounts.Int64())
						return
					}

					suite.Equal(oldDeveloperRewardsBalanceAmounts[i].Add(expectedAllocation).Int64(), actualDeveloperRewardsBalanceAmounts.Int64())
				}
			})
		})
	}
}
