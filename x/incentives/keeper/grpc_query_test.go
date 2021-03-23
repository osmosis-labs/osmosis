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
	suite.Require().Equal(res, (*types.PotByIDResponse)(nil))

	// final check
	res, err = suite.app.IncentivesKeeper.PotByID(sdk.WrapSDKContext(suite.ctx), &types.PotByIDRequest{Id: potID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Pot, nil)
	suite.Require().Equal(res.Pot.Id, potID)
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

	// create a pot and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// initial check
	res, err := suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner: lockOwner,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create a pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// TODO: implement final check after implementation
	// res, err = suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
	// 	Owner: lockOwner,
	// })
	// suite.Require().NoError(err)
	// suite.Require().Equal(res.Coins, coins)
}

func (suite *KeeperTestSuite) TestGRPCToDistributeCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create a pot and locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	distrTo := types.DistrCondition{
		LockQueryType: types.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	potID := suite.CreatePot(addr1, coins, distrTo, startTime, 2)
	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	suite.Require().NotNil(pot)

	// check after pot creation
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check pot changes after distribution
	pot, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	suite.Require().NotNil(pot)
	suite.Require().Equal(pot.FilledEpochs, uint64(1))
	suite.Require().Equal(pot.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *pot)
	suite.Require().NoError(err)

	// check after distribution
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins.Sub(distrCoins))

	// distribute second round to stakers
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 6)})

	// final check
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))
}

func (suite *KeeperTestSuite) TestGRPCDistributedCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create a pot and locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	distrTo := types.DistrCondition{
		LockQueryType: types.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	potID := suite.CreatePot(addr1, coins, distrTo, startTime, 2)
	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	suite.Require().NotNil(pot)

	// check after pot creation
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *pot)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check pot changes after distribution
	pot, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	suite.Require().NotNil(pot)
	suite.Require().Equal(pot.FilledEpochs, uint64(1))
	suite.Require().Equal(pot.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check after distribution
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, distrCoins)

	// distribute second round to stakers
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 6)})

	// final check
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}
