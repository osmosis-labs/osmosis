package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	_ types.PoolI = &Pool{}
)

func (p Pool) GetAddress() sdk.AccAddress {
	return sdk.AccAddress{}
}

func (p Pool) String() string {
	return ""
}

func (p Pool) GetId() uint64 {
	return 0
}
func (p Pool) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}
func (p Pool) GetExitFee(ctx sdk.Context) sdk.Dec {
	return sdk.Dec{}
}
func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}
func (p Pool) GetTotalShares() sdk.Int {
	return sdk.Int{}
}

func (p Pool) SwapOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

// this only works on a single directional trade, will implement bi directional trade in next milestone
func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) CalcInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p Pool) SpotPrice(ctx sdk.Context, baseAssetDenom string, quoteAssetDenom string) (sdk.Dec, error) {
	return sdk.Dec{}, nil
}
