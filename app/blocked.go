package app

import (
	"strings"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *OsmosisApp) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	// Deleted all OFAC blocked addresses. To be decentralized, access should remain permissionless. Preemptive compliance
	// is a dangerous precedent that will destroy the philosophical underpinnings of decentralized applications.
	ofacRawEthAddrs := []string{}
	for _, addr := range ofacRawEthAddrs {
		blockedAddrs[addr] = true
		blockedAddrs[strings.ToLower(addr)] = true
	}

	return blockedAddrs
}
