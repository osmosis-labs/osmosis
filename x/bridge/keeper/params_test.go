package keeper_test

import (
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	const votesNeeded = 100
	fee := math.LegacyNewDecWithPrec(8, 1)

	s.Run("valid params", func() {
		s.SetupTest()

		newParams := types.Params{
			Signers:     []string{addr1, addr2},
			Assets:      []types.Asset{asset1, asset2},
			VotesNeeded: votesNeeded,
			Fee:         fee,
		}

		_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
			Sender:    s.authority,
			NewParams: newParams,
		})
		s.Require().NoError(err)

		// Params have changed
		actualParams := s.GetParams()
		s.Require().Equal(newParams, actualParams)

		// Check the event was emitted
		eventType := proto.MessageName(new(types.EventUpdateParams))
		s.Require().True(s.HasEvent(eventType))

		// Get all tf denoms created by the bridge module
		tfDenoms := s.GetBridgeTFDenoms()
		// There should be 3 denoms since outdated denoms are not deleted from the tokenfactory
		s.Require().Len(tfDenoms, 3)

		// Get all denoms based on the assets stored in the module params
		bridgeDenoms := s.GetBridgeDenoms()
		// There should be 2 denoms since outdated denoms are deleted from the bridge
		s.Require().Len(bridgeDenoms, 2)
	})

	s.Run("sender is not the authority", func() {
		s.SetupTest()

		// Get initial params
		initialParams := s.GetParams()

		newParams := types.Params{
			Signers:     []string{addr1},
			Assets:      []types.Asset{asset1, asset2}, // duplicated assets
			VotesNeeded: votesNeeded,
			Fee:         fee,
		}

		// Update params
		_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
			Sender:    addr1, // just random addr
			NewParams: newParams,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)

		// Params have not changed
		actualParams := s.GetParams()
		s.Require().Equal(initialParams, actualParams)

		eventType := proto.MessageName(new(types.EventUpdateParams))
		s.Require().False(s.HasEvent(eventType))
	})

	s.Run("invalid params", func() {
		s.SetupTest()

		// Get initial params
		initialParams := s.GetParams()

		newParams := types.Params{
			Signers:     []string{addr1, addr1},
			Assets:      []types.Asset{asset1, asset2},
			VotesNeeded: votesNeeded,
			Fee:         fee,
		}

		_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{
			Sender:    s.authority,
			NewParams: newParams,
		})
		s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)

		// Params have not changed
		actualParams := s.GetParams()
		s.Require().Equal(initialParams, actualParams)

		eventType := proto.MessageName(new(types.EventUpdateParams))
		s.Require().False(s.HasEvent(eventType))
	})
}
