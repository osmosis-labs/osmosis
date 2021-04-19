package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/pool-incentives/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (suite *KeeperTestSuite) TestGetAllocatableAsset() {
	keeper := suite.app.PoolIncentivesKeeper

	// Params would be set as the stake coin with 20% ratio from the default genesis state.

	// In this time, fee collector doesn't have any assets.
	// So, it should be return the empty coins.
	coin := keeper.GetAllocatableAsset(suite.ctx)
	suite.Equal("0stake", coin.String())

	// Mint the stake coin to the fee collector.
	err := suite.app.BankKeeper.AddCoins(
		suite.ctx,
		suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName),
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000))),
	)
	suite.NoError(err)

	// In this time, should return the 20% of 100000stake
	coin = keeper.GetAllocatableAsset(suite.ctx)
	suite.Equal("20000stake", coin.String())

	// Mint some random coins to the fee collector.
	err = suite.app.BankKeeper.AddCoins(
		suite.ctx,
		suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName),
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1481290)), sdk.NewCoin("test", sdk.NewInt(12389190))),
	)
	suite.NoError(err)

	coin = keeper.GetAllocatableAsset(suite.ctx)
	suite.Equal("316258stake", coin.String())
}

func (suite *KeeperTestSuite) TestAllocateAssetToCommunityPool() {
	keeper := suite.app.PoolIncentivesKeeper

	// Mint the stake coin to the fee collector.
	err := suite.app.BankKeeper.AddCoins(
		suite.ctx,
		suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName),
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(200000))),
	)
	suite.NoError(err)

	// In this time, there is no distr record, so this asset should be allocated to the community pool.
	err = keeper.AllocateAsset(suite.ctx, sdk.NewCoin("stake", sdk.NewInt(100000)))
	suite.NoError(err)

	feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("100000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(feePool.CommunityPool.String(), sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(100000))).String())
	suite.Equal(
		sdk.NewCoin("stake", sdk.NewInt(100000)),
		suite.app.BankKeeper.GetBalance(
			suite.ctx,
			suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName),
			"stake"),
	)

	// Community pool should be increased
	err = keeper.AllocateAsset(suite.ctx, sdk.NewCoin("stake", sdk.NewInt(100000)))
	suite.NoError(err)

	feePool = suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("0stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(feePool.CommunityPool.String(), sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(200000))).String())
	suite.Equal(
		sdk.NewCoin("stake", sdk.NewInt(200000)),
		suite.app.BankKeeper.GetBalance(
			suite.ctx,
			suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName),
			"stake"),
	)
}

func (suite *KeeperTestSuite) TestAllocateAsset() {
	keeper := suite.app.PoolIncentivesKeeper

	// Mint the stake coin to the fee collector.
	err := suite.app.BankKeeper.AddCoins(
		suite.ctx,
		suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName),
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(200000))),
	)
	suite.NoError(err)

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
	err = keeper.AddDistrRecords(suite.ctx, types.DistrRecord{
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
	err = keeper.AllocateAsset(suite.ctx, sdk.NewCoin("stake", sdk.NewInt(100000)))
	suite.NoError(err)
	// It has very small margin of error.
	suite.Equal("100001stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())

	pot1, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot1Id)
	suite.NoError(err)
	suite.Equal("16666stake", pot1.Coins.String())

	pot2, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot2Id)
	suite.NoError(err)
	suite.Equal("33333stake", pot2.Coins.String())

	pot3, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot3Id)
	suite.NoError(err)
	suite.Equal("50000stake", pot3.Coins.String())

	// Allocate more.
	err = keeper.AllocateAsset(suite.ctx, sdk.NewCoin("stake", sdk.NewInt(50000)))
	suite.NoError(err)
	// It has very small margin of error.
	suite.Equal("50002stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())

	// Allocated assets should be increased.
	pot1, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot1Id)
	suite.NoError(err)
	suite.Equal("24999stake", pot1.Coins.String())

	pot2, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot2Id)
	suite.NoError(err)
	suite.Equal("49999stake", pot2.Coins.String())

	pot3, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, pot3Id)
	suite.NoError(err)
	suite.Equal("75000stake", pot3.Coins.String())
}

func (suite *KeeperTestSuite) TestAddDistrRecords() {
	suite.SetupTest()

	keeper := suite.app.PoolIncentivesKeeper

	// Not existent pot.
	err := keeper.AddDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  1,
		Weight: sdk.NewInt(100),
	})
	suite.Error(err)

	poolId := suite.preparePool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	potId, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	err = keeper.AddDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  potId,
		Weight: sdk.NewInt(100),
	})
	suite.NoError(err)
	distrInfo := keeper.GetDistrInfo(suite.ctx)
	suite.Equal(1, len(distrInfo.Records))
	suite.Equal(potId, distrInfo.Records[0].PotId)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(100), distrInfo.TotalWeight)

	err = keeper.AddDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  potId,
		Weight: sdk.NewInt(100),
	}, types.DistrRecord{
		PotId:  potId,
		Weight: sdk.NewInt(200),
	})
	suite.NoError(err)
	distrInfo = keeper.GetDistrInfo(suite.ctx)
	suite.Equal(3, len(distrInfo.Records))
	suite.Equal(potId, distrInfo.Records[0].PotId)
	suite.Equal(potId, distrInfo.Records[1].PotId)
	suite.Equal(potId, distrInfo.Records[2].PotId)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[1].Weight)
	suite.Equal(sdk.NewInt(200), distrInfo.Records[2].Weight)
	suite.Equal(sdk.NewInt(400), distrInfo.TotalWeight)
}

func (suite *KeeperTestSuite) TestEditDistrRecords() {
	suite.SetupTest()

	keeper := suite.app.PoolIncentivesKeeper

	// Not existent record.
	err := keeper.EditDistrRecords(suite.ctx, types.EditPoolIncentivesProposal_DistrRecordWithIndex{
		Index: 0,
		Record: types.DistrRecord{
			PotId:  1,
			Weight: sdk.NewInt(100),
		},
	})
	suite.Error(err)

	// Create 3 records
	suite.TestAddDistrRecords()

	keeper = suite.app.PoolIncentivesKeeper

	// Not existent record.
	err = keeper.EditDistrRecords(suite.ctx, types.EditPoolIncentivesProposal_DistrRecordWithIndex{
		Index: 3,
		Record: types.DistrRecord{
			PotId:  1,
			Weight: sdk.NewInt(100),
		},
	})
	suite.Error(err)

	priorDistrInfo := keeper.GetDistrInfo(suite.ctx)
	firstRecord := priorDistrInfo.Records[0]
	// Try change the first record
	err = keeper.EditDistrRecords(suite.ctx, types.EditPoolIncentivesProposal_DistrRecordWithIndex{
		Index: 0,
		Record: types.DistrRecord{
			PotId:  firstRecord.PotId,
			Weight: sdk.NewInt(200),
		},
	})
	suite.NoError(err)
	distrInfo := keeper.GetDistrInfo(suite.ctx)
	suite.Equal(3, len(distrInfo.Records))
	suite.Equal(firstRecord.PotId, distrInfo.Records[0].PotId)
	suite.Equal(sdk.NewInt(200), distrInfo.Records[0].Weight)
	suite.Equal(priorDistrInfo.TotalWeight.Add(sdk.NewInt(100)), distrInfo.TotalWeight)

	// Try change the invalid index record.
	err = keeper.EditDistrRecords(suite.ctx, types.EditPoolIncentivesProposal_DistrRecordWithIndex{
		Index: 3,
		Record: types.DistrRecord{
			PotId:  firstRecord.PotId,
			Weight: sdk.NewInt(200),
		},
	})
	suite.Error(err)

	// Try change the record with mismatched pot id.
	err = keeper.EditDistrRecords(suite.ctx, types.EditPoolIncentivesProposal_DistrRecordWithIndex{
		Index: 0,
		Record: types.DistrRecord{
			PotId:  firstRecord.PotId + 1,
			Weight: sdk.NewInt(300),
		},
	})
	suite.Error(err)

	priorDistrInfo = keeper.GetDistrInfo(suite.ctx)
	lastRecord := priorDistrInfo.Records[0]
	// Try change the last record.
	err = keeper.EditDistrRecords(suite.ctx, types.EditPoolIncentivesProposal_DistrRecordWithIndex{
		Index: 2,
		Record: types.DistrRecord{
			PotId:  lastRecord.PotId,
			Weight: sdk.NewInt(100),
		},
	})
	suite.NoError(err)
	distrInfo = keeper.GetDistrInfo(suite.ctx)
	suite.Equal(3, len(distrInfo.Records))
	suite.Equal(lastRecord.PotId, distrInfo.Records[2].PotId)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[2].Weight)
	suite.Equal(priorDistrInfo.TotalWeight.Sub(sdk.NewInt(100)), distrInfo.TotalWeight)
}

func (suite *KeeperTestSuite) TestRemoveDistrRecords() {
	suite.SetupTest()

	keeper := suite.app.PoolIncentivesKeeper

	// Not existent record.
	err := keeper.RemoveDistrRecords(suite.ctx, 0)
	suite.Error(err)

	// Create 3 records
	suite.TestAddDistrRecords()

	keeper = suite.app.PoolIncentivesKeeper

	// Not existent record.
	err = keeper.RemoveDistrRecords(suite.ctx, 3)
	suite.Error(err)

	priorDistrInfo := keeper.GetDistrInfo(suite.ctx)
	// Try remove the first record
	err = keeper.RemoveDistrRecords(suite.ctx, 0)
	suite.NoError(err)
	distrInfo := keeper.GetDistrInfo(suite.ctx)
	suite.Equal(2, len(distrInfo.Records))
	suite.Equal(priorDistrInfo.TotalWeight.Sub(sdk.NewInt(100)), distrInfo.TotalWeight)

	// Try remove the invalid index record.
	err = keeper.RemoveDistrRecords(suite.ctx, 3)
	suite.Error(err)

	priorDistrInfo = keeper.GetDistrInfo(suite.ctx)
	// Try remove the last record.
	err = keeper.RemoveDistrRecords(suite.ctx, 1)
	suite.NoError(err)
	distrInfo = keeper.GetDistrInfo(suite.ctx)
	suite.Equal(1, len(distrInfo.Records))
	suite.Equal(priorDistrInfo.TotalWeight.Sub(sdk.NewInt(200)), distrInfo.TotalWeight)

	priorDistrInfo = keeper.GetDistrInfo(suite.ctx)
	// Finally, try remove the remaining record.
	err = keeper.RemoveDistrRecords(suite.ctx, 0)
	suite.NoError(err)
	distrInfo = keeper.GetDistrInfo(suite.ctx)
	suite.Equal(0, len(distrInfo.Records))
	suite.Equal(priorDistrInfo.TotalWeight.Sub(sdk.NewInt(100)).String(), distrInfo.TotalWeight.String())
	suite.Equal(sdk.NewInt(0), distrInfo.TotalWeight)
}
