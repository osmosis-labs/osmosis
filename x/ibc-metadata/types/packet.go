package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewIbcPacketData contructs a new IbcPacketData instance
func NewFungibleTokenPacketData(sender, receiver, amount, denom string, metadata []byte) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Amount:   amount,
		Denom:    denom,
		Sender:   sender,
		Receiver: receiver,
		Metadata: metadata,
	}
}

// GetBytes is a helper for serialising
func (pd FungibleTokenPacketData) GetBytes() []byte {
	bz, err := json.Marshal(&pd)
	if err != nil {
		panic("FungibleTokenPacketData.GetBytes: " + err.Error())
	}
	return sdk.MustSortJSON(bz)
}

// GetBytes is a helper for serialising
func (pd FungibleTokenPacketData) GetSafeBytes() ([]byte, error) {
	bz, err := json.Marshal(&pd)
	if err != nil {
		return nil, err
	}

	return sdk.SortJSON(bz)
}
