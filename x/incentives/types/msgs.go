package types

import (
	"errors"
	"fmt"
	"time"

	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgCreateGauge = "create_gauge"
	TypeMsgAddToGauge  = "add_to_gauge"
)

var _ sdk.Msg = &MsgCreateGauge{}

// NewMsgCreateGauge creates a message to create a gauge with the provided parameters.
func NewMsgCreateGauge(isPerpetual bool, owner sdk.AccAddress, distributeTo lockuptypes.QueryCondition, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64, poolId uint64) *MsgCreateGauge {
	return &MsgCreateGauge{
		IsPerpetual:       isPerpetual,
		Owner:             owner.String(),
		DistributeTo:      distributeTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		PoolId:            poolId,
	}
}

// Route takes a create gauge message, then returns the RouterKey used for slashing.
func (m MsgCreateGauge) Route() string { return RouterKey }

// Type takes a create gauge message, then returns a create gauge message type.
func (m MsgCreateGauge) Type() string { return TypeMsgCreateGauge }

// ValidateBasic checks that the create gauge message is valid.
func (m MsgCreateGauge) ValidateBasic() error {
	lockType := m.DistributeTo.LockQueryType
	isNoLockGauge := lockType == lockuptypes.NoLock

	if m.Owner == "" {
		return errors.New("owner should be set")
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

	if lockType == lockuptypes.ByTime {
		return errors.New("start time distr conditions is an obsolete codepath slated for deletion")
	}

	if isNoLockGauge {
		if m.PoolId == 0 {
			return errors.New("pool id should be set for no lock distr condition")
		}

		if m.DistributeTo.Denom != "" {
			return errors.New(`no lock gauge denom should be unset. It will be automatically set to the NoLockExternalGaugeDenom(<pool id>)
			 format internally, allowing for querying the gauges by denom with this prefix`)
		}

		if m.DistributeTo.Duration != 0 {
			return fmt.Errorf("'no lock' gauge must have duration set to 0, was (%d)", m.DistributeTo.Duration)
		}
	} else {
		if m.PoolId != 0 {
			return errors.New("pool id should not be set for duration distr condition")
		}

		// For no lock type, the denom must be empty and we check that above.
		if err := sdk.ValidateDenom(m.DistributeTo.Denom); err != nil {
			return fmt.Errorf("denom should be valid for the condition, %s", err)
		}
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
