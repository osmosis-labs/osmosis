package types

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"testing"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgSwap(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
	}

	overflowOfferAmt, _ := sdk.NewIntFromString("100000000000000000000000000000000000000000000000000000000")

	tests := []struct {
		trader      sdk.AccAddress
		offerCoin   sdk.Coin
		askDenom    string
		expectedErr string
	}{
		{addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, osmomath.OneInt()), assets.StakeDenom, ""},
		{sdk.AccAddress{}, sdk.NewCoin(appParams.BaseCoinUnit, osmomath.OneInt()), assets.StakeDenom, "Invalid trader address (empty address string is not allowed): invalid address"},
		{addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, osmomath.ZeroInt()), assets.StakeDenom, "0note: invalid coins"},
		{addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, overflowOfferAmt), assets.StakeDenom, "100000000000000000000000000000000000000000000000000000000note: invalid coins"},
		{addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, osmomath.OneInt()), appParams.BaseCoinUnit, "note: recursive swap"},
	}

	for _, tc := range tests {
		msg := NewMsgSwap(tc.trader, tc.offerCoin, tc.askDenom)
		if tc.expectedErr == "" {
			require.Nil(t, msg.ValidateBasic())
		} else {
			require.EqualError(t, msg.ValidateBasic(), tc.expectedErr)
		}
	}
}

func TestMsgSwapSend(t *testing.T) {
	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
		sdk.AccAddress([]byte("addr2_______________")),
	}

	overflowOfferAmt, _ := sdk.NewIntFromString("100000000000000000000000000000000000000000000000000000000")

	tests := []struct {
		fromAddress sdk.AccAddress
		toAddress   sdk.AccAddress
		offerCoin   sdk.Coin
		askDenom    string
		expectedErr string
	}{
		{addrs[0], addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, osmomath.OneInt()), assets.StakeDenom, ""},
		{addrs[0], sdk.AccAddress{}, sdk.NewCoin(appParams.BaseCoinUnit, osmomath.OneInt()), assets.StakeDenom, "Invalid to address (empty address string is not allowed): invalid address"},
		{sdk.AccAddress{}, addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, osmomath.OneInt()), assets.StakeDenom, "Invalid from address (empty address string is not allowed): invalid address"},
		{addrs[0], addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, osmomath.ZeroInt()), assets.StakeDenom, "0note: invalid coins"},
		{addrs[0], addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, overflowOfferAmt), assets.StakeDenom, "100000000000000000000000000000000000000000000000000000000note: invalid coins"},
		{addrs[0], addrs[0], sdk.NewCoin(appParams.BaseCoinUnit, osmomath.OneInt()), appParams.BaseCoinUnit, "note: recursive swap"},
	}

	for _, tc := range tests {
		msg := NewMsgSwapSend(tc.fromAddress, tc.toAddress, tc.offerCoin, tc.askDenom)
		if tc.expectedErr == "" {
			require.Nil(t, msg.ValidateBasic())
		} else {
			require.EqualError(t, msg.ValidateBasic(), tc.expectedErr)
		}
	}
}
