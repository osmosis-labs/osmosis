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

	// select a pseudo-random denom using alpha-numeric characters
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	// denom can be no greater than 44 characters
	denomLen := simulation.RandLTBound(sim, 44)

	b := make([]rune, denomLen)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	subdenom := string(b)

	return &types.MsgCreateDenom{
		Sender:   sender.Address.String(),
		Subdenom: subdenom,
	}, nil
}
