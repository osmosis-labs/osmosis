package osmomath

import "testing"

var (
	testAmount = NewInt(1234567890323344555)
)

// go test -benchmem -run=^$ -bench ^BenchmarkGetPrecomputeOrderOfMagnitude$ github.com/osmosis-labs/osmosis/osmomath -count=6
func BenchmarkGetPrecomputeOrderOfMagnitude(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = GetPrecomputeOrderOfMagnitude(testAmount)
	}
}

// go test -benchmem -run=^$ -benchmem -bench ^BenchmarkOrderOfMagnitude$ github.com/osmosis-labs/osmosis/osmomath -count=6 > old
func BenchmarkOrderOfMagnitude(b *testing.B) {
	// Create a test amount. Change this based on what you expect to be typical inputs for your application.
	testAmount := NewInt(1234567890323344555) // You can adjust this value for different benchmark scenarios.

	for i := 0; i < b.N; i++ {
		_ = OrderOfMagnitude(testAmount.ToLegacyDec())
	}
}
