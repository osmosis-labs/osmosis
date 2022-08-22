package keeper_test

import (
	"testing"

	"github.com/cosmos/btcutil/bech32"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v11/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

const otherDenom = "other"

var (
	testAddressOne   = sdk.AccAddress([]byte("addr1---------------"))
	testAddressTwo   = sdk.AccAddress([]byte("addr2---------------"))
	testAddressThree = sdk.AccAddress([]byte("addr3---------------"))
	testAddressFour  = sdk.AccAddress([]byte("addr4---------------"))
	baseAmount       = sdk.NewInt(10000)
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

func (suite *KeeperTestSuite) ValidateSupplyAndMintModuleAccounts(expectedDeveloperVestingAccountBalance sdk.Int, expectedMintModuleAccountBalance, expectedMintedTruncated sdk.Int) {
	bankKeeper := suite.App.BankKeeper
	accountKeeper := suite.App.AccountKeeper
	ctx := suite.Ctx

	// Developer vesting module account balance
	actualDeveloperVestingAccountBalance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), sdk.DefaultBondDenom)
	suite.Require().Equal(expectedDeveloperVestingAccountBalance.String(), actualDeveloperVestingAccountBalance.Amount.String())

	// Mint module account balance
	actualMintModuleAccountBalance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)
	suite.Require().Equal(expectedMintModuleAccountBalance.String(), actualMintModuleAccountBalance.Amount.String())

	// Supply offset.
	actualSupplyOffset := bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom)
	suite.Require().Equal(expectedDeveloperVestingAccountBalance.Neg(), actualSupplyOffset)

	// Minted supply.
	suite.Require().Equal(sdk.NewInt(keeper.DeveloperVestingAmount).Add(expectedMintedTruncated).String(), bankKeeper.GetSupply(ctx, sdk.DefaultBondDenom).Amount.String())

	// Supply with offset (minted supply - supply offset)
	expectedSupplyWithOffset := sdk.NewInt(keeper.DeveloperVestingAmount).Add(expectedMintedTruncated).Sub(expectedDeveloperVestingAccountBalance)
	suite.Require().Equal(expectedSupplyWithOffset, bankKeeper.GetSupplyWithOffset(ctx, sdk.DefaultBondDenom).Amount)
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
		expected      sdk.Dec
		expectedError error
		minted        sdk.Dec
	}{
		{
			name:     "0 * 0.2 = 0",
			minted:   sdk.ZeroDec(),
			ratio:    sdk.NewDecWithPrec(2, 1),
			expected: sdk.ZeroDec(),
		},
		{
			name:     "100000 * 0.2 = 20000",
			minted:   sdk.NewDec(100000),
			ratio:    sdk.NewDecWithPrec(2, 1),
			expected: sdk.NewDec(100000).Quo(sdk.NewDec(5)),
		},
		{
			name:     "123456 * 2/3 = 82304",
			minted:   sdk.NewDec(123456),
			ratio:    sdk.NewDecWithPrec(2, 1).Quo(sdk.NewDecWithPrec(3, 1)),
			expected: sdk.NewDec(123456).Mul(sdk.NewDecWithPrec(2, 1).Quo(sdk.NewDecWithPrec(3, 1))),
		},
		{
			name:     "54617981 * .131/.273 approx = 2.62",
			minted:   sdk.NewDec(54617981),
			ratio:    complexRatioDec, // .131/.273
			expected: sdk.NewDec(54617981).Mul(complexRatioDec),
		},
		{
			name:     "1 * 1 = 1",
			minted:   sdk.OneDec(),
			ratio:    sdk.OneDec(),
			expected: sdk.OneDec(),
		},
		{
			name:   "1 * 1.01 - error, ratio must be <= 1",
			minted: sdk.ZeroDec(),
			ratio:  sdk.NewDecWithPrec(101, 2),

			expectedError: keeper.ErrInvalidRatio{ActualRatio: sdk.NewDecWithPrec(101, 2)},
		},
		{
			name:     "123456.789 * .131/.273 = 59241.1698",
			minted:   sdk.NewDec(123456789).Quo(sdk.NewDec(1000)),
			ratio:    complexRatioDec, // .131/.273
			expected: sdk.NewDec(123456789).Quo(sdk.NewDec(1000)).Mul(complexRatioDec),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			actual, err := keeper.GetProportions(tc.minted, tc.ratio)

			if tc.expectedError != nil {
				suite.Require().Equal(tc.expectedError, err)
				suite.Require().Equal(sdk.Dec{}, actual)
				return
			}

			suite.Require().NoError(err)
			suite.Require().Equal(tc.expected, actual)
		})
	}
}

func (suite *KeeperTestSuite) TestDistributeInflationProvisions() {
	tests := []struct {
		preExistingTruncationDelta sdk.Dec

		name                string
		proportions         types.DistributionProportions
		inflationProvisions sdk.DecCoin

		expectedDistributed sdk.Int
		// N.B.: if floor(preExistingTruncationDelta + inflationProvisions.Amount) > floor(inflationProvisions.Amount),
		// the difference goes to the community pool.
		expectedCommunityPoolDistributionsFromTruncations sdk.Int
		expectedLeftOverTruncationsDelta                  sdk.Dec
		expectNoOp                                        bool
	}{
		{
			name:                "default proportions",
			proportions:         types.DefaultParams().DistributionProportions,
			inflationProvisions: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(10000)),

			expectedDistributed: sdk.NewInt(10000),
			expectedCommunityPoolDistributionsFromTruncations: sdk.ZeroInt(),
			expectedLeftOverTruncationsDelta:                  sdk.ZeroDec(),
		},
		{
			name:                "custom proportions",
			proportions:         defaultDistributionProportions,
			inflationProvisions: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(12345)),

			expectedDistributed: sdk.NewInt(12345),
			expectedCommunityPoolDistributionsFromTruncations: sdk.ZeroInt(),
			expectedLeftOverTruncationsDelta:                  sdk.ZeroDec(),
		},
		{
			name:                "default proportions - truncated decimal amount; no truncation delta pre-existing",
			proportions:         defaultDistributionProportions,
			inflationProvisions: sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(1000099, 2)),

			expectedDistributed: sdk.NewInt(10000),
			expectedCommunityPoolDistributionsFromTruncations: sdk.ZeroInt(),
			expectedLeftOverTruncationsDelta:                  sdk.NewDecWithPrec(99, 2),
		},
		{
			name:                       "default proportions - truncated decimal amount; truncation delta of 0.1 pre-existing",
			preExistingTruncationDelta: sdk.NewDecWithPrec(11, 2),
			proportions:                defaultDistributionProportions,
			inflationProvisions:        sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(100009, 1)),

			expectedDistributed: sdk.NewInt(10000).Add(sdk.NewInt(1)),
			expectedCommunityPoolDistributionsFromTruncations: sdk.OneInt(),
			expectedLeftOverTruncationsDelta:                  sdk.NewDecWithPrec(1, 2),
		},
		{
			name:                "xero provisions - no-op",
			proportions:         types.DefaultParams().DistributionProportions,
			inflationProvisions: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.ZeroInt()),

			expectedDistributed: sdk.ZeroInt(),
			expectedCommunityPoolDistributionsFromTruncations: sdk.ZeroInt(),
			expectedLeftOverTruncationsDelta:                  sdk.ZeroDec(),
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Setup()

			ctx := suite.Ctx

			bankKeeper := suite.App.BankKeeper
			accountKeeper := suite.App.AccountKeeper

			mintKeeper := suite.App.MintKeeper

			inflationAmount := tc.inflationProvisions.Amount

			if !tc.preExistingTruncationDelta.IsNil() {
				suite.NoError(mintKeeper.SetTruncationDelta(ctx, types.TruncatedInflationDeltaKey, tc.preExistingTruncationDelta))
			}

			// The mint coins are created from the mint module account exclusive of developer
			// rewards. Developer rewards are distributed from the developer vesting module account.
			// As a result, we exclude the developer proportions from calculations of mint distributions.
			nonDeveloperRewardsProportion := sdk.OneDec().Sub(tc.proportions.DeveloperRewards)

			expectedPoolIncentivesAmount := inflationAmount.Mul(tc.proportions.PoolIncentives.Quo(nonDeveloperRewardsProportion)).TruncateInt()
			expectedStakingAmount := inflationAmount.Mul(tc.proportions.Staking.Quo(nonDeveloperRewardsProportion)).TruncateInt()

			// Community pool might receive more than the expected amount since it receives all the remaining coins after
			// estimating and truncating values for pool incentives and staking.
			expectedCommunityPoolAmount := inflationAmount.Sub(expectedPoolIncentivesAmount.ToDec()).Sub(expectedStakingAmount.ToDec()).TruncateInt()
			expectedCommunityPoolAmount = expectedCommunityPoolAmount.Add(tc.expectedCommunityPoolDistributionsFromTruncations)

			// System under test.
			actualDistributed, err := mintKeeper.DistributeInflationProvisions(ctx, tc.inflationProvisions, tc.proportions)
			suite.Require().NoError(err)

			// validate that the returned amount is correct.
			suite.Require().Equal(tc.expectedDistributed, actualDistributed)

			// validate distributions to fee collector.
			feeCollectorBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(authtypes.FeeCollectorName), sdk.DefaultBondDenom).Amount
			suite.Require().Equal(expectedStakingAmount, feeCollectorBalanceAmount)

			// validate pool incentives distributions.
			actualPoolIncentivesBalance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(poolincentivestypes.ModuleName), sdk.DefaultBondDenom).Amount
			suite.Require().Equal(expectedPoolIncentivesAmount, actualPoolIncentivesBalance)

			// validate distributions to community pool.
			actualCommunityPoolBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), sdk.DefaultBondDenom).Amount
			suite.Require().Equal(expectedCommunityPoolAmount, actualCommunityPoolBalanceAmount)

			// Validate left over truncations delta
			actualLeftOverTruncationsDelta := mintKeeper.GetTruncationDelta(ctx, types.TruncatedInflationDeltaKey)
			suite.Require().Equal(tc.expectedLeftOverTruncationsDelta, actualLeftOverTruncationsDelta)

			// N.B:
			// Developer vesting module account is unaffected.
			// Mint module account balance is decreased by the distributed amount.
			// We mint the amount equal to tc.preMintCoin that increases the supply
			suite.ValidateSupplyAndMintModuleAccounts(sdk.NewInt(keeper.DeveloperVestingAmount), sdk.ZeroInt(), tc.expectedDistributed)
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
			expectedError: sdkerrors.Wrap(types.ErrInvalidAmount, "amount cannot be nil or zero"),
		},
		"zero amount": {
			blockHeight:   0,
			amount:        sdk.NewCoin("stake", sdk.NewInt(0)),
			expectedError: sdkerrors.Wrap(types.ErrInvalidAmount, "amount cannot be nil or zero"),
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

// TestDistributeToModule tests that distribution from mint module to another module helper
// function is working as expected.
func (suite *KeeperTestSuite) TestDistributeToModule() {
	const (
		denomDoesNotExist         = "denomDoesNotExist"
		moduleAccountDoesNotExist = "moduleAccountDoesNotExist"
	)

	tests := map[string]struct {
		preMintCoin sdk.Coin

		recepientModule  string
		distributionCoin sdk.DecCoin
		proportion       sdk.Dec

		expectedError bool
		expectPanic   bool
	}{
		"pre-mint == distribute - poolincentives module - full amount - success": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),

			recepientModule:  poolincentivestypes.ModuleName,
			distributionCoin: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:       sdk.NewDec(1),
		},
		"pre-mint > distribute - developer vesting module - two thirds - success": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(101)),

			recepientModule:  poolincentivestypes.ModuleName,
			distributionCoin: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:       sdk.NewDecWithPrec(2, 1).Quo(sdk.NewDecWithPrec(3, 1)),
		},
		"pre-mint < distribute (0) - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),

			recepientModule:  poolincentivestypes.ModuleName,
			distributionCoin: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:       sdk.NewDecWithPrec(2, 1).Quo(sdk.NewDecWithPrec(3, 1)),

			expectedError: true,
		},
		"denom does not exist - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),

			recepientModule:  poolincentivestypes.ModuleName,
			distributionCoin: sdk.NewDecCoin(denomDoesNotExist, sdk.NewInt(100)),
			proportion:       sdk.NewDec(1),

			expectedError: true,
		},
		"invalid module account - panic": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),

			recepientModule:  moduleAccountDoesNotExist,
			distributionCoin: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:       sdk.NewDec(1),

			expectPanic: true,
		},
		"proportion greater than 1 - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(300)),

			recepientModule:  poolincentivestypes.ModuleName,
			distributionCoin: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			proportion:       sdk.NewDec(2),

			expectedError: true,
		},
		"floor(distributionCoin) == pre-mint; round down; success": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),

			recepientModule:  poolincentivestypes.ModuleName,
			distributionCoin: sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10099, 2)),
			proportion:       sdk.NewDec(1),
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			suite.Setup()
			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() {
				mintKeeper := suite.App.MintKeeper
				bankKeeper := suite.App.BankKeeper
				accountKeeper := suite.App.AccountKeeper
				ctx := suite.Ctx

				// Setup.
				suite.MintCoins(sdk.NewCoins(tc.preMintCoin))

				expectedDistributed := tc.distributionCoin.Amount.Mul(tc.proportion).TruncateInt()
				oldMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.distributionCoin.Denom).Amount
				oldRecepientModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(tc.recepientModule), tc.distributionCoin.Denom).Amount

				// Test.
				actualDistributed, err := mintKeeper.DistributeToModule(ctx, tc.recepientModule, tc.distributionCoin, tc.proportion)

				// Assertions.
				actualMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.distributionCoin.Denom).Amount
				actualRecepientModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(tc.recepientModule), tc.distributionCoin.Denom).Amount

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

				// N.B:
				// Developer vesting module account is unaffected.
				// Mint module account balance is decreased by the distributed amount.
				// We mint the amount equal to tc.preMintCoin that increases the supply
				suite.ValidateSupplyAndMintModuleAccounts(sdk.NewInt(keeper.DeveloperVestingAmount), oldMintModuleBalanceAmount.Sub(actualDistributed), tc.preMintCoin.Amount)

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
		validLargeDistributionAmount       = sdk.NewInt(keeper.DeveloperVestingAmount)
		validLargeDistributionAmountAddOne = sdk.NewInt(keeper.DeveloperVestingAmount).Add(sdk.OneInt())

		validDevRewardsProportion = sdk.NewDecWithPrec(153, 3)

		distributionCoin = sdk.NewDecCoin(sdk.DefaultBondDenom, validLargeDistributionAmount.ToDec().Mul(validDevRewardsProportion).TruncateInt())
	)

	tests := map[string]struct {
		distribution       sdk.DecCoin
		proportion         sdk.Dec
		recepientAddresses []types.WeightedAddress

		expectedError error
		expectPanic   bool
		// See testcases with this flag set to true for details.
		allowBalanceChange bool
		// See testcases with this flag set to true for details.
		expectSameAddresses bool
		// expected distributions to community pool other than truncation.
		// truncation logic is estimated and tested separately.
		// an example of the value tracked by this field is when
		// a weighted address is empty.
		expectedCommunityPoolNonTruncationDistributions int64
	}{
		"valid case with 1 weighted address": {
			distribution: distributionCoin,
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
		},
		"valid case with 3 weighted addresses and custom large mint amount under pre mint": {
			distribution: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(939_123_546_789)),
			proportion:   sdk.NewDecWithPrec(31347, 5),
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
			distribution: distributionCoin,
			proportion:   sdk.NewDecWithPrec(123, 3),
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
			distribution: distributionCoin,
			proportion:   sdk.NewDecWithPrec(153, 3),
		},
		"valid case with 0 amount of total minted coin": {
			distribution: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.ZeroInt()),
			proportion:   sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
		},
		"invalid address in developer reward receivers - error": {
			distribution: distributionCoin,
			proportion:   sdk.NewDecWithPrec(153, 3),
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
		"distribute * proportion > developer vesting amount - error": {
			distribution: sdk.NewDecCoin(sdk.DefaultBondDenom, validLargeDistributionAmountAddOne),
			proportion:   sdk.OneDec(),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDec(1),
				},
			},
			expectedError: keeper.ErrInsufficientDevVestingBalance{ActualBalance: validLargeDistributionAmount, AttemptedDistribution: validLargeDistributionAmountAddOne.ToDec()},
		},
		"valid case with 1 empty string weighted address - distributes to community pool": {
			distribution: distributionCoin,
			proportion:   sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: keeper.EmptyWeightedAddressReceiver,
					Weight:  sdk.NewDec(1),
				},
			},

			expectedCommunityPoolNonTruncationDistributions: distributionCoin.Amount.TruncateInt64(),
		},
		"valid case with 2 addresses - empty string (distributes to community pool) and regular address (distributes to the address)": {
			distribution: distributionCoin,
			proportion:   sdk.NewDecWithPrec(153, 3),
			recepientAddresses: []types.WeightedAddress{
				{
					Address: keeper.EmptyWeightedAddressReceiver,
					Weight:  sdk.NewDecWithPrec(5, 1),
				},
				{
					Address: testAddressOne.String(),
					Weight:  sdk.NewDecWithPrec(5, 1),
				},
			},

			// expectedCommunityPoolNonTruncationDistributions = distribution * proportion * empty weighted address weight
			expectedCommunityPoolNonTruncationDistributions: distributionCoin.Amount.TruncateInt().ToDec().Mul(sdk.NewDecWithPrec(5, 1)).TruncateInt64(),
		},

		// TODO: case with decimal number (not whole number)
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			suite.Setup()
			mintKeeper := suite.App.MintKeeper
			bankKeeper := suite.App.BankKeeper
			accountKeeper := suite.App.AccountKeeper
			ctx := suite.Ctx

			// Setup.
			expectedDistributed := tc.distribution.Amount

			oldMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.distribution.Denom).Amount
			oldDeveloperVestingModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), tc.distribution.Denom).Amount
			oldCommunityPoolBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), tc.distribution.Denom).Amount
			oldSupplyOffsetAmount := bankKeeper.GetSupplyOffset(ctx, tc.distribution.Denom)
			oldDeveloperRewardsBalanceAmounts := make([]sdk.Int, len(tc.recepientAddresses))
			for i, weightedAddress := range tc.recepientAddresses {
				if weightedAddress.Address == keeper.EmptyWeightedAddressReceiver {
					continue
				}

				// No error check to be able to test invalid addresses.
				address, _ := sdk.AccAddressFromBech32(weightedAddress.Address)
				oldDeveloperRewardsBalanceAmounts[i] = bankKeeper.GetBalance(ctx, address, tc.distribution.Denom).Amount
			}

			var (
				actualDistributed sdk.Int
				err               error
			)

			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() {
				// Test.
				actualDistributed, err = mintKeeper.DistributeDeveloperRewards(ctx, tc.distribution, tc.recepientAddresses)
			})

			// Assertions.
			actualMintModuleBalance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.distribution.Denom)
			actualDeveloperVestingModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), tc.distribution.Denom).Amount
			actualCommunityPoolModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), tc.distribution.Denom).Amount
			actualSupplyOffsetAmount := bankKeeper.GetSupplyOffset(ctx, tc.distribution.Denom)

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
			suite.Require().Equal(expectedDistributed.TruncateInt(), actualDistributed)

			// Updated balances.

			// Allocate to community pool when no addresses are provided.
			if len(tc.recepientAddresses) == 0 {
				suite.Require().Equal(oldDeveloperVestingModuleBalanceAmount.Sub(expectedDistributed.TruncateInt()).Int64(), actualDeveloperVestingModuleBalanceAmount.Int64())
				suite.Require().Equal(oldCommunityPoolBalanceAmount.Add(expectedDistributed.TruncateInt()).Int64(), actualCommunityPoolModuleBalanceAmount.Int64())
				return
			}

			expectedDistributedCommunityPool := sdk.NewInt(0)
			expectedTruncationDelta := sdk.ZeroDec()

			// Suppply offset delta is equal to the dev rewards
			suite.Require().Equal(oldSupplyOffsetAmount.Add(tc.distribution.Amount.TruncateInt()).Int64(), actualSupplyOffsetAmount.Int64())

			for i, weightedAddress := range tc.recepientAddresses {
				expectedAllocation := expectedDistributed.Mul(tc.recepientAddresses[i].Weight)
				expectedAllocationTruncated := expectedAllocation.TruncateInt()
				expectedTruncationDelta = expectedTruncationDelta.Add(expectedAllocation.Sub(expectedAllocationTruncated.ToDec()))

				if weightedAddress.Address == keeper.EmptyWeightedAddressReceiver {
					expectedDistributedCommunityPool = expectedDistributedCommunityPool.Add(expectedAllocationTruncated)
					continue
				}

				address, err := sdk.AccAddressFromBech32(weightedAddress.Address)
				suite.Require().NoError(err)

				actualDeveloperRewardsBalanceAmounts := bankKeeper.GetBalance(ctx, address, tc.distribution.Denom).Amount

				// Edge case. See testcases with this flag set to true for details.
				if tc.expectSameAddresses {
					suite.Require().Equal(oldDeveloperRewardsBalanceAmounts[i].Add(expectedAllocationTruncated.Mul(sdk.NewInt(2))).Int64(), actualDeveloperRewardsBalanceAmounts.Int64())
					return
				}

				suite.Require().Equal(oldDeveloperRewardsBalanceAmounts[i].Add(expectedAllocationTruncated).Int64(), actualDeveloperRewardsBalanceAmounts.Int64())
			}

			// N.B:
			// Developer vesting module account balance decreases by the distribution amount.
			// Mint module account balance is unchanged.
			// We do not mint any amount from mint module account.
			suite.ValidateSupplyAndMintModuleAccounts(oldDeveloperVestingModuleBalanceAmount.Sub(expectedDistributed.TruncateInt()), oldMintModuleBalanceAmount, sdk.ZeroInt())

			// All truncation delta gets rounded down and set to community pool.
			expectedDistributedCommunityPool = expectedTruncationDelta.TruncateInt()

			// Supply should not change since all distributions are from the developer vesting module account.
			// suite.ValidateSupply(sdk.ZeroInt())
			suite.Require().Equal(oldCommunityPoolBalanceAmount.Add(expectedDistributedCommunityPool).Add(sdk.NewInt(tc.expectedCommunityPoolNonTruncationDistributions)).Int64(), actualCommunityPoolModuleBalanceAmount.Int64())
		})
	}
}

// TestGetInflationAmount tests that the minted amount is calculated correctly.
func (suite *KeeperTestSuite) TestGetInflationAmount() {
	tests := map[string]struct {
		denom          string
		expectedAmount sdk.Int
		preMinted      sdk.Coins
		// preVest is necessary to ensure that vesting does not
		// affect the minted amount.
		preVest sdk.DecCoin
	}{
		"zero minted": {
			denom: sdk.DefaultBondDenom,

			expectedAmount: sdk.ZeroInt(),
		},
		"pre minted equals to minted": {
			denom:     sdk.DefaultBondDenom,
			preMinted: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, baseAmount)),

			expectedAmount: baseAmount,
		},
		"vesting does not affect": {
			denom:   sdk.DefaultBondDenom,
			preVest: sdk.NewDecCoin(sdk.DefaultBondDenom, baseAmount),

			expectedAmount: sdk.ZeroInt(),
		},
		"pre minted other denom": {
			denom:     sdk.DefaultBondDenom,
			preMinted: sdk.NewCoins(sdk.NewCoin(otherDenom, baseAmount)),

			expectedAmount: sdk.ZeroInt(),
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			// Setup.
			suite.Setup()
			mintKeeper := suite.App.MintKeeper
			bankKeeper := suite.App.BankKeeper
			ctx := suite.Ctx

			if !tc.preMinted.Empty() {
				err := bankKeeper.MintCoins(ctx, types.ModuleName, tc.preMinted)
				suite.Require().NoError(err)
			}

			if tc.preVest.IsValid() {
				_, err := mintKeeper.DistributeDeveloperRewards(ctx, tc.preVest, mintKeeper.GetParams(ctx).WeightedDeveloperRewardsReceivers)
				suite.Require().NoError(err)
			}

			// System under test
			actualAmount := mintKeeper.GetInflationAmount(ctx, tc.denom)

			// Assertions.
			suite.Require().Equal(tc.expectedAmount.String(), actualAmount.String())
		})
	}
}

// TestGetDeveloperVestedAmount tests that the vested amount is calculated correctly.
func (suite *KeeperTestSuite) TestGetDeveloperVestedAmount() {
	tests := map[string]struct {
		denom string
		// preminted is necessary to ensure that minting does not
		// affect the minted amount.
		preMinted sdk.Coins
		preVest   sdk.DecCoin

		expectedAmount sdk.Int
	}{
		"zero vested": {
			denom: sdk.DefaultBondDenom,

			expectedAmount: sdk.ZeroInt(),
		},
		"minting does not affect vesting": {
			denom:     sdk.DefaultBondDenom,
			preMinted: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, baseAmount)),

			expectedAmount: sdk.ZeroInt(),
		},
		"pre-vested amount equals to vested": {
			denom:   sdk.DefaultBondDenom,
			preVest: sdk.NewDecCoin(sdk.DefaultBondDenom, baseAmount),

			expectedAmount: baseAmount,
		},
		"max vested": {
			denom:   sdk.DefaultBondDenom,
			preVest: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(keeper.DeveloperVestingAmount)),

			expectedAmount: sdk.NewInt(keeper.DeveloperVestingAmount),
		},
		"pre vested other denom": {
			denom:   sdk.DefaultBondDenom,
			preVest: sdk.NewDecCoin(otherDenom, baseAmount),

			expectedAmount: sdk.ZeroInt(),
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			// Setup.
			suite.Setup()
			mintKeeper := suite.App.MintKeeper
			bankeKeeper := suite.App.BankKeeper
			ctx := suite.Ctx

			if !tc.preMinted.Empty() {
				err := bankeKeeper.MintCoins(ctx, types.ModuleName, tc.preMinted)
				suite.Require().NoError(err)
			}

			if tc.preVest.IsValid() {
				_, err := mintKeeper.DistributeDeveloperRewards(ctx, tc.preVest, mintKeeper.GetParams(ctx).WeightedDeveloperRewardsReceivers)
				suite.Require().NoError(err)
			}

			// System under test
			actualAmount := mintKeeper.GetDeveloperVestedAmount(ctx, tc.denom)

			// Assertions.
			suite.Require().Equal(tc.expectedAmount.String(), actualAmount.String())
		})
	}
}

// TestMintInflationCoins tests that minting inflation coins is
// functions as expected by updating supply and not affecting
// the minter's expected total minted amount. The minter should
// be updated separately by taking truncated amount into account.
func (suite *KeeperTestSuite) TestMintInflationCoins() {
	const mintDenom = sdk.DefaultBondDenom
	tests := map[string]struct {
		// Inputs
		// preminted is necessary to ensure that minting does not
		// affect the minted amount.
		mintedCoins sdk.Coins

		// Expected outputs
		expectedInflationAmount         sdk.Int
		expectedAccumulatorMintedAmount sdk.Dec
		expectNoOp                      bool
	}{
		"mint 1 coin native mint denom": {
			mintedCoins:                     sdk.NewCoins(sdk.NewCoin(mintDenom, sdk.OneInt())),
			expectedInflationAmount:         sdk.OneInt(),
			expectedAccumulatorMintedAmount: sdk.OneDec(),
		},
		"mint 0 coins - no-op": {
			mintedCoins: sdk.NewCoins(),
			expectNoOp:  true,
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			// Setup.
			suite.Setup()
			mintKeeper := suite.App.MintKeeper
			ctx := suite.Ctx

			// System under test
			err := mintKeeper.MintInflationCoin(ctx, tc.mintedCoins)
			suite.Require().NoError(err)
			// Assertions.
			if tc.expectNoOp {
				return
			}

			// Validate supply
			actualInflationAmount := mintKeeper.GetInflationAmount(ctx, mintDenom)
			suite.Require().Equal(tc.expectedInflationAmount.String(), actualInflationAmount.String())

			// Validate minter's expected total minted amount.

			suite.Require().Equal(sdk.ZeroDec(), mintKeeper.GetTruncationDelta(ctx, types.TruncatedInflationDeltaKey))
		})
	}
}

// TestCalculateTruncationsDelta tests that the per-epoch truncation delta is calculate
// correctly. That includes checking that the pre-existing delta is correctly
// extracted from the store at the desired index and summed with the new delta
// from the input values.
func (suite *KeeperTestSuite) TestCalculateTruncationDelta() {
	const mintDenom = sdk.DefaultBondDenom
	tests := map[string]struct {
		preExistingStoreDelta sdk.Dec

		key               []byte
		provisions        sdk.Dec
		amountDistributed sdk.Int

		expectedTruncationDelta sdk.Dec

		expectPanic bool
	}{
		"100.6 - 100 = 0.6; no pre-existing delta; inflation delta key": {
			key:               types.TruncatedInflationDeltaKey,
			provisions:        sdk.NewDecWithPrec(1006, 1),
			amountDistributed: sdk.NewInt(100),

			expectedTruncationDelta: sdk.NewDecWithPrec(6, 1),
		},
		"100.6 - 100  + 0.3 = 0.9; pre-existing delta of 0.3; inflation delta key": {
			preExistingStoreDelta: sdk.NewDecWithPrec(3, 1),

			key:               types.TruncatedInflationDeltaKey,
			provisions:        sdk.NewDecWithPrec(1006, 1),
			amountDistributed: sdk.NewInt(100),

			expectedTruncationDelta: sdk.NewDecWithPrec(9, 1),
		},
		"100.6 - 100  + 0.5 = 1.1; pre-existing delta of 1.1; inflation delta key": {
			preExistingStoreDelta: sdk.NewDecWithPrec(5, 1),

			key:               types.TruncatedInflationDeltaKey,
			provisions:        sdk.NewDecWithPrec(1006, 1),
			amountDistributed: sdk.NewInt(100),

			expectedTruncationDelta: sdk.NewDecWithPrec(11, 1),
		},
		"82191.178 - 17808  + 33.33 = 64416508; pre-existing delta of 33.33; developer vesting delta key": {
			preExistingStoreDelta: sdk.NewDecWithPrec(3333, 2),

			key:               types.TruncatedDeveloperVestingDeltaKey,
			provisions:        sdk.NewDecWithPrec(82191178, 3),
			amountDistributed: sdk.NewInt(17808),

			expectedTruncationDelta: sdk.NewDecWithPrec(64416508, 3),
		},
		"attempt to use invalid key - panic": {
			key:               types.LastReductionEpochKey,
			provisions:        sdk.NewDecWithPrec(1006, 1),
			amountDistributed: sdk.NewInt(100),

			expectPanic: true,
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			// Setup.
			suite.Setup()
			mintKeeper := suite.App.MintKeeper
			ctx := suite.Ctx

			if !tc.preExistingStoreDelta.IsNil() {
				mintKeeper.SetTruncationDelta(ctx, tc.key, tc.preExistingStoreDelta)
			}

			var actualTruncationDelta sdk.Dec

			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() {

				// System under test
				actualTruncationDelta = mintKeeper.CalculateTotalTruncationDelta(ctx, tc.key, tc.provisions, tc.amountDistributed)
			})

			suite.Require().Equal(tc.expectedTruncationDelta, actualTruncationDelta)
		})
	}
}

// TestGetTruncationStoreKeyForModuleAccount tests that the correct key
// is returned for a given module account.
func (suite *KeeperTestSuite) TestGetTruncationStoreKeyForModuleAccount() {
	const mintDenom = sdk.DefaultBondDenom
	tests := map[string]struct {
		moduleAccountName string

		expectedKey []byte
	}{
		"developer vesting module account -> developer vesting key": {
			moduleAccountName: types.DeveloperVestingModuleAcctName,

			expectedKey: types.TruncatedDeveloperVestingDeltaKey,
		},
		"mint module account -> inflation key": {
			moduleAccountName: types.ModuleName,

			expectedKey: types.TruncatedInflationDeltaKey,
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			keeper.GetTruncationStoreKeyForModuleAccount(tc.moduleAccountName)
		})
	}
}

// TestHandleTruncationDelta tests that the truncation delta is estimated
// and distributed as determined by the method spec.
func (suite *KeeperTestSuite) TestHandleTruncationDelta() {
	tests := map[string]struct {
		// Inputs
		moduleAccountName string
		preExistingDelta  sdk.Dec
		provisions        sdk.DecCoin
		amountDistributed sdk.Int

		// Expected outputs
		expectedDistributedAmount sdk.Int
		expectedPersistedInStore  sdk.Dec
		expectErr                 bool
	}{
		"valid case: 100.6 - 100 + 0.2 = 0.8; none-distrubuted and persisted until next epoch": {
			moduleAccountName: types.DeveloperVestingModuleAcctName,
			preExistingDelta:  sdk.NewDecWithPrec(2, 1),
			provisions:        sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(1006, 1)),
			amountDistributed: sdk.NewInt(100),

			expectedDistributedAmount: sdk.ZeroInt(),
			expectedPersistedInStore:  sdk.NewDecWithPrec(8, 1),
		},
		"valid case: 100.6 - 100 + 0.4 = 1; all distributed and none persisted": {
			moduleAccountName: types.DeveloperVestingModuleAcctName,
			preExistingDelta:  sdk.NewDecWithPrec(4, 1),
			provisions:        sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(1006, 1)),
			amountDistributed: sdk.NewInt(100),

			expectedDistributedAmount: sdk.OneInt(),
			expectedPersistedInStore:  sdk.ZeroDec(),
		},
		"valid case: 100.8 - 100 + 0.4 = 1.2; 1 distributed, 0.2 persisted in store; mint module account": {
			moduleAccountName: types.ModuleName,
			preExistingDelta:  sdk.NewDecWithPrec(4, 1),
			provisions:        sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(1008, 1)),
			amountDistributed: sdk.NewInt(100),

			expectedDistributedAmount: sdk.OneInt(),
			expectedPersistedInStore:  sdk.NewDecWithPrec(2, 1),
		},
		"invalid module account name - error": {
			moduleAccountName: poolincentivestypes.ModuleName,
			preExistingDelta:  sdk.NewDecWithPrec(2, 1),
			provisions:        sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(1006, 1)),
			amountDistributed: sdk.NewInt(100),

			expectErr: true,
		},
		"provisions are less than amount distributed - error": {
			moduleAccountName: types.ModuleName,
			preExistingDelta:  sdk.NewDecWithPrec(2, 1),
			provisions:        sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(1006, 1)),
			amountDistributed: sdk.NewInt(101),

			expectErr: true,
		},
		"attempted to distribute more than developer vesting module account balance - error": {
			moduleAccountName: types.DeveloperVestingModuleAcctName,
			preExistingDelta:  sdk.NewDecFromInt(sdk.NewInt(keeper.DeveloperVestingAmount)).Add(sdk.OneDec()),
			provisions:        sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(1006, 1)),
			amountDistributed: sdk.NewInt(100),

			expectErr: true,
		},
	}
	for name, tc := range tests {
		suite.Run(name, func() {
			// Setup.
			suite.Setup()
			mintKeeper := suite.App.MintKeeper
			ctx := suite.Ctx

			storeKey := keeper.GetTruncationStoreKeyForModuleAccount(tc.moduleAccountName)

			if !tc.preExistingDelta.IsNil() {
				mintKeeper.SetTruncationDelta(ctx, storeKey, tc.preExistingDelta)
			}

			// System under test
			actualAmount, err := mintKeeper.HandleTruncationDelta(ctx, tc.moduleAccountName, tc.provisions, tc.amountDistributed)

			// Assertions.
			if tc.expectErr {
				suite.Require().Error(err)
				return
			}

			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedDistributedAmount.String(), actualAmount.String())
			suite.Require().Equal(tc.expectedPersistedInStore, mintKeeper.GetTruncationDelta(ctx, storeKey))
		})
	}
}
