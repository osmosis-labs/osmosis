package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

// RegisterRoutes register query and tx rest routes.
func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	registerQueryRoutes(clientCtx, rtr)
	registerTxHandlers(clientCtx, rtr)
}
