package types

import (
	"fmt"

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

	if msg.HotRoutes == nil {
		return fmt.Errorf("hot routes cannot be nil")
	}

	// Each token pair arb route must be valid
	seenTokenPairs := make(map[TokenPair]bool)
	for _, tokenPairArbRoutes := range msg.HotRoutes {
		if tokenPairArbRoutes == nil {
			return fmt.Errorf("nil token pair arb routes")
		}

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
	if msg.MaxPoolPointsPerTx <= 0 || msg.MaxPoolPointsPerTx > MaxPoolPointsPerTx {
		return fmt.Errorf("max pool points per tx must be in the range (0, %d]", MaxPoolPointsPerTx)
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
	if msg.MaxPoolPointsPerBlock <= 0 || msg.MaxPoolPointsPerBlock > MaxPoolPointsPerBlock {
		return fmt.Errorf("max pool points per block must be in the range (0, %d]", MaxPoolPointsPerBlock)
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
func NewMsgSetPoolWeights(admin string, poolWeights *PoolWeights) *MsgSetPoolWeights {
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

	if msg.PoolWeights == nil {
		return fmt.Errorf("pool weights cannot be nil")
	}

	if msg.PoolWeights.BalancerWeight == 0 || msg.PoolWeights.StableWeight == 0 || msg.PoolWeights.ConcentratedWeight == 0 {
		return fmt.Errorf("pool weights cannot be 0")
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
func NewMsgSetBaseDenoms(admin string, baseDenoms []*BaseDenom) *MsgSetBaseDenoms {
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
	if len(msg.BaseDenoms) == 0 || msg.BaseDenoms[0] == nil || msg.BaseDenoms[0].Denom != OsmosisDenomination {
		return fmt.Errorf("must have at least one base denom and first base denom must be osmo")
	}

	// Each base denom must be valid
	seenBaseDenoms := make(map[string]bool)
	for _, baseDenom := range msg.BaseDenoms {
		if baseDenom == nil {
			return fmt.Errorf("base denom cannot be nil")
		}

		// Validate the base denom step size
		if baseDenom.StepSize.LT(sdk.OneInt()) {
			return fmt.Errorf("base denom step size must be at least 1: got %s", baseDenom)
		}

		// Validate that the base denom is unique
		if _, ok := seenBaseDenoms[baseDenom.Denom]; ok {
			return fmt.Errorf("duplicate base denom: %s", baseDenom)
		}

		seenBaseDenoms[baseDenom.Denom] = true
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
