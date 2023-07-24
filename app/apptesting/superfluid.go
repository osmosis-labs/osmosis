package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gammtypes "github.com/osmosis-labs/osmosis/v17/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v17/x/superfluid/types"
)

func (s *KeeperTestHelper) SuperfluidDelegateToDefaultVal(sender sdk.AccAddress, poolId uint64, lockId uint64) error {
	valAddr := s.SetupValidator(stakingtypes.Bonded)

	poolDenom := gammtypes.GetPoolShareDenom(poolId)
	err := s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     poolDenom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})
	s.Require().NoError(err)

	return s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, sender.String(), lockId, valAddr.String())
}
