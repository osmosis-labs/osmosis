package balancer

import (
	"fmt"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

const (
	TypeMsgCreateBalancerPool = "create_balancer_pool"
)

func NewMsgCreateBalancerPool(sender string, balancerPoolParams BalancerPoolParamsI, poolAssets []types.PoolAsset) (*MsgCreateBalancerPool, error) {
	m := &MsgCreateBalancerPool{
		Sender:     sender,
		PoolAssets: poolAssets,
	}
	err := m.SetPoolParams(balancerPoolParams)
	if err != nil {
		return nil, err
	}
	return m, nil
}

var _ sdk.Msg = &MsgCreateBalancerPool{}

func (msg MsgCreateBalancerPool) Route() string { return types.RouterKey }
func (msg MsgCreateBalancerPool) Type() string  { return TypeMsgCreateBalancerPool }
func (msg MsgCreateBalancerPool) ValidateBasic() error {

	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = types.ValidateUserSpecifiedPoolAssets(msg.PoolAssets)
	if err != nil {
		return err
	}

	params := msg.GetBalancerPoolParams()
	err = params.Validate(msg.PoolAssets)

	// validation for future owner
	if err = ValidateFutureGovernor(msg.FuturePoolGovernor); err != nil {
		return err
	}

	return nil
}
func (msg MsgCreateBalancerPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(&msg))
}
func (msg MsgCreateBalancerPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg MsgCreateBalancerPool) GetBalancerPoolParams() BalancerPoolParams {
	balancerPoolParams, ok := msg.PoolParams.GetCachedValue().(BalancerPoolParams)
	if !ok {
		return BalancerPoolParams{}
	}
	return balancerPoolParams
}

func (msg MsgCreateBalancerPool) SetPoolParams(balancerPoolParams BalancerPoolParamsI) error {
	m, ok := balancerPoolParams.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshl &T", m)
	}
	any, err := codectypes.NewAnyWithValue(m)
	if err != nil {
		return err
	}
	msg.PoolParams = *any
	return nil
}

func ValidateFutureGovernor(governor string) error {
	// allow empty governor
	if governor == "" {
		return nil
	}

	// validation for future owner
	// "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
	_, err := sdk.AccAddressFromBech32(governor)
	if err == nil {
		return nil
	}

	lockTimeStr := ""
	splits := strings.Split(governor, ",")
	if len(splits) > 2 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid future governor: %s", governor))
	}

	// token,100h
	if len(splits) == 2 {
		lpTokenStr := splits[0]
		if sdk.ValidateDenom(lpTokenStr) != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid future governor: %s", governor))
		}
		lockTimeStr = splits[1]
	}

	// 100h
	if len(splits) == 1 {
		lockTimeStr = splits[0]
	}

	// Note that a duration of 0 is allowed
	_, err = time.ParseDuration(lockTimeStr)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid future governor: %s", governor))
	}
	return nil
}
