package concentrated_liquidity

// TickBitmap defines a bitmap used to represent price ticks. It contains a
// mapping of 64-bit words, where each bit in a word corresponds to a unique
// tick. A set bit, i.e. a bit set to 1, indicates liquidity for that tick.
// Conversely, an unset bit means there is no liquidity for that tick.
type TickBitmap struct {
	bitmap map[int16]uint64
}

// Position returns the word and bit position in the tick bitmap given a tick
// index.
func (tb *TickBitmap) Position(tickIndex int64) (wordPos uint16, bitPos uint8) {
	// Perform an arithmetic right shift operation identical to integer division
	// by 64. Word position is the integer part of a tick index divided by 64.
	wordPos = uint16(tickIndex >> 6)

	// find the bit position in the word that corresponds to the tick.
	bitPos = uint8(uint64(tickIndex % 64))

	return wordPos, bitPos
}
