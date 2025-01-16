package writelistener_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v29/app/apptesting"
	"github.com/osmosis-labs/osmosis/v29/ingest/common/pooltracker"
	"github.com/osmosis-labs/osmosis/v29/ingest/common/writelistener"
	"github.com/osmosis-labs/osmosis/v29/x/concentrated-liquidity/model"
	concentratedtypes "github.com/osmosis-labs/osmosis/v29/x/concentrated-liquidity/types"
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

// Tests that the concentrated write listener correctly tracks pool and tick updates
// It ignores error cases as they are simple and unlikely giving app wiring.
func (s *WriteListenerTestSuite) TestWriteListener_Concentrated() {

	// Set up chain state once per test
	s.Setup()

	concentratedPool := s.PrepareConcentratedPool()
	concentratedPoolModel, ok := concentratedPool.(*model.Pool)
	s.Require().True(ok)

	concentratedPoolModelBz, err := concentratedPoolModel.Marshal()
	s.Require().NoError(err)

	testCases := []struct {
		name string

		key      []byte
		value    []byte
		isDelete bool

		expectedPoolUpdate     bool
		expectedPoolTickUpdate bool
	}{
		{
			name: "concentrated write unrelated to pool state, no-op",

			key:   concentratedtypes.KeyAuthorizedTickSpacing,
			value: someValue,
		},
		{
			name: "write concentrated pool index",

			key:   concentratedtypes.KeyPool(concentratedPoolModel.Id),
			value: concentratedPoolModelBz,

			expectedPoolUpdate: true,
		},
		{
			name: "delete concentrated pool index",

			key:   concentratedtypes.KeyPool(concentratedPoolModel.Id),
			value: concentratedPoolModelBz,

			expectedPoolUpdate: true,
		},
		{
			name: "write concentrated tick index",

			key:   concentratedtypes.KeyTick(concentratedPoolModel.Id, 1),
			value: concentratedPoolModelBz,

			expectedPoolTickUpdate: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {

			poolTracker := pooltracker.NewMemory()

			concentratedWriteListener := writelistener.NewConcentrated(poolTracker)

			concentratedKVStore := s.App.GetKey(concentratedtypes.ModuleName)

			err := concentratedWriteListener.OnWrite(concentratedKVStore, tc.key, tc.value, tc.isDelete)
			s.Require().NoError(err)

			// All non-concentrated pool updates should not be tracked.
			s.Require().Empty(poolTracker.GetCFMMPools())
			s.Require().Empty(poolTracker.GetCosmWasmPools())

			if tc.expectedPoolUpdate {
				concentratedPools := poolTracker.GetConcentratedPools()
				s.Require().Len(concentratedPools, 1)

				s.Require().Equal(concentratedPoolModel, concentratedPools[0])
			} else {
				concentratedPools := poolTracker.GetConcentratedPools()
				s.Require().Len(concentratedPools, 0)
			}

			if tc.expectedPoolTickUpdate {
				concentratedPoolIDTickChange := poolTracker.GetConcentratedPoolIDTickChange()
				s.Require().Len(concentratedPoolIDTickChange, 1)

				// Check that the pool ID is the one we expect
				_, ok := concentratedPoolIDTickChange[concentratedPoolModel.Id]
				s.Require().True(ok)
			} else {
				concentratedPoolIDTickChange := poolTracker.GetConcentratedPoolIDTickChange()
				s.Require().Len(concentratedPoolIDTickChange, 0)
			}
		})
	}
}
