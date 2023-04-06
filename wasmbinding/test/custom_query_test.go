package wasmbinding

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app"
	"github.com/osmosis-labs/osmosis/v15/wasmbinding/bindings"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
)

// we must pay this many uosmo for every pool we create
var poolFee int64 = 1000000000

var defaultFunds = sdk.NewCoins(
	sdk.NewInt64Coin("uatom", 333000000),
	sdk.NewInt64Coin("uosmo", 555000000+2*poolFee),
	sdk.NewInt64Coin("ustar", 999000000),
)

func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.OsmosisApp, sdk.Context) {
	osmosis, ctx := CreateTestInput()
	wasmKeeper := osmosis.WasmKeeper

	storeReflectCode(t, ctx, osmosis, addr)

	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return osmosis, ctx
}

func TestQueryFullDenom(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

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

func assertValidShares(t *testing.T, shares wasmvmtypes.Coin, poolID uint64) {
	// sanity check: check the denom and ensure at least 18 decimal places
	denom := fmt.Sprintf("gamm/pool/%d", poolID)
	require.Equal(t, denom, shares.Denom)
	require.Greater(t, len(shares.Amount), 18)
}

func storeReflectCode(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, addr sdk.AccAddress) {
	govKeeper := osmosis.GovKeeper
	wasmCode, err := os.ReadFile("../testdata/osmo_reflect.wasm")
	require.NoError(t, err)

	src := wasmtypes.StoreCodeProposalFixture(func(p *wasmtypes.StoreCodeProposal) {
		p.RunAs = addr.String()
		p.WASMByteCode = wasmCode
		checksum := sha256.Sum256(wasmCode)
		p.CodeHash = checksum[:]
	})

	// when stored
	storedProposal, err := govKeeper.SubmitProposal(ctx, src, false)
	require.NoError(t, err)

	// and proposal execute
	handler := govKeeper.Router().GetRoute(storedProposal.ProposalRoute())
	err = handler(ctx, storedProposal.GetContent())
	require.NoError(t, err)
}

func instantiateReflectContract(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, funder sdk.AccAddress) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(osmosis.WasmKeeper)
	codeID := uint64(1)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}

func fundAccount(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, addr sdk.AccAddress, coins sdk.Coins) {
	err := simapp.FundAccount(
		osmosis.BankKeeper,
		ctx,
		addr,
		coins,
	)
	require.NoError(t, err)
}

func preparePool(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, addr sdk.AccAddress, funds []sdk.Coin) uint64 {
	var assets []balancer.PoolAsset
	for _, coin := range funds {
		assets = append(assets, balancer.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  coin,
		})
	}

	poolParams := balancer.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	}

	msg := balancer.NewMsgCreateBalancerPool(addr, poolParams, assets, "")
	poolId, err := osmosis.PoolManagerKeeper.CreatePool(ctx, &msg)
	require.NoError(t, err)
	return poolId
}
