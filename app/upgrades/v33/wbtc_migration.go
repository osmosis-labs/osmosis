package v33

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	v3 "github.com/osmosis-labs/osmosis/v32/x/cosmwasmpool/cosmwasm/msg/v3"
)

// wbtcUnindexedHolders lists the WBTC holders that the bank denom index does not list. The index was
// never backfilled on osmosis-1, so it misses accounts whose balance has not changed since it was
// added back. Such an account is added to the index on its next balance change, so this list (taken
// from a full scan of the balances store) and the index together cover every holder.
//
//go:embed wbtc_unindexed_holders.txt
var wbtcUnindexedHolders string

// WBTCMigration configures MigrateWBTCToAllBTC.
type WBTCMigration struct {
	// Transmuter is the alloy transmuter that WBTC is deposited into.
	Transmuter string
	// Admin is the admin of Transmuter. The migration lifts the WBTC limiters as this account while
	// it deposits, and restores them afterwards (see MigrateWBTCToAllBTC).
	Admin sdk.AccAddress
	// WBTCDenom is the denom migrated.
	WBTCDenom string
	// AlloyedDenom is the denom Transmuter mints. Each holder must receive exactly the amount of
	// WBTCDenom it deposited.
	AlloyedDenom string
	// Excluded accounts are never migrated.
	Excluded []sdk.AccAddress
	// UnindexedHolders are holders of WBTCDenom that the bank denom index does not list.
	UnindexedHolders []sdk.AccAddress
}

// WBTCMigrationResult summarizes a WBTC to allBTC migration.
type WBTCMigrationResult struct {
	// Migrated is the number of accounts whose WBTC was deposited into the transmuter.
	Migrated int
	// Skipped is the number of holders that are excluded, are not user accounts, or have no spendable
	// WBTC.
	Skipped int
	// Failed is the number of user accounts whose deposit failed and was rolled back.
	Failed int
	// Total is the amount of WBTC deposited into the transmuter.
	Total osmomath.Int
}

// migrateWBTCAndLog runs the mainnet WBTC migration and logs the outcome. The migration is not
// critical to the chain, so a failure is logged instead of halting the upgrade.
func migrateWBTCAndLog(ctx sdk.Context, keepers *keepers.AppKeepers) {
	migration, err := mainnetWBTCMigration(ctx, keepers)
	if err != nil {
		ctx.Logger().Error("v33 upgrade: WBTC to allBTC migration skipped", "error", err)
		return
	}
	result, err := MigrateWBTCToAllBTC(ctx, keepers, migration)
	if err != nil {
		ctx.Logger().Error("v33 upgrade: WBTC to allBTC migration skipped", "error", err)
		return
	}
	ctx.Logger().Info("v33 upgrade: WBTC migrated to allBTC",
		"migrated", result.Migrated, "skipped", result.Skipped, "failed", result.Failed,
		"total", sdk.NewCoin(migration.WBTCDenom, result.Total).String())
}

// mainnetWBTCMigration returns the osmosis-1 migration: the issuer's merchants and custodian are
// excluded, the holders missing from the denom index are added, and the transmuter's limiters are
// lifted as the gov module, which is the allBTC transmuter's admin.
func mainnetWBTCMigration(ctx sdk.Context, keepers *keepers.AppKeepers) (WBTCMigration, error) {
	unindexed, err := parseAddresses(wbtcUnindexedHolders)
	if err != nil {
		return WBTCMigration{}, fmt.Errorf("unindexed WBTC holders: %w", err)
	}
	excluded, err := wbtcIssuerMembers(ctx, keepers, WBTCIssuerAddress)
	if err != nil {
		return WBTCMigration{}, err
	}
	return WBTCMigration{
		Transmuter:       AllBTCTransmuterAddress,
		Admin:            keepers.AccountKeeper.GetModuleAddress(govtypes.ModuleName),
		WBTCDenom:        WBTCDenom,
		AlloyedDenom:     AllBTCDenom,
		Excluded:         excluded,
		UnindexedHolders: unindexed,
	}, nil
}

// MigrateWBTCToAllBTC deposits the spendable WBTC of every user account into the alloy transmuter
// with join_pool, executed as the account, so the account receives the alloyed denom in return. A
// user account is one with a public key that is not excluded, a module account or a contract (see
// isUserAccount). Each deposit runs in its own cache context: a failed deposit is logged and rolled
// back without affecting the others.
//
// The transmuter's limiters reject a deposit that pushes WBTC's weight in the pool past their bound.
// On osmosis-1 WBTC is already above its static limiter, so every deposit would fail. The migration
// therefore lifts the WBTC limiters as m.Admin before depositing and restores their configured bounds
// afterwards. Limiters of the other assets are left alone: depositing WBTC only lowers their weight,
// and the transmuter only enforces a limiter when the weight rises. Lifting, depositing and restoring
// are applied together or not at all, so the limiters are never left lifted.
//
// An error is returned, with no state changed, if the transmuter fails pre-flight (it does not
// accept WBTC, m.Admin is not its admin, or a WBTC limiter has an unknown type), or the limiters cannot be lifted or
// restored.
func MigrateWBTCToAllBTC(ctx sdk.Context, keepers *keepers.AppKeepers, m WBTCMigration) (WBTCMigrationResult, error) {
	noResult := WBTCMigrationResult{Total: osmomath.ZeroInt()}
	transmuterAddr, err := sdk.AccAddressFromBech32(m.Transmuter)
	if err != nil {
		return noResult, fmt.Errorf("transmuter address: %w", err)
	}
	if err := checkTransmuterAcceptsDenom(ctx, keepers, transmuterAddr, m.WBTCDenom); err != nil {
		return noResult, err
	}
	if err := checkTransmuterAdmin(ctx, keepers, transmuterAddr, m.Admin); err != nil {
		return noResult, err
	}
	limiters, err := denomLimiters(ctx, keepers, transmuterAddr, m.WBTCDenom)
	if err != nil {
		return noResult, err
	}
	// Collect the holders before changing any balance, since every deposit updates the denom index.
	holders, err := collectHolders(ctx, keepers, m.WBTCDenom, m.UnindexedHolders)
	if err != nil {
		return noResult, err
	}

	var result WBTCMigrationResult
	err = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
		if err := setLimiterBounds(cacheCtx, keepers, transmuterAddr, m.Admin, m.WBTCDenom, limiters, true); err != nil {
			return fmt.Errorf("lifting %s limiters: %w", m.WBTCDenom, err)
		}
		result = depositHolders(cacheCtx, keepers, m, transmuterAddr, holders)
		if err := setLimiterBounds(cacheCtx, keepers, transmuterAddr, m.Admin, m.WBTCDenom, limiters, false); err != nil {
			return fmt.Errorf("restoring %s limiters: %w", m.WBTCDenom, err)
		}
		return nil
	})
	if err != nil {
		return noResult, err
	}
	return result, nil
}

// depositHolders deposits the spendable WBTC of every user account among holders into the
// transmuter. See MigrateWBTCToAllBTC.
func depositHolders(ctx sdk.Context, keepers *keepers.AppKeepers, m WBTCMigration, transmuter sdk.AccAddress, holders []sdk.AccAddress) WBTCMigrationResult {
	result := WBTCMigrationResult{Total: osmomath.ZeroInt()}
	excluded := make(map[string]struct{}, len(m.Excluded))
	for _, addr := range m.Excluded {
		excluded[string(addr)] = struct{}{}
	}

	for _, holder := range holders {
		if _, isExcluded := excluded[string(holder)]; isExcluded || !isUserAccount(ctx, keepers, holder) {
			result.Skipped++
			continue
		}
		coin := keepers.BankKeeper.SpendableCoin(ctx, holder, m.WBTCDenom)
		if !coin.IsPositive() {
			result.Skipped++
			continue
		}
		err := osmoutils.ApplyFuncIfNoErrorLogToDebug(ctx, func(cacheCtx sdk.Context) error {
			return joinAsHolder(cacheCtx, keepers, transmuter, holder, coin, m.AlloyedDenom)
		})
		if err != nil {
			ctx.Logger().Error("v33 upgrade: WBTC migration failed for account", "address", holder.String(), "amount", coin.String(), "error", err)
			result.Failed++
			continue
		}
		result.Migrated++
		result.Total = result.Total.Add(coin.Amount)
	}
	return result
}

// liftedLimiterBound is the bound a limiter is set to while the migration deposits. A weight is at
// most 1, so a static limiter with an upper limit of 1 never rejects a deposit, and neither does a
// change limiter with a boundary offset of 1, whose bound is its moving average plus the offset.
const liftedLimiterBound = "1"

// transmuterLimiter is a limiter the transmuter enforces on one denom. Exactly one of
// StaticUpperLimit and ChangeBoundaryOffset is set, depending on the limiter's type.
type transmuterLimiter struct {
	Label                string
	StaticUpperLimit     string
	ChangeBoundaryOffset string
}

// checkTransmuterAdmin returns an error unless admin is the transmuter's admin, the only account
// allowed to change its limiters.
func checkTransmuterAdmin(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter, admin sdk.AccAddress) error {
	var resp struct {
		Admin string `json:"admin"`
	}
	if err := querySmart(ctx, keepers, transmuter, `{"get_admin": {}}`, &resp); err != nil {
		return err
	}
	if resp.Admin != admin.String() {
		return fmt.Errorf("transmuter %s has admin %q, expected %s", transmuter, resp.Admin, admin)
	}
	return nil
}

// denomLimiters returns the limiters the transmuter enforces on denom. It returns an error if one of
// them is neither a static nor a change limiter, since it could not be lifted.
func denomLimiters(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter sdk.AccAddress, denom string) ([]transmuterLimiter, error) {
	// Each entry is [[denom, label], {"<limiter type>": {...}}].
	var resp struct {
		Limiters [][2]json.RawMessage `json:"limiters"`
	}
	if err := querySmart(ctx, keepers, transmuter, `{"list_limiters": {}}`, &resp); err != nil {
		return nil, err
	}

	var limiters []transmuterLimiter
	for _, entry := range resp.Limiters {
		var key [2]string
		if err := json.Unmarshal(entry[0], &key); err != nil {
			return nil, fmt.Errorf("decoding limiter key %s: %w", entry[0], err)
		}
		if key[0] != denom {
			continue
		}
		var params struct {
			Static *struct {
				UpperLimit string `json:"upper_limit"`
			} `json:"static_limiter"`
			Change *struct {
				BoundaryOffset string `json:"boundary_offset"`
			} `json:"change_limiter"`
		}
		if err := json.Unmarshal(entry[1], &params); err != nil {
			return nil, fmt.Errorf("decoding limiter %q of %s: %w", key[1], denom, err)
		}
		limiter := transmuterLimiter{Label: key[1]}
		switch {
		case params.Static != nil && params.Static.UpperLimit != "":
			limiter.StaticUpperLimit = params.Static.UpperLimit
		case params.Change != nil && params.Change.BoundaryOffset != "":
			limiter.ChangeBoundaryOffset = params.Change.BoundaryOffset
		default:
			return nil, fmt.Errorf("limiter %q of %s has an unknown type: %s", key[1], denom, entry[1])
		}
		limiters = append(limiters, limiter)
	}
	return limiters, nil
}

// setLimiterBounds sets the bound of each of denom's limiters, executed as the transmuter's admin:
// liftedLimiterBound if lift is set, otherwise the limiter's configured bound.
func setLimiterBounds(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter, admin sdk.AccAddress, denom string, limiters []transmuterLimiter, lift bool) error {
	for _, limiter := range limiters {
		method, field, bound := "set_static_limiter_upper_limit", "upper_limit", limiter.StaticUpperLimit
		if limiter.StaticUpperLimit == "" {
			method, field, bound = "set_change_limiter_boundary_offset", "boundary_offset", limiter.ChangeBoundaryOffset
		}
		if lift {
			bound = liftedLimiterBound
		}
		bz, err := json.Marshal(map[string]map[string]string{method: {"denom": denom, "label": limiter.Label, field: bound}})
		if err != nil {
			return err
		}
		if _, err := keepers.ContractKeeper.Execute(ctx, transmuter, admin, bz, nil); err != nil {
			return fmt.Errorf("%s for limiter %q: %w", method, limiter.Label, err)
		}
	}
	return nil
}

// checkTransmuterAcceptsDenom returns an error unless transmuter is a contract that lists denom in
// its asset configs.
func checkTransmuterAcceptsDenom(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter sdk.AccAddress, denom string) error {
	if !keepers.WasmKeeper.HasContractInfo(ctx, transmuter) {
		return fmt.Errorf("transmuter %s is not a contract", transmuter)
	}
	var resp v3.ListAssetConfigsResponse
	if err := querySmart(ctx, keepers, transmuter, `{"list_asset_configs": {}}`, &resp); err != nil {
		return err
	}
	for _, config := range resp.AssetConfigs {
		if config.Denom == denom {
			return nil
		}
	}
	return fmt.Errorf("transmuter %s does not accept %s", transmuter, denom)
}

// wbtcIssuerMembers returns the merchants and the custodian of the WBTC controller contract issuer.
func wbtcIssuerMembers(ctx sdk.Context, keepers *keepers.AppKeepers, issuer string) ([]sdk.AccAddress, error) {
	issuerAddr, err := sdk.AccAddressFromBech32(issuer)
	if err != nil {
		return nil, fmt.Errorf("issuer address: %w", err)
	}
	var merchants struct {
		Merchants []string `json:"merchants"`
	}
	query := fmt.Sprintf(`{"list_merchants": {"limit": %d}}`, maxWBTCIssuerMembers)
	if err := querySmart(ctx, keepers, issuerAddr, query, &merchants); err != nil {
		return nil, err
	}
	// A full page may be truncated, and a missed merchant would be migrated.
	if len(merchants.Merchants) >= maxWBTCIssuerMembers {
		return nil, fmt.Errorf("issuer %s has at least %d merchants", issuer, maxWBTCIssuerMembers)
	}
	var custodian struct {
		Address string `json:"address"`
	}
	if err := querySmart(ctx, keepers, issuerAddr, `{"get_custodian": {}}`, &custodian); err != nil {
		return nil, err
	}
	if custodian.Address == "" {
		return nil, fmt.Errorf("issuer %s returned no custodian", issuer)
	}

	members := make([]sdk.AccAddress, 0, len(merchants.Merchants)+1)
	for _, member := range append(merchants.Merchants, custodian.Address) {
		addr, err := sdk.AccAddressFromBech32(member)
		if err != nil {
			return nil, fmt.Errorf("issuer member %q: %w", member, err)
		}
		members = append(members, addr)
	}
	return members, nil
}

// querySmart runs a smart query against contract and decodes the response into resp.
func querySmart(ctx sdk.Context, keepers *keepers.AppKeepers, contract sdk.AccAddress, query string, resp any) error {
	bz, err := keepers.WasmKeeper.QuerySmart(ctx, contract, []byte(query))
	if err != nil {
		return fmt.Errorf("querying %s with %s: %w", contract, query, err)
	}
	if err := json.Unmarshal(bz, resp); err != nil {
		return fmt.Errorf("decoding %s response of %s: %w", query, contract, err)
	}
	return nil
}

// collectHolders returns the addresses with a balance of denom: those in the bank denom index, in
// address order, followed by the addresses of unindexed that are not in the index and still hold
// denom. Holders are not capped: one that is not a user account is skipped by a single account
// lookup, so spreading dust to many addresses costs the upgrade little, and a cap would let anyone
// cancel the migration by exceeding it.
func collectHolders(ctx sdk.Context, keepers *keepers.AppKeepers, denom string, unindexed []sdk.AccAddress) ([]sdk.AccAddress, error) {
	holders := make([]sdk.AccAddress, 0, 2048)
	seen := make(map[string]struct{}, 2048)
	add := func(addr sdk.AccAddress) {
		if _, ok := seen[string(addr)]; ok {
			return
		}
		seen[string(addr)] = struct{}{}
		holders = append(holders, addr)
	}

	iter, err := keepers.BankKeeper.Balances.Indexes.Denom.MatchExact(ctx, denom)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		pk, err := iter.PrimaryKey()
		if err != nil {
			return nil, err
		}
		add(pk.K1())
	}

	for _, addr := range unindexed {
		if keepers.BankKeeper.GetBalance(ctx, addr, denom).IsZero() {
			continue
		}
		add(addr)
	}
	return holders, nil
}

// isUserAccount reports whether addr is an account controlled by a key: an existing account with a
// public key that is neither a module account nor a wasm contract. Accounts without a public key are
// protocol accounts (pools, IBC escrows, CL spread reward and incentive addresses, ICA host accounts,
// contracts) whose balance is owed to others, or accounts that have never signed a transaction.
func isUserAccount(ctx sdk.Context, keepers *keepers.AppKeepers, addr sdk.AccAddress) bool {
	acc := keepers.AccountKeeper.GetAccount(ctx, addr)
	if acc == nil || acc.GetPubKey() == nil {
		return false
	}
	if _, isModule := acc.(sdk.ModuleAccountI); isModule {
		return false
	}
	return !keepers.WasmKeeper.HasContractInfo(ctx, addr)
}

// joinAsHolder deposits coin into the transmuter on behalf of holder and checks that holder spent
// exactly coin and received exactly the same amount of alloyedDenom in return.
func joinAsHolder(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter, holder sdk.AccAddress, coin sdk.Coin, alloyedDenom string) error {
	wbtcBefore := keepers.BankKeeper.GetBalance(ctx, holder, coin.Denom).Amount
	alloyedBefore := keepers.BankKeeper.GetBalance(ctx, holder, alloyedDenom).Amount

	if _, err := keepers.ContractKeeper.Execute(ctx, transmuter, holder, []byte(`{"join_pool": {}}`), sdk.NewCoins(coin)); err != nil {
		return err
	}

	if spent := wbtcBefore.Sub(keepers.BankKeeper.GetBalance(ctx, holder, coin.Denom).Amount); !spent.Equal(coin.Amount) {
		return fmt.Errorf("spent %s %s, expected %s", spent, coin.Denom, coin.Amount)
	}
	received := keepers.BankKeeper.GetBalance(ctx, holder, alloyedDenom).Amount.Sub(alloyedBefore)
	if !received.Equal(coin.Amount) {
		return fmt.Errorf("received %s %s, expected %s", received, alloyedDenom, coin.Amount)
	}
	return nil
}

// parseAddresses parses one bech32 address per line, ignoring blank lines and lines starting with #.
func parseAddresses(list string) ([]sdk.AccAddress, error) {
	addrs := make([]sdk.AccAddress, 0, 512)
	for _, line := range strings.Split(list, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		addr, err := sdk.AccAddressFromBech32(line)
		if err != nil {
			return nil, fmt.Errorf("address %q: %w", line, err)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}
