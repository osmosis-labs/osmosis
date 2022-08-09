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
	// List updated with tornado cash addresses Aug 9 2022
	// we should ensure that each build of osmosis contains all blocked addresses in this file.
	// tooling to parse the 55mb ofac list is available here: https://github.com/0xB10C/ofac-sanctioned-digital-currency-addresses
	// an updated list of ethereum addresses is available here: https://raw.githubusercontent.com/0xB10C/ofac-sanctioned-digital-currency-addresses/lists/sanctioned_addresses_ETH.txt
	// as of Aug 9 2022, this file contains all 68 sanctioned ethereum addresses.
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
		"0x8589427373D6D84E98730D7795D8f6f8731FDA16",
		"0x722122dF12D4e14e13Ac3b6895a86e84145b6967",
		"0xDD4c48C0B24039969fC16D1cdF626eaB821d3384",
		"0xd90e2f925DA726b50C4Ed8D0Fb90Ad053324F31b",
		"0xd96f2B1c14Db8458374d9Aca76E26c3D18364307",
		"0x4736dCf1b7A3d580672CcE6E7c65cd5cc9cFBa9D",
		"0xD4B88Df4D29F5CedD6857912842cff3b20C8Cfa3",
		"0x910Cbd523D972eb0a6f4cAe4618aD62622b39DbF",
		"0xA160cdAB225685dA1d56aa342Ad8841c3b53f291",
		"0xFD8610d20aA15b7B2E3Be39B396a1bC3516c7144",
		"0xF60dD140cFf0706bAE9Cd734Ac3ae76AD9eBC32A",
		"0x22aaA7720ddd5388A3c0A3333430953C68f1849b",
		"0xBA214C1c1928a32Bffe790263E38B4Af9bFCD659",
		"0xb1C8094B234DcE6e03f10a5b673c1d8C69739A00",
		"0x527653eA119F3E6a1F5BD18fbF4714081D7B31ce",
		"0x58E8dCC13BE9780fC42E8723D8EaD4CF46943dF2",
		"0xD691F27f38B395864Ea86CfC7253969B409c362d",
		"0xaEaaC358560e11f52454D997AAFF2c5731B6f8a6",
		"0x1356c899D8C9467C7f71C195612F8A395aBf2f0a",
		"0xA60C772958a3eD56c1F15dD055bA37AC8e523a0D",
		"0x169AD27A470D064DEDE56a2D3ff727986b15D52B",
		"0x0836222F2B2B24A3F36f98668Ed8F0B38D1a872f",
		"0xF67721A2D8F736E75a49FdD7FAd2e31D8676542a",
		"0x9AD122c22B14202B4490eDAf288FDb3C7cb3ff5E",
		"0x905b63Fff465B9fFBF41DeA908CEb12478ec7601",
		"0x07687e702b410Fa43f4cB4Af7FA097918ffD2730",
		"0x94A1B5CdB22c43faab4AbEb5c74999895464Ddaf",
		"0xb541fc07bC7619fD4062A54d96268525cBC6FfEF",
		"0x12D66f87A04A9E220743712cE6d9bB1B5616B8Fc",
		"0x47CE0C6eD5B0Ce3d3A51fdb1C52DC66a7c3c2936",
		"0x23773E65ed146A459791799d01336DB287f25334",
		"0xD21be7248e0197Ee08E0c20D4a96DEBdaC3D20Af",
		"0x610B717796ad172B316836AC95a2ffad065CeaB4",
		"0x178169B423a011fff22B9e3F3abeA13414dDD0F1",
		"0xbB93e510BbCD0B7beb5A853875f9eC60275CF498",
		"0x2717c5e28cf931547B621a5dddb772Ab6A35B701",
		"0x03893a7c7463AE47D46bc7f091665f1893656003",
		"0xCa0840578f57fE71599D29375e16783424023357",
		"0x58E8dCC13BE9780fC42E8723D8EaD4CF46943dF2",
		"0x8589427373D6D84E98730D7795D8f6f8731FDA16",
		"0x722122dF12D4e14e13Ac3b6895a86e84145b6967",
		"0xDD4c48C0B24039969fC16D1cdF626eaB821d3384",
		"0xd90e2f925DA726b50C4Ed8D0Fb90Ad053324F31b",
		"0xd96f2B1c14Db8458374d9Aca76E26c3D18364307",
		"0x4736dCf1b7A3d580672CcE6E7c65cd5cc9cFBa9D",
	}
	for _, addr := range ofacRawEthAddrs {
		blockedAddrs[addr] = true
		blockedAddrs[strings.ToLower(addr)] = true
	}

	return blockedAddrs
}
