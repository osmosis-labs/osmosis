package stableswap

import sdk "github.com/cosmos/cosmos-sdk/types"

func (pa Pool) GetScaledPoolAmts(denoms ...string) ([]sdk.Int, error) {
	return pa.getPoolAmts(denoms...)
}

func (pa Pool) GetDescaledPoolAmt(denom string, amount sdk.Dec) (sdk.Dec, error) {
	return pa.getDescaledPoolAmt(denom, amount)
}
