package concentrated_liquidity_test

import (
	"fmt"
	"testing"
)

func TestFoo(t *testing.T) {
	// 	function position(int24 tick) private pure returns (int16 wordPos, uint8 bitPos) {
	//     wordPos = int16(tick >> 8);
	//     bitPos = uint8(uint24(tick % 256));
	// }

	// tick = 85176
	// word_pos = tick >> 8 # or tick // 2**8
	// bit_pos = tick % 256
	// print(f"Word {word_pos}, bit {bit_pos}")
	// # Word 332, bit 184

	tick := 85176
	wordPos := tick >> 8

	fmt.Println(wordPos)
}
