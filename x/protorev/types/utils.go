package types

// CreateSeacherRoutes creates a new TokenPairArbRoutes for testing purposes
func CreateSeacherRoutes(numRoutes int, swapIn, swapOut, tokenInDenom, tokenOutDenom string) TokenPairArbRoutes {
	routes := make([]*Route, numRoutes)
	for i := 0; i < numRoutes; i++ {
		trades := make([]*Trade, 3)

		firstTrade := NewTrade(0, tokenInDenom, swapIn)
		trades[0] = &firstTrade

		secondTrade := NewTrade(1, swapIn, swapOut)
		trades[1] = &secondTrade

		thirdTrade := NewTrade(2, swapOut, tokenOutDenom)
		trades[2] = &thirdTrade

		newRoutes := NewRoutes(trades)
		routes[i] = &newRoutes
	}

	return NewTokenPairArbRoutes(routes, swapIn, swapOut)
}
