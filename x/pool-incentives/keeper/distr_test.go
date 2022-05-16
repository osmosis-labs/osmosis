package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/osmosis-labs/osmosis/v8/x/mint/types"
	"github.com/osmosis-labs/osmosis/v8/x/pool-incentives/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestAllocateAssetToCommunityPoolWhenNoDistrRecords() {
	mintKeeper := suite.app.MintKeeper
	params := suite.app.MintKeeper.GetParams(suite.ctx)
	params.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{
		{
			Address: sdk.AccAddress([]byte("addr1---------------")).String(),
			Weight:  sdk.NewDec(1),
		},
	}
	suite.app.MintKeeper.SetParams(suite.ctx, params)

	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoin := sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins := sdk.Coins{mintCoin}
	err := mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)

	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin) // this calls AllocateAsset via hook
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("40000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(40000))).String(), feePool.CommunityPool.String())
	suite.Equal("40000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName), "stake").String())

	// Community pool should be increased
	mintCoin = sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins = sdk.Coins{mintCoin}
	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin) // this calls AllocateAsset via hook
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	feePool = suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("80000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(feePool.CommunityPool.String(), sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(80000))).String())
	suite.Equal(sdk.NewCoin("stake", sdk.NewInt(80000)), suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName), "stake"))
}

func (suite *KeeperTestSuite) TestAllocateAsset() {
	keeper := suite.app.PoolIncentivesKeeper
	mintKeeper := suite.app.MintKeeper
	params := suite.app.MintKeeper.GetParams(suite.ctx)
	params.WeightedDeveloperRewardsReceivers = []minttypes.WeightedAddress{
		{
			Address: sdk.AccAddress([]byte("addr1---------------")).String(),
			Weight:  sdk.NewDec(1),
		},
	}
	suite.app.MintKeeper.SetParams(suite.ctx, params)

	poolId := suite.prepareBalancerPool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	for i, duration := range lockableDurations {
		suite.Equal(duration, types.DefaultGenesisState().GetLockableDurations()[i])
	}

	gauge1Id, err := keeper.GetPoolGaugeId(suite.ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	gauge2Id, err := keeper.GetPoolGaugeId(suite.ctx, poolId, lockableDurations[1])
	suite.NoError(err)

	gauge3Id, err := keeper.GetPoolGaugeId(suite.ctx, poolId, lockableDurations[2])
	suite.NoError(err)

	// Create 3 records
	err = keeper.ReplaceDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: gauge1Id,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gauge2Id,
		Weight:  sdk.NewInt(200),
	}, types.DistrRecord{
		GaugeId: gauge3Id,
		Weight:  sdk.NewInt(300),
	})
	suite.NoError(err)

	// In this time, there are 3 records, so the assets to be allocated to the gauges proportionally.
	mintCoin := sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins := sdk.Coins{mintCoin}
	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)

	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin) // this calls AllocateAsset via hook
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	suite.Equal("40000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())

	gauge1, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gauge1Id)
	suite.NoError(err)
	suite.Equal("5000stake", gauge1.Coins.String())

	gauge2, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gauge2Id)
	suite.NoError(err)
	suite.Equal("9999stake", gauge2.Coins.String())

	gauge3, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gauge3Id)
	suite.NoError(err)
	suite.Equal("15000stake", gauge3.Coins.String())

	// Allocate more.
	mintCoin = sdk.NewCoin("stake", sdk.NewInt(50000))
	mintCoins = sdk.Coins{mintCoin}
	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin) // this calls AllocateAsset via hook
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	// It has very small margin of error.
	suite.Equal("60000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())

	// Allocated assets should be increased.
	gauge1, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gauge1Id)
	suite.NoError(err)
	suite.Equal("7500stake", gauge1.Coins.String())

	gauge2, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gauge2Id)
	suite.NoError(err)
	suite.Equal("14999stake", gauge2.Coins.String())

	gauge3, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gauge3Id)
	suite.NoError(err)
	suite.Equal("22500stake", gauge3.Coins.String())

	// ------------ test community pool distribution when gaugeId is zero ------------ //

	// record original community pool balance
	feePoolOrigin := suite.app.DistrKeeper.GetFeePool(suite.ctx)

	// Create 3 records including community pool
	err = keeper.ReplaceDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: 0,
		Weight:  sdk.NewInt(700),
	}, types.DistrRecord{
		GaugeId: gauge1Id,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gauge2Id,
		Weight:  sdk.NewInt(200),
	})
	suite.NoError(err)

	// In this time, there are 3 records, so the assets to be allocated to the gauges proportionally.
	mintCoin = sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins = sdk.Coins{mintCoin}
	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin) // this calls AllocateAsset via hook
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	// check community pool balance increase
	feePoolNew := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal(feePoolOrigin.CommunityPool.Add(sdk.NewDecCoin("stake", sdk.NewInt(31000))), feePoolNew.CommunityPool)

	// ------------ test community pool distribution when no distribution records are set ------------ //

	// record original community pool balance
	feePoolOrigin = suite.app.DistrKeeper.GetFeePool(suite.ctx)

	// set empty records set
	err = keeper.ReplaceDistrRecords(suite.ctx)
	suite.NoError(err)

	// In this time, all should be allocated to community pool
	mintCoin = sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins = sdk.Coins{mintCoin}
	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin) // this calls AllocateAsset via hook
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	// check community pool balance increase
	feePoolNew = suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal(feePoolOrigin.CommunityPool.Add(sdk.NewDecCoin("stake", sdk.NewInt(40001))), feePoolNew.CommunityPool)
}

func (suite *KeeperTestSuite) TestReplaceDistrRecords() uint64 {
	suite.SetupTest()

	keeper := suite.app.PoolIncentivesKeeper

	// Not existent gauge.
	err := keeper.ReplaceDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: 1,
		Weight:  sdk.NewInt(100),
	})
	suite.Error(err)

	poolId := suite.prepareBalancerPool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state for testing
	lockableDurations := keeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	gaugeId, err := keeper.GetPoolGaugeId(suite.ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	err = keeper.ReplaceDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: gaugeId,
		Weight:  sdk.NewInt(100),
	})
	suite.NoError(err)
	distrInfo := keeper.GetDistrInfo(suite.ctx)
	suite.Equal(1, len(distrInfo.Records))
	suite.Equal(gaugeId, distrInfo.Records[0].GaugeId)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(100), distrInfo.TotalWeight)

	// adding two of the same gauge id at once should error
	err = keeper.ReplaceDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: gaugeId,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gaugeId,
		Weight:  sdk.NewInt(200),
	})
	suite.Error(err)

	gaugeId2 := gaugeId + 1
	gaugeId3 := gaugeId + 2

	err = keeper.ReplaceDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: gaugeId2,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gaugeId3,
		Weight:  sdk.NewInt(200),
	})
	suite.NoError(err)

	distrInfo = keeper.GetDistrInfo(suite.ctx)
	suite.Equal(2, len(distrInfo.Records))
	suite.Equal(gaugeId2, distrInfo.Records[0].GaugeId)
	suite.Equal(gaugeId3, distrInfo.Records[1].GaugeId)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(200), distrInfo.Records[1].Weight)
	suite.Equal(sdk.NewInt(300), distrInfo.TotalWeight)

	// Can replace the registered gauge id
	err = keeper.ReplaceDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: gaugeId2,
		Weight:  sdk.NewInt(100),
	})
	suite.NoError(err)

	return gaugeId
}

func (suite *KeeperTestSuite) TestUpdateDistrRecords() uint64 {
	suite.SetupTest()

	keeper := suite.app.PoolIncentivesKeeper

	// Not existent gauge.
	err := keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: 1,
		Weight:  sdk.NewInt(100),
	})
	suite.Error(err)

	poolId := suite.prepareBalancerPool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state for testing
	lockableDurations := keeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	gaugeId, err := keeper.GetPoolGaugeId(suite.ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: 0,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gaugeId,
		Weight:  sdk.NewInt(100),
	})
	suite.NoError(err)
	distrInfo := keeper.GetDistrInfo(suite.ctx)
	suite.Equal(2, len(distrInfo.Records))
	suite.Equal(sdk.NewInt(100), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(100), distrInfo.Records[1].Weight)
	suite.Equal(sdk.NewInt(200), distrInfo.TotalWeight)

	// adding two of the same gauge id at once should error
	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		GaugeId: 0,
		Weight:  sdk.NewInt(0),
	}, types.DistrRecord{
		GaugeId: gaugeId,
		Weight:  sdk.NewInt(200),
	}, types.DistrRecord{
		GaugeId: gaugeId + 1,
		Weight:  sdk.NewInt(150),
	})
	suite.NoError(err)
	distrInfo = keeper.GetDistrInfo(suite.ctx)
	suite.Equal(2, len(distrInfo.Records))
	suite.Equal(sdk.NewInt(200), distrInfo.Records[0].Weight)
	suite.Equal(sdk.NewInt(150), distrInfo.Records[1].Weight)
	suite.Equal(sdk.NewInt(350), distrInfo.TotalWeight, distrInfo.TotalWeight.String())

	return gaugeId
}
