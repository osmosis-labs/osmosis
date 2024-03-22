package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v23/x/market/client/queryproto"

	appParams "github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/stretchr/testify/require"
)

func TestQueryParams(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)

	querier := NewQuerier(input.MarketKeeper)
	res, err := querier.Params(ctx, &queryproto.QueryParamsRequest{})
	require.NoError(t, err)

	require.Equal(t, input.MarketKeeper.GetParams(input.Ctx), res.Params)
}

func TestQuerySwap(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)
	querier := NewQuerier(input.MarketKeeper)

	//price := sdk.NewDecWithPrec(17, 1)
	//input.OracleKeeper.SetLunaExchangeRate(input.Ctx, appParams.MicroSDRDenom, price)

	var err error

	// empty request cause error
	_, err = querier.Swap(ctx, &queryproto.QuerySwapRequest{})
	require.Error(t, err)

	// empty ask denom cause error
	_, err = querier.Swap(ctx, &queryproto.QuerySwapRequest{OfferCoin: sdk.Coin{Denom: appParams.MicroSDRDenom, Amount: sdk.NewInt(100)}.String()})
	require.Error(t, err)

	// empty offer coin cause error
	_, err = querier.Swap(ctx, &queryproto.QuerySwapRequest{AskDenom: appParams.MicroSDRDenom})
	require.Error(t, err)

	// recursive query
	offerCoin := sdk.NewCoin(appParams.BaseCoinUnit, sdk.NewInt(10)).String()
	_, err = querier.Swap(ctx, &queryproto.QuerySwapRequest{OfferCoin: offerCoin, AskDenom: appParams.BaseCoinUnit})
	require.Error(t, err)

	// overflow query
	overflowAmt, _ := sdk.NewIntFromString("1000000000000000000000000000000000")
	overflowOfferCoin := sdk.NewCoin(appParams.BaseCoinUnit, overflowAmt).String()
	_, err = querier.Swap(ctx, &queryproto.QuerySwapRequest{OfferCoin: overflowOfferCoin, AskDenom: appParams.MicroSDRDenom})
	require.Error(t, err)

	// valid query
	res, err := querier.Swap(ctx, &queryproto.QuerySwapRequest{OfferCoin: offerCoin, AskDenom: appParams.MicroSDRDenom})
	require.NoError(t, err)

	require.Equal(t, appParams.MicroSDRDenom, res.ReturnCoin.Denom)
	require.True(t, sdk.NewInt(17).GTE(res.ReturnCoin.Amount))
	require.True(t, res.ReturnCoin.Amount.IsPositive())
}

func TestQueryMintPoolDelta(t *testing.T) {
	input := CreateTestInput(t)
	ctx := sdk.WrapSDKContext(input.Ctx)
	querier := NewQuerier(input.MarketKeeper)

	poolDelta := sdk.NewDecWithPrec(17, 1)
	input.MarketKeeper.SetTerraPoolDelta(input.Ctx, poolDelta)

	res, errRes := querier.TerraPoolDelta(ctx, &queryproto.QueryTerraPoolDeltaRequest{})
	require.NoError(t, errRes)

	require.Equal(t, poolDelta, res.TerraPoolDelta)
}
