package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// ------------ types/functions to handle a SetHotRoutes CLI TX ------------ //
type Trade struct {
	Pool     uint64 `json:"pool"`
	TokenIn  string `json:"token_in"`
	TokenOut string `json:"token_out"`
}

type ArbRoutes struct {
	Trades   []Trade `json:"trades"`
	StepSize uint64  `json:"step_size"`
}

type hotRoutesInput struct {
	TokenIn   string      `json:"token_in"`
	TokenOut  string      `json:"token_out"`
	ArbRoutes []ArbRoutes `json:"arb_routes"`
}

type createArbRoutesInput []hotRoutesInput

type XCreateHotRoutesInputs hotRoutesInput

type XCreateHotRoutesExceptions struct {
	XCreateHotRoutesInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createArbRoutesInput) UnmarshalJSON(data []byte) error {
	var createHotRoutesE []XCreateHotRoutesExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createHotRoutesE); err != nil {
		return err
	}

	routes := make([]hotRoutesInput, 0)
	for _, route := range createHotRoutesE {
		routes = append(routes, hotRoutesInput(route.XCreateHotRoutesInputs))
	}

	*release = createArbRoutesInput(routes)
	return nil
}

// extractTokenPairArbRoutes builds all of the TokenPairArbRoutes that were extracted from the json file
func (release *createArbRoutesInput) extractTokenPairArbRoutes() []types.TokenPairArbRoutes {
	if release == nil {
		return nil
	}

	tokenPairArbRoutes := make([]types.TokenPairArbRoutes, 0)

	// Iterate through each hot route and construct the token pair arb routes
	for _, hotRoute := range *release {
		current := types.TokenPairArbRoutes{}
		current.TokenIn = hotRoute.TokenIn
		current.TokenOut = hotRoute.TokenOut

		for _, arbRoute := range hotRoute.ArbRoutes {
			currentArbRoute := types.Route{}
			currentArbRoute.StepSize = sdk.NewIntFromUint64(arbRoute.StepSize)

			for _, trade := range arbRoute.Trades {
				currentTrade := types.Trade{}
				currentTrade.Pool = trade.Pool
				currentTrade.TokenIn = trade.TokenIn
				currentTrade.TokenOut = trade.TokenOut
				currentArbRoute.Trades = append(currentArbRoute.Trades, currentTrade)
			}

			current.ArbRoutes = append(current.ArbRoutes, currentArbRoute)
		}

		tokenPairArbRoutes = append(tokenPairArbRoutes, current)
	}

	return tokenPairArbRoutes
}

// BuildSetHotRoutesMsg builds a MsgSetHotRoutes from the provided json file
func BuildSetHotRoutesMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("must provide a json file")
	}

	// Read the json file
	input := &createArbRoutesInput{}
	path := args[0]
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the json file
	if err := input.UnmarshalJSON(contents); err != nil {
		return nil, err
	}

	// Build the msg
	tokenPairArbRoutes := input.extractTokenPairArbRoutes()
	admin := clientCtx.GetFromAddress().String()
	return &types.MsgSetHotRoutes{
		Admin:     admin,
		HotRoutes: tokenPairArbRoutes,
	}, nil
}

// ------------ types/functions to handle a SetPoolWeights CLI TX ------------ //
type createPoolWeightsInput types.PoolWeights

type XCreatePoolWeightsInputs createPoolWeightsInput

type XCreatePoolWeightsExceptions struct {
	XCreatePoolWeightsInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createPoolWeightsInput) UnmarshalJSON(data []byte) error {
	var createPoolWeightsE XCreatePoolWeightsExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createPoolWeightsE); err != nil {
		return err
	}

	*release = createPoolWeightsInput(createPoolWeightsE.XCreatePoolWeightsInputs)
	return nil
}

// BuildSetPoolWeightsMsg builds a MsgSetPoolWeights from the provided json file
func BuildSetPoolWeightsMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("must provide a json file")
	}

	// Read the json file
	input := &createPoolWeightsInput{}
	path := args[0]
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the json file
	if err := input.UnmarshalJSON(contents); err != nil {
		return nil, err
	}

	// Build the msg
	admin := clientCtx.GetFromAddress().String()
	return &types.MsgSetPoolWeights{
		Admin:       admin,
		PoolWeights: types.PoolWeights(*input),
	}, nil
}

// ------------ types/functions to handle a SetBaseDenoms CLI TX ------------ //
type baseDenomInput struct {
	Denom    string `json:"denom"`
	StepSize uint64 `json:"step_size"`
}

type createBaseDenomsInput []baseDenomInput

type XCreateBaseDenomsInputs baseDenomInput

type XCreateBaseDenomsException struct {
	XCreateBaseDenomsInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createBaseDenomsInput) UnmarshalJSON(data []byte) error {
	var createBaseDenomsE []XCreateBaseDenomsException
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createBaseDenomsE); err != nil {
		return err
	}

	baseDenoms := make([]baseDenomInput, 0)
	for _, denom := range createBaseDenomsE {
		baseDenoms = append(baseDenoms, baseDenomInput(denom.XCreateBaseDenomsInputs))
	}

	*release = createBaseDenomsInput(baseDenoms)

	return nil
}

// BuildSetBaseDenomsMsg builds a MsgSetBaseDenoms from the provided json file
func BuildSetBaseDenomsMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("must provide a json file")
	}

	// Read the json file
	input := &createBaseDenomsInput{}
	path := args[0]
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the json file
	if err := input.UnmarshalJSON(contents); err != nil {
		return nil, err
	}

	// Build the base denoms
	baseDenoms := make([]types.BaseDenom, 0)
	for _, baseDenom := range *input {
		baseDenoms = append(baseDenoms, types.BaseDenom{
			Denom:    baseDenom.Denom,
			StepSize: sdk.NewIntFromUint64(baseDenom.StepSize),
		})
	}

	// Build the msg
	admin := clientCtx.GetFromAddress().String()
	return &types.MsgSetBaseDenoms{
		Admin:      admin,
		BaseDenoms: baseDenoms,
	}, nil
}
