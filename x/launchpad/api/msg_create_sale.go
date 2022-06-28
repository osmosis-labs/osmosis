package api

import (
	"fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ROUND = time.Second

var _ sdk.Msg = &MsgCreateSale{}

func (msg *MsgCreateSale) ValidateBasic() error {
	return errorStringsToError(msg.validate())
}

func (msg *MsgCreateSale) validate() []string {
	var errmsgs []string
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		errmsgs = append(errmsgs, fmt.Sprintf("Invalid creator address (%s)", err))
	}
	if _, err := sdk.AccAddressFromBech32(msg.Treasury); err != nil {
		errmsgs = append(errmsgs, fmt.Sprintf("Invalid treasury address (%s)", err))
	}

	var d = int64(msg.Duration / ROUND)
	if d < 10 {
		errmsgs = append(errmsgs, "`duration` must be at least 10 rounds")
	}
	const maxDuration = ROUND * 24 * 3600 * 356 * 10
	if d > int64(maxDuration) {
		errmsgs = append(errmsgs, "`duration` must not be bigger than "+maxDuration.String())
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
	if msg.InitialDeposit.Denom != msg.TokenOut {
		errmsgs = append(errmsgs, "`initial_deposit` denom must be the same as `token_out`")
	}
	if msg.InitialDeposit.Amount.LTE(sdk.NewInt(d)) {
		errmsgs = append(errmsgs, "`initial_deposit` amount must be positive and must be bigger than duration in seconds")
	}

	return errmsgs
}

func (msg *MsgCreateSale) Validate(now time.Time) error {
	errmsgs := msg.validate()
	if msg.StartTime.Before(now) {
		errmsgs = append(errmsgs, fmt.Sprint("`start` must be after ", now))
	}

	return errorStringsToError(errmsgs)
}

func (msg *MsgCreateSale) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{a}
}
