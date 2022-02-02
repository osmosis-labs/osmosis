package wasmtest

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
)

func TestNoStorageWithoutProposal(t *testing.T) {
	// we use default config
	osmosis, ctx := CreateTestInput()

	wasmKeeper := osmosis.WasmKeeper
	// this wraps wasmKeeper, providing interfaces exposed to external messages
	contractKeeper := keeper.NewDefaultPermissionKeeper(&wasmKeeper)

	_, _, creator := keyPubAddr()

	// upload reflect code
	wasmCode, err := ioutil.ReadFile("./testdata/hackatom.wasm")
	require.NoError(t, err)
	_, err = contractKeeper.Create(ctx, creator, wasmCode, nil)
	require.Error(t, err)
}

func TestStoreCodeProposal(t *testing.T) {
	osmosis, ctx := CreateTestInput()

	govKeeper, wasmKeeper := osmosis.GovKeeper, osmosis.WasmKeeper
	wasmCode, err := ioutil.ReadFile("./testdata/hackatom.wasm")
	require.NoError(t, err)

	myActorAddress := RandomBech32AccountAddress()

	src := types.StoreCodeProposalFixture(func(p *types.StoreCodeProposal) {
		p.RunAs = myActorAddress
		p.WASMByteCode = wasmCode
	})

	// when stored
	storedProposal, err := govKeeper.SubmitProposal(ctx, src)
	require.NoError(t, err)

	// and proposal execute
	handler := govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
	err = handler(ctx, storedProposal.GetContent())
	require.NoError(t, err)

	// then
	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)
	assert.Equal(t, myActorAddress, cInfo.Creator)
	assert.True(t, wasmKeeper.IsPinnedCode(ctx, 1))

	storedCode, err := wasmKeeper.GetByteCode(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, wasmCode, storedCode)
}
