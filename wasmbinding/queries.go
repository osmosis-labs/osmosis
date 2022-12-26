package wasmbinding

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/wasmbinding/bindings"
<<<<<<< HEAD
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v13/x/tokenfactory/keeper"
=======
	gammkeeper "github.com/osmosis-labs/osmosis/v13/x/gamm/keeper"
	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v13/x/tokenfactory/keeper"
	twapkeeper "github.com/osmosis-labs/osmosis/v13/x/twap"
>>>>>>> concentrated-liquidity-main
)

type QueryPlugin struct {
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(tfk *tokenfactorykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		tokenFactoryKeeper: tfk,
	}
}

// GetDenomAdmin is a query to get denom admin.
func (qp QueryPlugin) GetDenomAdmin(ctx sdk.Context, denom string) (*bindings.DenomAdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin for denom: %s", denom)
	}

	return &bindings.DenomAdminResponse{Admin: metadata.Admin}, nil
}
<<<<<<< HEAD
=======

// GetPoolState is a query to get pool liquidity and amount of each denoms' pool shares.
func (qp QueryPlugin) GetPoolState(ctx sdk.Context, poolID uint64) (*bindings.PoolAssets, error) {
	poolData, err := qp.gammKeeper.GetPoolAndPoke(ctx, poolID)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm get pool")
	}

	return &bindings.PoolAssets{
		Assets: poolData.GetTotalPoolLiquidity(ctx),
		Shares: sdk.Coin{
			Denom:  gammtypes.GetPoolShareDenom(poolID),
			Amount: poolData.GetTotalShares(),
		},
	}, nil
}

// GetSpotPrice is a query to get spot price of denoms.
func (qp QueryPlugin) GetSpotPrice(ctx sdk.Context, spotPrice *bindings.SpotPrice) (*sdk.Dec, error) {
	if spotPrice == nil {
		return nil, wasmvmtypes.InvalidRequest{Err: "gamm spot price null"}
	}

	poolId := spotPrice.Swap.PoolId
	denomIn := spotPrice.Swap.DenomIn
	denomOut := spotPrice.Swap.DenomOut
	withSwapFee := spotPrice.WithSwapFee

	price, err := qp.gammKeeper.CalculateSpotPrice(ctx, poolId, denomIn, denomOut)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm get spot price")
	}

	if withSwapFee {
		poolData, err := qp.gammKeeper.GetPoolAndPoke(ctx, poolId)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "gamm get pool")
		}

		price = price.Mul(sdk.OneDec().Sub(poolData.GetSwapFee(ctx)))
	}

	return &price, nil
}

func (qp QueryPlugin) ArithmeticTwap(ctx sdk.Context, arithmeticTwap *bindings.ArithmeticTwap) (*sdk.Dec, error) {
	if arithmeticTwap == nil {
		return nil, wasmvmtypes.InvalidRequest{Err: "gamm arithmetic twap null"}
	}

	poolId := arithmeticTwap.PoolId
	quoteAssetDenom := arithmeticTwap.QuoteAssetDenom
	baseAssetDenom := arithmeticTwap.BaseAssetDenom
	startTime := time.UnixMilli(arithmeticTwap.StartTime)
	endTime := time.UnixMilli(arithmeticTwap.EndTime)

	twap, err := qp.twapKeeper.GetArithmeticTwap(ctx, poolId, quoteAssetDenom, baseAssetDenom, startTime, endTime)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm arithmetic twap")
	}

	return &twap, nil
}

func (qp QueryPlugin) ArithmeticTwapToNow(ctx sdk.Context, arithmeticTwap *bindings.ArithmeticTwapToNow) (*sdk.Dec, error) {
	if arithmeticTwap == nil {
		return nil, wasmvmtypes.InvalidRequest{Err: "gamm arithmetic twap null"}
	}

	poolId := arithmeticTwap.PoolId
	quoteAssetDenom := arithmeticTwap.QuoteAssetDenom
	baseAssetDenom := arithmeticTwap.BaseAssetDenom
	startTime := time.UnixMilli(arithmeticTwap.StartTime)

	twap, err := qp.twapKeeper.GetArithmeticTwapToNow(ctx, poolId, quoteAssetDenom, baseAssetDenom, startTime)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gamm arithmetic twap")
	}

	return &twap, nil
}
>>>>>>> concentrated-liquidity-main
