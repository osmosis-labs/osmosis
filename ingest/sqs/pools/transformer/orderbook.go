package poolstransformer

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	"github.com/osmosis-labs/sqs/sqsdomain"
)

const (
	allTicksQueryString = `{"all_ticks":{}}`
	orderbookKey        = `orderbook`
)

type allTicksResponse struct {
	Ticks []sqsdomain.TickIdAndState `json:"ticks"`
}

type orderbook struct {
	QuoteDenom  string `json:"quote_denom"`
	BaseDenom   string `json:"base_denom"`
	CurrentTick int64  `json:"current_tick"`
	NextBidTick int64  `json:"next_bid_tick"`
	NextAskTick int64  `json:"next_ask_tick"`
}

// updateOrderbookInfo updates cosmwasmPoolModel with orderbook specific info.
// - It queries all ticks of the pool and constructs `OrderbookData`.
func (pi *poolTransformer) updateOrderbookInfo(
	ctx sdk.Context,
	poolId uint64,
	contractAddress sdk.AccAddress,
	cosmWasmPoolModel *sqsdomain.CosmWasmPoolModel,
) error {
	orderbook, err := pi.orderbookOrderbookRaw(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	ticks, err := pi.orderbookAllTicks(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	cosmWasmPoolModel.Data.Orderbook = &sqsdomain.OrderbookData{
		QuoteDenom:  orderbook.QuoteDenom,
		BaseDenom:   orderbook.BaseDenom,
		NextBidTick: orderbook.NextBidTick,
		NextAskTick: orderbook.NextAskTick,
		Ticks:       ticks,
	}

	return nil
}

func (pi *poolTransformer) orderbookAllTicks(
	ctx sdk.Context,
	wasmKeeper domain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) ([]sqsdomain.TickIdAndState, error) {
	bz, err := wasmKeeper.QuerySmart(ctx, contractAddress, []byte(allTicksQueryString))
	if err != nil {
		return nil, fmt.Errorf(
			"error querying all_ticks pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	var allTicksResponse allTicksResponse

	if err := json.Unmarshal(bz, &allTicksResponse); err != nil {
		return nil, fmt.Errorf(
			"error unmarshalling all_ticks response for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return allTicksResponse.Ticks, nil
}

func (pi *poolTransformer) orderbookOrderbookRaw(
	ctx sdk.Context,
	wasmKeeper domain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) (orderbook, error) {
	bz := wasmKeeper.QueryRaw(ctx, contractAddress, []byte(orderbookKey))

	if bz == nil || len(bz) == 0 {
		return orderbook{}, fmt.Errorf(
			"error querying orderbook for pool (%d) contrat_address (%s): not found",
			poolId, contractAddress,
		)
	}

	var orderbook orderbook

	if err := json.Unmarshal(bz, &orderbook); err != nil {
		return orderbook, fmt.Errorf(
			"error unmarshalling orderbook for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return orderbook, nil
}
