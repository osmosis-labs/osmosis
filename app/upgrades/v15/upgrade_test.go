package v15_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"testing"
	// "time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v15/x/ibc-rate-limit/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	v15 "github.com/osmosis-labs/osmosis/v15/app/upgrades/v15"
	gamm "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	balancer "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	// oldbalancer "github.com/osmosis-labs/osmosis/v15/x/gamm/v2types/balancer"
	// oldstableswap "github.com/osmosis-labs/osmosis/v15/x/gamm/v2types/stableswap"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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

	gammKeeperType := reflect.TypeOf(&gamm.Keeper{})

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
	swapFee := sdk.MustNewDecFromStr("0.003")
	poolID, err := suite.App.PoolManagerKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[0],
			balancer.PoolParams{
				SwapFee: swapFee,
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
	suite.Require().NoError(err)

	// join the pool
	shareOutAmount := sdk.NewInt(1_000_000_000_000_000)
	tokenInMaxs := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(5000000)), sdk.NewCoin("bar", sdk.NewInt(5000000)))
	tokenIn, sharesOut, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, testAccount, poolID, shareOutAmount, tokenInMaxs)
	suite.Require().NoError(err)

	// shares before migration
	balancerPool, err := gammKeeper.GetPool(suite.Ctx, poolID)
	suite.Require().NoError(err)
	balancerShares := balancerPool.GetTotalShares()
	balancerLiquidity := balancerPool.GetTotalPoolLiquidity(ctx).String()
	// check balancer pool liquidity using the bank module
	balancerBalances := suite.App.BankKeeper.GetAllBalances(ctx, balancerPool.GetAddress())

	// test migrating the balancer pool to a stable pool
	v15.MigrateBalancerPoolToSolidlyStable(ctx, gammKeeper, poolmanagerKeeper, suite.App.BankKeeper, poolID)

	// check that the pool is now a stable pool
	stablepool, err := gammKeeper.GetPool(ctx, poolID)
	suite.Require().NoError(err)
	suite.Require().Equal(stablepool.GetType(), poolmanagertypes.Stableswap)
	// check that the number of stableswap LP shares is the same as the number of balancer LP shares
	suite.Require().Equal(balancerShares.String(), stablepool.GetTotalShares().String())
	// check that the pool liquidity is the same
	suite.Require().Equal(balancerLiquidity, stablepool.GetTotalPoolLiquidity(ctx).String())
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

	state, err = suite.App.WasmKeeper.QuerySmart(suite.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"}}`))
	suite.Require().Greaterf(len(state), 0, "state should not be empty")

	// This is the last one. If the others failed the upgrade would've panicked before adding this one
	state, err = suite.App.WasmKeeper.QuerySmart(suite.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1"}}`))
	suite.Require().Greaterf(len(state), 0, "state should not be empty")
}

func (suite *UpgradeTestSuite) validateCons(coinsA, coinsB sdk.Coins) {
	suite.Require().Equal(len(coinsA), len(coinsB))
	for _, coinA := range coinsA {
		coinBAmount := coinsB.AmountOf(coinA.Denom)
		// minor tolerance due to fees and rounding
		osmoassert.DecApproxEq(suite.T(), coinBAmount.ToDec(), coinA.Amount.ToDec(), sdk.NewDec(2))
	}
}

func (suite *UpgradeTestSuite) TestRemoveExitFee() {
	suite.SetupTest() // reset

	store := suite.Ctx.KVStore(suite.App.GAMMKeeper.GetStoreKey(suite.Ctx))

	// Set up balancer pool 2 with zero exit fee
	pool2Bz, err := hex.DecodeString("0a1a2f6f736d6f7369732e67616d6d2e763162657461312e506f6f6c12d7010a3f6f736d6f316d773061633672776c70357238776170776b337a73366732396838666373637871616b647a7739656d6b6e65366338776a70397130743376387410011a260a113130303030303030303030303030303030121131303030303030303030303030303030302a240a0b67616d6d2f706f6f6c2f311215313030303030303030303030303030303030303030321d0a0f0a096e6f6465746f6b656e12023130120a3130373337343138323432190a0b0a057374616b6512023130120a313037333734313832343a0a32313437343833363438")
	store.Set(gammtypes.GetKeyPrefixPools(1), pool2Bz)
	
	// Set up balancer pool 553 with non zero exit fee
	pool553Bz, err := hex.DecodeString("0a1a2f6f736d6f7369732e67616d6d2e763162657461312e506f6f6c12de010a3f6f736d6f316d3830666e71767673643833776538676e68393938726535356a71386371326d646b6363636c716d7571773878706d36396167736d386666643610a9041a260a113235303030303030303030303030303030121132353030303030303030303030303030302a260a0d67616d6d2f706f6f6c2f3535331215313030303030303030303030303030303030303030321c0a0c0a0362617212053130303030120c313037333734313832343030321c0a0c0a03666f6f12053130303030120c3130373337343138323430303a0c323134373438333634383030")
	store.Set(gammtypes.GetKeyPrefixPools(553), pool553Bz)

	// Set up stableswap pool 596 with non zero exit fee
	pool596Bz, err := hex.DecodeString("0a302f6f736d6f7369732e67616d6d2e706f6f6c6d6f64656c732e737461626c65737761702e763162657461312e506f6f6c12b4010a3f6f736d6f316a747a6b7a32333833636567676138707a7137617a6d377470336c63757465703935757270767571787a3378383573667077377373617170633510d4041a260a113130303030303030303030303030303030121131303030303030303030303030303030302a260a0d67616d6d2f706f6f6c2f3539361215313030303030303030303030303030303030303030320c0a0362617212053130303030320c0a03666f6f120531303030303a020101")
	store.Set(gammtypes.GetKeyPrefixPools(596), pool596Bz)

	// Pool with zero exit fee should not be removed
	pool2, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, 1)
	fmt.Println("pool2", pool2, err)
	suite.Require().NoError(err)
	suite.Require().NotNil(pool2)
	fmt.Println("pool2", pool2)

	// Pool 553 & 596 should be removed
	pool553, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, 553)
	suite.Require().Error(err)
	suite.Require().Nil(pool553)

	pool596, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, 596)
	suite.Require().Error(err)
	suite.Require().Nil(pool596)
}
