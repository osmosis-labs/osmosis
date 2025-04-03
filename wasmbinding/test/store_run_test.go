package wasmbinding

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
)

func TestNoStorageWithoutProposal(t *testing.T) {
	// we use default config
	osmosis, ctx, homeDir := CreateTestInput()
	defer os.RemoveAll(homeDir)

	wasmKeeper := osmosis.WasmKeeper
	// this wraps wasmKeeper, providing interfaces exposed to external messages
	contractKeeper := keeper.NewDefaultPermissionKeeper(wasmKeeper)

	_, _, creator := keyPubAddr()

	// upload reflect code
	wasmCode, err := os.ReadFile("../testdata/hackatom.wasm")
	require.NoError(t, err)
	_, _, err = contractKeeper.Create(ctx, creator, wasmCode, nil)
	require.Error(t, err)
}

func storeCodeViaProposal(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, addr sdk.AccAddress) {
	t.Helper()
	govKeeper := osmosis.GovKeeper
	wasmCode, err := os.ReadFile("../testdata/hackatom.wasm")
	require.NoError(t, err)

	msgStoreCode := wasmtypes.MsgStoreCode{Sender: addr.String(), WASMByteCode: wasmCode, InstantiatePermission: &types.AccessConfig{Permission: types.AccessTypeEverybody}}
	msgStoreCodeSlice := []sdk.Msg{&msgStoreCode}

	storedProposal, err := govKeeper.SubmitProposal(ctx, msgStoreCodeSlice, "", "title", "summary", addr, false)
	require.NoError(t, err)

	messages, err := storedProposal.GetMsgs()
	require.NoError(t, err)

	for _, msg := range messages {
		handler := govKeeper.Router().Handler(msg)
		_, err = handler(ctx, msg)
		require.NoError(t, err)
	}
}

func TestStoreCodeProposal(t *testing.T) {
	apptesting.SkipIfWSL(t)
	osmosis, ctx, homeDir := CreateTestInput()
	defer os.RemoveAll(homeDir)

	wasmKeeper := osmosis.WasmKeeper

	govModuleAccount := osmosis.AccountKeeper.GetModuleAccount(ctx, govtypes.ModuleName).GetAddress()
	storeCodeViaProposal(t, ctx, osmosis, govModuleAccount)

	// then
	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)
	assert.Equal(t, govModuleAccount.String(), cInfo.Creator)
	// UNFORKINGTODO C: It seems like we no longer pin contracts when executing a gov proposal, want to confirm this is okay
	// assert.True(t, wasmKeeper.IsPinnedCode(ctx, 1))

	storedCode, err := wasmKeeper.GetByteCode(ctx, 1)
	require.NoError(t, err)
	wasmCode, err := os.ReadFile("../testdata/hackatom.wasm")
	require.NoError(t, err)
	assert.Equal(t, wasmCode, storedCode)
}

type HackatomExampleInitMsg struct {
	Verifier    sdk.AccAddress `json:"verifier"`
	Beneficiary sdk.AccAddress `json:"beneficiary"`
}

func TestInstantiateContract(t *testing.T) {
	apptesting.SkipIfWSL(t)
	osmosis, ctx, homeDir := CreateTestInput()
	defer os.RemoveAll(homeDir)

	instantiator := RandomAccountAddress()
	benefit, arb := RandomAccountAddress(), RandomAccountAddress()
	FundAccount(t, ctx, osmosis, instantiator)

	govModuleAccount := osmosis.AccountKeeper.GetModuleAccount(ctx, govtypes.ModuleName).GetAddress()

	storeCodeViaProposal(t, ctx, osmosis, govModuleAccount)
	contractKeeper := keeper.NewDefaultPermissionKeeper(osmosis.WasmKeeper)
	codeID := uint64(1)

	initMsg := HackatomExampleInitMsg{
		Verifier:    arb,
		Beneficiary: benefit,
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(t, err)

	funds := sdk.NewInt64Coin(appparams.BaseCoinUnit, 123456)
	_, _, err = contractKeeper.Instantiate(ctx, codeID, instantiator, instantiator, initMsgBz, "demo contract", sdk.Coins{funds})
	require.NoError(t, err)
}
