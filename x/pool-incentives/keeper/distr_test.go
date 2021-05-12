package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (suite *KeeperTestSuite) TestAllocateAssetToCommunityPool() {
	mintKeeper := suite.app.MintKeeper

	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoins := sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err := mintKeeper.AddCollectedFees(suite.ctx, mintCoins) // this calls AllocateAsset via hook
	suite.NoError(err)

	feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("60000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(feePool.CommunityPool.String(), sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(20000))).String())
	suite.Equal(
		sdk.NewCoin("stake", sdk.NewInt(20000)),
		suite.app.BankKeeper.GetBalance(
			suite.ctx,
			suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName),
			"stake"),
	)

	// Community pool should be increased
	mintCoins = sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err = mintKeeper.AddCollectedFees(suite.ctx, mintCoins) // this calls AllocateAsset via hook
	suite.NoError(err)

	feePool = suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("76000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(feePool.CommunityPool.String(), sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(72000))).String())
	suite.Equal(
		sdk.NewCoin("stake", sdk.NewInt(72000)),
		suite.app.BankKeeper.GetBalance(
			suite.ctx,
			suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName),
			"stake"),
	)
}

func (suite *KeeperTestSuite) TestAllocateAsset() {
	keeper := suite.app.PoolIncentivesKeeper
	mintKeeper := suite.app.MintKeeper

	poolId := suite.preparePool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	pot1Id, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	pot2Id, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[1])
	suite.NoError(err)

	pot3Id, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[2])
	suite.NoError(err)

	// Create 3 records
	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  pot1Id,
		Weight: sdk.NewInt(100),
	}, types.DistrRecord{
		PotId:  pot2Id,
		Weight: sdk.NewInt(200),
	}, types.DistrRecord{
		PotId:  pot3Id,
		Weight: sdk.NewInt(300),
	})
	suite.NoError(err)

	// In this time, there are 3 records, so the assets to be allocated to the pots proportionally.
	mintCoins := sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err = mintKeeper.AddCollectedFees(suite.ctx, mintCoins) // this calls AllocateAsset via hook
	suite.NoError(err)
	suite.Equal("80000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())

	pot1, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot1Id)
	suite.NoError(err)
	suite.Equal("3333stake", pot1.Coins.String())

	pot2, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot2Id)
	suite.NoError(err)
	suite.Equal("6666stake", pot2.Coins.String())

	pot3, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot3Id)
	suite.NoError(err)
	suite.Equal("10000stake", pot3.Coins.String())

	// Allocate more.
	mintCoins = sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(50000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err = mintKeeper.AddCollectedFees(suite.ctx, mintCoins) // this calls AllocateAsset via hook
	suite.NoError(err)
	// It has very small margin of error.
	suite.Equal("104000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())

	// Allocated assets should be increased.
	pot1, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot1Id)
	suite.NoError(err)
	suite.Equal("7666stake", pot1.Coins.String())

	pot2, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot2Id)
	suite.NoError(err)
	suite.Equal("15332stake", pot2.Coins.String())

	pot3, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot3Id)
	suite.NoError(err)
	suite.Equal("23000stake", pot3.Coins.String())

	// ------------ test community pool distribution when potId is zero ------------ //

	// record original community pool balance
	feePoolOrigin := suite.app.DistrKeeper.GetFeePool(suite.ctx)

	// Create 3 records including community pool
	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  pot1Id,
		Weight: sdk.NewInt(100),
	}, types.DistrRecord{
		PotId:  pot2Id,
		Weight: sdk.NewInt(200),
	}, types.DistrRecord{
		PotId:  0,
		Weight: sdk.NewInt(700),
	})
	suite.NoError(err)

	// In this time, there are 3 records, so the assets to be allocated to the pots proportionally.
	mintCoins = sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err = mintKeeper.AddCollectedFees(suite.ctx, mintCoins) // this calls AllocateAsset via hook
	suite.NoError(err)

	// check community pool balance increase
	feePoolNew := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal(feePoolOrigin.CommunityPool.Add(sdk.NewDecCoin("stake", sdk.NewInt(28561))), feePoolNew.CommunityPool)

	// ------------ test community pool distribution when no distribution records are set ------------ //

	// record original community pool balance
	feePoolOrigin = suite.app.DistrKeeper.GetFeePool(suite.ctx)

	// set empty records set
	err = keeper.UpdateDistrRecords(suite.ctx)
	suite.NoError(err)

	// In this time, all should be allocated to community pool
	mintCoins = sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err = mintKeeper.AddCollectedFees(suite.ctx, mintCoins) // this calls AllocateAsset via hook
	suite.NoError(err)

	// check community pool balance increase
	feePoolNew = suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal(feePoolOrigin.CommunityPool.Add(sdk.NewDecCoin("stake", sdk.NewInt(75489))), feePoolNew.CommunityPool)
}

func (suite *KeeperTestSuite) TestUpdateDistrRecords() uint64 {
	suite.SetupTest()

	keeper := suite.app.PoolIncentivesKeeper

	// Not existent pot.
	err := keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  1,
		Weight: sdk.NewInt(100),
	})
	suite.NoError(err)

	poolId := suite.preparePool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state for testing
	lockableDurations := keeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	potId, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  potId,
		Weight: sdk.NewInt(100),
	})
	suite.NoError(err)
	distrInfo := keeper.GetDistrInfo(suite.ctx)
	suite.Equal(1, len(distrInfo.Records))
	suite.Equal(potId, distrInfo.Records[0].PotId)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(100), distrInfo.TotalWeight)

	// adding two of the same pot id at once should error
	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  potId,
		Weight: sdk.NewInt(100),
	}, types.DistrRecord{
		PotId:  potId,
		Weight: sdk.NewInt(200),
	})
	suite.Error(err)

	potId2 := potId + 1
	potId3 := potId + 2

	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  potId2,
		Weight: sdk.NewInt(100),
	}, types.DistrRecord{
		PotId:  potId3,
		Weight: sdk.NewInt(200),
	})
	suite.NoError(err)

	distrInfo = keeper.GetDistrInfo(suite.ctx)
	suite.Equal(2, len(distrInfo.Records))
	suite.Equal(potId2, distrInfo.Records[0].PotId)
	suite.Equal(potId3, distrInfo.Records[1].PotId)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(200), distrInfo.Records[1].Weight)
	suite.Equal(sdk.NewInt(300), distrInfo.TotalWeight)

	// Can update the registered pot id
	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  potId2,
		Weight: sdk.NewInt(100),
	})
	suite.NoError(err)

	return potId
}
