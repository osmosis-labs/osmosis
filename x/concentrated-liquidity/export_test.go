package concentrated_liquidity

// SetInitialPoolDenoms sets the pool denoms of a cl pool
func (p *Pool) SetInitialPoolDenoms(poolDenoms []string) error {
	return p.setInitialPoolDenoms(poolDenoms)
}
