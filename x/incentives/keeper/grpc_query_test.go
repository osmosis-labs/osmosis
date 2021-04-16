package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGRPCPotByID() {
	suite.SetupTest()

	// create a pot
	potID, _, coins, startTime := suite.SetupNewPot(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// Ensure that a querying for a pot with an ID that doesn't exist returns an error
	res, err := suite.app.IncentivesKeeper.PotByID(sdk.WrapSDKContext(suite.ctx), &types.PotByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, (*types.PotByIDResponse)(nil))

	// Check that querying a pot with an ID that exists returns the pot.
	res, err = suite.app.IncentivesKeeper.PotByID(sdk.WrapSDKContext(suite.ctx), &types.PotByIDRequest{Id: potID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Pot, nil)
	expectedPot := types.Pot{
		Id:          potID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Pot.String(), expectedPot.String())
}

func (suite *KeeperTestSuite) TestGRPCPots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.Pots(sdk.WrapSDKContext(suite.ctx), &types.PotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	potID, _, coins, startTime := suite.SetupNewPot(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.app.IncentivesKeeper.Pots(sdk.WrapSDKContext(suite.ctx), &types.PotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedPot := types.Pot{
		Id:          potID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Data[0].String(), expectedPot.String())
}

func (suite *KeeperTestSuite) TestGRPCActivePots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ActivePots(sdk.WrapSDKContext(suite.ctx), &types.ActivePotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	potID, pot, coins, startTime := suite.SetupNewPot(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	suite.ctx = suite.ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *pot)
	suite.Require().NoError(err)

	// final check
	res, err = suite.app.IncentivesKeeper.ActivePots(sdk.WrapSDKContext(suite.ctx), &types.ActivePotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedPot := types.Pot{
		Id:          potID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Data[0].String(), expectedPot.String())
}

func (suite *KeeperTestSuite) TestGRPCUpcomingPots() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.UpcomingPots(sdk.WrapSDKContext(suite.ctx), &types.UpcomingPotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a pot
	potID, _, coins, startTime := suite.SetupNewPot(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.app.IncentivesKeeper.UpcomingPots(sdk.WrapSDKContext(suite.ctx), &types.UpcomingPotsRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedPot := types.Pot{
		Id:          potID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(res.Data[0].String(), expectedPot.String())
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
	lockOwner, _, coins, _ := suite.SetupLockAndPot(false)

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
	potID, _, coins, startTime := suite.SetupNewPot(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
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
	potID, _, coins, startTime := suite.SetupNewPot(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
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
