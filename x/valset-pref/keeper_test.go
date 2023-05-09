package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v15/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/valset-pref/types"

	valPref "github.com/osmosis-labs/osmosis/v15/x/valset-pref"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

// PrepareDelegateToValidatorSet generates 4 validators for the valsetpref.
// We self assign weights and round up to 2 decimal places in validateBasic.
func (suite *KeeperTestSuite) PrepareDelegateToValidatorSet() []types.ValidatorPreference {
	valAddrs := suite.SetupMultipleValidators(4)
	valPreferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         sdk.NewDecWithPrec(2, 1), // 0.2
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         sdk.NewDecWithPrec(332, 3), // 0.33
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         sdk.NewDecWithPrec(12, 2), // 0.12
		},
		{
			ValOperAddress: valAddrs[3],
			Weight:         sdk.NewDecWithPrec(348, 3), // 0.35
		},
	}

	return valPreferences
}

func (suite *KeeperTestSuite) GetDelegationRewards(ctx sdk.Context, valAddrStr string, delegator sdk.AccAddress) (sdk.DecCoins, stakingtypes.Validator) {
	valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
	suite.Require().NoError(err)

	validator, found := suite.App.StakingKeeper.GetValidator(ctx, valAddr)
	suite.Require().True(found)

	endingPeriod := suite.App.DistrKeeper.IncrementValidatorPeriod(ctx, validator)

	delegation, found := suite.App.StakingKeeper.GetDelegation(ctx, delegator, valAddr)
	suite.Require().True(found)

	rewards := suite.App.DistrKeeper.CalculateDelegationRewards(ctx, validator, delegation, endingPeriod)

	return rewards, validator
}

func (suite *KeeperTestSuite) SetupDelegationReward(delegator sdk.AccAddress, preferences []types.ValidatorPreference, existingValAddrStr string, setValSetDel, setExistingdel bool) {
	var ctx sdk.Context
	// incrementing the blockheight by 1 for reward
	ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeight() + 1)

	if setValSetDel {
		// only necessary if there are tokens delegated
		for _, val := range preferences {
			suite.AllocateRewards(ctx, delegator, val.ValOperAddress)
		}
	}

	if setExistingdel {
		suite.AllocateRewards(ctx, delegator, existingValAddrStr)
	}
}

func (suite *KeeperTestSuite) AllocateRewards(ctx sdk.Context, delegator sdk.AccAddress, valAddrStr string) {
	// check that there is enough reward to withdraw
	_, validator := suite.GetDelegationRewards(ctx, valAddrStr, delegator)

	// allocate some rewards
	tokens := sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 10))
	suite.App.DistrKeeper.AllocateTokensToValidator(ctx, validator, tokens)

	rewardsAfterAllocation, _ := suite.GetDelegationRewards(ctx, valAddrStr, delegator)
	suite.Require().NotNil(rewardsAfterAllocation)
	suite.Require().NotZero(rewardsAfterAllocation[0].Amount)
}

// PrepareExistingDelegations sets up existing delegation by creating a certain number of validators and delegating tokenAmt to them.
func (suite *KeeperTestSuite) PrepareExistingDelegations(ctx sdk.Context, valAddrs []string, delegator sdk.AccAddress, tokenAmt sdk.Int) error {
	for i := 0; i < len(valAddrs); i++ {
		valAddr, err := sdk.ValAddressFromBech32(valAddrs[i])
		if err != nil {
			return fmt.Errorf("validator address not formatted")
		}

		validator, found := suite.App.StakingKeeper.GetValidator(ctx, valAddr)
		if !found {
			return fmt.Errorf("validator not found %s", validator)
		}

		// Delegate the unbonded tokens
		_, err = suite.App.StakingKeeper.Delegate(ctx, delegator, tokenAmt, stakingtypes.Unbonded, validator, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (suite *KeeperTestSuite) TestGetDelegationPreference() {
	suite.SetupTest()

	// prepare existing delegations validators
	valAddrs := suite.SetupMultipleValidators(3)

	// prepare validators to delegate to valset
	preferences := suite.PrepareDelegateToValidatorSet()

	tests := []struct {
		name                   string
		setValSet              bool
		delegator              sdk.AccAddress
		setExistingDelegations bool
		expectPass             bool
	}{
		{
			name:       "ValSet exist, existing delegations does not exist",
			delegator:  sdk.AccAddress([]byte("addr1---------------")),
			setValSet:  true,
			expectPass: true,
		},
		{
			name:                   "ValSet exists, existing delegations exist",
			delegator:              sdk.AccAddress([]byte("addr2---------------")),
			setValSet:              true,
			setExistingDelegations: true,
			expectPass:             true,
		},
		{
			name:                   "ValSet does not exist, but existing delegations exist",
			delegator:              sdk.AccAddress([]byte("addr3---------------")),
			setExistingDelegations: true,
			expectPass:             true,
		},
		{
			name:       "ValSet does not exist, no existing delegations",
			delegator:  sdk.AccAddress([]byte("addr4---------------")),
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			amountToFund := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)} // 100 osmo

			suite.FundAcc(test.delegator, amountToFund)

			if test.setValSet {
				_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
				suite.Require().NoError(err)
			}

			if test.setExistingDelegations {
				err := suite.PrepareExistingDelegations(suite.Ctx, valAddrs, test.delegator, sdk.NewInt(10_000_000))
				suite.Require().NoError(err)
			}

			_, err := suite.App.ValidatorSetPreferenceKeeper.GetDelegationPreferences(suite.Ctx, test.delegator.String())
			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) SetupValidatorsAndDelegations() ([]string, []types.ValidatorPreference, sdk.Coins) {
	// prepare existing delegations validators
	valAddrs := suite.SetupMultipleValidators(3)

	// prepare validators to delegate to valset
	preferences := suite.PrepareDelegateToValidatorSet()

	amountToFund := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)}

	return valAddrs, preferences, amountToFund
}

func (suite *KeeperTestSuite) SetupLocks(delegator sdk.AccAddress) []lockuptypes.PeriodLock {
	// create a pool with uosmo
	locks := []lockuptypes.PeriodLock{}
	// Setup lock
	coinsToLock := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000)}
	osmoToLock := sdk.Coins{sdk.NewInt64Coin(appParams.BaseCoinUnit, 10_000_000)}
	multipleCoinsToLock := sdk.Coins{coinsToLock[0], osmoToLock[0]}
	suite.FundAcc(delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000), sdk.NewInt64Coin(appParams.BaseCoinUnit, 100_000_000)})

	// lock with osmo
	twoWeekDuration, err := time.ParseDuration("336h")
	suite.Require().NoError(err)
	workingLock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, delegator, osmoToLock, twoWeekDuration)
	suite.Require().NoError(err)

	locks = append(locks, workingLock)

	// locking with stake denom instead of osmo denom
	stakeDenomLock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, delegator, coinsToLock, twoWeekDuration)
	suite.Require().NoError(err)

	locks = append(locks, stakeDenomLock)

	// lock case where lock owner != delegator
	suite.FundAcc(sdk.AccAddress([]byte("addr5---------------")), osmoToLock)
	lockWithDifferentOwner, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, sdk.AccAddress([]byte("addr5---------------")), osmoToLock, twoWeekDuration)
	suite.Require().NoError(err)

	locks = append(locks, lockWithDifferentOwner)

	// lock case where the duration != <= 2 weeks
	morethanTwoWeekDuration, err := time.ParseDuration("337h")
	suite.Require().NoError(err)
	maxDurationLock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, delegator, osmoToLock, morethanTwoWeekDuration)
	suite.Require().NoError(err)

	locks = append(locks, maxDurationLock)

	// unbonding locks
	unbondingLocks, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, delegator, osmoToLock, twoWeekDuration)
	suite.Require().NoError(err)

	_, err = suite.App.LockupKeeper.BeginUnlock(suite.Ctx, unbondingLocks.ID, nil)
	suite.Require().NoError(err)

	locks = append(locks, unbondingLocks)

	// synthetic locks
	syntheticLocks, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, delegator, osmoToLock, twoWeekDuration)
	suite.Require().NoError(err)

	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, syntheticLocks.ID, "uosmo", time.Minute, true)
	suite.Require().NoError(err)

	locks = append(locks, syntheticLocks)

	// multiple asset lock
	multiassetLock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, delegator, multipleCoinsToLock, twoWeekDuration)
	suite.Require().NoError(err)

	locks = append(locks, multiassetLock)

	return locks
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
