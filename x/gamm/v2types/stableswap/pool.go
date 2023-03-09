package stableswap

import (
	

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

var (
	_ poolmanagertypes.PoolI       = &Pool{}
	_ types.CFMMPoolI              = &Pool{}
)


// GetAddress returns the address of a pool.
// If the pool address is not bech32 valid, it returns an empty address.
func (p Pool) GetAddress() sdk.AccAddress {
	return nil
}

func (p Pool) GetId() uint64 {
	return p.Id
}

func (p Pool) GetSwapFee(_ sdk.Context) sdk.Dec {
	return p.PoolParams.SwapFee
}

func (p Pool) GetTotalPoolLiquidity(_ sdk.Context) sdk.Coins {
	return sdk.Coins{}
}

func (p Pool) GetTotalShares() sdk.Int {
	return p.TotalShares.Amount
}

func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

func (p Pool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.Stableswap
}

func (p Pool) SpotPrice(ctx sdk.Context, quoteAsset, baseAsset string) (spotPrice sdk.Dec, err error) {
	return sdk.ZeroDec(), nil
}

func (p *Pool) CalcExitPoolCoinsFromShares(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return sdk.Coins{}, err
}

func (p Pool) CalcInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (
	tokenIn sdk.Coin, err error,
) {
	return sdk.Coin{}, err
}

func (p *Pool) CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, tokensJoined sdk.Coins, err error) {
	return sdk.ZeroInt(), sdk.Coins{}, nil
}

func (p *Pool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (numShares sdk.Int, newLiquidity sdk.Coins, err error) {
	return sdk.ZeroInt(), sdk.Coins{}, nil
}

func (p Pool) CalcOutAmtGivenIn(ctx sdk.Context, tokenIn sdk.Coins, tokenOutDenom string, swapFee sdk.Dec) (tokenOut sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p *Pool) ExitPool(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	return sdk.Coins{}, err
}

func (p *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (sdk.Int, error) {
	return sdk.ZeroInt(), nil
}

func (p *Pool) JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, swapFee sdk.Dec) (sdk.Int, error) {
	return sdk.ZeroInt(), nil
}

func (p *Pool) SwapInAmtGivenOut(ctx sdk.Context, tokenOut sdk.Coins, tokenInDenom string, swapFee sdk.Dec) (tokenIn sdk.Coin, err error) {
	return sdk.Coin{}, nil
}

func (p *Pool) SwapOutAmtGivenIn(
	ctx sdk.Context,
	tokensIn sdk.Coins,
	tokenOutDenom string,
	swapFee sdk.Dec,
) (
	tokenOut sdk.Coin, err error,
) {
	return sdk.Coin{}, nil
}