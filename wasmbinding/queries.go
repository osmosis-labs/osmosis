package wasmbinding

import (
	"fmt"

	interchainquerieskeeper "github.com/osmosis-labs/osmosis/v20/x/interchainqueries/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/wasmbinding/bindings"
	tokenfactorykeeper "github.com/osmosis-labs/osmosis/v20/x/tokenfactory/keeper"
)

type QueryPlugin struct {
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
	icqKeeper          *interchainquerieskeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(tfk *tokenfactorykeeper.Keeper, icqk *interchainquerieskeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		tokenFactoryKeeper: tfk,
		icqKeeper:          icqk,
	}
}

// GetDenomAdmin is a query to get denom admin.
func (qp QueryPlugin) GetDenomAdmin(ctx sdk.Context, denom string) (*bindings.DenomAdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin for denom: %s", denom)
	}

	return &bindings.DenomAdminResponse{Admin: metadata.Admin}, nil
}

//func (qp *QueryPlugin) GetInterchainQueryResult(ctx sdk.Context, queryID uint64) (*bindings.QueryRegisteredQueryResultResponse, error) {
//	grpcResp, err := qp.icqKeeper.GetQueryResultByID(ctx, queryID)
//	if err != nil {
//		return nil, err
//	}
//	resp := bindings.QueryResult{
//		KvResults: make([]*bindings.StorageValue, 0, len(grpcResp.KvResults)),
//		Height:    grpcResp.GetHeight(),
//		Revision:  grpcResp.GetRevision(),
//	}
//	for _, grpcKv := range grpcResp.GetKvResults() {
//		kv := bindings.StorageValue{
//			StoragePrefix: grpcKv.GetStoragePrefix(),
//			Key:           grpcKv.GetKey(),
//			Value:         grpcKv.GetValue(),
//		}
//		resp.KvResults = append(resp.KvResults, &kv)
//	}
//
//	return &bindings.QueryRegisteredQueryResultResponse{Result: &resp}, nil
//}
