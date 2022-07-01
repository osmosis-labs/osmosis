package types

import (
	"fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ROUND = time.Second

var _ sdk.Msg = &MsgCreateSale{}

func (msg *MsgCreateSale) ValidateBasic() error {
	_, errs := msg.validate()
	return errorStringsToError(errs)
}

func (msg *MsgCreateSale) validate() (sdk.AccAddress, []string) {
	var errmsgs []string
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		errmsgs = append(errmsgs, fmt.Sprintf("Invalid creator address (%s)", err))
	}
	if msg.Recipient != "" {
		if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
			errmsgs = append(errmsgs, fmt.Sprintf("Invalid treasury address (%s)", err))
		}
	}
	var d = int64(msg.Duration / ROUND)
	if d < 10 {
		errmsgs = append(errmsgs, "`duration` must be at least 10 rounds")
	}
	const maxDuration = ROUND * 24 * 3600 * 356 * 10
	if d > int64(maxDuration) {
		errmsgs = append(errmsgs, "`duration` must not be bigger than "+maxDuration.String())
	}
	if msg.TokenOut == nil || msg.TokenOut.IsZero() {
		errmsgs = append(errmsgs, "`token_out` amount must be positive")
	}
	if msg.TokenIn == msg.TokenOut.Denom {
		errmsgs = append(errmsgs, "`token_in` must be different than `token_out`")
	}
	if err = sdk.ValidateDenom(msg.TokenIn); err != nil {
		errmsgs = append(errmsgs, "`token_in` must be a proper denom, "+err.Error())
	}
	if err = msg.TokenOut.Validate(); err != nil {
		errmsgs = append(errmsgs, "`token_out` must be well defined, "+err.Error())
	}
	if msg.TokenOut.IsZero() {
		errmsgs = append(errmsgs, "`token_out` amount must be positive")
	}

	return creator, errmsgs
}

func (msg *MsgCreateSale) Validate(now time.Time, minDuration, minDurationUntilStart time.Duration) (sdk.AccAddress, error) {
	creator, errmsgs := msg.validate()
	minStart := now.Add(minDurationUntilStart)
	if msg.StartTime.Before(minStart) {
		errmsgs = append(errmsgs, fmt.Sprint("`start` must be after ", minStart))
	}
	if msg.Duration < minDuration {
		errmsgs = append(errmsgs, fmt.Sprint("Sale duration must be at least ", minDuration.String()))
	}

	return creator, errorStringsToError(errmsgs)
}

func (msg *MsgCreateSale) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{a}
}
