package types

import "encoding/binary"

// Converts a uint64 to a []byte
func UInt64ToBytes(number uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, number)
	return bz
}

// Converts a []byte into a uint64
func BytesToUInt64(bz []byte) uint64 {
	return binary.LittleEndian.Uint64(bz)
}

// Checks if the matching variable matches one of the tokens and if so returns the other and true
func CheckMatchAndReturnOther(tokenA, tokenB, match string) (string, bool) {
	if tokenA == match {
		return tokenB, true
	} else if tokenB == match {
		return tokenA, true
	}
	return "", false
}

func CreateSeacherRoutes(numRoutes int, tokenInDenom, tokenOutDenom string) TokenPairArbRoutes {
	routes := make([]*Route, numRoutes)
	for i := 0; i < numRoutes; i++ {
		trades := make([]*Trade, 3)

		for j := 0; j < 3; j++ {
			trade := NewTrade(uint64((j+1)*(1+i)), "a", "b")
			trades[j] = &trade
		}
		newRoutes := NewRoutes(trades)
		routes[i] = &newRoutes
	}

	return NewTokenPairArbRoutes(routes, tokenInDenom, tokenOutDenom)
}
