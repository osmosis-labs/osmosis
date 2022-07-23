package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MsgMinFeeExtension is an extension of the messages that defines a minimum fee
// denominated in base denom.
type MsgMinFeeExtension interface {
	// GetRequiredMinBaseFee returns minimum fee for a message denominated in the base
	// fee denom
	GetRequiredMinBaseFee() sdk.Int
}
