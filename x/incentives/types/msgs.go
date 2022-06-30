package types

import (
	"errors"
	"time"

	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgCreateGauge = "create_gauge"
	TypeMsgAddToGauge  = "add_to_gauge"
)

var _ sdk.Msg = &MsgCreateGauge{}

// Creates a message to create a gauge with the provided parameters.
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

// Given a create gauge message, returns the RouterKey used for slashing.
func (m MsgCreateGauge) Route() string { return RouterKey }

// Given a create gauge message, returns a create gauge message type.
func (m MsgCreateGauge) Type() string { return TypeMsgCreateGauge }

// Checks that the create gauge message is valid.
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

// Takes a create gauge message and turns it into a byte array.
func (m MsgCreateGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// Take a create gauge message and returns the owner in a byte array.
func (m MsgCreateGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgAddToGauge{}

// Creates a message to add rewards to a specific gauge.
func NewMsgAddToGauge(owner sdk.AccAddress, gaugeId uint64, rewards sdk.Coins) *MsgAddToGauge {
	return &MsgAddToGauge{
		Owner:   owner.String(),
		GaugeId: gaugeId,
		Rewards: rewards,
	}
}

// Given an add to gauge message, returns the RouterKey used for slashing.
func (m MsgAddToGauge) Route() string { return RouterKey }

// Given an add to gauge message, returns an add to gauge message type.
func (m MsgAddToGauge) Type() string { return TypeMsgAddToGauge }

// Checks that the add to gauge message is valid.
func (m MsgAddToGauge) ValidateBasic() error {
	if m.Owner == "" {
		return errors.New("owner should be set")
	}
	if m.Rewards.Empty() {
		return errors.New("additional rewards should not be empty")
	}

	return nil
}

// Takes an add to gauge message and turns it into a byte array.
func (m MsgAddToGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// Take an add to gauge message and returns the owner in a byte array.
func (m MsgAddToGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}
