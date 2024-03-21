package cosmwasm

import (
	"encoding/json"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DefaultContractCallGasLimit = 30_000_000

// ContracKeeper defines the interface needed to be fulfilled for
// the ContractKeeper.
type ContractKeeper interface {
	Instantiate(
		ctx sdk.Context,
		codeID uint64,
		creator, admin sdk.AccAddress,
		initMsg []byte,
		label string,
		deposit sdk.Coins,
	) (sdk.AccAddress, []byte, error)

	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)

	Execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) ([]byte, error)
}

// WasmKeeper defines the interface needed to be fulfilled for
// the WasmKeeper.
type WasmKeeper interface {
	QuerySmart(ctx sdk.Context, contractAddress sdk.AccAddress, queryMsg []byte) ([]byte, error)
	QueryGasLimit() storetypes.Gas
}

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
func Query[T any, K any](ctx sdk.Context, wasmKeeper WasmKeeper, contractAddress string, request T) (response K, err error) {
	bz, err := json.Marshal(request)
	if err != nil {
		return response, err
	}

	childCtx := ctx.WithGasMeter(storetypes.NewGasMeter(wasmKeeper.QueryGasLimit()))
	responseBz, err := wasmKeeper.QuerySmart(childCtx, sdk.MustAccAddressFromBech32(contractAddress), bz)
	if err != nil {
		return response, err
	}

	ctx.GasMeter().ConsumeGas(childCtx.GasMeter().GasConsumed(), "query smart")

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
func MustQuery[T any, K any](ctx sdk.Context, wasmKeeper WasmKeeper, contractAddress string, request T) (response K) {
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
func Sudo[T any, K any](ctx sdk.Context, contractKeeper ContractKeeper, contractAddress string, request T) (response K, err error) {
	bz, err := json.Marshal(request)
	if err != nil {
		return response, err
	}

	// Defer to catch panics in case the sudo call runs out of gas.
	defer func() {
		if r := recover(); r != nil {
			var emptyResponse K
			response = emptyResponse
			err = fmt.Errorf("contract call ran out of gas")
		}
	}()

	// Make contract call with a gas limit of 30M to ensure contracts cannot run unboundedly
	gasLimit := min(ctx.GasMeter().Limit(), DefaultContractCallGasLimit)
	childCtx := ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))
	responseBz, err := contractKeeper.Sudo(childCtx, sdk.MustAccAddressFromBech32(contractAddress), bz)
	if err != nil {
		return response, err
	}

	// Consume gas used for calling contract to the parent ctx
	ctx.GasMeter().ConsumeGas(childCtx.GasMeter().GasConsumed(), "Track contract call gas")

	// valid empty response
	if len(responseBz) == 0 {
		return response, nil
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
func MustSudo[T any, K any](ctx sdk.Context, contractKeeper ContractKeeper, contractAddress string, request T) (response K) {
	response, err := Sudo[T, K](ctx, contractKeeper, contractAddress, request)
	if err != nil {
		panic(err)
	}

	return response
}

// Execute is a generic function to execute a contract call on a given contract address with a specified request.
// It accepts a context, contract keeper, contract address, caller address, coins, and request data.
// This function works with any data type for the request and response (T and K).
// It marshals the request data into JSON format, executes the contract call, and then unmarshals the response
// data back into the specified response type (K). In case of any error, it returns the zero value of the response
// type along with the error.
//
// Parameters:
// - ctx: The SDK context.
// - contractKeeper: An instance of the contract keeper to manage contract interactions.
// - contractAddress: The bech32 address of the contract to be executed.
// - caller: The address of the account making the call.
// - coins: The coins to be transferred during the call.
// - request: The request data, can be of any data type.
//
// Returns:
// - response: The response data, can be of any data type (K). Returns the zero value of K in case of an error.
// - err: An error object that indicates any error during the contract execution or data marshalling/unmarshalling process.
func Execute[T any, K any](ctx sdk.Context, contractKeeper ContractKeeper, contractAddress string, caller sdk.AccAddress, coins sdk.Coins, request T) (response K, err error) {
	bz, err := json.Marshal(request)
	if err != nil {
		return response, err
	}

	responseBz, err := contractKeeper.Execute(ctx, sdk.MustAccAddressFromBech32(contractAddress), caller, bz, coins)
	if err != nil {
		return response, err
	}

	// valid empty response
	if len(responseBz) == 0 {
		return response, nil
	}

	if err := json.Unmarshal(responseBz, &response); err != nil {
		return response, err
	}

	return response, nil
}

// MustExecute is a wrapper around the Execute function, which provides a more concise API for
// executing a contract call on a given contract address with a specified request.
// It works with any data type for the request and response (T and K).
// This function panics if an error occurs during the contract execution or data marshalling/unmarshalling process.
// Use this function when you're confident that the contract execution will not encounter any errors.
//
// Parameters:
// - ctx: The SDK context.
// - contractKeeper: An instance of the contract keeper to manage contract interactions.
// - contractAddress: The bech32 address of the contract to be executed.
// - caller: The address of the account making the call.
// - coins: The coins to be transferred during the call.
// - request: The request data, can be of any data type.
//
// Returns:
// - response: The response data, can be of any data type (K).
func MustExecute[T any, K any](ctx sdk.Context, contractKeeper ContractKeeper, contractAddress string, caller sdk.AccAddress, coins sdk.Coins, request T) (response K) {
	response, err := Execute[T, K](ctx, contractKeeper, contractAddress, caller, coins, request)
	if err != nil {
		panic(err)
	}

	return response
}
