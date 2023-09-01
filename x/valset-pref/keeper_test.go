package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v19/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v19/x/valset-pref/types"

	valPref "github.com/osmosis-labs/osmosis/v19/x/valset-pref"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

// PrepareDelegateToValidatorSet generates 4 validators for the valsetpref.
// We self assign weights and round up to 2 decimal places in validateBasic.
func (s *KeeperTestSuite) PrepareDelegateToValidatorSet() []types.ValidatorPreference {
	valAddrs := s.SetupMultipleValidators(4)
	valPreferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         osmomath.NewDecWithPrec(2, 1), // 0.2
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         osmomath.NewDecWithPrec(332, 3), // 0.33
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         osmomath.NewDecWithPrec(12, 2), // 0.12
		},
		{
			ValOperAddress: valAddrs[3],
			Weight:         osmomath.NewDecWithPrec(348, 3), // 0.35
		},
	}

	return valPreferences
}

func (s *KeeperTestSuite) GetDelegationRewards(ctx sdk.Context, valAddrStr string, delegator sdk.AccAddress) (sdk.DecCoins, stakingtypes.Validator) {
	valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
	s.Require().NoError(err)

	validator, found := s.App.StakingKeeper.GetValidator(ctx, valAddr)
	s.Require().True(found)

	endingPeriod := s.App.DistrKeeper.IncrementValidatorPeriod(ctx, validator)

	delegation, found := s.App.StakingKeeper.GetDelegation(ctx, delegator, valAddr)
	s.Require().True(found)

	rewards := s.App.DistrKeeper.CalculateDelegationRewards(ctx, validator, delegation, endingPeriod)

	return rewards, validator
}

func (s *KeeperTestSuite) SetupDelegationReward(delegator sdk.AccAddress, preferences []types.ValidatorPreference, existingValAddrStr string, setValSetDel, setExistingdel bool) {
	var ctx sdk.Context
	// incrementing the blockheight by 1 for reward
	ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)

	if setValSetDel {
		// only necessary if there are tokens delegated
		for _, val := range preferences {
			s.AllocateRewards(ctx, delegator, val.ValOperAddress)
		}
	}

	if setExistingdel {
		s.AllocateRewards(ctx, delegator, existingValAddrStr)
	}
}

// AllocateRewards allocates rewards to a delegator
func (s *KeeperTestSuite) AllocateRewards(ctx sdk.Context, delegator sdk.AccAddress, valAddrStr string) {
	// check that there is enough reward to withdraw
	_, validator := s.GetDelegationRewards(ctx, valAddrStr, delegator)

	// allocate some rewards
	tokens := sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 10))
	s.App.DistrKeeper.AllocateTokensToValidator(ctx, validator, tokens)

	rewardsAfterAllocation, _ := s.GetDelegationRewards(ctx, valAddrStr, delegator)
	s.Require().NotNil(rewardsAfterAllocation)
	s.Require().NotZero(rewardsAfterAllocation[0].Amount)
}

// Pres *KeeperTestSuites sets up existing delegation by creating a certain number of validators and delegating tokenAmt to them.
func (s *KeeperTestSuite) PrepareExistingDelegations(ctx sdk.Context, valAddrs []string, delegator sdk.AccAddress, tokenAmt osmomath.Int) error {
	for i := 0; i < len(valAddrs); i++ {
		valAddr, err := sdk.ValAddressFromBech32(valAddrs[i])
		if err != nil {
			return fmt.Errorf("validator address not formatted")
		}

		validator, found := s.App.StakingKeeper.GetValidator(ctx, valAddr)
		if !found {
			return fmt.Errorf("validator not found %s", validator)
		}

		// Delegate the unbonded tokens
		_, err = s.App.StakingKeeper.Delegate(ctx, delegator, tokenAmt, stakingtypes.Unbonded, validator, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// TestGetDelegationPreference tests the GetDelegationPreference function
func (s *KeeperTestSuite) TestGetDelegationPreference() {
	s.SetupTest()

	// prepare existing delegations validators
	valAddrs := s.SetupMultipleValidators(3)

	// prepare validators to delegate to valset
	preferences := s.PrepareDelegateToValidatorSet()

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
		s.Run(test.name, func() {
			msgServer := valPref.NewMsgServerImpl(s.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(s.Ctx)

			amountToFund := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)} // 100 osmo

			s.FundAcc(test.delegator, amountToFund)

			if test.setValSet {
				_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
				s.Require().NoError(err)
			}

			if test.setExistingDelegations {
				err := s.PrepareExistingDelegations(s.Ctx, valAddrs, test.delegator, osmomath.NewInt(10_000_000))
				s.Require().NoError(err)
			}

			_, err := s.App.ValidatorSetPreferenceKeeper.GetDelegationPreferences(s.Ctx, test.delegator.String())
			if test.expectPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// SetupValidatorsAndDelegations sets up existing delegation by creating a certain number of validators and delegating tokenAmt to them.
func (s *KeeperTestSuite) SetupValidatorsAndDelegations() ([]string, []types.ValidatorPreference, sdk.Coins) {
	// prepare existing delegations validators
	valAddrs := s.SetupMultipleValidators(3)

	// prepare validators to delegate to valset
	preferences := s.PrepareDelegateToValidatorSet()

	amountToFund := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)}

	return valAddrs, preferences, amountToFund
}

// SetupLocks sets up locks for a delegator
func (s *KeeperTestSuite) SetupLocks(delegator sdk.AccAddress) []lockuptypes.PeriodLock {
	// create a pool with uosmo
	locks := []lockuptypes.PeriodLock{}
	// Setup lock
	coinsToLock := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000)}
	osmoToLock := sdk.Coins{sdk.NewInt64Coin(appParams.BaseCoinUnit, 10_000_000)}
	multipleCoinsToLock := sdk.Coins{coinsToLock[0], osmoToLock[0]}
	s.FundAcc(delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000), sdk.NewInt64Coin(appParams.BaseCoinUnit, 100_000_000)})

	// lock with osmo
	twoWeekDuration, err := time.ParseDuration("336h")
	s.Require().NoError(err)
	workingLock, err := s.App.LockupKeeper.CreateLock(s.Ctx, delegator, osmoToLock, twoWeekDuration)
	s.Require().NoError(err)

	locks = append(locks, workingLock)

	// locking with stake denom instead of osmo denom
	stakeDenomLock, err := s.App.LockupKeeper.CreateLock(s.Ctx, delegator, coinsToLock, twoWeekDuration)
	s.Require().NoError(err)

	locks = append(locks, stakeDenomLock)

	// lock case where lock owner != delegator
	s.FundAcc(sdk.AccAddress([]byte("addr5---------------")), osmoToLock)
	lockWithDifferentOwner, err := s.App.LockupKeeper.CreateLock(s.Ctx, sdk.AccAddress([]byte("addr5---------------")), osmoToLock, twoWeekDuration)
	s.Require().NoError(err)

	locks = append(locks, lockWithDifferentOwner)

	// lock case where the duration != <= 2 weeks
	morethanTwoWeekDuration, err := time.ParseDuration("337h")
	s.Require().NoError(err)
	maxDurationLock, err := s.App.LockupKeeper.CreateLock(s.Ctx, delegator, osmoToLock, morethanTwoWeekDuration)
	s.Require().NoError(err)

	locks = append(locks, maxDurationLock)

	// unbonding locks
	unbondingLocks, err := s.App.LockupKeeper.CreateLock(s.Ctx, delegator, osmoToLock, twoWeekDuration)
	s.Require().NoError(err)

	_, err = s.App.LockupKeeper.BeginUnlock(s.Ctx, unbondingLocks.ID, nil)
	s.Require().NoError(err)

	locks = append(locks, unbondingLocks)

	// synthetic locks
	syntheticLocks, err := s.App.LockupKeeper.CreateLock(s.Ctx, delegator, osmoToLock, twoWeekDuration)
	s.Require().NoError(err)

	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, syntheticLocks.ID, "uosmo", time.Minute, true)
	s.Require().NoError(err)

	locks = append(locks, syntheticLocks)

	// multiple asset lock
	multiassetLock, err := s.App.LockupKeeper.CreateLock(s.Ctx, delegator, multipleCoinsToLock, twoWeekDuration)
	s.Require().NoError(err)

	locks = append(locks, multiassetLock)

	return locks
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
