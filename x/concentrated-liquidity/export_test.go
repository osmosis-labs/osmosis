package concentrated_liquidity

// SetInitialPoolDenoms sets the pool denoms of a cl pool
func (p *Pool) SetInitialPoolDenoms(denom0, denom1 string) error {
	return p.orderInitialPoolDenoms(denom0, denom1)
}
