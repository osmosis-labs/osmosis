package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newKeeper() (*simapp.Simapp, Keeper, sdk.Context) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	keeper := NewKeeper(
		app.AppCodec(),
		app.GetKey(types.StoreKey),
		nil, // TODO
	)
	return app, keeper, ctx
}

func TestRegisterCell(t *testing.T) {
	app, keeper, ctx := newKeeper()
	require.NoError(t, k.RegisterCell(ctx, 0, types.ExampleCellState{}, nil))
}

func TestExecuteExpression(t *testing.T) {

}
