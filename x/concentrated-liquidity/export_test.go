package concentrated_liquidity

import (
	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	return cltypes.OrderInitialPoolDenoms(denom0, denom1)
}
