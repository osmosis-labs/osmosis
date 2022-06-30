package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

type CreateSaleInputs createSaleInputs

// UnmarshalJSON should error if there are fields unexpected.
func (inputs *createSaleInputs) UnmarshalJSON(data []byte) error {
	var saleInputs CreateSaleInputs
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force

	if err := dec.Decode(&saleInputs); err != nil {
		return err
	}

	*inputs = createSaleInputs(saleInputs)
	return nil
}

func (inputs *createSaleInputs) ToMsgCreateSale(creator string) (*types.MsgCreateSale, error) {
	_, err := sdk.AccAddressFromBech32(inputs.Recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipient address: %s", inputs.Recipient)
	}
	duration, err := time.ParseDuration(inputs.Duration)
	if err != nil {
		return nil, err
	}
	tOut, err := sdk.ParseCoinNormalized(inputs.TokenOut)
	if err != nil {
		return nil, err
	}
	msg := &types.MsgCreateSale{
		TokenIn:   inputs.TokenIn,
		TokenOut:  &tOut,
		StartTime: inputs.StartTime,
		Duration:  duration,
		Recipient: inputs.Recipient,
		Creator:   creator,
	}
	return msg, msg.ValidateBasic()
}

func parseCreateSaleFlags(fs *pflag.FlagSet) (*createSaleInputs, error) {
	sale := &createSaleInputs{}
	saleFile, _ := fs.GetString(FlagSaleFile)

	if saleFile == "" {
		return nil, fmt.Errorf("must pass in a sale json using the --%s flag", FlagSaleFile)
	}

	contents, err := ioutil.ReadFile(saleFile)
	if err != nil {
		return nil, err
	}

	// make exception if unknown field exists
	err = sale.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}

	return sale, nil
}
