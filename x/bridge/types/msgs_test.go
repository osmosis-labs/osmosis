package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// Test authz serialize and de-serializes for bridge msg.
func TestAuthzMsg(t *testing.T) {
	var (
		pk1     = ed25519.GenPrivKey().PubKey()
		pk2     = ed25519.GenPrivKey().PubKey()
		addr1   = sdk.AccAddress(pk1.Address()).String()
		addr2   = sdk.AccAddress(pk2.Address()).String()
		assetID = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		}
	)
	asset := types.DefaultAssets()[0]
	asset.Id = assetID

	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgInboundTransfer",
			msg: &types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
		},
		{
			name: "MsgOutboundTransfer",
			msg: &types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
		},
		{
			name: "MsgUpdateParams",
			msg: &types.MsgUpdateParams{
				Sender: addr1,
				NewParams: types.Params{
					Signers: []string{"s1", "s2", "s3"},
					Assets:  []types.Asset{asset},
				},
			},
		},
		{
			name: "MsgChangeAssetStatus",
			msg: &types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID,
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
	var (
		pk1        = ed25519.GenPrivKey().PubKey()
		addr1Bytes = sdk.AccAddress(pk1.Address())
		addr1      = addr1Bytes.String()

		pk2        = ed25519.GenPrivKey().PubKey()
		addr2Bytes = sdk.AccAddress(pk2.Address())
		addr2      = addr2Bytes.String()

		assetID = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		}
	)
	asset := types.DefaultAssets()[0]
	asset.Id = assetID

	var testCases = []struct {
		name            string
		msg             types.MsgInboundTransfer
		expectedSigners []sdk.AccAddress
		expectedValid   bool
	}{
		{
			name: "valid",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   true,
		},
		{
			name: "empty sender",
			msg: types.MsgInboundTransfer{
				Sender:   "",
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{sdk.AccAddress("")},
			expectedValid:   false,
		},
		{
			name: "invalid sender",
			msg: types.MsgInboundTransfer{
				Sender:   "qwerty",
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{nil},
			expectedValid:   false,
		},
		{
			name: "empty destination addr",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: "",
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
		{
			name: "invalid destination addr",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: "qwerty",
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
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
			expectedValid:   false,
		},
		{
			name: "zero amount",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(0),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
		{
			name: "negative amount",
			msg: types.MsgInboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(-100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			if tc.expectedValid {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			} else {
				require.Error(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			}
		})
	}
}

// TestMsgOutboundTransfer tests if MsgOutboundTransfer messages are properly validated
// and contain proper signers.
func TestMsgOutboundTransfer(t *testing.T) {
	var (
		pk1        = ed25519.GenPrivKey().PubKey()
		addr1Bytes = sdk.AccAddress(pk1.Address())
		addr1      = addr1Bytes.String()

		pk2        = ed25519.GenPrivKey().PubKey()
		addr2Bytes = sdk.AccAddress(pk2.Address())
		addr2      = addr2Bytes.String()

		assetID = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		}
	)
	asset := types.DefaultAssets()[0]
	asset.Id = assetID

	var testCases = []struct {
		name            string
		msg             types.MsgOutboundTransfer
		expectedSigners []sdk.AccAddress
		expectedValid   bool
	}{
		{
			name: "valid",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   true,
		},
		{
			name: "empty sender",
			msg: types.MsgOutboundTransfer{
				Sender:   "",
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{sdk.AccAddress("")},
			expectedValid:   false,
		},
		{
			name: "invalid sender",
			msg: types.MsgOutboundTransfer{
				Sender:   "qwerty",
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{nil},
			expectedValid:   false,
		},
		{
			name: "empty destination addr",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: "",
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
		{
			name: "invalid destination addr",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: "qwerty",
				AssetId:  assetID,
				Amount:   math.NewInt(100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
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
			expectedValid:   false,
		},
		{
			name: "zero amount",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(0),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
		{
			name: "negative amount",
			msg: types.MsgOutboundTransfer{
				Sender:   addr1,
				DestAddr: addr2,
				AssetId:  assetID,
				Amount:   math.NewInt(-100),
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			if tc.expectedValid {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			} else {
				require.Error(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			}
		})
	}
}

// TestMsgUpdateParams tests if MsgUpdateParams messages are properly validated
// and contain proper signers.
func TestMsgUpdateParams(t *testing.T) {
	var (
		pk1        = ed25519.GenPrivKey().PubKey()
		addr1Bytes = sdk.AccAddress(pk1.Address())
		addr1      = addr1Bytes.String()

		pk2        = ed25519.GenPrivKey().PubKey()
		addr2Bytes = sdk.AccAddress(pk2.Address())
		addr2      = addr2Bytes.String()

		assetID1 = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc1",
		}
		assetID2 = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc2",
		}
	)
	asset1 := types.DefaultAssets()[0]
	asset1.Id = assetID1

	asset2 := types.DefaultAssets()[0]
	asset2.Id = assetID2

	// don't check the invalid asset case here since it is already tested in a different case
	var testCases = []struct {
		name            string
		msg             types.MsgUpdateParams
		expectedSigners []sdk.AccAddress
		expectedValid   bool
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
			expectedValid:   true,
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
			expectedValid:   false,
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
			expectedValid:   false,
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
			expectedValid:   true,
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
			expectedValid:   false,
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
			expectedValid:   false,
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
			expectedValid:   false,
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
			expectedValid:   false,
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
			expectedValid:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			if tc.expectedValid {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			} else {
				require.Error(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			}
		})
	}
}

// TestMsgChangeAssetStatus tests if MsgChangeAssetStatus messages are properly validated
// and contain proper signers.
func TestMsgChangeAssetStatus(t *testing.T) {
	var (
		pk1        = ed25519.GenPrivKey().PubKey()
		addr1Bytes = sdk.AccAddress(pk1.Address())
		addr1      = addr1Bytes.String()

		assetID = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc",
		}
	)
	asset := types.DefaultAssets()[0]
	asset.Id = assetID

	// don't check the invalid asset case here since it is already tested in a different case
	var testCases = []struct {
		name            string
		msg             types.MsgChangeAssetStatus
		expectedSigners []sdk.AccAddress
		expectedValid   bool
	}{
		{
			name: "valid",
			msg: types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID,
				NewStatus: types.AssetStatus_ASSET_STATUS_OK,
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   true,
		},
		{
			name: "empty sender",
			msg: types.MsgChangeAssetStatus{
				Sender:    "",
				AssetId:   assetID,
				NewStatus: types.AssetStatus_ASSET_STATUS_OK,
			},
			expectedSigners: []sdk.AccAddress{sdk.AccAddress("")},
			expectedValid:   false,
		},
		{
			name: "invalid sender",
			msg: types.MsgChangeAssetStatus{
				Sender:    "qwerty",
				AssetId:   assetID,
				NewStatus: types.AssetStatus_ASSET_STATUS_OK,
			},
			expectedSigners: []sdk.AccAddress{nil},
			expectedValid:   false,
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
			expectedValid:   false,
		},
		{
			name: "invalid asset status",
			msg: types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID,
				NewStatus: 10,
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
		{
			name: "unspecified asset status",
			msg: types.MsgChangeAssetStatus{
				Sender:    addr1,
				AssetId:   assetID,
				NewStatus: types.AssetStatus_ASSET_STATUS_UNSPECIFIED,
			},
			expectedSigners: []sdk.AccAddress{addr1Bytes},
			expectedValid:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.ElementsMatch(t, tc.msg.GetSigners(), tc.expectedSigners, "test: %v", tc.name)

			if tc.expectedValid {
				require.NoError(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			} else {
				require.Error(t, tc.msg.ValidateBasic(), "test: %v", tc.name)
			}
		})
	}
}
