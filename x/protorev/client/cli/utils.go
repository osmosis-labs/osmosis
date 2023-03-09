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

type Trade struct {
	Pool     uint64 `json:"pool"`
	TokenIn  string `json:"token_in"`
	TokenOut string `json:"token_out"`
}

type ArbRoutes struct {
	Trades   []Trade `json:"trades"`
	StepSize sdk.Int `json:"step_size"`
}

type hotRoutesInput struct {
	TokenIn   string      `json:"token_in"`
	TokenOut  string      `json:"token_out"`
	ArbRoutes []ArbRoutes `json:"arb_routes"`
}

type createHotRoutesInput struct {
	HotRoutes []hotRoutesInput
}

type XCreateHotRoutesInputs createHotRoutesInput

type XCreateHotRoutesExceptions struct {
	XCreateHotRoutesInputs
	Other *string // Other won't raise an error
}

// UnmarshalJSON should error if there are fields unexpected.
func (release *createHotRoutesInput) UnmarshalJSON(data []byte) error {
	var createHotRoutesE XCreateHotRoutesExceptions
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&createHotRoutesE); err != nil {
		return err
	}

	*release = createHotRoutesInput(createHotRoutesE.XCreateHotRoutesInputs)
	return nil
}

// CreateHotRoutesMsg builds a set hot routes message from the provided input object
func (release *createHotRoutesInput) extractTokenPairArbRoutes() []types.TokenPairArbRoutes {
	tokenPairArbRoutes := make([]types.TokenPairArbRoutes, 0)

	// Iterate through each hot route and construct the token pair arb routes
	for _, hotRoute := range release.HotRoutes {
		current := types.TokenPairArbRoutes{}
		current.TokenIn = hotRoute.TokenIn
		current.TokenOut = hotRoute.TokenOut

		for _, arbRoute := range hotRoute.ArbRoutes {
			currentArbRoute := types.Route{}
			currentArbRoute.StepSize = arbRoute.StepSize

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

// BuildSetHotRoutesMsg builds a set hot routes message from the provided json file
func BuildSetHotRoutesMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("must provide a json file")
	}

	// Read the json file
	input := &createHotRoutesInput{}
	path := args[0]
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the json file
	err = input.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}

	// Extract and build the msg
	tokenPairArbRoutes := input.extractTokenPairArbRoutes()
	return types.NewMsgSetHotRoutes(clientCtx.GetFromAddress().String(), tokenPairArbRoutes), nil
}
