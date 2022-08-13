package wasm

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/app"
)

func TestNoStorageWithoutProposal(t *testing.T) {
	// we use default config
	osmosis, ctx := CreateTestInput()

	wasmKeeper := osmosis.WasmKeeper
	// this wraps wasmKeeper, providing interfaces exposed to external messages
	contractKeeper := keeper.NewDefaultPermissionKeeper(wasmKeeper)

	_, _, creator := keyPubAddr()

	// upload reflect code
	wasmCode, err := ioutil.ReadFile("../testdata/hackatom.wasm")
	require.NoError(t, err)
	_, err = contractKeeper.Create(ctx, creator, wasmCode, nil)
	require.Error(t, err)
}

func storeCodeViaProposal(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, addr sdk.AccAddress) {
	govKeeper := osmosis.GovKeeper
	wasmCode, err := ioutil.ReadFile("../testdata/hackatom.wasm")
	require.NoError(t, err)

	src := types.StoreCodeProposalFixture(func(p *types.StoreCodeProposal) {
		p.RunAs = addr.String()
		p.WASMByteCode = wasmCode
	})

	// when stored
	storedProposal, err := govKeeper.SubmitProposal(ctx, src)
	require.NoError(t, err)

	// and proposal execute
	handler := govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
	err = handler(ctx, storedProposal.GetContent())
	require.NoError(t, err)
}

func TestStoreCodeProposal(t *testing.T) {
	osmosis, ctx := CreateTestInput()
	myActorAddress := RandomAccountAddress()
	wasmKeeper := osmosis.WasmKeeper

	storeCodeViaProposal(t, ctx, osmosis, myActorAddress)

	// then
	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)
	assert.Equal(t, myActorAddress.String(), cInfo.Creator)
	assert.True(t, wasmKeeper.IsPinnedCode(ctx, 1))

	storedCode, err := wasmKeeper.GetByteCode(ctx, 1)
	require.NoError(t, err)
	wasmCode, err := ioutil.ReadFile("../testdata/hackatom.wasm")
	require.NoError(t, err)
	assert.Equal(t, wasmCode, storedCode)
}

type HackatomExampleInitMsg struct {
	Verifier    sdk.AccAddress `json:"verifier"`
	Beneficiary sdk.AccAddress `json:"beneficiary"`
}

func TestInstantiateContract(t *testing.T) {
	osmosis, ctx := CreateTestInput()
	funder := RandomAccountAddress()
	benefit, arb := RandomAccountAddress(), RandomAccountAddress()
	FundAccount(t, ctx, osmosis, funder)

	storeCodeViaProposal(t, ctx, osmosis, funder)
	contractKeeper := keeper.NewDefaultPermissionKeeper(osmosis.WasmKeeper)
	codeID := uint64(1)

	initMsg := HackatomExampleInitMsg{
		Verifier:    arb,
		Beneficiary: benefit,
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(t, err)

	funds := sdk.NewInt64Coin("uosmo", 123456)
	_, _, err = contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", sdk.Coins{funds})
	require.NoError(t, err)
}
