package types

// Checks if the matching variable matches one of the tokens and if so returns the other and true
func CheckOsmoAtomDenomMatch(tokenA, tokenB, match string) (string, bool) {
	if tokenA == match {
		return tokenB, true
	} else if tokenB == match {
		return tokenA, true
	}
	return "", false
}

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
