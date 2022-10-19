package concentrated_liquidity

import (
	"fmt"
	"math"
	"math/bits"
)

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

// NextInitializedTickWithinOneWord returns the next initialized tick contained
// in the same word (or adjacent word) as the tick that is either
// to the left (less than or equal to) or right (greater than) of the given tick.
//
// In other words, it returns the next initialized or uninitialized tick up to
// 64 ticks away from the current tick and whether that next tick is initialized,
// as the function only searches within up to 64 ticks.
func (tb *TickBitmap) NextInitializedTickWithinOneWord(tickIndex, tickSpacing int32, lte bool) (next int32, initialized bool) {
	compressed := tickIndex / tickSpacing

	// round towards negative infinity
	if tickIndex < 0 && tickIndex%tickSpacing != 0 {
		compressed--
	}

	if lte {
		wordPos, bitPos := tickPosition(compressed)

		// all the 1s at or to the right of the current bitPos
		bitMask := uint64((1 << bitPos) - 1 + (1 << bitPos))
		masked := tb.bitmap[wordPos] & bitMask

		// If there are no initialized ticks to the right of or at the current tick,
		// return rightmost in the word.
		initialized = masked != 0

		// Note, overflow/underflow is possible, but prevented externally by
		// limiting both tickSpacing and tick.
		if initialized {
			msbIndex := uint8(64 - bits.LeadingZeros64(masked) - 1)
			next = (compressed - int32(uint32(bitPos-msbIndex))) * tickSpacing
		} else {
			next = (compressed - int32(uint32(bitPos))) * tickSpacing
		}

		return next, initialized
	}

	// Start from the word of the next tick, since the current tick state doesn't
	// matter.
	wordPos, bitPos := tickPosition(compressed + 1)

	// all the 1s at or to the left of the bitPos
	bitMask := uint64(^((1 << bitPos) - 1))
	masked := tb.bitmap[wordPos] & bitMask

	// If there are no initialized ticks to the left of the current tick, return
	// leftmost in the word.
	initialized = masked != 0

	// Note, overflow/underflow is possible, but prevented externally by limiting
	// both tickSpacing and tick.
	if initialized {
		lsbIndex := uint8(bits.TrailingZeros64(masked))
		next = (compressed + 1 + int32(uint32((lsbIndex - bitPos)))) * tickSpacing
	} else {
		next = (compressed + 1 + int32(uint32(math.MaxUint8-bitPos))) * tickSpacing
	}

	return next, initialized
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
