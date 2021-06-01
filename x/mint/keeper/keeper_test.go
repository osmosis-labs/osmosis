package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/osmosis-labs/osmosis/app"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.OsmosisApp
	ctx sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, Time: time.Now().UTC()})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestMintCoinsToFeeCollectorAndGetProportions() {
	mintKeeper := suite.app.MintKeeper

	// When coin is minted to the fee collector
	fees := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(0)))
	coin := mintKeeper.GetProportions(suite.ctx, fees, sdk.NewDecWithPrec(2, 1))
	suite.Equal("0stake", coin.String())

	// When mint the 100K stake coin to the fee collector
	fees = sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000)))
	err := suite.app.BankKeeper.AddCoins(
		suite.ctx,
		suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName),
		fees,
	)
	suite.NoError(err)

	// check propotion for 20%
	coin = mintKeeper.GetProportions(suite.ctx, fees, sdk.NewDecWithPrec(2, 1))
	suite.Equal(fees[0].Amount.Quo(sdk.NewInt(5)), coin.Amount)
}

func (suite *KeeperTestSuite) TestDistrAssetToDeveloperRewardsAddrWhenNotEmpty() {
	mintKeeper := suite.app.MintKeeper
	params := suite.app.MintKeeper.GetParams(suite.ctx)
	devRewardsReceiver := sdk.AccAddress([]byte("addr1---------------"))
	potCreator := sdk.AccAddress([]byte("addr2---------------"))
	params.DeveloperRewardsReceiver = devRewardsReceiver.String()
	suite.app.MintKeeper.SetParams(suite.ctx, params)

	// Create record
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	suite.app.BankKeeper.SetBalances(suite.ctx, potCreator, coins)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	potId, err := suite.app.IncentivesKeeper.CreatePot(suite.ctx, true, potCreator, coins, distrTo, time.Now(), 1)
	suite.NoError(err)
	err = suite.app.PoolIncentivesKeeper.UpdateDistrRecords(suite.ctx, poolincentivestypes.DistrRecord{
		PotId:  potId,
		Weight: sdk.NewInt(100),
	})
	suite.NoError(err)

	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoins := sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err = mintKeeper.DistributeMintedCoins(suite.ctx, mintCoins)
	suite.NoError(err)

	feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("40000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal("10000.000000000000000000stake", feePool.CommunityPool.String())
	suite.Equal("20000stake", suite.app.BankKeeper.GetBalance(suite.ctx, devRewardsReceiver, "stake").String())
}

func (suite *KeeperTestSuite) TestDistrAssetToCommunityPoolWhenNoDeveloperRewardsAddr() {
	mintKeeper := suite.app.MintKeeper

	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoins := sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err := mintKeeper.DistributeMintedCoins(suite.ctx, mintCoins)
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, suite.app.DistrKeeper)

	feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("40000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(60000))).String(), feePool.CommunityPool.String())
	suite.Equal("60000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName), "stake").String())

	// Community pool should be increased
	mintCoins = sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(100000))}
	mintKeeper.MintCoins(suite.ctx, mintCoins)
	err = mintKeeper.DistributeMintedCoins(suite.ctx, mintCoins)
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, suite.app.DistrKeeper)

	feePool = suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal("80000stake", suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), "stake").String())
	suite.Equal(feePool.CommunityPool.String(), sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdk.NewInt(120000))).String())
	suite.Equal(sdk.NewCoin("stake", sdk.NewInt(120000)), suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName), "stake"))
}
