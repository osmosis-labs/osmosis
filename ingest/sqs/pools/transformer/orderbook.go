package poolstransformer

import (
	"encoding/json"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sqscosmwasmpool "github.com/osmosis-labs/sqs/sqsdomain/cosmwasmpool"

	"github.com/osmosis-labs/osmosis/osmomath"
	commondomain "github.com/osmosis-labs/osmosis/v28/ingest/common/domain"
)

const (
	allTicksQueryString = `{"all_ticks":{}}`
	orderbookKey        = `orderbook`
)

// allTicksResponse is `all_ticks` query response.
type allTicksResponse struct {
	Ticks []orderbookTickIdAndState `json:"ticks"`
}

type orderbookTickIdAndState struct {
	TickId    int64              `json:"tick_id"`
	TickState orderbookTickState `json:"tick_state"`
}

// orderbookTickState represents the state of the orderbook tick in both bid and ask directions.
type orderbookTickState struct {
	// Values for the bid direction of the tick
	BidValues orderbookTickValues `json:"bid_values"`
	// Values for the ask direction of the tick
	AskValues orderbookTickValues `json:"ask_values"`
}

// orderbookTickValues represents the values of the orderbook tick.
// Other values are present in the response but omitted on purpose since it's not being used in sqs.
type orderbookTickValues struct {
	TotalAmountOfLiquidity osmomath.BigDec `json:"total_amount_of_liquidity"`
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
	cosmWasmPoolModel *sqscosmwasmpool.CosmWasmPoolModel,
) error {
	orderbook, err := pi.getOrderbookRaw(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	ticks, err := pi.orderbookAllTicks(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	nextBidTickIndex := tickIndexById(ticks, orderbook.NextBidTick)
	nextAskTickIndex := tickIndexById(ticks, orderbook.NextAskTick)

	bidAmountToExhaustAskLiquidity, err := sqscosmwasmpool.CalcAmountInToExhaustOrderbookLiquidity(sqscosmwasmpool.BID, nextAskTickIndex, ticks)
	if err != nil {
		return fmt.Errorf("error calculating bid amount to exhaust ask liquidity: %w", err)
	}

	askAmountToExhaustBidLiquidity, err := sqscosmwasmpool.CalcAmountInToExhaustOrderbookLiquidity(sqscosmwasmpool.ASK, nextBidTickIndex, ticks)
	if err != nil {
		return fmt.Errorf("error calculating ask amount to exhaust bid liquidity: %w", err)
	}

	cosmWasmPoolModel.Data.Orderbook = &sqscosmwasmpool.OrderbookData{
		QuoteDenom:                     orderbook.QuoteDenom,
		BaseDenom:                      orderbook.BaseDenom,
		NextBidTickIndex:               nextBidTickIndex,
		NextAskTickIndex:               nextAskTickIndex,
		BidAmountToExhaustAskLiquidity: bidAmountToExhaustAskLiquidity,
		AskAmountToExhaustBidLiquidity: askAmountToExhaustBidLiquidity,
		Ticks:                          ticks,
	}

	return nil
}

func (pi *poolTransformer) orderbookAllTicks(
	ctx sdk.Context,
	wasmKeeper commondomain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) ([]sqscosmwasmpool.OrderbookTick, error) {
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

	ticks := make([]sqscosmwasmpool.OrderbookTick, len(allTicksResponse.Ticks))

	for i, tick := range allTicksResponse.Ticks {
		ticks[i] = sqscosmwasmpool.OrderbookTick{
			TickId: tick.TickId,
			TickLiquidity: sqscosmwasmpool.OrderbookTickLiquidity{
				BidLiquidity: tick.TickState.BidValues.TotalAmountOfLiquidity,
				AskLiquidity: tick.TickState.AskValues.TotalAmountOfLiquidity,
			},
		}
	}

	return ticks, nil
}

func (pi *poolTransformer) getOrderbookRaw(
	ctx sdk.Context,
	wasmKeeper commondomain.WasmKeeper,
	poolId uint64,
	contractAddress sdk.AccAddress,
) (orderbook, error) {
	bz := wasmKeeper.QueryRaw(ctx, contractAddress, []byte(orderbookKey))

	if len(bz) == 0 {
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

// tickIndexById find the index of the tick with the given tickId.
// return -1 and no errors if not found. Since next bid and next ask
// can be min/max tick id in case there is no bid/ask tick.
func tickIndexById(ticks []sqscosmwasmpool.OrderbookTick, tickId int64) int {
	// binary search for the tickId
	index := sort.Search(len(ticks), func(i int) bool {
		return ticks[i].TickId >= tickId
	})

	// if tickId is found, return the index
	if index < len(ticks) && ticks[index].TickId == tickId {
		return index
	}

	// else return -1
	return -1
}

// dirty
