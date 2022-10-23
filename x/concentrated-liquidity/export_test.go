package concentrated_liquidity

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func (k Keeper) OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	return k.orderInitialPoolDenoms(denom0, denom1)
}
