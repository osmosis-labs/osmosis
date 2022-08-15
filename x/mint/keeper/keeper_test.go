package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/btcutil/bech32"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/osmoutils"
	"github.com/osmosis-labs/osmosis/v11/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

type mintHooksMock struct {
	hookCallCount int
}

func (hm *mintHooksMock) AfterDistributeMintedCoin(ctx sdk.Context) {
	hm.hookCallCount++
}

var _ types.MintHooks = (*mintHooksMock)(nil)

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
	params := suite.App.MintKeeper.GetParams(suite.Ctx)
	params.ReductionPeriodInEpochs = 10
	suite.App.MintKeeper.SetParams(suite.Ctx, params)
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
		suite.Require().Equal(sdk.ZeroInt(), bankKeeper.GetSupplyOffset(suite.Ctx, sdk.DefaultBondDenom))
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
			coin, err := keeper.GetProportions(tc.mintedCoin, tc.ratio)

			if tc.expectedError != nil {
				suite.Require().Equal(tc.expectedError, err)
				suite.Require().Equal(sdk.Coin{}, coin)
				return
			}

			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedCoin, coin)
		})
	}
}

func (suite *KeeperTestSuite) TestDistributeMintedCoin() {
	const (
		mintAmount = 10000
	)

	var (
		params = types.DefaultParams()
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
			mintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(mintAmount)),
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
			mintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(mintAmount)),
		},
		{
			name:              "nil dev reward address - dev rewards go to community pool",
			weightedAddresses: nil,
			mintCoin:          sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(mintAmount)),
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Setup()

			ctx := suite.Ctx

			bankKeeper := suite.App.BankKeeper
			accountKeeper := suite.App.AccountKeeper

			mintKeeper := suite.App.MintKeeper
			// We reset the hooks with a mock to simplify the assertions
			// about the results of the call to DistributeMintedCoin.
			// The goal is to assert that AfterDistributeMintedCoin
			// is called once.
			mintKeeper.SetMintHooksUnsafe(&mintHooksMock{})

			mintAmount := tc.mintCoin.Amount.ToDec()

			// set WeightedDeveloperRewardsReceivers
			params.WeightedDeveloperRewardsReceivers = tc.weightedAddresses
			mintKeeper.SetParams(ctx, params)

			expectedCommunityPoolAmount := mintAmount.Mul((params.DistributionProportions.CommunityPool))
			expectedDevRewardsAmount := mintAmount.Mul(params.DistributionProportions.DeveloperRewards)
			expectedPoolIncentivesAmount := mintAmount.Mul(params.DistributionProportions.PoolIncentives)
			expectedStakingAmount := tc.mintCoin.Amount.ToDec().Mul(params.DistributionProportions.Staking)

			// distributions go to community pool because nil dev reward addresses.
			if tc.weightedAddresses == nil {
				expectedCommunityPoolAmount = expectedCommunityPoolAmount.Add(expectedDevRewardsAmount)
			}

			// mints coins so supply exists on chain
			err := mintKeeper.MintCoins(ctx, sdk.NewCoins(tc.mintCoin))
			suite.Require().NoError(err)

			// System under test.
			err = mintKeeper.DistributeMintedCoin(ctx, tc.mintCoin)
			suite.Require().NoError(err)

			// validate that AfterDistributeMintedCoin hook was called once.
			suite.Require().Equal(1, mintKeeper.GetMintHooksUnsafe().(*mintHooksMock).hookCallCount)

			// validate distributions to fee collector.
			feeCollectorBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(authtypes.FeeCollectorName), sdk.DefaultBondDenom).Amount.ToDec()
			suite.Require().Equal(
				expectedStakingAmount,
				feeCollectorBalanceAmount)

			// validate pool incentives distributions.
			actualPoolIncentivesBalance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(poolincentivestypes.ModuleName), sdk.DefaultBondDenom).Amount.ToDec()
			suite.Require().Equal(expectedPoolIncentivesAmount, actualPoolIncentivesBalance)

			// validate distributions to community pool.
			actualCommunityPoolBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), sdk.DefaultBondDenom).Amount.ToDec()
			suite.Require().Equal(expectedCommunityPoolAmount, actualCommunityPoolBalanceAmount)

			// validate distributions to developer addresses.
			for i, weightedAddress := range tc.weightedAddresses {
				devRewardsReceiver, _ := sdk.AccAddressFromBech32(weightedAddress.GetAddress())
				suite.Require().Equal(
					expectedDevRewardsAmount.Mul(params.WeightedDeveloperRewardsReceivers[i].Weight).TruncateInt(),
					bankKeeper.GetBalance(ctx, devRewardsReceiver, sdk.DefaultBondDenom).Amount)
			}
		})
	}
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
			expectedError:                   sdkerrors.Wrapf(types.ErrModuleAccountAlreadyExist, "%s vesting module account already exist", types.DeveloperVestingModuleAcctName),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.setupDeveloperVestingModuleAccountTest(tc.blockHeight, tc.isDeveloperModuleAccountCreated)
			mintKeeper := suite.App.MintKeeper

			// Test
			actualError := mintKeeper.CreateDeveloperVestingModuleAccount(suite.Ctx, tc.amount)

			if tc.expectedError != nil {
				suite.Require().Error(actualError)
				suite.Require().ErrorIs(actualError, tc.expectedError)
				return
			}
			suite.Require().NoError(actualError)
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
				suite.Require().Error(actualError)
				suite.Require().ErrorIs(actualError, tc.expectedError)

				suite.Require().Equal(supplyWithOffsetBefore.Amount, bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)
				suite.Require().Equal(supplyOffsetBefore, bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom))
				return
			}
			suite.Require().NoError(actualError)
			suite.Require().Equal(supplyWithOffsetBefore.Amount.Sub(sdk.NewInt(keeper.DeveloperVestingAmount)), bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)
			suite.Require().Equal(supplyOffsetBefore.Sub(sdk.NewInt(keeper.DeveloperVestingAmount)), bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom))
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
				suite.Require().NoError(mintKeeper.MintCoins(ctx, sdk.NewCoins(tc.preMintCoin)))

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
					suite.Require().Error(err)
					suite.Require().Equal(actualDistributed, sdk.Int{})
					// Old balances should not change.
					suite.Require().Equal(oldMintModuleBalanceAmount.Int64(), actualMintModuleBalanceAmount.Int64())
					suite.Require().Equal(oldRecepientModuleBalanceAmount.Int64(), actualRecepientModuleBalanceAmount.Int64())
					return
				}

				suite.Require().NoError(err)
				suite.Require().Equal(expectedDistributed, actualDistributed)

				// Updated balances.
				suite.Require().Equal(oldMintModuleBalanceAmount.Sub(actualDistributed).Int64(), actualMintModuleBalanceAmount.Int64())
				suite.Require().Equal(oldRecepientModuleBalanceAmount.Add(actualDistributed).Int64(), actualRecepientModuleBalanceAmount.Int64())
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
			expectedError: keeper.ErrInsufficientDevVestingBalance{ActualBalance: validPreMintCoin.Amount, AttemptedDistribution: validPreMintAmountAddOne},
		},
		"valid case with 1 empty string weighted address - distributes to community pool": {
			preMintCoin: validPreMintCoin,

			mintedCoin: validPreMintCoin,
			proportion: sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: keeper.EmptyWeightedAddressReceiver,
					Weight:  sdk.NewDec(1),
				},
			},
		},
		"valid case with 2 addresses - empty string (distributes to community pool) and regular address (distributes to the address)": {
			preMintCoin: validPreMintCoin,

			mintedCoin: validPreMintCoin,
			proportion: sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: keeper.EmptyWeightedAddressReceiver,
					Weight:  sdk.NewDec(1),
				},
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
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
				suite.Require().NoError(mintKeeper.MintCoins(ctx, sdk.NewCoins(tc.preMintCoin)))

				// TODO: Should not be truncated. Remove truncation after rounding errors are addressed and resolved.
				// Ref: https://github.com/osmosis-labs/osmosis/issues/1917
				expectedDistributed := tc.mintedCoin.Amount.ToDec().Mul(tc.proportion).TruncateInt()

				oldMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.mintedCoin.Denom).Amount
				oldDeveloperVestingModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), tc.mintedCoin.Denom).Amount
				oldCommunityPoolBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), tc.mintedCoin.Denom).Amount
				oldDeveloperRewardsBalanceAmounts := make([]sdk.Int, len(tc.recepientAddresses))
				for i, weightedAddress := range tc.recepientAddresses {
					if weightedAddress.Address == keeper.EmptyWeightedAddressReceiver {
						continue
					}

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
					suite.Require().Error(err)
					suite.Require().Equal(tc.expectedError.Error(), err.Error())
					suite.Require().Equal(actualDistributed, sdk.Int{})

					// See testcases with this flag set to true for details.
					if tc.allowBalanceChange {
						return
					}
					// Old balances should not change.
					suite.Require().Equal(oldMintModuleBalanceAmount.Int64(), actualMintModuleBalance.Amount.Int64())
					suite.Require().Equal(oldDeveloperVestingModuleBalanceAmount.Int64(), actualDeveloperVestingModuleBalanceAmount.Int64())
					suite.Require().Equal(oldCommunityPoolBalanceAmount.Int64(), actualCommunityPoolModuleBalanceAmount.Int64())
					return
				}

				suite.Require().NoError(err)
				suite.Require().Equal(expectedDistributed, actualDistributed)

				// Updated balances.

				// Burn from mint module account. We over-allocate.
				// To be fixed: https://github.com/osmosis-labs/osmosis/issues/2025
				suite.Require().Equal(oldMintModuleBalanceAmount.Sub(expectedDistributed).Int64(), actualMintModuleBalance.Amount.Int64())

				// Allocate to community pool when no addresses are provided.
				if len(tc.recepientAddresses) == 0 {
					suite.Require().Equal(oldDeveloperVestingModuleBalanceAmount.Sub(expectedDistributed).Int64(), actualDeveloperVestingModuleBalanceAmount.Int64())
					suite.Require().Equal(oldCommunityPoolBalanceAmount.Add(expectedDistributed).Int64(), actualCommunityPoolModuleBalanceAmount.Int64())
					return
				}

				// TODO: these should be equal, slightly off due to known rounding issues: https://github.com/osmosis-labs/osmosis/issues/1917
				// suite.Require().Equal(oldDeveloperVestingModuleBalanceAmount.Sub(expectedDistributed).Int64(), actualDeveloperVestingModuleBalanceAmount.Int64())

				expectedDistributedCommunityPool := sdk.NewInt(0)

				for i, weightedAddress := range tc.recepientAddresses {
					// TODO: truncation should not occur: https://github.com/osmosis-labs/osmosis/issues/1917
					expectedAllocation := expectedDistributed.ToDec().Mul(tc.recepientAddresses[i].Weight).TruncateInt()

					if weightedAddress.Address == keeper.EmptyWeightedAddressReceiver {
						expectedDistributedCommunityPool = expectedDistributedCommunityPool.Add(expectedAllocation)
						continue
					}

					address, err := sdk.AccAddressFromBech32(weightedAddress.Address)
					suite.Require().NoError(err)

					actualDeveloperRewardsBalanceAmounts := bankKeeper.GetBalance(ctx, address, tc.mintedCoin.Denom).Amount

					// Edge case. See testcases with this flag set to true for details.
					if tc.expectSameAddresses {
						suite.Require().Equal(oldDeveloperRewardsBalanceAmounts[i].Add(expectedAllocation.Mul(sdk.NewInt(2))).Int64(), actualDeveloperRewardsBalanceAmounts.Int64())
						return
					}

					suite.Require().Equal(oldDeveloperRewardsBalanceAmounts[i].Add(expectedAllocation).Int64(), actualDeveloperRewardsBalanceAmounts.Int64())
				}

				suite.Require().Equal(oldCommunityPoolBalanceAmount.Add(expectedDistributedCommunityPool).Int64(), actualCommunityPoolModuleBalanceAmount.Int64())
			})
		})
	}
}
