package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
	"net/http"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	fmt.Printf("Register")
	r.HandleFunc(
		"/hello",
		helloHandlerFn(clientCtx),
	).Methods("GET")
}

func helloHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("Hello World"))
	}
}
