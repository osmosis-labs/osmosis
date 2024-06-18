package poolstransformer_test

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	sqscosmwasmpool "github.com/osmosis-labs/sqs/sqsdomain/cosmwasmpool"
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
			QuoteDenom:  USDC,
			BaseDenom:   UOSMO,
			NextBidTick: -108000000,
			NextAskTick: 182402823,
			Ticks:       []sqscosmwasmpool.OrderbookTickIdAndState{},
		},
	}, cosmWasmPoolModel.Data)

	// Place a limit order
	quantity := osmomath.NewInt(10000)
	msg := ExecuteMsg{
		PlaceLimit: &PlaceLimitMsg{
			TickID:         9,
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
			QuoteDenom:  USDC,
			BaseDenom:   UOSMO,
			NextBidTick: 9,
			NextAskTick: 182402823,
			Ticks: []sqscosmwasmpool.OrderbookTickIdAndState{{
				TickId: 9,
				TickState: sqscosmwasmpool.OrderbookTickState{
					AskValues: sqscosmwasmpool.OrderbookTickValues{TotalAmountOfLiquidity: osmomath.ZeroBigDec()},
					BidValues: sqscosmwasmpool.OrderbookTickValues{TotalAmountOfLiquidity: osmomath.BigDecFromSDKInt(quantity)},
				},
			}},
		},
	}, cosmWasmPoolModel.Data)
}
