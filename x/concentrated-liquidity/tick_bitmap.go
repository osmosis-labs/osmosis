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

func (tb *TickBitmap) FlipTick(tickIndex, tickSpacing int32) error {
	//   /// @notice Flips the initialized state for a given tick from false to true, or vice versa
	//   /// @param self The mapping in which to flip the tick
	//   /// @param tick The tick to flip
	//   /// @param tickSpacing The spacing between usable ticks
	//   function flipTick(
	// 		mapping(int16 => uint256) storage self,
	// 		int24 tick,
	// 		int24 tickSpacing
	// ) internal {
	// 		require(tick % tickSpacing == 0); // ensure that the tick is spaced
	// 		(int16 wordPos, uint8 bitPos) = position(tick / tickSpacing);
	// 		uint256 mask = 1 << bitPos;
	// 		self[wordPos] ^= mask;
	// }
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
