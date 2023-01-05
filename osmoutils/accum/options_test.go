package accum_test

import "github.com/osmosis-labs/osmosis/osmoutils/accum"

// TestOptionsValidate tests that the options are validated correctly.
func (suite *AccumTestSuite) TestOptionsValidate() {

	tests := map[string]struct {
		options     *accum.Options
		expectError error
	}{
		"nil options - success": {
			options: nil,
		},
		"non-nil options - success": {
			options: &accum.Options{},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()

			err := tc.options.Validate()

			if tc.expectError != nil {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)
		})
	}
}
