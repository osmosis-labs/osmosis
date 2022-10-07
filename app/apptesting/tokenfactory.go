package apptesting

import (
	"io/ioutil"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v10/x/tokenfactory/types"
)

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

func (s *KeeperTestHelper) SetBasicTokenFactoryDenom() string {
	s.App.TokenFactoryKeeper.CreateModuleAccount(s.Ctx)

	s.contractKeeper = wasmkeeper.NewPermissionedKeeper(s.App.WasmKeeper, SudoAuthorizationPolicy{})

	// fund account with token creation fee
	s.FundAcc(s.TestAccs[0], sdk.Coins{sdk.NewInt64Coin("uosmo", 10000000)})
	tokenFactoryMsgSerer := tokenfactorykeeper.NewMsgServerImpl(*s.App.TokenFactoryKeeper)
	res, err := tokenFactoryMsgSerer.CreateDenom(sdk.WrapSDKContext(s.Ctx), tokenfactorytypes.NewMsgCreateDenom(s.TestAccs[0].String(), "tokenfactorydenom"))
	s.Require().NoError(err)
	newDenom := res.GetNewTokenDenom()

	// mint new coins to the creator
	_, err = tokenFactoryMsgSerer.Mint(sdk.WrapSDKContext(s.Ctx), tokenfactorytypes.NewMsgMint(s.TestAccs[0].String(), sdk.NewInt64Coin(newDenom, 1000000000)))
	s.Require().NoError(err)
	return newDenom
}

func (s *KeeperTestHelper) SetBasicTokenFacotryListener(denom string) {
	// TODO: change this to relative path
	wasmFile := "../../../app/apptesting/testdata/no100.wasm"
	wasmCode, err := ioutil.ReadFile(wasmFile)
	s.Require().NoError(err)

	codeID, err := s.contractKeeper.Create(s.Ctx, s.TestAccs[0], wasmCode, nil)
	s.Require().NoError(err)

	cosmwasmAddress, _, err := s.contractKeeper.Instantiate(s.Ctx, codeID, s.TestAccs[0], s.TestAccs[0], []byte("{}"), "", sdk.NewCoins())
	s.Require().NoError(err)
	tokenFactoryMsgSerer := tokenfactorykeeper.NewMsgServerImpl(*s.App.TokenFactoryKeeper)

	_, err = tokenFactoryMsgSerer.SetBeforeSendListener(sdk.WrapSDKContext(s.Ctx), tokenfactorytypes.NewMsgSetBeforeSendListener(s.TestAccs[0].String(), denom, cosmwasmAddress.String()))
	s.Require().NoError(err)
}
