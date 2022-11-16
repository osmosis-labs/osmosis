package keeper

// IterateRoutes checks the profitability of every single route that is passed in
// and returns the optimal route if there is one
func (k Keeper) IterateRoutes(routes []string) (string, error) {
	var optimalRoute string
	var localMax uint64

	for _, route := range routes {
		optimalInputAmount := k.CalculateInputAmount(route)
		profit := k.CalculateProfit(route, optimalInputAmount)

		if profit > localMax {
			optimalRoute = route
			localMax = profit
		}
	}

	return optimalRoute, nil
}

// CalculateInputAmount returns the optimal amount that should be inputted for a given route
func (k Keeper) CalculateInputAmount(route string) uint64 {
	return 0
}

// CalculateProfit calculates the profit that would be achieved if this route is taken
func (k Keeper) CalculateProfit(route string, amountIn uint64) uint64 {
	return 0
}

// ExecuteTrade inputs a route, amount in, and rebalances the pool
func (k Keeper) ExecuteTrade() {

}

// FindRoutes that should be traversed in determining the optimal cyclic arbitrage opportunities
// Given a pool id and input denomination, this will return a list of routes that are a union of
// avaiable routes in the hot routes and the 3-hop routes.
func (k Keeper) FindRoutes() {

}
