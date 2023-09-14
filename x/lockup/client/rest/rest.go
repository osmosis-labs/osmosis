package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

// REST Variable names

const (
	LockID           = "lock-id"
	RestOwnerAddress = "owner"
	RestDenom        = "denom"
	RestTimestamp    = "timestamp"
	RestDuration     = "duration"
)

// RegisterRoutes register query and tx rest routes.
func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	registerQueryRoutes(clientCtx, rtr)
	registerTxHandlers(clientCtx, rtr)
}
