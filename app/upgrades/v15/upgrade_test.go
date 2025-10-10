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
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	ibcratelimittypes "github.com/osmosis-labs/osmosis/v31/x/ibc-rate-limit/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v31/app/params"
	v15 "github.com/osmosis-labs/osmosis/v31/app/upgrades/v15"
	gamm "github.com/osmosis-labs/osmosis/v31/x/gamm/keeper"
	balancer "github.com/osmosis-labs/osmosis/v31/x/gamm/pool-models/balancer"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

var DefaultAcctFunds sdk.Coins = sdk.NewCoins(
	sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10000000000)),
	sdk.NewCoin("foo", osmomath.NewInt(10000000)),
	sdk.NewCoin("bar", osmomath.NewInt(10000000)),
	sdk.NewCoin("baz", osmomath.NewInt(10000000)),
)

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
	s.SkipIfWSL()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestMigrateNextPoolIdAndCreatePool() {
	s.SetupTest() // reset

	const (
		expectedNextPoolId uint64 = 1
	)

	gammKeeperType := reflect.TypeOf(&gamm.Keeper{})

	ctx := s.Ctx
	gammKeeper := s.App.GAMMKeeper
	poolmanagerKeeper := s.App.PoolManagerKeeper

	nextPoolId := gammKeeper.GetNextPoolId(ctx) //nolint:staticcheck // we're using the deprecated version for testing.
	s.Require().Equal(expectedNextPoolId, nextPoolId)

	// system under test.
	v15.MigrateNextPoolId(ctx, gammKeeper, poolmanagerKeeper)

	// validate poolmanager's next pool id.
	actualNextPoolId := poolmanagerKeeper.GetNextPoolId(ctx)
	s.Require().Equal(expectedNextPoolId, actualNextPoolId)

	// create a pool after migration.
	actualCreatedPoolId := s.PrepareBalancerPool()
	s.Require().Equal(expectedNextPoolId, actualCreatedPoolId)

	// validate that module route mapping has been created for each pool id.
	for poolId := uint64(1); poolId < expectedNextPoolId; poolId++ {
		swapModule, err := poolmanagerKeeper.GetPoolModule(ctx, poolId)
		s.Require().NoError(err)

		s.Require().Equal(gammKeeperType, reflect.TypeOf(swapModule))
	}

	// validate params
	gammPoolCreationFee := gammKeeper.GetParams(ctx).PoolCreationFee
	poolmanagerPoolCreationFee := poolmanagerKeeper.GetParams(ctx).PoolCreationFee
	s.Require().Equal(gammPoolCreationFee, poolmanagerPoolCreationFee)
}

func (s *UpgradeTestSuite) TestMigrateBalancerToStablePools() {
	s.SetupTest() // reset

	ctx := s.Ctx
	gammKeeper := s.App.GAMMKeeper
	// bankKeeper := s.App.BankKeeper
	testAccount := s.TestAccs[0]

	// Mint some assets to the accounts.
	s.FundAcc(testAccount, DefaultAcctFunds)

	// Create the balancer pool
	spreadFactor := osmomath.MustNewDecFromStr("0.003")
	exitFee := osmomath.ZeroDec()
	poolID, err := s.App.PoolManagerKeeper.CreatePool(
		s.Ctx,
		balancer.NewMsgCreateBalancerPool(s.TestAccs[0],
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
	s.Require().NoError(err)

	// join the pool
	shareOutAmount := osmomath.NewInt(1_000_000_000_000_000)
	tokenInMaxs := sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(5000000)), sdk.NewCoin("bar", osmomath.NewInt(5000000)))
	tokenIn, sharesOut, err := s.App.GAMMKeeper.JoinPoolNoSwap(s.Ctx, testAccount, poolID, shareOutAmount, tokenInMaxs)
	s.Require().NoError(err)

	// shares before migration
	balancerPool, err := gammKeeper.GetCFMMPool(s.Ctx, poolID)
	s.Require().NoError(err)
	balancerLiquidity, err := gammKeeper.GetTotalPoolLiquidity(s.Ctx, balancerPool.GetId())
	s.Require().NoError(err)

	balancerShares := balancerPool.GetTotalShares()
	// check balancer pool liquidity using the bank module
	balancerBalances := s.App.BankKeeper.GetAllBalances(ctx, balancerPool.GetAddress())

	// test migrating the balancer pool to a stable pool
	v15.MigrateBalancerPoolToSolidlyStable(ctx, gammKeeper, s.App.BankKeeper, poolID)

	// check that the pool is now a stable pool
	stablepool, err := gammKeeper.GetCFMMPool(ctx, poolID)
	s.Require().NoError(err)
	s.Require().Equal(stablepool.GetType(), poolmanagertypes.Stableswap)

	// check that the number of stableswap LP shares is the same as the number of balancer LP shares
	s.Require().Equal(balancerShares.String(), stablepool.GetTotalShares().String())
	// check that the pool liquidity is the same
	stableLiquidity, err := gammKeeper.GetTotalPoolLiquidity(s.Ctx, balancerPool.GetId())
	s.Require().NoError(err)
	s.Require().Equal(balancerLiquidity.String(), stableLiquidity.String())
	// check pool liquidity using the bank module
	stableBalances := s.App.BankKeeper.GetAllBalances(ctx, stablepool.GetAddress())
	s.Require().Equal(balancerBalances, stableBalances)

	// exit the pool
	exitCoins, err := s.App.GAMMKeeper.ExitPool(s.Ctx, testAccount, poolID, sharesOut, sdk.NewCoins())
	s.Require().NoError(err)

	s.validateCons(exitCoins, tokenIn)

	// join again
	tokenInStable, _, err := s.App.GAMMKeeper.JoinPoolNoSwap(s.Ctx, testAccount, poolID, shareOutAmount, tokenInMaxs)
	s.Require().NoError(err)

	s.validateCons(tokenInStable, tokenIn)
}

func (s *UpgradeTestSuite) TestRegisterOsmoIonMetadata() {
	s.SetupTest() // reset

	expectedUosmodenom := appparams.BaseCoinUnit
	expectedUiondenom := "uion"

	ctx := s.Ctx
	bankKeeper := s.App.BankKeeper

	// meta data should not be found pre-registration of meta data
	_, found := s.App.BankKeeper.GetDenomMetaData(ctx, appparams.BaseCoinUnit)
	s.Require().False(found)

	_, found = s.App.BankKeeper.GetDenomMetaData(ctx, "uion")
	s.Require().False(found)

	// system under test.
	v15.RegisterOsmoIonMetadata(ctx, bankKeeper)

	uosmoMetadata, found := s.App.BankKeeper.GetDenomMetaData(ctx, appparams.BaseCoinUnit)
	s.Require().True(found)

	uionMetadata, found := s.App.BankKeeper.GetDenomMetaData(ctx, "uion")
	s.Require().True(found)

	s.Require().Equal(expectedUosmodenom, uosmoMetadata.Base)
	s.Require().Equal(expectedUiondenom, uionMetadata.Base)
}

func (s *UpgradeTestSuite) TestSetICQParams() {
	s.SetupTest() // reset

	// system under test.
	v15.SetICQParams(s.Ctx, s.App.ICQKeeper)

	s.Require().True(s.App.ICQKeeper.IsHostEnabled(s.Ctx))
	// commented out for historical reasons since v15 upgrade is now over.
	// s.Require().Len(s.App.ICQKeeper.GetAllowQueries(s.Ctx), 65)
}

func (s *UpgradeTestSuite) TestSetRateLimits() {
	s.SetupTest() // reset
	accountKeeper := s.App.AccountKeeper
	govModule := accountKeeper.GetModuleAddress(govtypes.ModuleName)

	code, err := os.ReadFile("../v13/rate_limiter.wasm")
	s.Require().NoError(err)
	contractKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	instantiateConfig := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeAnyOfAddresses, Addresses: []string{govModule.String()}}
	codeID, _, err := contractKeeper.Create(s.Ctx, govModule, code, &instantiateConfig)
	s.Require().NoError(err)
	transferModule := accountKeeper.GetModuleAddress(transfertypes.ModuleName)
	initMsgBz := []byte(fmt.Sprintf(`{
           "gov_module":  "%s",
           "ibc_module":"%s",
           "paths": []
        }`,
		govModule, transferModule))

	addr, _, err := contractKeeper.Instantiate(s.Ctx, codeID, govModule, govModule, initMsgBz, "rate limiting contract", nil)
	s.Require().NoError(err)
	addrStr, err := sdk.Bech32ifyAddressBytes("osmo", addr)
	s.Require().NoError(err)
	params, err := ibcratelimittypes.NewParams(addrStr)
	s.Require().NoError(err)
	paramSpace, ok := s.App.ParamsKeeper.GetSubspace(ibcratelimittypes.ModuleName)
	s.Require().True(ok)
	paramSpace.SetParamSet(s.Ctx, &params)

	// system under test.
	v15.SetRateLimits(s.Ctx, accountKeeper, s.App.RateLimitingICS4Wrapper, s.App.WasmKeeper)

	state, err := s.App.WasmKeeper.QuerySmart(s.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"}}`))
	s.Require().Greaterf(len(state), 0, "state should not be empty")
	s.Require().NoError(err)

	state, err = s.App.WasmKeeper.QuerySmart(s.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858"}}`))
	s.Require().Greaterf(len(state), 0, "state should not be empty")
	s.Require().NoError(err)

	// This is the last one. If the others failed the upgrade would've panicked before adding this one
	state, err = s.App.WasmKeeper.QuerySmart(s.Ctx, addr, []byte(`{"get_quotas": {"channel_id": "any", "denom": "ibc/E6931F78057F7CC5DA0FD6CEF82FF39373A6E0452BF1FD76910B93292CF356C1"}}`))
	s.Require().Greaterf(len(state), 0, "state should not be empty")
	s.Require().NoError(err)
}

func (s *UpgradeTestSuite) validateCons(coinsA, coinsB sdk.Coins) {
	s.Require().Equal(len(coinsA), len(coinsB))
	for _, coinA := range coinsA {
		coinBAmount := coinsB.AmountOf(coinA.Denom)
		// minor tolerance due to fees and rounding
		osmoassert.DecApproxEq(s.T(), coinBAmount.ToLegacyDec(), coinA.Amount.ToLegacyDec(), osmomath.NewDec(2))
	}
}
