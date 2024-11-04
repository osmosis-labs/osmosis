package keeper_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"

	valPref "github.com/osmosis-labs/osmosis/v27/x/valset-pref"
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

	validator, err := s.App.StakingKeeper.GetValidator(ctx, valAddr)
	s.Require().NoError(err)

	endingPeriod, err := s.App.DistrKeeper.IncrementValidatorPeriod(ctx, validator)
	s.Require().NoError(err)

	delegation, err := s.App.StakingKeeper.GetDelegation(ctx, delegator, valAddr)
	s.Require().NoError(err)

	rewards, err := s.App.DistrKeeper.CalculateDelegationRewards(ctx, validator, delegation, endingPeriod)
	s.Require().NoError(err)

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

// PrepareExistingDelegations sets up existing delegation by creating a certain number of validators and delegating tokenAmt to them.
func (s *KeeperTestSuite) PrepareExistingDelegations(ctx sdk.Context, valAddrs []string, delegator sdk.AccAddress, tokenAmt osmomath.Int) error {
	for i := 0; i < len(valAddrs); i++ {
		valAddr, err := sdk.ValAddressFromBech32(valAddrs[i])
		if err != nil {
			return errors.New("validator address not formatted")
		}

		validator, err := s.App.StakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			return fmt.Errorf("validator not found %s", validator.String())
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
			c := s.Ctx

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

func (s *KeeperTestSuite) TestGetValSetPreferencesWithDelegations() {
	s.SetupTest()

	defaultDelegateAmt := osmomath.NewInt(1000)

	tests := []struct {
		name          string
		setPref       bool
		setDelegation bool
		expectedErr   bool
	}{
		{
			name:    "valset preference exists, no existing delegation",
			setPref: true,
		},
		{
			name:        "no valset preference, no existing delegation",
			expectedErr: true,
		},
		{
			name:          "valset preference exists, existing delegation exists",
			setPref:       true,
			setDelegation: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Setup()

			delegator := s.TestAccs[0]

			// prepare existing delegations validators
			valAddrs := s.SetupMultipleValidators(3)
			defaultValPrefs := types.ValidatorSetPreferences{
				Preferences: []types.ValidatorPreference{
					{
						ValOperAddress: valAddrs[0],
						Weight:         osmomath.NewDecWithPrec(5, 1), // 0.5
					},
					{
						ValOperAddress: valAddrs[1],
						Weight:         osmomath.NewDecWithPrec(5, 1), // 0.5
					},
				},
			}

			var expectedValsetPref types.ValidatorSetPreferences
			if test.setPref {
				s.App.ValidatorSetPreferenceKeeper.SetValidatorSetPreferences(s.Ctx, delegator.String(), defaultValPrefs)
				expectedValsetPref = defaultValPrefs
			}

			// set two delegation with different weights to test delegation -> val set pref conversion
			if test.setDelegation {
				valAddr0, err := sdk.ValAddressFromBech32(valAddrs[0])
				s.Require().NoError(err)
				validator0, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr0)
				s.Require().NoError(err)
				bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
				s.Require().NoError(err)

				s.FundAcc(delegator, sdk.NewCoins(sdk.NewCoin(bondDenom, defaultDelegateAmt)))
				_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, defaultDelegateAmt, stakingtypes.Unbonded, validator0, true)
				s.Require().NoError(err)

				valAddr1, err := sdk.ValAddressFromBech32(valAddrs[1])
				s.Require().NoError(err)
				validator1, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr1)
				s.Require().NoError(err)

				s.FundAcc(delegator, sdk.NewCoins(sdk.NewCoin(bondDenom, defaultDelegateAmt.Mul(osmomath.NewInt(2)))))
				_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, defaultDelegateAmt.Mul(osmomath.NewInt(2)), stakingtypes.Unbonded, validator1, true)
				s.Require().NoError(err)

				expectedValsetPref = types.ValidatorSetPreferences{
					Preferences: []types.ValidatorPreference{
						{
							ValOperAddress: validator0.OperatorAddress,
							Weight:         osmomath.MustNewDecFromStr("0.333333333333333333"),
						},
						{
							Weight:         osmomath.MustNewDecFromStr("0.666666666666666667"),
							ValOperAddress: validator1.OperatorAddress,
						},
					},
				}
			}

			// system under test
			valsetPref, err := s.App.ValidatorSetPreferenceKeeper.GetValSetPreferencesWithDelegations(s.Ctx, delegator.String())

			if test.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				for _, valset := range valsetPref.Preferences {
					for _, expectedvalSet := range expectedValsetPref.Preferences {
						if valset.ValOperAddress == expectedvalSet.ValOperAddress {
							s.Require().True(valset.Weight.Equal(expectedvalSet.Weight))
						}
					}
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestFormatToValPrefArr() {
	s.SetupTest()

	tests := map[string]struct {
		delegationShares       []osmomath.Dec
		expectedValPrefWeights []osmomath.Dec
		invalidDelegation      bool

		expectedError error
	}{
		"Single delegation": {
			delegationShares:       []osmomath.Dec{osmomath.NewDec(100)},
			expectedValPrefWeights: []osmomath.Dec{osmomath.NewDec(1)},
		},
		"Multiple Delegations": {
			delegationShares: []osmomath.Dec{osmomath.NewDec(100), osmomath.NewDec(200)},
			expectedValPrefWeights: []osmomath.Dec{
				osmomath.MustNewDecFromStr("0.333333333333333333"),
				osmomath.MustNewDecFromStr("0.666666666666666667"),
			},
		},
		"No Delegation": {
			expectedValPrefWeights: []osmomath.Dec{},
		},
		"Invalid delegation (validator doesn't exist)": {
			invalidDelegation: true,
			expectedError:     types.ValidatorNotFoundError{},
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.Setup()
			defaultDelegator := s.TestAccs[0]
			bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
			s.Require().NoError(err)

			// --- Setup ---

			// Prepare delegations by setting up validators and delegating the appropriate amount to each.
			// Since `SetupMultipleValidators` is nondeterministic, we do this setup logic in each test case.
			valAddrs := s.SetupMultipleValidators(len(test.delegationShares))
			delegations, expectedValPrefs := []stakingtypes.Delegation{}, []types.ValidatorPreference{}
			for i, delegationShare := range test.delegationShares {
				// Get validator to delegate to
				valAddr, err := sdk.ValAddressFromBech32(valAddrs[i])
				s.Require().NoError(err)
				validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
				s.Require().NoError(err)

				// Fund delegator and execute delegation
				s.FundAcc(defaultDelegator, sdk.NewCoins(sdk.NewCoin(bondDenom, delegationShare.RoundInt())))
				_, err = s.App.StakingKeeper.Delegate(s.Ctx, defaultDelegator, delegationShare.RoundInt(), stakingtypes.Unbonded, validator, true)
				s.Require().NoError(err)

				// Build list of delegations to pass into SUT
				delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, defaultDelegator, valAddr)
				s.Require().NoError(err)
				delegations = append(delegations, delegation)

				// Build expected validator preferences
				expectedValPrefs = append(expectedValPrefs, types.ValidatorPreference{
					ValOperAddress: valAddrs[i],
					Weight:         test.expectedValPrefWeights[i],
				})
			}

			// Add invalid delegation if specified by test case
			if test.invalidDelegation {
				// Add invalid delegation
				delegations = append(delegations, stakingtypes.Delegation{
					DelegatorAddress: defaultDelegator.String(),
					// Generate random but valid validator address
					ValidatorAddress: sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					Shares:           osmomath.NewDec(100),
				})
			}

			// --- System under test ---

			actualPrefArr, err := s.App.ValidatorSetPreferenceKeeper.FormatToValPrefArr(s.Ctx, delegations)

			// --- Assertions ---

			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().IsType(test.expectedError, err)
				return
			}

			s.Require().NoError(err)
			if test.delegationShares == nil {
				expectedValPrefs = []types.ValidatorPreference{}
			}
			s.Require().Equal(actualPrefArr, expectedValPrefs)
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

	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, syntheticLocks.ID, appparams.BaseCoinUnit, time.Minute, true)
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
