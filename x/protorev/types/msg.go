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
	TypeMsgSetHotRoutes         = "set_hot_routes"
	TypeMsgSetDeveloperAccount  = "set_developer_account"
	TypeMsgSetMaxRoutesPerTx    = "set_max_routes_per_tx"
	TypeMsgSetMaxRoutesPerBlock = "set_max_routes_per_block"
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

// ---------------------- Interface for MsgSetMaxRoutesPerTx ---------------------- //
// NewMsgSetMaxRoutesPerTx creates a new MsgSetMaxRoutesPerTx instance
func NewMsgSetMaxRoutesPerTx(admin string, maxRoutesPerTx uint64) *MsgSetMaxRoutesPerTx {
	return &MsgSetMaxRoutesPerTx{
		Admin:          admin,
		MaxRoutesPerTx: maxRoutesPerTx,
	}
}

// Route returns the name of the module
func (msg MsgSetMaxRoutesPerTx) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetMaxRoutesPerTx) Type() string {
	return TypeMsgSetMaxRoutesPerTx
}

// ValidateBasic validates the MsgSetMaxRoutesPerTx
func (msg MsgSetMaxRoutesPerTx) ValidateBasic() error {
	// Account must be a valid bech32 address
	_, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	if msg.MaxRoutesPerTx > MaxIterableRoutesPerTx || msg.MaxRoutesPerTx == 0 {
		return fmt.Errorf("max routes per tx must be less than or equal to %d and greater than 0", MaxIterableRoutesPerTx)
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetMaxRoutesPerTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetMaxRoutesPerTx) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}

// ---------------------- Interface for MsgSetMaxRoutesPerBlock ---------------------- //
// NewMsgSetMaxRoutesPerBlock creates a new MsgSetMaxRoutesPerBlock instance
func NewMsgSetMaxRoutesPerBlock(admin string, maxRoutesPerBlock uint64) *MsgSetMaxRoutesPerBlock {
	return &MsgSetMaxRoutesPerBlock{
		Admin:             admin,
		MaxRoutesPerBlock: maxRoutesPerBlock,
	}
}

// Route returns the name of the module
func (msg MsgSetMaxRoutesPerBlock) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetMaxRoutesPerBlock) Type() string {
	return TypeMsgSetMaxRoutesPerBlock
}

// ValidateBasic validates the MsgSetMaxRoutesPerBlock
func (msg MsgSetMaxRoutesPerBlock) ValidateBasic() error {
	// Account must be a valid bech32 address
	_, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	if msg.MaxRoutesPerBlock > MaxIterableRoutesPerBlock || msg.MaxRoutesPerBlock == 0 {
		return fmt.Errorf("max routes per block must be less than or equal to %d and greater than 0", MaxIterableRoutesPerBlock)
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetMaxRoutesPerBlock) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetMaxRoutesPerBlock) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}
