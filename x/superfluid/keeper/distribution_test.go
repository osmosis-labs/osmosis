package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
)

func (suite *KeeperTestSuite) TestMoveSuperfluidDelegationRewardToGauges() {
	valAddr, _ := suite.SetupSuperfluidDelegate()

	validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	suite.Require().True(found)

	// allocate reward tokens to distribution module
	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20000))}
	suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, distrtypes.ModuleName, coins)

	// allocate rewards to validator
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	decTokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(20000)}}
	suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, decTokens)
	suite.app.DistrKeeper.IncrementValidatorPeriod(suite.ctx, validator)

	// move intermediary account delegation rewards to gauges
	suite.app.SuperfluidKeeper.MoveSuperfluidDelegationRewardToGauges(suite.ctx)

	// check gauge balance
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, uint64(1))
	suite.Require().NoError(err)
	suite.Require().Equal(gauge.Id, uint64(1))
	suite.Require().Equal(gauge.IsPerpetual, true)
	suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "gamm/pool/1" + keeper.StakingSuffix(valAddr.String()),
		Duration:      time.Hour * 24 * 14,
	})
	suite.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsPositive())
	suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
	suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
	suite.Require().Equal(gauge.FilledEpochs, uint64(0))
	suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))
}
