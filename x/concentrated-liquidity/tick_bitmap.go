package concentrated_liquidity

import fmt "fmt"

// TickBitmap defines a bitmap used to represent price ticks. It contains a
// mapping of 64-bit words, where each bit in a word corresponds to a unique
// tick. A set bit, i.e. a bit set to 1, indicates liquidity for that tick.
// Conversely, an unset bit means there is no liquidity for that tick. Note,
// ticks are in the range [−887272, 887272].
//
// Ref: https://uniswapv3book.com/docs/introduction/uniswap-v3/#ticks
// Ref: https://uniswapv3book.com/docs/milestone_2/tick-bitmap-index/#bitmap
type TickBitmap struct {
	bitmap map[int16]uint64
}

// FlipTick flips the tick for the given tick index from false (no liquidity) to
// true (liquidity) or vice versa. The tickSpacing parameter defines the spacing
// between usable ticks and must be a multiple of the tick index.
func (tb *TickBitmap) FlipTick(tickIndex, tickSpacing int32) error {
	if tickIndex%tickSpacing != 0 {
		return fmt.Errorf("tickIndex %d is not a multiple of tickSpacing %d", tickIndex, tickSpacing)
	}

	wordPos, bitPos := tickPosition(tickIndex / tickSpacing)
	bitMask := uint64(1 << bitPos)
	tb.bitmap[wordPos] ^= bitMask

	return nil
}

// tickPosition returns the word and bit position in the tick bitmap given a
// tick index.
func tickPosition(tickIndex int32) (wordPos int16, bitPos uint8) {
	// Perform an arithmetic right shift operation identical to integer division
	// by 64. Word position is the integer part of a tick index divided by 64.
	wordPos = int16(tickIndex >> 6)

	// find the bit position in the word that corresponds to the tick.
	bitPos = uint8(uint32(tickIndex % 64))

	return wordPos, bitPos
}
