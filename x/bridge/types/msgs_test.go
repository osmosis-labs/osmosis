package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// Test authz serialize and de-serializes for bridge msg.
func TestAuthzMsg(t *testing.T) {
	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgInboundTransfer",
			msg: &types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
		},
		{
			name: "MsgOutboundTransfer",
			msg: &types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
		},
		{
			name: "MsgUpdateParams",
			msg: &types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers: []string{"s1", "s2", "s3"},
					Assets:  []types.Asset{asset1},
				},
			},
		},
		{
			name: "MsgChangeAssetStatus",
			msg: &types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID1,
				NewStatus: types.AssetStatus_ASSET_STATUS_BLOCKED_BOTH,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, tc.msg)
		})
	}
}

// TestMsgInboundTransfer tests if MsgInboundTransfer messages are properly validated
// and contain proper signers.
func TestMsgInboundTransfer(t *testing.T) {
	var testCases = []struct {
		name            string
		msg             types.MsgInboundTransfer
		expectedSigners []sdk.AccAddress
		expectedErr     error
	}{
		{
			name: "valid",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     nil,
		},
		{
			name: "empty sender",
			msg: types.MsgInboundTransfer{
				Sender:   "",
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{sdk.AccAddress("")},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid sender",
			msg: types.MsgInboundTransfer{
				Sender:   "qwerty",
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{nil},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty destination addr",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: "",
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid destination addr",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: "qwerty",
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid asset id",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId: types.AssetID{
					SourceChain: "",
					Denom:       "btc",
				},
				Amount: math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidAssetID,
		},
		{
			name: "zero amount",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(0),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidCoins,
		},
		{
			name: "negative amount",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(-100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			err := tc.msg.ValidateBasic()
			require.ErrorIsf(t, err, tc.expectedErr, "test: %v", tc.name)
		})
	}
}

// TestMsgOutboundTransfer tests if MsgOutboundTransfer messages are properly validated
// and contain proper signers.
func TestMsgOutboundTransfer(t *testing.T) {
	var testCases = []struct {
		name            string
		msg             types.MsgOutboundTransfer
		expectedSigners []sdk.AccAddress
		expectedErr     error
	}{
		{
			name: "valid",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     nil,
		},
		{
			name: "empty sender",
			msg: types.MsgOutboundTransfer{
				Sender:   "",
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{sdk.AccAddress("")},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid sender",
			msg: types.MsgOutboundTransfer{
				Sender:   "qwerty",
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{nil},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty destination addr",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: "",
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid destination addr",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: "qwerty",
				AssetId:  assetID1,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty asset id",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId: types.AssetID{
					SourceChain: "",
					Denom:       "btc",
				},
				Amount: math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidAssetID,
		},
		{
			name: "zero amount",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(0),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidCoins,
		},
		{
			name: "negative amount",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID1,
				Amount:   math.NewInt(-100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			err := tc.msg.ValidateBasic()
			require.ErrorIsf(t, err, tc.expectedErr, "test: %v", tc.name)
		})
	}
}

// TestMsgUpdateParams tests if MsgUpdateParams messages are properly validated
// and contain proper signers.
func TestMsgUpdateParams(t *testing.T) {
	var testCases = []struct {
		name            string
		msg             types.MsgUpdateParams
		expectedSigners []sdk.AccAddress
		expectedErr     error
	}{
		{
			name: "valid",
			msg: types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers:     []string{addr1},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     nil,
		},
		{
			name: "empty sender",
			msg: types.MsgUpdateParams{
				Sender: "",
				NewParams: types.Params{
					Signers:     []string{addr1},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{sdk.AccAddress("")},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid sender",
			msg: types.MsgUpdateParams{
				Sender: "qwerty",
				NewParams: types.Params{
					Signers:     []string{addr1},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{nil},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty signers are valid",
			msg: types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers:     []string{},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     nil,
		},
		{
			name: "invalid signer",
			msg: types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers:     []string{"qwerty"},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidParams,
		},
		{
			name: "duplicated signers",
			msg: types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers:     []string{addr2, addr2},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidParams,
		},
		{
			name: "empty assets",
			msg: types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers:     []string{addr1, addr2},
					Assets:      []types.Asset{},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidParams,
		},
		{
			name: "invalid asset",
			msg: types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers: []string{addr1, addr2},
					Assets: []types.Asset{{
						Id:       assetID1,
						Status:   types.AssetStatus_ASSET_STATUS_UNSPECIFIED, // invalid status
						Exponent: types.DefaultBitcoinExponent,
					}},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidParams,
		},
		{
			name: "duplicated assets",
			msg: types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers:     []string{addr1, addr2},
					Assets:      []types.Asset{asset1, asset1},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidParams,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			err := tc.msg.ValidateBasic()
			require.ErrorIsf(t, err, tc.expectedErr, "test: %v", tc.name)
		})
	}
}

// TestMsgChangeAssetStatus tests if MsgChangeAssetStatus messages are properly validated
// and contain proper signers.
func TestMsgChangeAssetStatus(t *testing.T) {
	var testCases = []struct {
		name            string
		msg             types.MsgChangeAssetStatus
		expectedSigners []sdk.AccAddress
		expectedErr     error
	}{
		{
			name: "valid",
			msg: types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID1,
				NewStatus: types.AssetStatus_ASSET_STATUS_OK,
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     nil,
		},
		{
			name: "empty sender",
			msg: types.MsgChangeAssetStatus{
				Sender:    "",
				AssetId:   assetID1,
				NewStatus: types.AssetStatus_ASSET_STATUS_OK,
			},
			expectedSigners: []sdk.AccAddress{sdk.AccAddress("")},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid sender",
			msg: types.MsgChangeAssetStatus{
				Sender:    "qwerty",
				AssetId:   assetID1,
				NewStatus: types.AssetStatus_ASSET_STATUS_OK,
			},
			expectedSigners: []sdk.AccAddress{nil},
			expectedErr:     sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid asset id",
			msg: types.MsgChangeAssetStatus{
				Sender: addr1,
				AssetId: types.AssetID{
					SourceChain: "",
					Denom:       "btc",
				},
				NewStatus: 10,
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidAssetID,
		},
		{
			name: "invalid asset status",
			msg: types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID1,
				NewStatus: 10,
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidAssetStatus,
		},
		{
			name: "unspecified asset status",
			msg: types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID1,
				NewStatus: types.AssetStatus_ASSET_STATUS_UNSPECIFIED,
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedErr:     types.ErrInvalidAssetStatus,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			err := tc.msg.ValidateBasic()
			require.ErrorIsf(t, err, tc.expectedErr, "test: %v", tc.name)
		})
	}
}
