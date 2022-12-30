package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgSetHotRoutes{}
	_ sdk.Msg = &MsgSetDeveloperAccount{}
)

const (
	TypeMsgSetHotRoutes        = "set_hot_routes"
	TypeMsgSetDeveloperAccount = "set_developer_account"
)

// ---------------------- Interface for MsgSetHotRoutes ---------------------- //
// NewMsgSetHotRoutes creates a new MsgSetHotRoutes instance
func NewMsgSetHotRoutes(admin string, tokenPairArbRoutes []*TokenPairArbRoutes) *MsgSetHotRoutes {
	return &MsgSetHotRoutes{
		Admin:     admin,
		HotRoutes: tokenPairArbRoutes,
	}
}

// Route returns the name of the module
func (msg MsgSetHotRoutes) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetHotRoutes) Type() string {
	return TypeMsgSetHotRoutes
}

// ValidateBasic validates the MsgSetHotRoutes
func (msg MsgSetHotRoutes) ValidateBasic() error {
	// Account must be a valid bech32 address
	_, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	// Each token pair arb route must be valid
	seenTokenPairs := make(map[TokenPair]bool)
	for _, tokenPairArbRoutes := range msg.HotRoutes {
		// Validate the arb routes
		if err := tokenPairArbRoutes.Validate(); err != nil {
			return err
		}

		tokenPair := TokenPair{
			TokenA: tokenPairArbRoutes.TokenIn,
			TokenB: tokenPairArbRoutes.TokenOut,
		}
		// Validate that the token pair is unique
		if _, ok := seenTokenPairs[tokenPair]; ok {
			return fmt.Errorf("duplicate token pair: %s", tokenPair)
		}

		seenTokenPairs[tokenPair] = true
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetHotRoutes) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetHotRoutes) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}

// ---------------------- Interface for MsgSetDeveloperAccount ---------------------- //
// NewMsgSetDeveloperAccount creates a new MsgSetDeveloperAccount instance
func NewMsgSetDeveloperAccount(admin string, developerAccount string) *MsgSetDeveloperAccount {
	return &MsgSetDeveloperAccount{
		Admin:            admin,
		DeveloperAccount: developerAccount,
	}
}

// Route returns the name of the module
func (msg MsgSetDeveloperAccount) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetDeveloperAccount) Type() string {
	return TypeMsgSetDeveloperAccount
}

// ValidateBasic validates the MsgSetDeveloperAccount
func (msg MsgSetDeveloperAccount) ValidateBasic() error {
	// Account must be a valid bech32 address
	_, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	// Account must be a valid bech32 address
	_, err = sdk.AccAddressFromBech32(msg.DeveloperAccount)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid developer account address (must be bech32)")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetDeveloperAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetDeveloperAccount) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}
