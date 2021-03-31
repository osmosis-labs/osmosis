package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) CreatePot(addr sdk.AccAddress, coins sdk.Coins, distrTo types.DistrCondition, startTime time.Time, numEpoch uint64) uint64 {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	potID, err := suite.app.IncentivesKeeper.CreatePot(suite.ctx, addr, coins, distrTo, startTime, numEpoch)
	suite.Require().NoError(err)
	return potID
}

func (suite *KeeperTestSuite) AddToPot(addr sdk.AccAddress, coins sdk.Coins, potID uint64) uint64 {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	err := suite.app.IncentivesKeeper.AddToPot(suite.ctx, addr, coins, potID)
	suite.Require().NoError(err)
	return potID
}

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	_, err := suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	// test for module get pots
	suite.SetupTest()

	// initial check
	coins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// create a pot and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// create pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	potCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	distrTo := types.DistrCondition{
		LockQueryType: types.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	potID := suite.CreatePot(addr1, potCoins, distrTo, startTime, 2)

	// check after pot creation
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, potCoins)

	// add to pot and check
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToPot(addr1, addCoins, potID)
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, potCoins.Add(addCoins...))

	// check after creating another pot from another address
	addr2 := sdk.AccAddress([]byte("addr1---------------"))
	startTime2 := time.Now()
	potCoins2 := sdk.Coins{sdk.NewInt64Coin("stake", 1000)}
	suite.CreatePot(addr2, potCoins2, distrTo, startTime2, 2)
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, potCoins.Add(addCoins...).Add(potCoins2...))

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *pot)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check pot changes after distribution
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, potCoins.Add(addCoins...).Add(potCoins2...).Sub(distrCoins))
}

func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	suite.SetupTest()

	// initial check
	coins := suite.app.IncentivesKeeper.GetModuleDistributedCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// create a pot and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	distrTo := types.DistrCondition{
		LockQueryType: types.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}

	// create pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	potCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	potID := suite.CreatePot(addr1, potCoins, distrTo, startTime, 2)

	// check after pot creation
	coins = suite.app.IncentivesKeeper.GetModuleDistributedCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *pot)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 5)})

	// check after distribution
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, distrCoins)
}

func (suite *KeeperTestSuite) TestPotOperations() {
	// test for module get pots
	suite.SetupTest()

	// initial module pots check
	pots := suite.app.IncentivesKeeper.GetPots(suite.ctx)
	suite.Require().Len(pots, 0)

	// create a pot and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// create pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	distrTo := types.DistrCondition{
		LockQueryType: types.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	potID := suite.CreatePot(addr1, coins, distrTo, startTime, 2)

	// check pots
	pots = suite.app.IncentivesKeeper.GetPots(suite.ctx)
	suite.Require().Len(pots, 1)
	suite.Require().Equal(pots[0].Id, potID)
	suite.Require().Equal(pots[0].Coins, coins)
	suite.Require().Equal(pots[0].NumEpochs, uint64(2))
	suite.Require().Equal(pots[0].FilledEpochs, uint64(0))
	suite.Require().Equal(pots[0].DistributedCoins, sdk.Coins{})
	suite.Require().Equal(pots[0].StartTime.Unix(), startTime.Unix())

	// check rewards estimation
	rewardsEst := suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, []types.Pot{}, 100)
	suite.Require().Equal(coins.String(), rewardsEst.String())

	// add to pot
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToPot(addr1, addCoins, potID)

	// check pots
	pots = suite.app.IncentivesKeeper.GetPots(suite.ctx)
	suite.Require().Len(pots, 1)
	suite.Require().Equal(pots[0].Id, potID)
	suite.Require().Equal(pots[0].Coins, coins.Add(addCoins...))
	suite.Require().Equal(pots[0].NumEpochs, uint64(2))
	suite.Require().Equal(pots[0].FilledEpochs, uint64(0))
	suite.Require().Equal(pots[0].DistributedCoins, sdk.Coins{})
	suite.Require().Equal(pots[0].StartTime.Unix(), startTime.Unix())

	// check upcoming pots
	pots = suite.app.IncentivesKeeper.GetUpcomingPots(suite.ctx)
	suite.Require().Len(pots, 1)

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *pot)
	suite.Require().NoError(err)

	// check upcoming pots
	pots = suite.app.IncentivesKeeper.GetUpcomingPots(suite.ctx)
	suite.Require().Len(pots, 0)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check active pots
	pots = suite.app.IncentivesKeeper.GetActivePots(suite.ctx)
	suite.Require().Len(pots, 1)

	// finish distribution
	err = suite.app.IncentivesKeeper.FinishDistribution(suite.ctx, *pot)
	suite.Require().NoError(err)

	// check finished pots
	pots = suite.app.IncentivesKeeper.GetFinishedPots(suite.ctx)
	suite.Require().Len(pots, 1)

	// check pot by ID
	pot, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	suite.Require().NotNil(pot)
	suite.Require().Equal(*pot, pots[0])

	// check invalid pot ID
	_, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID+1000)
	suite.Require().Error(err)

	// check rewards estimation
	rewardsEst = suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, []types.Pot{}, 100)
	suite.Require().Equal(sdk.Coins{}, rewardsEst)
}
