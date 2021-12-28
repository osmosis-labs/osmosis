package keepers

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	claimtypes "github.com/osmosis-labs/osmosis/x/claim/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	txfeestypes "github.com/osmosis-labs/osmosis/x/txfees/types"
)

var (
	// module account permissions
	// TODO: Delete this on next upgrade, all of these module permissions should be set in their apps.
	// Or at minimum, we should make a test for consistency here.
	MaccPerms = map[string][]string{
		authtypes.FeeCollectorName:               nil,
		distrtypes.ModuleName:                    nil,
		minttypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
		minttypes.DeveloperVestingModuleAcctName: nil,
		stakingtypes.BondedPoolName:              {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:           {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                      {authtypes.Burner},
		ibctransfertypes.ModuleName:              {authtypes.Minter, authtypes.Burner},
		claimtypes.ModuleName:                    {authtypes.Minter, authtypes.Burner},
		gammtypes.ModuleName:                     {authtypes.Minter, authtypes.Burner},
		incentivestypes.ModuleName:               {authtypes.Minter, authtypes.Burner},
		lockuptypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
		poolincentivestypes.ModuleName:           nil,
		txfeestypes.ModuleName:                   nil,
	}
)
