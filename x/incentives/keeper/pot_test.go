package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
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

func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	// test for module get pots
	suite.SetupTest()

	// initial check
	coins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// create pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	potCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	// TODO: create locks and add those locks for distribution condition after distribution code finish
	potID := suite.CreatePot(addr1, potCoins, types.DistrCondition{}, startTime, 2)

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
	suite.CreatePot(addr2, potCoins2, types.DistrCondition{}, startTime2, 2)
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, potCoins.Add(addCoins...).Add(potCoins2...))

	// TODO: implement check after distribution
	// BeginDistribution()
	// Distribute()
	// FinishDistribution()
}

func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	// test for module get pots
	suite.SetupTest()

	// initial check
	coins := suite.app.IncentivesKeeper.GetModuleDistributedCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// TODO: implement distribution test
	// BeginDistribution()
	// Distribute()
	// FinishDistribution()
}

func (suite *KeeperTestSuite) TestPotOperations() {
	// test for module get pots
	suite.SetupTest()

	// initial module pots check
	pots := suite.app.IncentivesKeeper.GetPots(suite.ctx)
	suite.Require().Len(pots, 0)

	// create pot
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	startTime := time.Now()
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	// TODO: create locks and add those locks for distribution condition after distribution code finish
	potID := suite.CreatePot(addr1, coins, types.DistrCondition{}, startTime, 2)

	// check pots
	pots = suite.app.IncentivesKeeper.GetPots(suite.ctx)
	suite.Require().Len(pots, 1)
	suite.Require().Equal(pots[0].Id, potID)
	suite.Require().Equal(pots[0].Coins, coins)
	suite.Require().Equal(pots[0].NumEpochs, uint64(2))
	suite.Require().Equal(pots[0].FilledEpochs, uint64(0))
	suite.Require().Equal(pots[0].DistributedCoins, sdk.Coins{})
	suite.Require().Equal(pots[0].StartTime.Unix(), startTime.Unix())

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

	// TODO: add test for distribution
	// BeginDistribution()
	// Distribute()
	// FinishDistribution()

	// check pot by ID
	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID)
	suite.Require().NoError(err)
	suite.Require().NotNil(pot)
	suite.Require().Equal(*pot, pots[0])

	// check invalid pot ID
	_, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, potID+1000)
	suite.Require().Error(err)

	// TODO: check active pots - GetActivePots()
	// TODO: check upcoming pots - GetUpcomingPots()
	// TODO: check finished pots - GetFinishedPots()
	// TODO: check rewards estimation - GetRewardsEst()
}
