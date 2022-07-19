package apptesting

import "testing"

func BenchmarkSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := new(KeeperTestHelper)
		s.Setup()
	}
}
