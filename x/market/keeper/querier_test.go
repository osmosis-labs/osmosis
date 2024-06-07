package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v23/app/apptesting/assets"
	appParams "github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/osmosis-labs/osmosis/v23/x/market/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

func (s *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(s.Ctx)

	querier := keeper.NewQuerier(*s.App.MarketKeeper)
	res, err := querier.Params(ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)

	s.Require().Equal(s.App.MarketKeeper.GetParams(s.Ctx), res.Params)
}

func (s *KeeperTestSuite) TestQuerySwap() {
	ctx := sdk.WrapSDKContext(s.Ctx)
	querier := keeper.NewQuerier(*s.App.MarketKeeper)

	price := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, price)

	var err error

	// empty request cause error
	_, err = querier.Swap(ctx, &types.QuerySwapRequest{})
	s.Require().Error(err)

	// empty ask denom cause error
	_, err = querier.Swap(ctx, &types.QuerySwapRequest{OfferCoin: sdk.Coin{Denom: assets.MicroSDRDenom, Amount: sdk.NewInt(100)}.String()})
	s.Require().Error(err)

	// empty offer coin cause error
	_, err = querier.Swap(ctx, &types.QuerySwapRequest{AskDenom: assets.MicroSDRDenom})
	s.Require().Error(err)

	// recursive query
	offerCoin := sdk.NewCoin(appParams.BaseCoinUnit, sdk.NewInt(10)).String()
	_, err = querier.Swap(ctx, &types.QuerySwapRequest{OfferCoin: offerCoin, AskDenom: appParams.BaseCoinUnit})
	s.Require().Error(err)

	// overflow query
	overflowAmt, _ := sdk.NewIntFromString("1000000000000000000000000000000000")
	overflowOfferCoin := sdk.NewCoin(appParams.BaseCoinUnit, overflowAmt).String()
	_, err = querier.Swap(ctx, &types.QuerySwapRequest{OfferCoin: overflowOfferCoin, AskDenom: assets.MicroSDRDenom})
	s.Require().Error(err)

	// valid query
	res, err := querier.Swap(ctx, &types.QuerySwapRequest{OfferCoin: offerCoin, AskDenom: assets.MicroSDRDenom})
	s.Require().NoError(err)

	s.Require().Equal(assets.MicroSDRDenom, res.ReturnCoin.Denom)
	s.Require().True(sdk.NewInt(17).GTE(res.ReturnCoin.Amount))
	s.Require().True(res.ReturnCoin.Amount.IsPositive())
}

func (s *KeeperTestSuite) TestQueryMintPoolDelta() {
	ctx := sdk.WrapSDKContext(s.Ctx)
	querier := keeper.NewQuerier(*s.App.MarketKeeper)

	poolDelta := sdk.NewDecWithPrec(17, 1)
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, poolDelta)

	res, errRes := querier.OsmosisPoolDelta(ctx, &types.QueryOsmosisPoolDeltaRequest{})
	s.Require().NoError(errRes)

	s.Require().Equal(poolDelta, res.OsmosisPoolDelta)
}
