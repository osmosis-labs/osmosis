package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
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
	r := clientrest.WithHTTPDeprecationHeaders(rtr)
	registerQueryRoutes(clientCtx, r)
	registerTxHandlers(clientCtx, r)
}

// LockTokensReq defines the properties of a MsgLockTokens request.
type LockTokensReq struct {
	BaseReq  rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Owner    sdk.AccAddress `json:"owner,omitempty" yaml:"owner"`
	Duration string         `json:"duration,omitempty" yaml:"duration"`
	Coins    sdk.Coins      `json:"coins" yaml:"coins"`
}
