package osmosisibctesting

import (
	"io/ioutil"

	"github.com/stretchr/testify/require"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/osmosis-labs/osmosis/v12/x/ibc-rate-limit/types"
)

func (chain *TestChain) StoreContractCode(contractFileName string) error {
	osmosisApp := chain.GetOsmosisApp()

	govKeeper := osmosisApp.GovKeeper
	wasmCode, err := ioutil.ReadFile(contractFileName)
	if err != nil {
		return err
	}

	addr := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	src := wasmtypes.StoreCodeProposalFixture(func(p *wasmtypes.StoreCodeProposal) {
		p.RunAs = addr.String()
		p.WASMByteCode = wasmCode
	})

	// when stored
	storedProposal, err := govKeeper.SubmitProposal(chain.GetContext(), src, false)
	if err != nil {
		return err
	}

	// and proposal execute
	handler := govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
	err = handler(chain.GetContext(), storedProposal.GetContent())
	if err != nil {
		return err
	}
	return nil
}

func (chain *TestChain) InstantiateContract(codeId int, initMsgBz []byte, label string) (sdk.AccAddress, error) {
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	codeID := uint64(codeId)
	creator := osmosisApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(chain.GetContext(), codeID, creator, creator, initMsgBz, label, nil)
	return addr, err
}

func (chain *TestChain) RegisterRateLimitingContract(addr []byte) {
	addrStr, err := sdk.Bech32ifyAddressBytes("osmo", addr)
	require.NoError(chain.T, err)
	params, err := types.NewParams(addrStr)
	require.NoError(chain.T, err)
	osmosisApp := chain.GetOsmosisApp()
	paramSpace, ok := osmosisApp.AppKeepers.ParamsKeeper.GetSubspace(types.ModuleName)
	require.True(chain.T, ok)
	paramSpace.SetParamSet(chain.GetContext(), &params)
}

func (chain *TestChain) ExecuteContract(contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) ([]byte, error) {
	osmosisApp := chain.GetOsmosisApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(osmosisApp.WasmKeeper)
	result, err := contractKeeper.Execute(chain.GetContext(), contractAddress, caller, msg, coins)
	return result, err
}
