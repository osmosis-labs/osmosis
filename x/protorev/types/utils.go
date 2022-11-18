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

// Checks if both denoms match the Atom <-> Osmo pair
func CheckPerfectMatch(tokenA, tokenB string) bool {
	if tokenA == OsmosisDenomination && tokenB == AtomDenomination {
		return true
	} else if tokenA == AtomDenomination && tokenB == OsmosisDenomination {
		return true
	}

	return false
}

// Checks if one denom matches the tokenToMatch and returns a boolean and the other denom
func CheckMatch(tokenA, tokenB, tokenToMatch string) (string, bool) {
	if tokenA == tokenToMatch {
		return tokenB, true
	} else if tokenB == tokenToMatch {
		return tokenA, true
	}

	return "", false
}
