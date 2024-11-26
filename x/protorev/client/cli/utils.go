package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
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
			currentArbRoute.StepSize = osmomath.NewIntFromUint64(arbRoute.StepSize)

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
		return nil, errors.New("must provide a json file")
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

// ------------ types/functions to handle a SetInfoByPoolType CLI TX ------------ //
type InfoByPoolTypeInput struct {
	Stable       StablePoolInfoInput       `json:"stable"`
	Balancer     BalancerPoolInfoInput     `json:"balancer"`
	Concentrated ConcentratedPoolInfoInput `json:"concentrated"`
	Cosmwasm     CosmwasmPoolInfoInput     `json:"cosmwasm"`
}

type StablePoolInfoInput struct {
	Weight uint64 `json:"weight"`
}

type BalancerPoolInfoInput struct {
	Weight uint64 `json:"weight"`
}

type ConcentratedPoolInfoInput struct {
	Weight          uint64 `json:"weight"`
	MaxTicksCrossed uint64 `json:"max_ticks_crossed"`
}

type CosmwasmPoolInfoInput struct {
	WeightMap map[string]uint64 `json:"weight_map"`
}
type createInfoByPoolTypeInput types.InfoByPoolType

type XCreateInfoByPoolTypeInputs createInfoByPoolTypeInput

type XCreateInfoByPoolTypeExceptions struct {
	XCreateInfoByPoolTypeInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createInfoByPoolTypeInput) UnmarshalJSON(data []byte) error {
	var createInfoByPoolTypeE XCreateInfoByPoolTypeExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createInfoByPoolTypeE); err != nil {
		return err
	}

	*release = createInfoByPoolTypeInput(createInfoByPoolTypeE.XCreateInfoByPoolTypeInputs)
	return nil
}

// createInfoByPoolTypeInput converts the input to the types.InfoByPoolType type
func (release *createInfoByPoolTypeInput) convertToInfoByPoolType() types.InfoByPoolType {
	if release == nil {
		return types.InfoByPoolType{}
	}

	infoByPoolType := types.InfoByPoolType{}
	infoByPoolType.Stable = release.Stable
	infoByPoolType.Balancer = release.Balancer
	infoByPoolType.Concentrated = release.Concentrated
	infoByPoolType.Cosmwasm = release.Cosmwasm

	return infoByPoolType
}

// BuildSetInfoByPoolTypeMsg builds a MsgSetInfoByPoolType from the provided json file
func BuildSetInfoByPoolTypeMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	if len(args) == 0 {
		return nil, errors.New("must provide a json file")
	}

	// Read the json file
	input := &createInfoByPoolTypeInput{}
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
	return &types.MsgSetInfoByPoolType{
		Admin:          admin,
		InfoByPoolType: input.convertToInfoByPoolType(),
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
		return nil, errors.New("must provide a json file")
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
			StepSize: osmomath.NewIntFromUint64(baseDenom.StepSize),
		})
	}

	// Build the msg
	admin := clientCtx.GetFromAddress().String()
	return &types.MsgSetBaseDenoms{
		Admin:      admin,
		BaseDenoms: baseDenoms,
	}, nil
}
