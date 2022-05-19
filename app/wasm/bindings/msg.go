package wasmbindings

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"time"
)

type OsmosisMsg struct {
	/// Contracts can mint native tokens that have an auto-generated denom
	/// namespaced under the contract's address. A contract may create any number
	/// of independent sub-denoms.
	MintTokens *MintTokens `json:"mint_tokens,omitempty"`
	/// Swap over one or more pools
	Swap *SwapMsg `json:"swap,omitempty"`
	/// Lock gamm tokens
	LockTokens *LockTokensMsg `json:"lock_tokens,omitempty"`
}

type MintTokens struct {
	/// Must be 2-32 alphanumeric characters
	SubDenom  string  `json:"sub_denom"`
	Amount    sdk.Int `json:"amount"`
	Recipient string  `json:"recipient"`
}

type SwapMsg struct {
	First  Swap                `json:"first"`
	Route  []Step              `json:"route"`
	Amount SwapAmountWithLimit `json:"amount"`
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "invalid duration")
	}
}

type LockTokensMsg struct {
	Denom    string   `json:"denom"`
	Amount   sdk.Int  `json:"amount"`
	Duration Duration `json:"duration"`
}
