package twap_test

// TestTrackChangedPool takes a list of poolIds as test cases, and runs one list per block.
// Every simulated block, checks that all cumulatively tracked pool IDs thus far are not marked as changed.
// Then runs k.trackChangedPool on every item in the test case list.
// Appends these all to our cumulative list.
// Finally checks that hasPoolChangedThisBlock registers the items in our list.
//
// This achieves testing the functionality that we depend on, that this clears every end block.
func (s *TestSuite) TestTrackChangedPool() {
	cumulativeIds := map[uint64]bool{}
	tests := map[string][]uint64{
		"single":     {1},
		"duplicated": {1, 1},
		"four":       {1, 2, 3, 4},
	}
	for name, test := range tests {
		s.Run(name, func() {
			// Test that no cumulative ID registers as tracked
			for k := range cumulativeIds {
				s.Require().False(s.twapkeeper.HasPoolChangedThisBlock(s.Ctx, k))
			}
			// Track every pool in list
			for _, v := range test {
				cumulativeIds[v] = true
				s.twapkeeper.TrackChangedPool(s.Ctx, v)
			}
			for _, v := range test {
				s.Require().True(s.twapkeeper.HasPoolChangedThisBlock(s.Ctx, v))
			}
			s.Commit()
		})
	}
}
