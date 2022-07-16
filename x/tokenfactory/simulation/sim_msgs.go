package simulation

import (
	"math/rand"

	simulation "github.com/osmosis-labs/osmosis/v7/simulation/types"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RandomMsgCreateDenom creates a random tokenfactory denom that is no greater than 44 alphanumeric characters
func RandomMsgCreateDenom(k keeper.Keeper, sim *simulation.SimCtx, ctx sdk.Context) (*types.MsgCreateDenom, error) {
	// select a pseudo-random simulation account
	sender := sim.RandomSimAccount()

	subdenom := sim.RandStringOfLength(44)

	return &types.MsgCreateDenom{
		Sender:   sender.Address.String(),
		Subdenom: subdenom,
	}, nil
}
