package e2eTesting

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	proto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"

	"github.com/osmosis-labs/osmosis/v26/app"
)

var _ grpc.ClientConnInterface = (*grpcClient)(nil)

type grpcClient struct {
	app *app.OsmosisApp
}

func (c grpcClient) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	req := args.(proto.Message)
	resp, err := c.app.Query(ctx, &abci.RequestQuery{
		Data:   c.app.AppCodec().MustMarshal(req),
		Path:   method,
		Height: 0, // TODO: heightened queries
		Prove:  false,
	})
	if err != nil {
		return err
	}

	if resp.Code != abci.CodeTypeOK {
		return fmt.Errorf(resp.Log)
	}

	c.app.AppCodec().MustUnmarshal(resp.Value, reply.(proto.Message))

	return nil
}

func (c grpcClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("not supported")
}

func (chain *TestChain) Client() grpc.ClientConnInterface {
	return grpcClient{app: chain.app}
}
