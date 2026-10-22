package keeper_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v32/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v32/app/params"
	"github.com/osmosis-labs/osmosis/v32/x/mint/keeper"
	"github.com/osmosis-labs/osmosis/v32/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v32/x/pool-incentives/types"
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

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.queryClient = types.NewQueryClient(s.QueryHelper)
	params := s.App.MintKeeper.GetParams(s.Ctx)
	params.ReductionPeriodInEpochs = 10
	s.App.MintKeeper.SetParams(s.Ctx, params)
}

// setupDeveloperVestingModuleAccountTest sets up test cases that utilize developer vesting
// module account logic. It reverts some default logic added by s.Setup()
// Specifically, it removes the developer vesting module account
// from account keeper if isDeveloperModuleAccountCreated is true.
// Additionally, it initializes suite's Ctx with blockHeight
func (s *KeeperTestSuite) setupDeveloperVestingModuleAccountTest(blockHeight int64, isDeveloperModuleAccountCreated bool) {
	s.Setup()
	// Reset height to the desired value since test suite setup initialized
	// it to 1.
	bankKeeper := s.App.BankKeeper
	accountKeeper := s.App.AccountKeeper

	s.Ctx = s.Ctx.WithBlockHeader(tmproto.Header{Height: blockHeight})

	if !isDeveloperModuleAccountCreated {
		// Remove the developer vesting account since suite setup creates and initializes it.
		// This environment w/o the developer vesting account configured is necessary for
		// testing edge cases of multiple tests.
		developerVestingAccount := accountKeeper.GetAccount(s.Ctx, accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
		accountKeeper.RemoveAccount(s.Ctx, developerVestingAccount)
		err := bankKeeper.BurnCoins(s.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(keeper.DeveloperVestingAmount))))
		s.Require().Error(err)

		// If developer module account is created, the s.Setup() also sets the offset,
		// therefore, we should reset it to 0 to set up the environment truly w/o the module account.
		supplyOffset := bankKeeper.GetSupplyOffset(s.Ctx, sdk.DefaultBondDenom)
		bankKeeper.AddSupplyOffset(s.Ctx, sdk.DefaultBondDenom, supplyOffset.Mul(osmomath.NewInt(-1)))
		s.Require().Equal(osmomath.ZeroInt(), bankKeeper.GetSupplyOffset(s.Ctx, sdk.DefaultBondDenom))
	}
}

// TestGetProportions tests that mint allocations are computed as expected.
func (s *KeeperTestSuite) TestGetProportions() {
	complexRatioDec := osmomath.NewDecWithPrec(131, 3).Quo(osmomath.NewDecWithPrec(273, 3))

	tests := []struct {
		name          string
		ratio         osmomath.Dec
		expectedCoin  sdk.Coin
		expectedError error
		mintedCoin    sdk.Coin
	}{
		{
			name:         "0 * 0.2 = 0",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(0)),
			ratio:        osmomath.NewDecWithPrec(2, 1),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(0)),
		},
		{
			name:         "100000 * 0.2 = 20000",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100000)),
			ratio:        osmomath.NewDecWithPrec(2, 1),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100000).Quo(osmomath.NewInt(5))),
		},
		{
			name:         "123456 * 2/3 = 82304",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(123456)),
			ratio:        osmomath.NewDecWithPrec(2, 1).Quo(osmomath.NewDecWithPrec(3, 1)),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(82304)),
		},
		{
			name:       "54617981 * .131/.273osmomath.NewInt(.62",
			mintedCoin: sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(54617981)),
			ratio:      complexRatioDec, // .131/.273
			// TODO: Should not be truncated. Remove truncation after rounding errors are addressed and resolved.
			// Ref: https://github.com/osmosis-osmomath.NewInt(s/issues/1917
			expectedCoin: sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(54617981).ToLegacyDec().Mul(complexRatioDec).TruncateInt()),
		},
		{
			name:         "1 * 1 = 1",
			mintedCoin:   sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1)),
			ratio:        osmomath.NewDec(1),
			expectedCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(1)),
		},
		{
			name:       "1 * 1.01 - error, ratio must be <= 1",
			mintedCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(0)),
			ratio:      osmomath.NewDecWithPrec(101, 2),

			expectedError: keeper.ErrInvalidRatio{ActualRatio: osmomath.NewDecWithPrec(101, 2)},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			coin, err := keeper.GetProportions(tc.mintedCoin, tc.ratio)

			if tc.expectedError != nil {
				s.Require().Equal(tc.expectedError, err)
				s.Require().Equal(sdk.Coin{}, coin)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tc.expectedCoin, coin)
		})
	}
}

// TestDistributeMintedCoin tests that the minted coin is split between staking, pool incentives
// and the community pool, and that the developer vesting module account is not touched, even when
// the deprecated developer rewards receivers are set.
func (s *KeeperTestSuite) TestDistributeMintedCoin() {
	const (
		mintAmount = 10000
	)

	params := types.DefaultParams()

	tests := []struct {
		name              string
		weightedAddresses []types.WeightedAddress
		mintCoin          sdk.Coin
	}{
		{
			name:              "no dev reward addresses",
			weightedAddresses: nil,
			mintCoin:          sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(mintAmount)),
		},
		{
			name: "deprecated dev reward addresses are ignored",
			weightedAddresses: []types.WeightedAddress{
				{
					Address: testAddressThree.String(),
					Weight:  osmomath.NewDecWithPrec(6, 1),
				},
				{
					Address: testAddressFour.String(),
					Weight:  osmomath.NewDecWithPrec(4, 1),
				},
			},
			mintCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(mintAmount)),
		},
		{
			name:              "rounding remainder goes to community pool",
			weightedAddresses: nil,
			mintCoin:          sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(mintAmount+7)),
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Setup()

			ctx := s.Ctx

			bankKeeper := s.App.BankKeeper
			accountKeeper := s.App.AccountKeeper

			mintKeeper := s.App.MintKeeper
			// We reset the hooks with a mock to simplify the assertions
			// about the results of the call to DistributeMintedCoin.
			// The goal is to assert that AfterDistributeMintedCoin
			// is called once.
			mintKeeper.SetMintHooksUnsafe(&mintHooksMock{})

			params.WeightedDeveloperRewardsReceivers = tc.weightedAddresses
			mintKeeper.SetParams(ctx, params)

			mintAmount := tc.mintCoin.Amount.ToLegacyDec()
			expectedStakingAmount := mintAmount.Mul(params.DistributionProportions.Staking).TruncateInt()
			expectedPoolIncentivesAmount := mintAmount.Mul(params.DistributionProportions.PoolIncentives).TruncateInt()
			expectedCommunityPoolAmount := tc.mintCoin.Amount.Sub(expectedStakingAmount).Sub(expectedPoolIncentivesAmount)

			devVestingAddr := accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
			devVestingBalanceBefore := bankKeeper.GetBalance(ctx, devVestingAddr, sdk.DefaultBondDenom)
			supplyOffsetBefore := bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom)

			// mints coins so supply exists on chain
			s.MintCoins(sdk.NewCoins(tc.mintCoin))
			supplyBefore := bankKeeper.GetSupply(ctx, sdk.DefaultBondDenom)

			// System under test.
			err := mintKeeper.DistributeMintedCoin(ctx, tc.mintCoin)
			s.Require().NoError(err)

			// validate that AfterDistributeMintedCoin hook was called once.
			hooks, ok := mintKeeper.GetMintHooksUnsafe().(*mintHooksMock)
			s.Require().True(ok, "unexpected type of mint hooks")
			s.Require().Equal(1, hooks.hookCallCount)

			// validate distributions to fee collector.
			feeCollectorBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(authtypes.FeeCollectorName), sdk.DefaultBondDenom).Amount
			s.Require().Equal(expectedStakingAmount, feeCollectorBalanceAmount)

			// validate pool incentives distributions.
			actualPoolIncentivesBalance := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(poolincentivestypes.ModuleName), sdk.DefaultBondDenom).Amount
			s.Require().Equal(expectedPoolIncentivesAmount, actualPoolIncentivesBalance)

			// validate distributions to community pool.
			actualCommunityPoolBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(distributiontypes.ModuleName), sdk.DefaultBondDenom).Amount
			s.Require().Equal(expectedCommunityPoolAmount, actualCommunityPoolBalanceAmount)

			// nothing is left in the mint module and nothing is burned.
			s.Require().True(bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom).IsZero())
			s.Require().Equal(supplyBefore, bankKeeper.GetSupply(ctx, sdk.DefaultBondDenom))

			// the developer vesting module account and the supply offset are not touched.
			s.Require().Equal(devVestingBalanceBefore, bankKeeper.GetBalance(ctx, devVestingAddr, sdk.DefaultBondDenom))
			s.Require().Equal(supplyOffsetBefore, bankKeeper.GetSupplyOffset(ctx, sdk.DefaultBondDenom))

			// deprecated developer rewards receivers get nothing.
			for _, weightedAddress := range tc.weightedAddresses {
				devRewardsReceiver := sdk.MustAccAddressFromBech32(weightedAddress.GetAddress())
				s.Require().True(bankKeeper.GetBalance(ctx, devRewardsReceiver, sdk.DefaultBondDenom).IsZero())
			}
		})
	}
}

func (s *KeeperTestSuite) TestCreateDeveloperVestingModuleAccount() {
	testcases := map[string]struct {
		blockHeight                     int64
		amount                          sdk.Coin
		isDeveloperModuleAccountCreated bool

		expectedError error
	}{
		"valid call": {
			blockHeight: 0,
			amount:      sdk.NewCoin("stake", osmomath.NewInt(keeper.DeveloperVestingAmount)),
		},
		"nil amount": {
			blockHeight:   0,
			expectedError: errorsmod.Wrap(types.ErrAmountNilOrZero, "amount cannot be nil or zero"),
		},
		"zero amount": {
			blockHeight:   0,
			amount:        sdk.NewCoin("stake", osmomath.NewInt(0)),
			expectedError: errorsmod.Wrap(types.ErrAmountNilOrZero, "amount cannot be nil or zero"),
		},
		"module account is already created": {
			blockHeight:                     0,
			amount:                          sdk.NewCoin("stake", osmomath.NewInt(keeper.DeveloperVestingAmount)),
			isDeveloperModuleAccountCreated: true,
			expectedError:                   errorsmod.Wrapf(types.ErrModuleAccountAlreadyExist, "%s vesting module account already exist", types.DeveloperVestingModuleAcctName),
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.setupDeveloperVestingModuleAccountTest(tc.blockHeight, tc.isDeveloperModuleAccountCreated)
			mintKeeper := s.App.MintKeeper

			// Test
			actualError := mintKeeper.CreateDeveloperVestingModuleAccount(s.Ctx, tc.amount)

			if tc.expectedError != nil {
				s.Require().Error(actualError)
				s.Require().ErrorIs(actualError, tc.expectedError)
				return
			}
			s.Require().NoError(actualError)
		})
	}
}

// TestDistributeToModule tests that distribution from mint module to another module helper
// function is working as expected.
func (s *KeeperTestSuite) TestDistributeToModule() {
	const (
		denomDoesNotExist         = "denomDoesNotExist"
		moduleAccountDoesNotExist = "moduleAccountDoesNotExist"
	)

	tests := map[string]struct {
		preMintCoin sdk.Coin

		recepientModule string
		mintedCoin      sdk.Coin
		proportion      osmomath.Dec

		expectedError bool
		expectPanic   bool
	}{
		"pre-mint == distribute - poolincentives module osmomath.NewInt(nt - success": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),
			proportion:      osmomath.NewDec(1),
		},
		"pre-mint > distribute - developer vesting modulosmomath.NewInt(rds - success": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(101)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),
			proportion:      osmomath.NewDecWithPrec(2, 1).Quo(osmomath.NewDecWithPrec(3, 1)),
		},
		"pre-mint < distribute (0) - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(0)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),
			proportion:      osmomath.NewDecWithPrec(2, 1).Quo(osmomath.NewDecWithPrec(3, 1)),

			expectedError: true,
		},
		"denom does not exist - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(denomDoesNotExist, osmomath.NewInt(100)),
			proportion:      osmomath.NewDec(1),

			expectedError: true,
		},
		"invalid module account -panic": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),

			recepientModule: moduleAccountDoesNotExist,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),
			proportion:      osmomath.NewDec(1),

			expectPanic: true,
		},
		"proportion greater than 1 - error": {
			preMintCoin: sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(300)),

			recepientModule: poolincentivestypes.ModuleName,
			mintedCoin:      sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100)),
			proportion:      osmomath.NewDec(2),

			expectedError: true,
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.Setup()
			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() {
				mintKeeper := s.App.MintKeeper
				bankKeeper := s.App.BankKeeper
				accountKeeper := s.App.AccountKeeper
				ctx := s.Ctx

				// Setup.
				s.MintCoins(sdk.NewCoins(tc.preMintCoin))

				// TODO: Should not be truncated. Remove truncation after rounding errors are addressed and resolved.
				// Ref: https://github.com/osmosis-labs/osmosis/issues/1917
				expectedDistributed := tc.mintedCoin.Amount.ToLegacyDec().Mul(tc.proportion).TruncateInt()
				oldMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.mintedCoin.Denom).Amount
				oldRecepientModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(tc.recepientModule), tc.mintedCoin.Denom).Amount

				// Test.
				actualDistributed, err := mintKeeper.DistributeToModule(ctx, tc.recepientModule, tc.mintedCoin, tc.proportion)

				// Assertions.
				actualMintModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(types.ModuleName), tc.mintedCoin.Denom).Amount
				actualRecepientModuleBalanceAmount := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(tc.recepientModule), tc.mintedCoin.Denom).Amount

				if tc.expectedError {
					s.Require().Error(err)
					s.Require().Equal(actualDistributed, osmomath.Int{})
					// Old balances should not change.
					s.Require().Equal(oldMintModuleBalanceAmount.Int64(), actualMintModuleBalanceAmount.Int64())
					s.Require().Equal(oldRecepientModuleBalanceAmount.Int64(), actualRecepientModuleBalanceAmount.Int64())
					return
				}

				s.Require().NoError(err)
				s.Require().Equal(expectedDistributed, actualDistributed)

				// Updated balances.
				s.Require().Equal(oldMintModuleBalanceAmount.Sub(actualDistributed).Int64(), actualMintModuleBalanceAmount.Int64())
				s.Require().Equal(oldRecepientModuleBalanceAmount.Add(actualDistributed).Int64(), actualRecepientModuleBalanceAmount.Int64())
			})
		})
	}
}

func (s *KeeperTestSuite) TestGetInflation() {
	s.Setup()

	// Mainnet values after the v33 upgrade removed the developer rewards share:
	// epoch_provisions: 121765601217.656011811263318112 * 0.75 = 91324200913.242008858447488584
	// community_pool proportion: 0.67 / 0.75 = 0.893333333333333333
	// uOSMO supply: 744387309021205

	// Set epoch provisions in the minter
	epochProvisions := osmomath.MustNewDecFromStr("91324200913.242008858447488584")
	minter := types.Minter{
		EpochProvisions: epochProvisions,
	}
	s.App.MintKeeper.SetMinter(s.Ctx, minter)

	// Set up params with distribution proportions that match the real chain
	params := types.DefaultParams()
	params.DistributionProportions.Staking = osmomath.MustNewDecFromStr("0.106666666666666667")
	params.DistributionProportions.PoolIncentives = osmomath.MustNewDecFromStr("0.000000000000000000")
	params.DistributionProportions.DeveloperRewards = osmomath.ZeroDec()
	params.DistributionProportions.CommunityPool = osmomath.MustNewDecFromStr("0.893333333333333333")
	params.MintDenom = "uosmo"
	s.App.MintKeeper.SetParams(s.Ctx, params)

	// Mock the bank keeper to return the total supply
	totalSupply := osmomath.NewInt(744387309021205)
	mintCoins := sdk.NewCoins(sdk.NewCoin("uosmo", totalSupply))
	err := s.App.BankKeeper.MintCoins(s.Ctx, types.ModuleName, mintCoins)
	s.Require().NoError(err)

	// Get the inflation from the keeper
	inflation, err := s.App.MintKeeper.GetInflation(s.Ctx)
	s.Require().NoError(err)

	// ((Epoch provisions * (1 - Distribution:Community Pool)) * 365) / uOSMO Supply
	// = ((91324200913.242008858447488584 * (1 - 0.893333333333333333)) * 365) / 744387309021205 = 0.004776485993871599
	s.Require().Equal(osmomath.MustNewDecFromStr("0.004776485993871599"), inflation)
}
