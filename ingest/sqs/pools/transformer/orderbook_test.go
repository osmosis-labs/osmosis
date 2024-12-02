package poolstransformer_test

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sqscosmwasmpool "github.com/osmosis-labs/sqs/sqsdomain/cosmwasmpool"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v28/app/apptesting"
	poolstransformer "github.com/osmosis-labs/osmosis/v28/ingest/sqs/pools/transformer"
)

const (
	// Tick Price = 2
	LARGE_POSITIVE_TICK int64 = 1000000
)

type PlaceLimitMsg struct {
	TickID         int64        `json:"tick_id"`
	OrderDirection string       `json:"order_direction"` // 'bid' | 'ask'
	Quantity       osmomath.Int `json:"quantity"`
}

type ExecuteMsg struct {
	PlaceLimit *PlaceLimitMsg `json:"place_limit,omitempty"`
}

func (s *PoolTransformerTestSuite) TestUpdateOrderbookInfo() {
	s.Setup()

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(
		sdk.NewCoin(UOSMO, osmomath.NewInt(100000000)),
		sdk.NewCoin(USDC, osmomath.NewInt(100000000)),
	))

	pool := s.PrepareOrderbookPool(s.TestAccs[0], apptesting.OrderbookInstantiateMsg{
		BaseDenom:  UOSMO,
		QuoteDenom: USDC,
	})

	// Create OSMO / USDC pool
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

	// Initialize the pool ingester
	poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

	cosmWasmPoolModel := sqscosmwasmpool.CosmWasmPoolModel{}

	poolIngester.UpdateOrderbookInfo(s.Ctx, pool.GetId(), pool.GetAddress(), &cosmWasmPoolModel)

	// Check if the pool has been updated
	s.Equal(sqscosmwasmpool.CosmWasmPoolData{
		Orderbook: &sqscosmwasmpool.OrderbookData{
			QuoteDenom:                     USDC,
			BaseDenom:                      UOSMO,
			NextBidTickIndex:               -1,
			NextAskTickIndex:               -1,
			BidAmountToExhaustAskLiquidity: osmomath.ZeroBigDec(),
			AskAmountToExhaustBidLiquidity: osmomath.ZeroBigDec(),
			Ticks:                          []sqscosmwasmpool.OrderbookTick{},
		},
	}, cosmWasmPoolModel.Data)

	// Place a limit order
	quantity := osmomath.NewInt(10000)
	msg := ExecuteMsg{
		PlaceLimit: &PlaceLimitMsg{
			TickID:         LARGE_POSITIVE_TICK,
			OrderDirection: "bid",
			Quantity:       quantity,
		},
	}

	bz, err := json.Marshal(msg)
	s.NoError(err)

	_, err = s.App.ContractKeeper.Execute(s.Ctx, pool.GetAddress(), s.TestAccs[0], bz, sdk.NewCoins(sdk.NewCoin(USDC, osmomath.NewInt(10000))))
	s.NoError(err)

	poolIngester.UpdateOrderbookInfo(s.Ctx, pool.GetId(), pool.GetAddress(), &cosmWasmPoolModel)

	// Check if the pool has been updated
	s.Equal(sqscosmwasmpool.CosmWasmPoolData{
		AlloyTransmuter: nil,
		Orderbook: &sqscosmwasmpool.OrderbookData{
			QuoteDenom:                     USDC,
			BaseDenom:                      UOSMO,
			NextBidTickIndex:               0,
			NextAskTickIndex:               -1,
			BidAmountToExhaustAskLiquidity: osmomath.ZeroBigDec(),
			AskAmountToExhaustBidLiquidity: osmomath.BigDecFromSDKInt(quantity).Quo(osmomath.NewBigDec(2)),
			Ticks: []sqscosmwasmpool.OrderbookTick{{
				TickId: LARGE_POSITIVE_TICK,
				TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{
					AskLiquidity: osmomath.ZeroBigDec(),
					BidLiquidity: osmomath.BigDecFromSDKInt(quantity),
				},
			}},
		},
	}, cosmWasmPoolModel.Data)
}

func (s *PoolTransformerTestSuite) TestTickIndexById() {
	ticks := []sqscosmwasmpool.OrderbookTick{
		{TickId: -99, TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{}},
		{TickId: 1, TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{}},
		{TickId: 3, TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{}},
		{TickId: 7, TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{}},
		{TickId: 10, TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{}},
		{TickId: 15, TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{}},
	}

	tests := []struct {
		tickId   int64
		expected int
	}{
		{tickId: -99, expected: 0},
		{tickId: -98, expected: -1},
		{tickId: -1, expected: -1},
		{tickId: 1, expected: 1},
		{tickId: 2, expected: -1},
		{tickId: 3, expected: 2},
		{tickId: 7, expected: 3},
		{tickId: 10, expected: 4},
		{tickId: 15, expected: 5},
		{tickId: 20, expected: -1},
	}

	for _, tt := range tests {
		index := poolstransformer.TickIndexById(ticks, tt.tickId)
		s.Equal(tt.expected, index)
	}
}
