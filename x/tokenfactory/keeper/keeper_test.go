package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient    types.QueryClient
	msgServer      types.MsgServer
	contractKeeper wasmtypes.ContractOpsKeeper
	bankMsgServer  banktypes.MsgServer

	// defaultDenom is on the suite, as it depends on the creator test address.
	defaultDenom string
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

type SudoAuthorizationPolicy struct{}

func (p SudoAuthorizationPolicy) CanCreateCode(chainAccesscoConfig wasmtypes.ChainAccessConfigs, actor sdk.AccAddress, config wasmtypes.AccessConfig) bool {
	return true
}

func (p SudoAuthorizationPolicy) CanInstantiateContract(config wasmtypes.AccessConfig, actor sdk.AccAddress) bool {
	return true
}

func (p SudoAuthorizationPolicy) CanModifyContract(admin, actor sdk.AccAddress) bool {
	return true
}

func (p SudoAuthorizationPolicy) CanModifyCodeAccessConfig(creator, actor sdk.AccAddress, isSubset bool) bool {
	return true
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	// Fund every TestAcc with two denoms, one of which is the denom creation fee
	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(apptesting.SecondaryDenom, apptesting.SecondaryAmount))
	for _, acc := range s.TestAccs {
		s.FundAcc(acc, fundAccsAmount)
	}
	s.contractKeeper = wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.App.TokenFactoryKeeper)
	s.bankMsgServer = bankkeeper.NewMsgServerImpl(*s.App.BankKeeper)
}

func (s *KeeperTestSuite) CreateDefaultDenom() {
	res, _ := s.msgServer.CreateDenom(s.Ctx, types.NewMsgCreateDenom(s.TestAccs[0].String(), "bitcoin"))
	s.defaultDenom = res.GetNewTokenDenom()
}

func (s *KeeperTestSuite) TestCreateModuleAccount() {
	app := s.App

	// setup new next account number
	nextAccountNumber := app.AccountKeeper.NextAccountNumber(s.Ctx)

	// remove module account
	tokenfactoryModuleAccount := app.AccountKeeper.GetAccount(s.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	app.AccountKeeper.RemoveAccount(s.Ctx, tokenfactoryModuleAccount)

	// ensure module account was removed
	s.Ctx = app.BaseApp.NewContextLegacy(false, tmproto.Header{})
	tokenfactoryModuleAccount = app.AccountKeeper.GetAccount(s.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().Nil(tokenfactoryModuleAccount)

	// create module account
	app.TokenFactoryKeeper.CreateModuleAccount(s.Ctx)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = app.AccountKeeper.GetAccount(s.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().NotNil(tokenfactoryModuleAccount)

	// check that the account number of the module account is now initialized correctly
	s.Require().Equal(nextAccountNumber+1, tokenfactoryModuleAccount.GetAccountNumber())
}
