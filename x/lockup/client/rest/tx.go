package rest

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/osmosis-labs/osmosis/v11/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/lockup/locktokens", newLockTokensHandlerFn(clientCtx)).Methods("POST")
}

func newLockTokensHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LockTokensReq
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		duration, err := time.ParseDuration(req.Duration)
		if err != nil {
			return
		}

		msg := types.NewMsgLockTokens(
			req.Owner,
			duration,
			req.Coins,
		)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
