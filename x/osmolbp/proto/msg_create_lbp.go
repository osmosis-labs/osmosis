package proto

import (
	"fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ROUND = time.Second

var _ sdk.Msg = &MsgCreateLBP{}

func (msg *MsgCreateLBP) ValidateBasic() error {
	return errorStringsToError(msg.validate())
}

func (msg *MsgCreateLBP) validate() []string {
	var errmsgs []string
	var d = int64(msg.Duration / ROUND)
	if d > 1 {
		errmsgs = append(errmsgs, "`duration` must be at least 1 second")
	}
	if msg.TokenIn == msg.TokenOut {
		errmsgs = append(errmsgs, "`token_in` must be different than `token_out`")
	}
	if msg.TokenIn == "" {
		errmsgs = append(errmsgs, "`token_in` must be not empty")
	}
	if msg.TokenOut == "" {
		errmsgs = append(errmsgs, "`token_out` must be not empty")
	}
	if msg.TotalSale.GTE(sdk.NewInt(d)) {
		errmsgs = append(errmsgs, "`total_sale` must be positive and must be bigger then duration in seconds")
	}
	return errmsgs
}

func (msg *MsgCreateLBP) Validate(now time.Time) error {
	errmsgs := msg.validate()
	if msg.Start.Before(now) {
		errmsgs = append(errmsgs, fmt.Sprint("`start` must be after ", now))
	}

	return errorStringsToError(errmsgs)
}

func (msg *MsgCreateLBP) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{a}
}

// TODO: remove when updating to SDK v0.44+
// Deprecated methods

func (msg *MsgCreateLBP) GetSignBytes() []byte {
	panic("not implemented")
}

func (msg *MsgCreateLBP) Route() string {
	panic("not implemented")
}

func (msg *MsgCreateLBP) Type() string {
	panic("not implemented")
}
