package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"
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

func (p SudoAuthorizationPolicy) CanCreateCode(config wasmtypes.AccessConfig, actor sdk.AccAddress) bool {
	return true
}

func (p SudoAuthorizationPolicy) CanInstantiateContract(config wasmtypes.AccessConfig, actor sdk.AccAddress) bool {
	return true
}

func (p SudoAuthorizationPolicy) CanModifyContract(admin, actor sdk.AccAddress) bool {
	return true
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	// Fund every TestAcc with 100 denom creation fees.
	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, fundAccsAmount)
	}

	suite.SetupTokenFactory()

	suite.contractKeeper = wasmkeeper.NewPermissionedKeeper(suite.App.WasmKeeper, SudoAuthorizationPolicy{})

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
	suite.msgServer = keeper.NewMsgServerImpl(*suite.App.TokenFactoryKeeper)

	suite.bankMsgServer = bankkeeper.NewMsgServerImpl(suite.App.BankKeeper)
}

func (suite *KeeperTestSuite) CreateDefaultDenom() {
	res, _ := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.defaultDenom = res.GetNewTokenDenom()
}
