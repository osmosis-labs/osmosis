package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
)

type CWPoolGovTypesSuite struct {
	apptesting.KeeperTestHelper
}

func TestCWPoolGovTypesSuite(t *testing.T) {
	suite.Run(t, new(CWPoolGovTypesSuite))
}

// TestValidateMigrationProposalCondiguration tests that the migration proposal configuration is validated correctly.
// It validates the following cases:
// 1. Success: pool ids are set, code id is set and byte code is not.
// 2. Success: pool ids are set, code id is not set and byte code is set.
// 3. Error: pool ids are not set, code id is set and byte code is not.
// 4. Error: pool ids are set but both code id and byte code are set at the same time.
// 5. Error: pool ids are set but both code id and byte code are unset.
// See method spec for more details as to why these vectors are chosen.
func (s *CWPoolGovTypesSuite) TestValidateMigrationProposalCondiguration() {
	// Get valid transmuter code.
	validTransmuterCode := s.GetContractCode(apptesting.TransmuterContractName)

	const (
		preUploadCodeIdPlaceholder  uint64 = 1000
		zeroCodeId                  uint64 = 0
		defaultPoolCountToPreCreate uint64 = 3
	)

	var (
		emptyByteCode           []byte = []byte{}
		emptyPoolIds            []uint64
		defaultPoolIdsToMigrate = []uint64{1, 2, 3}
	)

	tests := []struct {
		name      string
		poolIds   []uint64
		newCodeId uint64
		byteCode  []byte

		expectedErr error
	}{
		{
			name:      "success: pool ids are set, code id is set and byte code is not",
			poolIds:   defaultPoolIdsToMigrate,
			newCodeId: preUploadCodeIdPlaceholder,
			byteCode:  emptyByteCode,
		},
		{
			name:     "success: pool ids are set, code id is not set and byte code is",
			poolIds:  defaultPoolIdsToMigrate,
			byteCode: validTransmuterCode,
		},
		{
			name:      "error: pool ids are not set, code id is set and byte code is not",
			poolIds:   emptyPoolIds,
			newCodeId: preUploadCodeIdPlaceholder,
			byteCode:  emptyByteCode,

			expectedErr: types.ErrEmptyPoolIds,
		},
		{
			name:      "error: pool ids are set but both code id and byte code are set at the same time",
			poolIds:   defaultPoolIdsToMigrate,
			newCodeId: preUploadCodeIdPlaceholder,
			byteCode:  validTransmuterCode,

			expectedErr: types.ErrBothOfCodeIdAndContractCodeSpecified,
		},
		{
			name:      "error: pool ids are set but both code id and byte code are NOT set",
			poolIds:   defaultPoolIdsToMigrate,
			newCodeId: 0,
			byteCode:  emptyByteCode,

			expectedErr: types.ErrNoneOfCodeIdAndContractCodeSpecified,
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			// System under test.
			err := types.ValidateMigrationProposalConfiguration(tc.poolIds, tc.newCodeId, tc.byteCode)

			if tc.expectedErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectedErr)
				return
			}

			s.Require().NoError(err)
		})
	}
}
