package osmosisibctesting

import (
	"fmt"
	"os"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/require"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v23/x/ibc-rate-limit/types"
)

func (chain *TestChain) StoreContractCode(suite *suite.Suite, path string) {
	chain.StoreContractCodeDirect(suite, path)
}

func (chain *TestChain) InstantiateRLContract(suite *suite.Suite, quotas string) sdk.AccAddress {
	symphonyApp := chain.GetSymphonyApp()
	transferModule := symphonyApp.AccountKeeper.GetModuleAddress(transfertypes.ModuleName)
	govModule := symphonyApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	initMsgBz := []byte(fmt.Sprintf(`{
           "gov_module":  "%s",
           "ibc_module":"%s",
           "paths": [%s]
        }`,
		govModule, transferModule, quotas))

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(symphonyApp.WasmKeeper)
	codeID := uint64(1)
	creator := symphonyApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(chain.GetContext(), codeID, creator, creator, initMsgBz, "rate limiting contract", nil)
	suite.Require().NoError(err)
	return addr
}

func (chain *TestChain) StoreContractCodeDirect(suite *suite.Suite, path string) uint64 {
	symphonyApp := chain.GetSymphonyApp()
	govKeeper := wasmkeeper.NewGovPermissionKeeper(symphonyApp.WasmKeeper)
	creator := symphonyApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	wasmCode, err := os.ReadFile(path)
	suite.Require().NoError(err)
	accessEveryone := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody}
	codeID, _, err := govKeeper.Create(chain.GetContext(), creator, wasmCode, &accessEveryone)
	suite.Require().NoError(err)
	return codeID
}

func (chain *TestChain) InstantiateContract(suite *suite.Suite, msg string, codeID uint64) sdk.AccAddress {
	symphonyApp := chain.GetSymphonyApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(symphonyApp.WasmKeeper)
	creator := symphonyApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(chain.GetContext(), codeID, creator, creator, []byte(msg), "contract", nil)
	suite.Require().NoError(err)
	return addr
}

func (chain *TestChain) QueryContract(suite *suite.Suite, contract sdk.AccAddress, key []byte) string {
	symphonyApp := chain.GetSymphonyApp()
	state, err := symphonyApp.WasmKeeper.QuerySmart(chain.GetContext(), contract, key)
	suite.Require().NoError(err)
	return string(state)
}

func (chain *TestChain) QueryContractJson(suite *suite.Suite, contract sdk.AccAddress, key []byte) gjson.Result {
	symphonyApp := chain.GetSymphonyApp()
	state, err := symphonyApp.WasmKeeper.QuerySmart(chain.GetContext(), contract, key)
	suite.Require().NoError(err)
	suite.Require().True(gjson.Valid(string(state)))
	json := gjson.Parse(string(state))
	suite.Require().NoError(err)
	return json
}

func (chain *TestChain) RegisterRateLimitingContract(addr []byte) {
	addrStr, err := sdk.Bech32ifyAddressBytes("symphony", addr)
	require.NoError(chain.T, err)
	params, err := types.NewParams(addrStr)
	require.NoError(chain.T, err)
	symphonyApp := chain.GetSymphonyApp()
	paramSpace, ok := symphonyApp.AppKeepers.ParamsKeeper.GetSubspace(types.ModuleName)
	require.True(chain.T, ok)
	paramSpace.SetParamSet(chain.GetContext(), &params)
}
