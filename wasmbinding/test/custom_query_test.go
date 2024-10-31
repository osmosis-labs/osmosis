package wasmbinding

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/osmosis-labs/osmosis/v27/app"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/wasmbinding/bindings"
)

func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.OsmosisApp, sdk.Context, string) {
	t.Helper()

	osmosis, ctx, homeDir := CreateTestInput()
	wasmKeeper := osmosis.WasmKeeper

	storeReflectCode(t, ctx, osmosis, addr)

	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return osmosis, ctx, homeDir
}

func TestQueryFullDenom(t *testing.T) {
	apptesting.SkipIfWSL(t)
	actor := RandomAccountAddress()
	osmosis, ctx, homeDir := SetupCustomApp(t, actor)
	defer os.RemoveAll(homeDir)

	reflect := instantiateReflectContract(t, ctx, osmosis, actor)
	require.NotEmpty(t, reflect)

	// query full denom
	query := bindings.OsmosisQuery{
		FullDenom: &bindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "ustart",
		},
	}
	resp := bindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	expected := fmt.Sprintf("factory/%s/ustart", reflect.String())
	require.EqualValues(t, expected, resp.Denom)

}

type ReflectQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

type ChainRequest struct {
	Request wasmvmtypes.QueryRequest `json:"request"`
}

type ChainResponse struct {
	Data []byte `json:"data"`
}

func queryCustom(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, contract sdk.AccAddress, request bindings.OsmosisQuery, response interface{}) {
	t.Helper()

	msgBz, err := json.Marshal(request)
	require.NoError(t, err)

	query := ReflectQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	require.NoError(t, err)

	resBz, err := osmosis.WasmKeeper.QuerySmart(ctx, contract, queryBz)
	require.NoError(t, err)
	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	require.NoError(t, err)
	err = json.Unmarshal(resp.Data, response)
	require.NoError(t, err)
}

func storeReflectCode(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, addr sdk.AccAddress) {
	t.Helper()
	wasmCode, err := os.ReadFile("../testdata/osmo_reflect.wasm")
	require.NoError(t, err)

	// Quick hack to allow code upload
	originalParams := osmosis.WasmKeeper.GetParams(ctx)
	temporaryParams := originalParams
	temporaryParams.CodeUploadAccess.Permission = wasmtypes.AccessTypeEverybody
	osmosis.WasmKeeper.SetParams(ctx, temporaryParams)

	msg := wasmtypes.MsgStoreCodeFixture(func(m *wasmtypes.MsgStoreCode) {
		m.WASMByteCode = wasmCode
		m.Sender = addr.String()
	})
	_, err = osmosis.MsgServiceRouter().Handler(msg)(ctx, msg)
	require.NoError(t, err)

	osmosis.WasmKeeper.SetParams(ctx, originalParams)
}

func instantiateReflectContract(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, funder sdk.AccAddress) sdk.AccAddress {
	t.Helper()

	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(osmosis.WasmKeeper)
	codeID := uint64(1)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}

func fundAccount(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, addr sdk.AccAddress, coins sdk.Coins) {
	t.Helper()
	err := testutil.FundAccount(
		ctx,
		osmosis.BankKeeper,
		addr,
		coins,
	)
	require.NoError(t, err)
}
