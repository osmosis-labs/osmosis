package wasmbinding

import (
	"sync"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	epochtypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
)

// StargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var StargateWhitelist sync.Map

func init() {
	StargateWhitelist.Store("/cosmos.auth.v1beta1.Query/Account", &authtypes.QueryAccountResponse{})
	StargateWhitelist.Store("/cosmos.bank.v1beta1.Query/AllBalances", &banktypes.QueryAllBalancesResponse{})

	StargateWhitelist.Store("/osmosis.epochs.v1beta1.Query/EpochInfos", &epochtypes.QueryCurrentEpochResponse{})
}
