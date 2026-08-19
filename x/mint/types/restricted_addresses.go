package types

// RestrictedAddresses is the curated set of foundation / investor / strategic
// wallet addresses (bech32) whose liquid, staked, and unbonding OSMO are
// excluded from circulating supply.
//
// This list is intentionally a compiled-in constant rather than a governance
// param: it changes rarely, and keeping it out of consensus state lets the
// supply endpoints ship as a point release (no migration, no upgrade handler).
// Updating the list is itself a code change shipped in a point release.
//
// The entries are kept as strings (not pre-parsed sdk.AccAddress) on purpose:
// the bech32 "osmo" prefix is configured by app params init (SetAddressPrefixes),
// which is not guaranteed to have run when this package initialises in isolation
// (e.g. in a unit test of the types package). Parsing therefore happens at query
// time in the mint keeper, after the app has configured the prefix.
//
// A malformed entry does NOT degrade gracefully: GetRestrictedSupply returns an
// error for it, which fails the restricted-supply, circulating-supply, and
// inflation queries until a corrected binary ships. TestRestrictedAddressesParse
// guards every entry at CI time; any list change must keep that test passing.
//
// Composition: the foundation address (the strategic-reserve holder traceable
// from the airdrop) plus the original genesis developer-rewards receiver set.
// Those genesis receivers were collapsed by governance into the single current
// WeightedDeveloperRewardsReceivers entry, so they no longer appear in mint
// params and are not otherwise counted. The current param receiver is NOT listed
// here: it is already counted via the params loop in GetRestrictedSupply, and
// the keeper de-duplicates against the param set to prevent double-counting if
// an address ever appears in both.
//
// MAINTENANCE NOTE: the current WeightedDeveloperRewardsReceivers param entry is
// intentionally omitted here precisely because the param loop counts it. If
// governance later removes an address from that param (as it did for this
// genesis set), that address stops being counted by either path and silently
// re-enters circulating supply. When dev-rewards receivers change, review
// whether the outgoing address should be added to this constant.
//
// Deliberately excluded (these circulate or are earmarked for sale, per the
// restricted-supply methodology): liquidity deployments, grant deployments, and
// community-pool liquidity positions.
var RestrictedAddresses = []string{
	// Foundation / strategic reserve.
	"osmo1ugku28hwyexpljrrmtet05nd6kjlrvr9jz6z00",

	// Original genesis developer-rewards receivers (no longer in mint params).
	"osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
	"osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
	"osmo13lt0hzc6u3htsk7z5rs6vuurmgg4hh2ecgxqkf",
	"osmo1kvc3he93ygc0us3ycslwlv2gdqry4ta73vk9hu",
	"osmo19qgldlsk7hdv3ddtwwpvzff30pxqe9phq9evxf",
	"osmo19fs55cx4594een7qr8tglrjtt5h9jrxg458htd",
	"osmo1ssp6px3fs3kwreles3ft6c07mfvj89a544yj9k",
	"osmo1c5yu8498yzqte9cmfv5zcgtl07lhpjrj0skqdx",
	"osmo1yhj3r9t9vw7qgeg22cehfzj7enwgklw5k5v7lj",
	"osmo18nzmtyn5vy5y45dmcdnta8askldyvehx66lqgm",
	"osmo1z2x9z58cg96ujvhvu6ga07yv9edq2mvkxpgwmc",
	"osmo1tvf3373skua8e6480eyy38avv8mw3hnt8jcxg9",
	"osmo1zs0txy03pv5crj2rvty8wemd3zhrka2ne8u05n",
	"osmo1djgf9p53n7m5a55hcn6gg0cm5mue4r5g3fadee",
	"osmo1488zldkrn8xcjh3z40v2mexq7d088qkna8ceze",
}
