package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

const (
	MethodGet = "GET"
)

// RegisterRoutes registers claim-related REST handlers to a router.
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
}
