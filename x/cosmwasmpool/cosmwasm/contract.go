package cosmwasm

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
)

// Query is a generic function to query a CosmWasm smart contract with the given request.
// The function marshals the request into JSON, performs a smart query using the Wasm keeper, and then
// unmarshals the response into the provided response type.
// It returns the response of the specified type and an error if the JSON marshaling, smart query,
// or JSON unmarshaling process fails.
//
// The function uses generics, and T and K represent the request and response types, respectively.
//
// Parameters:
// - ctx: The SDK context.
// - wasmKeeper: The Wasm keeper to query the smart contract.
// - contractAddress: The Bech32 address of the smart contract.
// - request: The request of type T to be sent to the smart contract.
//
// Returns:
// - response: The response of type K from the smart contract.
// - err: An error if the JSON marshaling, smart query, or JSON unmarshaling process fails; otherwise, nil.
func Query[T any, K any](ctx sdk.Context, wasmKeeper types.WasmKeeper, contractAddress string, request T) (response K, err error) {
	bz, err := json.Marshal(request)
	if err != nil {
		return response, err
	}

	responseBz, err := wasmKeeper.QuerySmart(ctx, sdk.MustAccAddressFromBech32(contractAddress), bz)
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(responseBz, &response); err != nil {
		return response, err
	}

	return response, nil
}

// MustQuery is a generic function to query a CosmWasm smart contract with the given request.
// It is similar to the Query function but panics if an error occurs during the query.
// The function uses the Query function to perform the smart query and panics if an error is returned.
//
// The function uses generics, and T and K represent the request and response types, respectively.
//
// Parameters:
// - ctx: The SDK context.
// - wasmKeeper: The Wasm keeper to query the smart contract.
// - contractAddress: The Bech32 address of the smart contract.
// - request: The request of type T to be sent to the smart contract.
//
// Returns:
// - response: The response of type K from the smart contract.
//
// Panics:
// - If an error occurs during the JSON marshaling, smart query, or JSON unmarshaling process.
func MustQuery[T any, K any](ctx sdk.Context, wasmKeeper types.WasmKeeper, contractAddress string, request T) (response K) {
	response, err := Query[T, K](ctx, wasmKeeper, contractAddress, request)
	if err != nil {
		panic(err)
	}

	return response
}

// Sudo is a generic function to execute a sudo message on a CosmWasm smart contract with the given request.
// The function marshals the request into JSON, performs a sudo call using the contract keeper, and then
// unmarshals the response into the provided response type.
// It returns the response of the specified type and an error if the JSON marshaling, sudo call,
// or JSON unmarshaling process fails.
//
// The function uses generics, and T and K represent the request and response types, respectively.
//
// Parameters:
// - ctx: The SDK context.
// - contractKeeper: The Contract keeper to execute the sudo call on the smart contract.
// - contractAddress: The Bech32 address of the smart contract.
// - request: The request of type T to be sent to the smart contract.
//
// Returns:
// - response: The response of type K from the smart contract.
// - err: An error if the JSON marshaling, sudo call, or JSON unmarshaling process fails; otherwise, nil.
func Sudo[T any, K any](ctx sdk.Context, contractKeeper types.ContractKeeper, contractAddress string, request T) (response K, err error) {
	bz, err := json.Marshal(request)
	if err != nil {
		return response, err
	}

	responseBz, err := contractKeeper.Sudo(ctx, sdk.MustAccAddressFromBech32(contractAddress), bz)
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(responseBz, &response); err != nil {
		return response, err
	}

	return response, nil
}

// MustSudo is a generic function to execute a sudo message on a CosmWasm smart contract with the given request.
// It is similar to the Sudo function but panics if an error occurs during the sudo call.
// The function uses the Sudo function to perform the sudo call and panics if an error is returned.
//
// The function uses generics, and T and K represent the request and response types, respectively.
//
// Parameters:
// - ctx: The SDK context.
// - contractKeeper: The Contract keeper to execute the sudo call on the smart contract.
// - contractAddress: The Bech32 address of the smart contract.
// - request: The request of type T to be sent to the smart contract.
//
// Returns:
// - response: The response of type K from the smart contract.
//
// Panics:
// - If an error occurs during the JSON marshaling, sudo call, or JSON unmarshaling process.
func MustSudo[T any, K any](ctx sdk.Context, contractKeeper types.ContractKeeper, contractAddress string, request T) (response K) {
	response, err := Sudo[T, K](ctx, contractKeeper, contractAddress, request)
	if err != nil {
		panic(err)
	}

	return response
}
