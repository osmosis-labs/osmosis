package types

import (
	"errors"
	"time"

	lockuptypes "github.com/osmosis-labs/osmosis/v9/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants.
const (
	TypeMsgCreateGauge = "create_gauge"
	TypeMsgAddToGauge  = "add_to_gauge"
)

var _ sdk.Msg = &MsgCreateGauge{}

// NewMsgCreateGauge creates a message to create a gauge.
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

func (m MsgCreateGauge) Route() string { return RouterKey }
func (m MsgCreateGauge) Type() string  { return TypeMsgCreateGauge }
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

func (m MsgCreateGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgCreateGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}

var _ sdk.Msg = &MsgAddToGauge{}

// NewMsgCreateGauge creates a message to create a gauge.
func NewMsgAddToGauge(owner sdk.AccAddress, gaugeId uint64, rewards sdk.Coins) *MsgAddToGauge {
	return &MsgAddToGauge{
		Owner:   owner.String(),
		GaugeId: gaugeId,
		Rewards: rewards,
	}
}

func (m MsgAddToGauge) Route() string { return RouterKey }
func (m MsgAddToGauge) Type() string  { return TypeMsgAddToGauge }
func (m MsgAddToGauge) ValidateBasic() error {
	if m.Owner == "" {
		return errors.New("owner should be set")
	}
	if m.Rewards.Empty() {
		return errors.New("additional rewards should not be empty")
	}

	return nil
}

func (m MsgAddToGauge) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgAddToGauge) GetSigners() []sdk.AccAddress {
	owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{owner}
}
