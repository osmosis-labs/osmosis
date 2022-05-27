package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v8/app/apptesting"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v8/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v8/x/pool-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestMintCoinsToFeeCollectorAndGetProportions() {
	mintKeeper := suite.App.MintKeeper

	// When coin is minted to the fee collector
	fee := sdk.NewCoin("stake", sdk.NewInt(0))
	fees := sdk.NewCoins(fee)
	coin := mintKeeper.GetProportions(suite.Ctx, fee, sdk.NewDecWithPrec(2, 1))
	suite.Equal("0stake", coin.String())

	// When mint the 100K stake coin to the fee collector
	fee = sdk.NewCoin("stake", sdk.NewInt(100000))
	fees = sdk.NewCoins(fee)

	err := simapp.FundModuleAccount(suite.App.BankKeeper,
		suite.Ctx,
		authtypes.FeeCollectorName,
		fees)
	suite.NoError(err)

	// check proportion for 20%
	coin = mintKeeper.GetProportions(suite.Ctx, fee, sdk.NewDecWithPrec(2, 1))
	suite.Equal(fees[0].Amount.Quo(sdk.NewInt(5)), coin.Amount)
}

func (suite *KeeperTestSuite) TestDistrAssetToDeveloperRewardsAddrWhenNotEmpty() {
	mintKeeper := suite.App.MintKeeper
	params := suite.App.MintKeeper.GetParams(suite.Ctx)
	devRewardsReceiver := sdk.AccAddress([]byte("addr1---------------"))
	gaugeCreator := sdk.AccAddress([]byte("addr2---------------"))
	devRewardsReceiver2 := sdk.AccAddress([]byte("addr3---------------"))
	devRewardsReceiver3 := sdk.AccAddress([]byte("addr4---------------"))
	params.WeightedDeveloperRewardsReceivers = []types.WeightedAddress{
		{
			Address: devRewardsReceiver.String(),
			Weight:  sdk.NewDec(1),
		},
	}
	suite.App.MintKeeper.SetParams(suite.Ctx, params)

	// Create record
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	suite.FundAcc(gaugeCreator, coins)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}

	// mints coins so supply exists on chain
	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	suite.FundAcc(gaugeCreator, mintLPtokens)

	gaugeId, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, true, gaugeCreator, coins, distrTo, time.Now(), 1)
	suite.NoError(err)
	err = suite.App.PoolIncentivesKeeper.UpdateDistrRecords(suite.Ctx, poolincentivestypes.DistrRecord{
		GaugeId: gaugeId,
		Weight:  sdk.NewInt(100),
	})
	suite.NoError(err)

	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoin := sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins := sdk.Coins{mintCoin}
	err = mintKeeper.MintCoins(suite.Ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.Ctx, mintCoin)
	suite.NoError(err)

	feePool := suite.App.DistrKeeper.GetFeePool(suite.Ctx)
	feeCollector := suite.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	suite.Equal(
		mintCoin.Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt(),
		suite.App.BankKeeper.GetAllBalances(suite.Ctx, feeCollector).AmountOf("stake"))
	suite.Equal(
		mintCoin.Amount.ToDec().Mul(params.DistributionProportions.CommunityPool),
		feePool.CommunityPool.AmountOf("stake"))
	suite.Equal(
		mintCoin.Amount.ToDec().Mul(params.DistributionProportions.DeveloperRewards).TruncateInt(),
		suite.App.BankKeeper.GetBalance(suite.Ctx, devRewardsReceiver, "stake").Amount)

	// Test for multiple dev reward addresses
	params.WeightedDeveloperRewardsReceivers = []types.WeightedAddress{
		{
			Address: devRewardsReceiver2.String(),
			Weight:  sdk.NewDecWithPrec(6, 1),
		},
		{
			Address: devRewardsReceiver3.String(),
			Weight:  sdk.NewDecWithPrec(4, 1),
		},
	}
	suite.App.MintKeeper.SetParams(suite.Ctx, params)

	err = mintKeeper.MintCoins(suite.Ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.Ctx, mintCoin)
	suite.NoError(err)

	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.DeveloperRewards).Mul(params.WeightedDeveloperRewardsReceivers[0].Weight).TruncateInt(),
		suite.App.BankKeeper.GetBalance(suite.Ctx, devRewardsReceiver2, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.DeveloperRewards).Mul(params.WeightedDeveloperRewardsReceivers[1].Weight).TruncateInt(),
		suite.App.BankKeeper.GetBalance(suite.Ctx, devRewardsReceiver3, "stake").Amount)
}

func (suite *KeeperTestSuite) TestDistrAssetToCommunityPoolWhenNoDeveloperRewardsAddr() {
	mintKeeper := suite.App.MintKeeper

	params := suite.App.MintKeeper.GetParams(suite.Ctx)
	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoin := sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins := sdk.Coins{mintCoin}
	err := mintKeeper.MintCoins(suite.Ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.Ctx, mintCoin)
	suite.NoError(err)

	distribution.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{}, *suite.App.DistrKeeper)

	feePool := suite.App.DistrKeeper.GetFeePool(suite.Ctx)
	feeCollector := suite.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	// PoolIncentives + DeveloperRewards + CommunityPool => CommunityPool
	proportionToCommunity := params.DistributionProportions.PoolIncentives.
		Add(params.DistributionProportions.DeveloperRewards).
		Add(params.DistributionProportions.CommunityPool)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt(),
		suite.App.BankKeeper.GetBalance(suite.Ctx, feeCollector, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(proportionToCommunity),
		feePool.CommunityPool.AmountOf("stake"))

	// Mint more and community pool should be increased
	err = mintKeeper.MintCoins(suite.Ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.Ctx, mintCoin)
	suite.NoError(err)

	distribution.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{}, *suite.App.DistrKeeper)

	feePool = suite.App.DistrKeeper.GetFeePool(suite.Ctx)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt().Mul(sdk.NewInt(2)),
		suite.App.BankKeeper.GetBalance(suite.Ctx, feeCollector, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(proportionToCommunity).Mul(sdk.NewDec(2)),
		feePool.CommunityPool.AmountOf("stake"))
}
