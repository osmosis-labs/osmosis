package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgSetHotRoutes{}
	_ sdk.Msg = &MsgSetDeveloperAccount{}
	_ sdk.Msg = &MsgSetMaxPoolPointsPerTx{}
	_ sdk.Msg = &MsgSetMaxPoolPointsPerBlock{}
	_ sdk.Msg = &MsgSetPoolWeights{}
	_ sdk.Msg = &MsgSetBaseDenoms{}
)

const (
	TypeMsgSetHotRoutes             = "set_hot_routes"
	TypeMsgSetDeveloperAccount      = "set_developer_account"
	TypeMsgSetMaxPoolPointsPerTx    = "set_max_pool_points_per_tx"
	TypeMsgSetMaxPoolPointsPerBlock = "set_max_pool_points_per_block"
	TypeMsgSetPoolWeights           = "set_pool_weights"
	TypeMsgSetBaseDenoms            = "set_base_denoms"
)

// ---------------------- Interface for MsgSetHotRoutes ---------------------- //
// NewMsgSetHotRoutes creates a new MsgSetHotRoutes instance
func NewMsgSetHotRoutes(admin string, tokenPairArbRoutes []TokenPairArbRoutes) *MsgSetHotRoutes {
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
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	// Validate the hot routes
	if err := ValidateTokenPairArbRoutes(msg.HotRoutes); err != nil {
		return err
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

// ---------------------- Interface for MsgSetMaxPoolPointsPerTx ---------------------- //
// NewMsgSetMaxPoolPointsPerTx creates a new MsgSetMaxPoolPointsPerTx instance
func NewMsgSetMaxPoolPointsPerTx(admin string, maxPoolPointsPerTx uint64) *MsgSetMaxPoolPointsPerTx {
	return &MsgSetMaxPoolPointsPerTx{
		Admin:              admin,
		MaxPoolPointsPerTx: maxPoolPointsPerTx,
	}
}

// Route returns the name of the module
func (msg MsgSetMaxPoolPointsPerTx) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetMaxPoolPointsPerTx) Type() string {
	return TypeMsgSetMaxPoolPointsPerTx
}

// ValidateBasic validates the MsgSetMaxPoolPointsPerTx
func (msg MsgSetMaxPoolPointsPerTx) ValidateBasic() error {
	// Account must be a valid bech32 address
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	// Max pool points per tx must be in the valid range
	if err := ValidateMaxPoolPointsPerTx(msg.MaxPoolPointsPerTx); err != nil {
		return err
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetMaxPoolPointsPerTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetMaxPoolPointsPerTx) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}

// ---------------------- Interface for MsgSetMaxPoolPointsPerBlock ---------------------- //
// NewMsgSetMaxPoolPointsPerBlock creates a new MsgSetMaxPoolPointsPerBlock instance
func NewMsgSetMaxPoolPointsPerBlock(admin string, maxPoolPointsPerBlock uint64) *MsgSetMaxPoolPointsPerBlock {
	return &MsgSetMaxPoolPointsPerBlock{
		Admin:                 admin,
		MaxPoolPointsPerBlock: maxPoolPointsPerBlock,
	}
}

// Route returns the name of the module
func (msg MsgSetMaxPoolPointsPerBlock) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetMaxPoolPointsPerBlock) Type() string {
	return TypeMsgSetMaxPoolPointsPerBlock
}

// ValidateBasic validates the MsgSetMaxPoolPointsPerBlock
func (msg MsgSetMaxPoolPointsPerBlock) ValidateBasic() error {
	// Account must be a valid bech32 address
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	// Max pool points per block must be in the valid range
	if err := ValidateMaxPoolPointsPerBlock(msg.MaxPoolPointsPerBlock); err != nil {
		return err
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetMaxPoolPointsPerBlock) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetMaxPoolPointsPerBlock) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}

// ---------------------- Interface for MsgSetPoolWeights ---------------------- //
// NewMsgSetPoolWeights creates a new MsgSetPoolWeights instance
func NewMsgSetPoolWeights(admin string, poolWeights PoolWeights) *MsgSetPoolWeights {
	return &MsgSetPoolWeights{
		Admin:       admin,
		PoolWeights: poolWeights,
	}
}

// Route returns the name of the module
func (msg MsgSetPoolWeights) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetPoolWeights) Type() string {
	return TypeMsgSetPoolWeights
}

// ValidateBasic validates the MsgSetPoolWeights
func (msg MsgSetPoolWeights) ValidateBasic() error {
	// Account must be a valid bech32 address
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	if err := msg.PoolWeights.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetPoolWeights) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetPoolWeights) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}

// ---------------------- Interface for MsgSetBaseDenoms ---------------------- //
// NewMsgSetBaseDenoms creates a new MsgSetBaseDenoms instance
func NewMsgSetBaseDenoms(admin string, baseDenoms []BaseDenom) *MsgSetBaseDenoms {
	return &MsgSetBaseDenoms{
		Admin:      admin,
		BaseDenoms: baseDenoms,
	}
}

// Route returns the name of the module
func (msg MsgSetBaseDenoms) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSetBaseDenoms) Type() string {
	return TypeMsgSetBaseDenoms
}

// ValidateBasic validates the MsgSetBaseDenoms
func (msg MsgSetBaseDenoms) ValidateBasic() error {
	// Account must be a valid bech32 address
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return sdkerrors.Wrap(err, "invalid admin address (must be bech32)")
	}

	// Check that there is at least one base denom and that first denom is osmo
	if err := ValidateBaseDenoms(msg.BaseDenoms); err != nil {
		return err
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetBaseDenoms) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetBaseDenoms) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Admin)
	return []sdk.AccAddress{addr}
}
