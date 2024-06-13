package writelistener_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/service/writelistener"
)

type WriteListenerTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

var (
	someValue = []byte("someValue")
)

func TestWriteListenerTestSuite(t *testing.T) {
	suite.Run(t, new(WriteListenerTestSuite))
}

func (s *WriteListenerTestSuite) TestWriteListener_Bank() {

	// Set up chain state once per test
	// TODO: potentially run setup in the loop
	s.Setup()

	testCases := []struct {
		name string

		key      []byte
		value    []byte
		isDelete bool

		// TODO: add expected fields for the test.
	}{
		// TODO: add tests
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {

			bankWriteListener := writelistener.NewBank()

			bankKVStore := s.App.GetKey(banktypes.ModuleName)

			err := bankWriteListener.OnWrite(bankKVStore, tc.key, tc.value, tc.isDelete)
			s.Require().NoError(err)

			// TODO: assertions
		})
	}
}
