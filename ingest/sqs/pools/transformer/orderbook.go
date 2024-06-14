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
)

// updateOrderbookInfo updates cosmwasmPoolModel with orderbook specific info.
// - It queries all ticks of the pool and constructs `OrderbookData`.
func (pi *poolTransformer) updateOrderbookInfo(
	ctx sdk.Context,
	poolId uint64,
	contractAddress sdk.AccAddress,
	cosmWasmPoolModel *sqsdomain.CosmWasmPoolModel,
) error {
	ticks, err := pi.orderbookAllTicks(ctx, pi.wasmKeeper, poolId, contractAddress)
	if err != nil {
		return err
	}

	cosmWasmPoolModel.Data.Orderbook = &sqsdomain.OrderbookData{
		Ticks: ticks,
	}

	return nil
}

func (pi *poolTransformer) orderbookAllTicks(ctx sdk.Context,
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

	allTicksResponse := struct {
		Ticks []sqsdomain.TickIdAndState `json:"ticks"`
	}{}

	if err := json.Unmarshal(bz, &allTicksResponse); err != nil {
		return nil, fmt.Errorf(
			"error unmarshalling all_ticks response for pool (%d) contrat_address (%s): %w",
			poolId, contractAddress, err,
		)
	}

	return allTicksResponse.Ticks, nil
}
