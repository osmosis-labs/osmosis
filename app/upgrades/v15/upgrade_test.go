package v15_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v19/x/ibc-rate-limit/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	v15 "github.com/osmosis-labs/osmosis/v19/app/upgrades/v15"
	gamm "github.com/osmosis-labs/osmosis/v19/x/gamm/keeper"
	balancer "github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

var DefaultAcctFunds sdk.Coins = sdk.NewCoins(
	sdk.NewCoin("uosmo", osmomath.NewInt(10000000000)),
	sdk.NewCoin("foo", osmomath.NewInt(10000000)),
	sdk.NewCoin("bar", osmomath.NewInt(10000000)),
	sdk.NewCoin("baz", osmomath.NewInt(10000000)),
)

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
	suite.SkipIfWSL()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestMigrateNextPoolIdAndCreatePool() {
	suite.SetupTest() // reset

	const (
		expectedNextPoolId uint64 = 1
	)

	gammKeeperType := reflect.TypeOf(&gamm.Keeper{})

	ctx := suite.Ctx
	gammKeeper := suite.App.GAMMKeeper
	poolmanagerKeeper := suite.App.PoolManagerKeeper

	nextPoolId := gammKeeper.GetNextPoolId(ctx) //nolint:staticcheck // we're using the deprecated version for testing.
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
	// bankKeeper := suite.App.BankKeeper
	testAccount := suite.TestAccs[0]

	// Mint some assets to the accounts.
	suite.FundAcc(testAccount, DefaultAcctFunds)

	// Create the balancer pool
	spreadFactor := osmomath.MustNewDecFromStr("0.003")
	exitFee := osmomath.ZeroDec()
	poolID, err := suite.App.PoolManagerKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[0],
			balancer.PoolParams{
				SwapFee: spreadFactor,
				ExitFee: exitFee,
			},
			[]balancer.PoolAsset{
				{
					Weight: osmomath.NewInt(100),
					Token:  sdk.NewCoin("foo", osmomath.NewInt(5000000)),
				},
				{
					Weight: osmomath.NewInt(200),
					Token:  sdk.NewCoin("bar", osmomath.NewInt(5000000)),
				},
			},
			""),
	)
	suite.Require().NoError(err)

	// join the pool
	shareOutAmount := osmomath.NewInt(1_000_000_000_000_000)
	tokenInMaxs := sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(5000000)), sdk.NewCoin("bar", osmomath.NewInt(5000000)))
	tokenIn, sharesOut, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, testAccount, poolID, shareOutAmount, tokenInMaxs)
	suite.Require().NoError(err)

	// shares before migration
	balancerPool, err := gammKeeper.GetCFMMPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	balancerLiquidity, err := gammKeeper.GetTotalPoolLiquidity(suite.Ctx, balancerPool.GetId())
	suite.Require().NoError(err)

	balancerShares := balancerPool.GetTotalShares()
	// check balancer pool liquidity using the bank module
	balancerBalances := suite.App.BankKeeper.GetAllBalances(ctx, balancerPool.GetAddress())

	// test migrating the balancer pool to a stable pool
	v15.MigrateBalancerPoolToSolidlyStable(ctx, gammKeeper, suite.App.BankKeeper, poolID)

	// check that the pool is now a stable pool
	stablepool, err := gammKeeper.GetCFMMPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(stablepool.GetType(), poolmanagertypes.Stableswap)

	// check that the number of stableswap LP shares is the same as the number of balancer LP shares
	suite.Require().Equal(balancerShares.String(), stablepool.GetTotalShares().String())
	// check that the pool liquidity is the same
	stableLiquidity, err := gammKeeper.GetTotalPoolLiquidity(suite.Ctx, balancerPool.GetId())
	suite.Require().NoError(err)
	suite.Require().Equal(balancerLiquidity.String(), stableLiquidity.String())
	// check pool liquidity using the bank module
	stableBalances := suite.App.BankKeeper.GetAllBalances(ctx, stablepool.GetAddress())
	suite.Require().Equal(balancerBalances, stableBalances)

	// exit the pool
	exitCoins, err := suite.App.GAMMKeeper.ExitPool(suite.Ctx, testAccount, poolID, sharesOut, sdk.NewCoins())
	suite.Require().NoError(err)

	suite.validateCons(exitCoins, tokenIn)

	// join again
	tokenInStable, _, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, testAccount, poolID, shareOutAmount, tokenInMaxs)
	suite.Require().NoError(err)

	suite.validateCons(tokenInStable, tokenIn)
}

func (suite *UpgradeTestSuite) TestRegisterOsmoIonMetadata() {
	suite.SetupTest() // reset

	expectedUosmodenom := "uosmo"
	expectedUiondenom := "uion"

	ctx := suite.Ctx
	bankKeeper := suite.App.BankKeeper

	// meta data should not be found pre-registration of meta data
	_, found := suite.App.BankKeeper.GetDenomMetaData(ctx, "uosmo")
	suite.Require().False(found)

	_, found = suite.App.BankKeeper.GetDenomMetaData(ctx, "uion")
	suite.Require().False(found)

	// system under test.
	v15.RegisterOsmoIonMetadata(ctx, *bankKeeper)

	uosmoMetadata, found := suite.App.BankKeeper.GetDenomMetaData(ctx, "uosmo")
	suite.Require().True(found)

	uionMetadata, found := suite.App.BankKeeper.GetDenomMetaData(ctx, "uion")
	suite.Require().True(found)

	suite.Require().Equal(expectedUosmodenom, uosmoMetadata.Base)
	suite.Require().Equal(expectedUiondenom, uionMetadata.Base)
}

func (suite *UpgradeTestSuite) TestSetICQParams() {
	suite.SetupTest() // reset

	// system under test.
	v15.SetICQParams(suite.Ctx, suite.App.ICQKeeper)

	suite.Require().True(suite.App.ICQKeeper.IsHostEnabled(suite.Ctx))
	// commented out for historical reasons since v15 upgrade is now over.
	// suite.Require().Len(suite.App.ICQKeeper.GetAllowQueries(suite.Ctx), 65)
}

func (suite *UpgradeTestSuite) TestSetRateLimits() {
	suite.SetupTest() // reset
	accountKeeper := suite.App.AccountKeeper
	govModule := accountKeeper.GetModuleAddress(govtypes.ModuleName)

	code, err := os.ReadFile("../v13/rate_limiter.wasm")
	suite.Require().NoError(err)
	contractKeeper := wasmkeeper.NewGovPermissionKeeper(suite.App.WasmKeeper)
	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeOnlyAddress, Address: govModule.String()}
	codeID, _, err := contractKeeper.Create(suite.Ctx, govModule, code, &instantiateConfig)
	suite.Require().NoError(err)
	transferModule := accountKeeper.GetModuleAddress(transfertypes.ModuleName)
	initMsgBz := []byte(fmt.Sprintf(`{
           "gov_module":  "%s",
           "ibc_module":"%s",
           "paths": []
        }`,
		govModule, transferModule))

	addr, _, err := contractKeeper.Instantiate(suite.Ctx, codeID, govModule, govModule, initMsgBz, "rate limiting contract", nil)
	suite.Require().NoError(err)
	addrStr, err := sdk.Bech32ifyAddressBytes("osmo", addr)
	suite.Require().NoError(err)
	params, err := ibcratelimittypes.NewParams(addrStr)
	suite.Require().NoError(err)
	paramSpace, ok := suite.App.ParamsKeeper.GetSubspace(ibcratelimittypes.ModuleName)
	suite.Require().True(ok)
	paramSpace.SetParamSet(suite.Ctx, &params)

	// system under test.
	v15.SetRateLimits(suite.Ctx, accountKeeper, suite.App.RateLimitingICS4Wrapper, suite.App.WasmKeeper)

	state, err := suite.App.WasmKeeper.QuerySmart(suite.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"}}`))
	suite.Require().Greaterf(len(state), 0, "state should not be empty")
	suite.Require().NoError(err)

	state, err = suite.App.WasmKeeper.QuerySmart(suite.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"}}`))
	suite.Require().Greaterf(len(state), 0, "state should not be empty")
	suite.Require().NoError(err)

	// This is the last one. If the others failed the upgrade would've panicked before adding this one
	state, err = suite.App.WasmKeeper.QuerySmart(suite.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1"}}`))
	suite.Require().Greaterf(len(state), 0, "state should not be empty")
	suite.Require().NoError(err)
}

func (suite *UpgradeTestSuite) validateCons(coinsA, coinsB sdk.Coins) {
	suite.Require().Equal(len(coinsA), len(coinsB))
	for _, coinA := range coinsA {
		coinBAmount := coinsB.AmountOf(coinA.Denom)
		// minor tolerance due to fees and rounding
		osmoassert.DecApproxEq(suite.T(), coinBAmount.ToLegacyDec(), coinA.Amount.ToLegacyDec(), osmomath.NewDec(2))
	}
}
