package types

import (
	"errors"
	"time"

	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v10/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgCreateGauge = "create_gauge"
	TypeMsgAddToGauge  = "add_to_gauge"

	createGageMinbaseFee = 2
	addToGaugeMinBaseFee = 1
)

var _ sdk.Msg = &MsgCreateGauge{}
var _ txfeestypes.MsgMinFeeExtension = &MsgCreateGauge{}
var _ txfeestypes.MsgMinFeeExtension = &MsgAddToGauge{}

// NewMsgCreateGauge creates a message to create a gauge with the provided parameters.
func NewMsgCreateGauge(isPerpetual bool, owner sdk.AccAddress, distributeTo lockuptypes.QueryCondition, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64) *MsgCreateGauge {
	return &MsgCreateGauge{
		IsPerpetual:       isPerpetual,
		Owner:             owner.String(),
		DistributeTo:      distributeTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
	}
}

func (m MsgCreateGauge) GetRequiredMinBaseFee() sdk.Int {
	return sdk.NewInt(createGageMinbaseFee)
}

func (m MsgAddToGauge) GetRequiredMinBaseFee() sdk.Int {
	return sdk.NewInt(addToGaugeMinBaseFee)
}

// Route takes a create gauge message, then returns the RouterKey used for slashing.
func (m MsgCreateGauge) Route() string { return RouterKey }

// Type takes a create gauge message, then returns a create gauge message type.
func (m MsgCreateGauge) Type() string { return TypeMsgCreateGauge }

// ValidateBasic checks that the create gauge message is valid.
func (m MsgCreateGauge) ValidateBasic() error {
	if m.Owner == "" {
		return errors.New("owner should be set")
	}
	if sdk.ValidateDenom(m.DistributeTo.Denom) != nil {
		return errors.New("denom should be valid for the condition")
	}
	if lockuptypes.LockQueryType_name[int32(m.DistributeTo.LockQueryType)] == "" {
		return errors.New("lock query type is invalid")
	}
	if m.StartTime.Equal(time.Time{}) {
		return errors.New("distribution start time should be set")
	}
	if m.NumEpochsPaidOver == 0 {
		return errors.New("distribution period should be at least 1 epoch")
	}
	if m.IsPerpetual && m.NumEpochsPaidOver != 1 {
		return errors.New("distribution period should be 1 epoch for perpetual gauge")
	}

	if lockuptypes.LockQueryType_name[int32(m.DistributeTo.LockQueryType)] != "ByDuration" {
		return errors.New("only duration query condition is allowed. Start time distr conditions is an obsolete codepath slated for deletion")
	}

	return nil
}

// GetSignBytes takes a create gauge message and turns it into a byte array.
func (m MsgCreateGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners takes a create gauge message and returns the owner in a byte array.
func (m MsgCreateGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgAddToGauge{}

// NewMsgAddToGauge creates a message to add rewards to a specific gauge.
func NewMsgAddToGauge(owner sdk.AccAddress, gaugeId uint64, rewards sdk.Coins) *MsgAddToGauge {
	return &MsgAddToGauge{
		Owner:   owner.String(),
		GaugeId: gaugeId,
		Rewards: rewards,
	}
}

// Route takes an add to gauge message, then returns the RouterKey used for slashing.
func (m MsgAddToGauge) Route() string { return RouterKey }

// Type takes an add to gauge message, then returns an add to gauge message type.
func (m MsgAddToGauge) Type() string { return TypeMsgAddToGauge }

// ValidateBasic checks that the add to gauge message is valid.
func (m MsgAddToGauge) ValidateBasic() error {
	if m.Owner == "" {
		return errors.New("owner should be set")
	}
	if m.Rewards.Empty() {
		return errors.New("additional rewards should not be empty")
	}

	return nil
}

// GetSignBytes takes an add to gauge message and turns it into a byte array.
func (m MsgAddToGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners takes an add to gauge message and returns the owner in a byte array.
func (m MsgAddToGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}
