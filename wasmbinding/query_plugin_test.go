package wasmbinding_test

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	epochtypes "github.com/osmosis-labs/osmosis/v11/x/epochs/types"

	"github.com/osmosis-labs/osmosis/v11/wasmbinding"
)

func (suite *StargateTestSuite) TestStargateQuerier() {
	testCases := []struct {
		name                   string
		path                   string
		requestData            func() []byte
		responseProtoStruct    interface{}
		expectedQuerierError   bool
		expectedUnMarshalError bool
	}{
		{
			name: "happy path",
			path: "/osmosis.epochs.v1beta1.Query/EpochInfos",
			requestData: func() []byte {
				epochrequest := epochtypes.QueryEpochsInfoRequest{}
				bz, err := proto.Marshal(&epochrequest)
				suite.Require().NoError(err)
				return bz
			},
			responseProtoStruct: &epochtypes.QueryEpochsInfoResponse{},
		},
		{
			name: "unregsitered path(not whitelisted)",
			path: "/osmosis.epochs.v1beta1.Query/CurrentEpoch",
			requestData: func() []byte {
				currentEpochRequest := epochtypes.QueryCurrentEpochRequest{}
				bz, err := proto.Marshal(&currentEpochRequest)
				suite.Require().NoError(err)
				return bz
			},
			expectedQuerierError: true,
		},
		{
			name: "unmatching path and data in request",
			path: "/osmosis.epochs.v1beta1.Query/EpochInfos",
			requestData: func() []byte {
				epochrequest := epochtypes.QueryCurrentEpochRequest{}
				bz, err := proto.Marshal(&epochrequest)
				suite.Require().NoError(err)
				return bz
			},
			responseProtoStruct:    &epochtypes.QueryCurrentEpochResponse{},
			expectedUnMarshalError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			stargateQuerier := wasmbinding.StargateQuerier(*suite.app.GRPCQueryRouter(), suite.app.AppCodec())
			stargateRequest := &wasmvmtypes.StargateQuery{
				Path: tc.path,
				Data: tc.requestData(),
			}
			stargateResponse, err := stargateQuerier(suite.ctx, stargateRequest)
			if tc.expectedQuerierError {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)

				protoResponse, ok := tc.responseProtoStruct.(proto.Message)
				suite.Require().True(ok)

				// test correctness by unmarshalling json response into proto struct
				err = suite.app.AppCodec().UnmarshalJSON(stargateResponse, protoResponse)
				if tc.expectedUnMarshalError {
					suite.Require().Error(err)
				} else {
					suite.Require().NoError(err)
					suite.Require().NotNil(protoResponse)
				}
			}
		})
	}
}
