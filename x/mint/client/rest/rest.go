package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/gorilla/mux"
)

// RegisterRoutes registers minting module REST handlers on the provided router.
func RegisterRoutes(ClientCtx client.Context, rtr *mux.Router) {
	r := rest.WithHTTPDeprecationHeaders(rtr)
	registerQueryRoutes(ClientCtx, r)
}
