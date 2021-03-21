package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGRPCPotByID() {
	suite.SetupTest()

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	potID := suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// not available pot
	res, err := suite.app.IncentivesKeeper.PotByID(sdk.WrapSDKContext(suite.ctx), &types.PotByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, nil)

	// final check
	res, err = suite.app.IncentivesKeeper.PotByID(sdk.WrapSDKContext(suite.ctx), &types.PotByIDRequest{Id: potID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Pot, nil)
	suite.Require().NotEqual(res.Pot.Id, potID)
}

func (suite *KeeperTestSuite) TestGRPCPots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.Pots(sdk.WrapSDKContext(suite.ctx), &types.PotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// final check
	res, err = suite.app.IncentivesKeeper.Pots(sdk.WrapSDKContext(suite.ctx), &types.PotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
}

func (suite *KeeperTestSuite) TestGRPCActivePots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ActivePots(sdk.WrapSDKContext(suite.ctx), &types.ActivePotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// final check
	res, err = suite.app.IncentivesKeeper.ActivePots(sdk.WrapSDKContext(suite.ctx), &types.ActivePotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
}

func (suite *KeeperTestSuite) TestGRPCUpcomingPots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.UpcomingPots(sdk.WrapSDKContext(suite.ctx), &types.UpcomingPotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// final check
	res, err = suite.app.IncentivesKeeper.UpcomingPots(sdk.WrapSDKContext(suite.ctx), &types.UpcomingPotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
}

func (suite *KeeperTestSuite) TestGRPCRewardsEst() {
	suite.SetupTest()

	// TODO: setup locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))

	// initial check
	res, err := suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner: lockOwner,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// final check
	res, err = suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner: lockOwner,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}

func (suite *KeeperTestSuite) TestGRPCToDistributeCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// final check
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}

func (suite *KeeperTestSuite) TestGRPCDistributedCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// final check
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})
}
