package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGRPCPotByID() {
	suite.SetupTest()

	// create a pot
	potID, coins, startTime := suite.SetupNewPot(sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// Ensure that a querying for a pot with an ID that doesn't exist returns an error
	res, err := suite.app.IncentivesKeeper.PotByID(sdk.WrapSDKContext(suite.ctx), &types.PotByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, (*types.PotByIDResponse)(nil))

	// Check that querying a pot with an ID that exists returns the pot.
	res, err = suite.app.IncentivesKeeper.PotByID(sdk.WrapSDKContext(suite.ctx), &types.PotByIDRequest{Id: potID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Pot, nil)
	suite.Require().Equal(res.Pot.Id, potID)
	suite.Require().Equal(res.Pot.Coins, coins)
	suite.Require().Equal(res.Pot.NumEpochs, uint64(2))
	suite.Require().Equal(res.Pot.FilledEpochs, uint64(0))
	suite.Require().Equal(res.Pot.DistributedCoins, sdk.Coins{})
	suite.Require().Equal(res.Pot.StartTime.Unix(), startTime.Unix())
}

func (suite *KeeperTestSuite) TestGRPCPots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.Pots(sdk.WrapSDKContext(suite.ctx), &types.PotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	_, coins, startTime := suite.SetupNewPot(sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.app.IncentivesKeeper.Pots(sdk.WrapSDKContext(suite.ctx), &types.PotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	suite.Require().Equal(res.Data[0].Coins, coins)
	suite.Require().Equal(res.Data[0].NumEpochs, uint64(2))
	suite.Require().Equal(res.Data[0].FilledEpochs, uint64(0))
	suite.Require().Equal(res.Data[0].DistributedCoins, sdk.Coins{})
	suite.Require().Equal(res.Data[0].StartTime.Unix(), startTime.Unix())
}

func (suite *KeeperTestSuite) TestGRPCActivePots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ActivePots(sdk.WrapSDKContext(suite.ctx), &types.ActivePotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	_, coins, startTime := suite.SetupNewPot(sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.app.IncentivesKeeper.ActivePots(sdk.WrapSDKContext(suite.ctx), &types.ActivePotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	suite.Require().Equal(res.Data[0].Coins, coins)
	suite.Require().Equal(res.Data[0].NumEpochs, uint64(2))
	suite.Require().Equal(res.Data[0].FilledEpochs, uint64(0))
	suite.Require().Equal(res.Data[0].DistributedCoins, sdk.Coins{})
	suite.Require().Equal(res.Data[0].StartTime.Unix(), startTime.Unix())
}

func (suite *KeeperTestSuite) TestGRPCUpcomingPots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.UpcomingPots(sdk.WrapSDKContext(suite.ctx), &types.UpcomingPotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	_, coins, startTime := suite.SetupNewPot(sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.app.IncentivesKeeper.UpcomingPots(sdk.WrapSDKContext(suite.ctx), &types.UpcomingPotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	suite.Require().Equal(res.Data[0].Coins, coins)
	suite.Require().Equal(res.Data[0].NumEpochs, uint64(2))
	suite.Require().Equal(res.Data[0].FilledEpochs, uint64(0))
	suite.Require().Equal(res.Data[0].DistributedCoins, sdk.Coins{})
	suite.Require().Equal(res.Data[0].StartTime.Unix(), startTime.Unix())
}

func (suite *KeeperTestSuite) TestGRPCRewardsEst() {
	suite.SetupTest()

	// initial check
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	res, err := suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner: lockOwner,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// setup lock and pot
	lockOwner, _, coins, _ := suite.SetupLockAndPot()

	res, err = suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner:    lockOwner,
		EndEpoch: 100,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}

func (suite *KeeperTestSuite) TestGRPCToDistributeCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a pot
	potID, coins, startTime := suite.SetupNewPot(sdk.Coins{sdk.NewInt64Coin("stake", 10)})
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

	// create locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a pot
	potID, coins, startTime := suite.SetupNewPot(sdk.Coins{sdk.NewInt64Coin("stake", 10)})
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
