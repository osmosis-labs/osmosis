package client

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
)

func emptyHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
