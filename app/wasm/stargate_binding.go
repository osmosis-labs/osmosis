package wasm

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"

	gammtypes "github.com/osmosis-labs/osmosis/v10/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v10/x/incentives/types"
)

func StargateQuerier(queryRouter *baseapp.GRPCQueryRouter, codec codec.Codec) func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
		// reqBinding, whitelisted := StargateLayerRequestBindings.Load(request.Path)
		// if !whitelisted {
		// 	return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", request.Path)}
		// }
		_, whitelisted := StargateLayerRequestBindings.Load(request.Path)
		if !whitelisted {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", request.Path)}
		}

		route := queryRouter.Route(request.Path)
		if route == nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("No route to query '%s'", request.Path)}
		}

		req := abci.RequestQuery{
			Data: request.Data,
			Path: request.Path,
		}
		res, err := route(ctx, req)
		if err != nil {
			return nil, err
		}

		resBinding, whitelisted := StargateLayerResponseBindings.Load(request.Path)
		if !whitelisted {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", request.Path)}
		}

		bz, err := NormalizeReponsesAndJsonfy(resBinding, res.Value, codec)
		if err != nil {
			return nil, err
		}
		return bz, nil
	}
}

func NormalizeRequestsAndUnjsonfy(binding interface{}, bz []byte, codec codec.Codec) ([]byte, error) {
	// all values are proto message
	message, ok := binding.(proto.Message)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}

	err := codec.UnmarshalJSON(bz, message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	bz, err = proto.Marshal(message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	return bz, nil
}

func NormalizeReponsesAndJsonfy(binding interface{}, bz []byte, codec codec.Codec) ([]byte, error) {
	// all values are proto message
	message, ok := binding.(proto.Message)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}

	// unmarshal binary into stargate response data structure
	err := proto.Unmarshal(bz, message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	// build new deterministic response
	_, err = proto.Marshal(message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	// clear proto message
	message.Reset()

	err = proto.Unmarshal(bz, message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	// jsonfy
	bz, err = codec.MarshalJSON(message)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	return bz, nil
}

var StargateLayerRequestBindings sync.Map
var StargateLayerResponseBindings sync.Map

func init() {
	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolRequest{})
	StargateLayerRequestBindings.Store("/osmosis.gamm.v1beta1.Query/Pools", &gammtypes.QueryPoolRequest{})
	StargateLayerRequestBindings.Store("/osmosis.incentives.v1beta1.Query/LockableDurations", &incentivestypes.QueryLockableDurationsRequest{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolResponse{})
	StargateLayerResponseBindings.Store("/osmosis.gamm.v1beta1.Query/Pool", &gammtypes.QueryPoolResponse{})
	StargateLayerResponseBindings.Store("/osmosis.incentives.v1beta1.Query/LockableDurations", &incentivestypes.QueryLockableDurationsResponse{})
}
