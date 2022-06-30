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
	if msg.TokenIn == msg.TokenOut {
		errmsgs = append(errmsgs, "`token_in` must be different than `token_out`")
	}
	if msg.TokenIn == "" {
		errmsgs = append(errmsgs, "`token_in` must be not empty")
	}
	if msg.TokenOut == "" {
		errmsgs = append(errmsgs, "`token_out` must be not empty")
	}
	if !msg.InitialDeposit.GT(sdk.NewInt(d)) {
		errmsgs = append(errmsgs, "`initial_deposit` amount must be positive and must be bigger than duration in seconds")
	}

	return creator, errmsgs
}

func (msg *MsgCreateSale) Validate(now time.Time) (sdk.AccAddress, error) {
	creator, errmsgs := msg.validate()
	if !msg.StartTime.After(now) {
		errmsgs = append(errmsgs, fmt.Sprint("`start` must be after ", now))
	}

	return creator, errorStringsToError(errmsgs)
}

func (msg *MsgCreateSale) GetSigners() []sdk.AccAddress {
	a, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{a}
}
