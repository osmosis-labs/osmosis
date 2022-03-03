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

	// We block all OFAC-blocked ETH addresses from receiving tokens as well
	// The list is sourced from: https://www.treasury.gov/ofac/downloads/sanctions/1.0/sdn_advanced.xml
	ofacRawEthAddrs := []string{
		"0x7F367cC41522cE07553e823bf3be79A889DEbe1B",
		"0xd882cfc20f52f2599d84b8e8d58c7fb62cfe344b",
		"0x901bb9583b24d97e995513c6778dc6888ab6870e",
		"0xa7e5d5a720f06526557c513402f2e6b5fa20b008",
		"0x8576acc5c05d6ce88f4e49bf65bdf0c62f91353c",
		"0x1da5821544e25c636c1417ba96ade4cf6d2f9b5a",
		"0x7Db418b5D567A4e0E8c59Ad71BE1FcE48f3E6107",
		"0x72a5843cc08275C8171E582972Aa4fDa8C397B2A",
		"0x7F19720A857F834887FC9A7bC0a0fBe7Fc7f8102",
		"0x9f4cda013e354b8fc285bf4b9a60460cee7f7ea9",
		"03cbded43efdaf0fc77b9c55f6fc9988fcc9b757d",
		"0x2f389ce8bd8ff92de3402ffce4691d17fc4f6535",
		"0x19aa5fe80d33a56d56c78e82ea5e50e5d80b4dff",
		"0xe7aa314c77f4233c18c6cc84384a9247c0cf367b",
		"0x308ed4b7b49797e1a98d3818bff6fe5385410370",
		"0x2f389ce8bd8ff92de3402ffce4691d17fc4f6535",
		"0x19aa5fe80d33a56d56c78e82ea5e50e5d80b4dff",
		"0x67d40EE1A85bf4a4Bb7Ffae16De985e8427B6b45",
		"0x6f1ca141a28907f78ebaa64fb83a9088b02a8352",
		"0x6acdfba02d390b97ac2b2d42a63e85293bcc160e",
		"0x48549a34ae37b12f6a30566245176994e17c6b4a",
		"0x5512d943ed1f7c8a43f3435c85f7ab68b30121b0",
		"0xc455f7fd3e0e12afd51fba5c106909934d8a0e4a",
		"0xfec8a60023265364d066a1212fde3930f6ae8da7",
	}
	for _, addr := range ofacRawEthAddrs {
		blockedAddrs[addr] = true
		blockedAddrs[strings.ToLower(addr)] = true
	}

	return blockedAddrs
}
