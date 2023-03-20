package cosmwasmpool_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/mocks"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

type PoolModuleSuite struct {
	apptesting.KeeperTestHelper
}

func TestPoolModuleSuite(t *testing.T) {
	suite.Run(t, new(PoolModuleSuite))
}

func (suite *PoolModuleSuite) SetupTest() {
	suite.Setup()
}

func (s *PoolModuleSuite) TestInitializePool() {
	var (
		defaultPoolId = uint64(1)
		validTestPool = &model.Pool{
			PoolStoreModel: model.PoolStoreModel{
				PoolAddress:     gammtypes.NewPoolAddress(defaultPoolId).String(),
				ContractAddress: "", // N.B.: to be set in InitializePool()
				PoolId:          defaultPoolId,
				CodeId:          1,
				InstantiateMsg:  []byte(nil),
			},
		}
	)

	tests := map[string]struct {
		mockInstantiateReturn struct {
			contractAddress sdk.AccAddress
			data            []byte
			err             error
		}
		isValidPool bool
		expectError error
	}{
		"valid pool": {
			isValidPool: true,
		},
		"invalid pool": {
			isValidPool: false,
			expectError: types.InvalidPoolTypeError{},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.SetupTest()
			cosmwasmPoolKeeper := s.App.CosmwasmPoolKeeper

			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			var testPool poolmanagertypes.PoolI
			if tc.isValidPool {
				testPool = validTestPool

				mockContractKeeper := mocks.NewMockContractKeeper(ctrl)
				mockContractKeeper.EXPECT().Instantiate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tc.mockInstantiateReturn.contractAddress, tc.mockInstantiateReturn.data, tc.mockInstantiateReturn.err)
				cosmwasmPoolKeeper.SetContractKeeper(mockContractKeeper)
			} else {
				testPool = s.PrepareConcentratedPool()
			}

			err := cosmwasmPoolKeeper.InitializePool(s.Ctx, testPool, s.TestAccs[0])

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectError)
				return
			}
			s.Require().NoError(err)

			pool, err := cosmwasmPoolKeeper.GetPool(s.Ctx, defaultPoolId)
			s.Require().NoError(err)

			cosmWasmPool, ok := pool.(*model.Pool)
			s.Require().True(ok)

			// Check that the pool's contract address is set
			s.Require().Equal(tc.mockInstantiateReturn.contractAddress.String(), cosmWasmPool.GetContractAddress())

			// Check that the pool's data is set
			expectedPool := validTestPool
			expectedPool.ContractAddress = tc.mockInstantiateReturn.contractAddress.String()
			s.Require().Equal(expectedPool.PoolStoreModel, cosmWasmPool.PoolStoreModel)
		})
	}
}
