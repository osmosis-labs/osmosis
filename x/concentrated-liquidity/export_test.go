package concentrated_liquidity

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func (p *Pool) OrderInitialPoolDenoms(denom0, denom1 string) error {
	return p.orderInitialPoolDenoms(denom0, denom1)
}
