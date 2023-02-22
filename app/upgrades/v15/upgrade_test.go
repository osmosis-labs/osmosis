package v15_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v14/app/apptesting"
	v15 "github.com/osmosis-labs/osmosis/v14/app/upgrades/v15"
	gamm "github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	balancer "github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

var DefaultAcctFunds sdk.Coins = sdk.NewCoins(
	sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
	sdk.NewCoin("foo", sdk.NewInt(10000000)),
	sdk.NewCoin("bar", sdk.NewInt(10000000)),
	sdk.NewCoin("baz", sdk.NewInt(10000000)),
)

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestMigrateNextPoolIdAndCreatePool() {
	suite.SetupTest() // reset

	const (
		expectedNextPoolId uint64 = 1
	)

	var (
		gammKeeperType = reflect.TypeOf(&gamm.Keeper{})
	)

	ctx := suite.Ctx
	gammKeeper := suite.App.GAMMKeeper
	poolmanagerKeeper := suite.App.PoolManagerKeeper

	nextPoolId := gammKeeper.GetNextPoolId(ctx)
	suite.Require().Equal(expectedNextPoolId, nextPoolId)

	// system under test.
	v15.MigrateNextPoolId(ctx, gammKeeper, poolmanagerKeeper)

	// validate poolmanager's next pool id.
	actualNextPoolId := poolmanagerKeeper.GetNextPoolId(ctx)
	suite.Require().Equal(expectedNextPoolId, actualNextPoolId)

	// create a pool after migration.
	actualCreatedPoolId := suite.PrepareBalancerPool()
	suite.Require().Equal(expectedNextPoolId, actualCreatedPoolId)

	// validate that module route mapping has been created for each pool id.
	for poolId := uint64(1); poolId < expectedNextPoolId; poolId++ {
		swapModule, err := poolmanagerKeeper.GetPoolModule(ctx, poolId)
		suite.Require().NoError(err)

		suite.Require().Equal(gammKeeperType, reflect.TypeOf(swapModule))
	}

	// validate params
	gammPoolCreationFee := gammKeeper.GetParams(ctx).PoolCreationFee
	poolmanagerPoolCreationFee := poolmanagerKeeper.GetParams(ctx).PoolCreationFee
	suite.Require().Equal(gammPoolCreationFee, poolmanagerPoolCreationFee)
}


func (suite *UpgradeTestSuite) TestMigrateBalancerToStablePools() {
	suite.SetupTest() // reset

	ctx := suite.Ctx
	gammKeeper := suite.App.GAMMKeeper
	poolmanagerKeeper := suite.App.PoolManagerKeeper
	// bankKeeper := suite.App.BankKeeper
	testAccount := suite.TestAccs[0]

	// Mint some assets to the accounts.
	suite.FundAcc(testAccount, DefaultAcctFunds)

	// Create the balancer pool
	swapFee, err := sdk.NewDecFromStr("0.003")
	exitFee, err := sdk.NewDecFromStr("0.025")
	poolID, err := suite.App.PoolManagerKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[0],
		balancer.PoolParams{
			SwapFee: swapFee,
			ExitFee: exitFee,
		},
		[]balancertypes.PoolAsset{
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
			},
			{
				Weight: sdk.NewInt(200),
				Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
			},
		},
		""),
	)
	_ = poolID
	suite.Require().NoError(err)

	balancerPool, err := gammKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	balancerShares := balancerPool.GetTotalShares()

	// join pool
	sharesRequested := types.OneShare.MulRaw(1)
	tokenInMaxs := sdk.Coins{}
	_, _, err = gammKeeper.JoinPoolNoSwap(suite.Ctx, testAccount, poolID, sharesRequested, tokenInMaxs)
	suite.Require().Equal(sharesRequested.String(), bankKeeper.GetBalance(suite.Ctx, testAccount, "gamm/pool/1").Amount.String())
	liquidity := gammKeeper.GetTotalLiquidity(suite.Ctx)
	suite.Require().Equal("15000bar,15000foo", liquidity.String())

	// test migrating the balancer pool to a stable pool
	v15.MigrateBalancerPoolToSolidlyStable(ctx, gammKeeper, poolmanagerKeeper, poolID)

	// check that the pool is now a stable pool
	stablepool, err := gammKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(stablepool.GetType(), poolmanagertypes.Stableswap)

	// check that a user can withdraw from the pool

	// check that a user can swap in the pool

	// check that the number of stableswap LP shares is the same as the number of balancer LP shares
	suite.Require().Equal(balancerShares.String(), stablepool.GetTotalShares().String())

}

func (suite *UpgradeTestSuite) TestRegisterOsmoIonMetadata() {
	suite.SetupTest() // reset

	expectedUosmodenom := "uosmo"
	expectedUiondenom := "uion"

	ctx := suite.Ctx
	bankKeeper := suite.App.BankKeeper

	// meta data should not be found pre-registration of meta data
	uosmoMetadata, found := suite.App.BankKeeper.GetDenomMetaData(ctx, "uosmo")
	suite.Require().False(found)

	uionMetadata, found := suite.App.BankKeeper.GetDenomMetaData(ctx, "uion")
	suite.Require().False(found)

	// system under test.
	v15.RegisterOsmoIonMetadata(ctx, *bankKeeper)

	uosmoMetadata, found = suite.App.BankKeeper.GetDenomMetaData(ctx, "uosmo")
	suite.Require().True(found)

	uionMetadata, found = suite.App.BankKeeper.GetDenomMetaData(ctx, "uion")
	suite.Require().True(found)

	suite.Require().Equal(expectedUosmodenom, uosmoMetadata.Base)
	suite.Require().Equal(expectedUiondenom, uionMetadata.Base)
}

func (suite *UpgradeTestSuite) TestSetICQParams() {
	suite.SetupTest() // reset

	// system under test.
	v15.SetICQParams(suite.Ctx, suite.App.ICQKeeper)

	suite.Require().True(suite.App.ICQKeeper.IsHostEnabled(suite.Ctx))
	suite.Require().Len(suite.App.ICQKeeper.GetAllowQueries(suite.Ctx), 63)
}
