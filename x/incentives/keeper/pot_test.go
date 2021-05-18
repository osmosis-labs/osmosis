package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) CreatePot(isPerpetual bool, addr sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpoch uint64) (uint64, *types.Pot) {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	potID, err := suite.app.IncentivesKeeper.CreatePot(suite.ctx, isPerpetual, addr, coins, distrTo, startTime, numEpoch)
	suite.Require().NoError(err)
	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	return potID, pot
}

func (suite *KeeperTestSuite) AddToPot(coins sdk.Coins, potID uint64) uint64 {
	addr := sdk.AccAddress([]byte("addrx---------------"))
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	err := suite.app.IncentivesKeeper.AddToPotRewards(suite.ctx, addr, coins, potID)
	suite.Require().NoError(err)
	return potID
}

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	_, err := suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) SetupNewPot(isPerpetual bool, coins sdk.Coins) (uint64, *types.Pot, sdk.Coins, time.Time) {
	addr2 := sdk.AccAddress([]byte("addr1---------------"))
	startTime2 := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	potID, pot := suite.CreatePot(isPerpetual, addr2, coins, distrTo, startTime2, numEpochsPaidOver)
	return potID, pot, coins, startTime2
}

func (suite *KeeperTestSuite) SetupLockAndPot(isPerpetual bool) (sdk.AccAddress, uint64, sdk.Coins, time.Time) {
	// create a pot and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// create pot
	potID, _, potCoins, startTime := suite.SetupNewPot(isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	return lockOwner, potID, potCoins, startTime
}

func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	// test for module get pots
	suite.SetupTest()

	// initial check
	coins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup lock and pot
	_, potID, potCoins, startTime := suite.SetupLockAndPot(false)

	// check after pot creation
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, potCoins)

	// add to pot and check
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToPot(addCoins, potID)
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, potCoins.Add(addCoins...))

	// check after creating another pot from another address
	_, _, potCoins2, _ := suite.SetupNewPot(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})

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

	// setup lock and pot
	_, potID, _, startTime := suite.SetupLockAndPot(false)

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

func (suite *KeeperTestSuite) TestNonPerpetualPotOperations() {
	// test for module get pots
	suite.SetupTest()

	// initial module pots check
	pots := suite.app.IncentivesKeeper.GetNotFinishedPots(suite.ctx)
	suite.Require().Len(pots, 0)

	// setup lock and pot
	lockOwner, potID, coins, startTime := suite.SetupLockAndPot(false)

	// check pots
	pots = suite.app.IncentivesKeeper.GetNotFinishedPots(suite.ctx)
	suite.Require().Len(pots, 1)
	suite.Require().Equal(pots[0].Id, potID)
	suite.Require().Equal(pots[0].Coins, coins)
	suite.Require().Equal(pots[0].NumEpochsPaidOver, uint64(2))
	suite.Require().Equal(pots[0].FilledEpochs, uint64(0))
	suite.Require().Equal(pots[0].DistributedCoins, sdk.Coins{})
	suite.Require().Equal(pots[0].StartTime.Unix(), startTime.Unix())

	// check rewards estimation
	rewardsEst := suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, []types.Pot{}, 100)
	suite.Require().Equal(coins.String(), rewardsEst.String())

	// add to pot
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToPot(addCoins, potID)

	// check pots
	pots = suite.app.IncentivesKeeper.GetNotFinishedPots(suite.ctx)
	suite.Require().Len(pots, 1)
	expectedPot := types.Pot{
		Id:          potID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins.Add(addCoins...),
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(pots[0].String(), expectedPot.String())

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

func (suite *KeeperTestSuite) TestPerpetualPotOperations() {
	// test for module get pots
	suite.SetupTest()

	// initial module pots check
	pots := suite.app.IncentivesKeeper.GetNotFinishedPots(suite.ctx)
	suite.Require().Len(pots, 0)

	// setup lock and pot
	lockOwner, potID, coins, startTime := suite.SetupLockAndPot(true)

	// check pots
	pots = suite.app.IncentivesKeeper.GetNotFinishedPots(suite.ctx)
	suite.Require().Len(pots, 1)
	expectedPot := types.Pot{
		Id:          potID,
		IsPerpetual: true,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(pots[0].String(), expectedPot.String())

	// check rewards estimation
	rewardsEst := suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, []types.Pot{}, 100)
	suite.Require().Equal(coins.String(), rewardsEst.String())

	// check pots
	pots = suite.app.IncentivesKeeper.GetNotFinishedPots(suite.ctx)
	suite.Require().Len(pots, 1)
	expectedPot = types.Pot{
		Id:          potID,
		IsPerpetual: true,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(pots[0].String(), expectedPot.String())

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

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// distributing twice without adding more for perpetual pot
	pot, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{})

	// add to pot
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToPot(addCoins, potID)

	// distributing twice with adding more for perpetual pot
	pot, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *pot)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 200)})

	// check active pots
	pots = suite.app.IncentivesKeeper.GetActivePots(suite.ctx)
	suite.Require().Len(pots, 1)

	// check finished pots
	pots = suite.app.IncentivesKeeper.GetFinishedPots(suite.ctx)
	suite.Require().Len(pots, 0)

	// check rewards estimation
	rewardsEst = suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, []types.Pot{}, 100)
	suite.Require().Equal(sdk.Coins(nil), rewardsEst)
}
