package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v14/app/apptesting"
	"github.com/osmosis-labs/osmosis/v14/x/valset-pref/types"
	"github.com/stretchr/testify/suite"
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

func (suite *KeeperTestSuite) SetupDelegationReward(ctx sdk.Context, delegator sdk.AccAddress, preferences []types.ValidatorPreference, existingValAddrStr string, setValSetDel, setExistingdel bool) {
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

// PrepareExistingDelegations prepares three existing non valset delegations
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
