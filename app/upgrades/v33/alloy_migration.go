package v33

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

// nobleUSDCUnindexedHolders lists the Noble USDC holders that the bank denom index does not list.
// The file is a placeholder until cmd/alloy-holder-scan is run against a mainnet export.
//
//go:embed noble_usdc_unindexed_holders.txt
var nobleUSDCUnindexedHolders string

// Bounds on contract responses. A longer list may be truncated, and a truncated corrupted-denom
// list could hide the denom being migrated, so the migration is skipped instead.
const (
	maxListedAssetConfigs = 64
	maxCorruptedDenoms    = 64
)

// AlloyMigration configures MigrateToAlloy for one native denom deposited into an alloy transmuter.
type AlloyMigration struct {
	// Name identifies the migration in logs.
	Name string
	// Transmuter is the alloy transmuter the denom is deposited into.
	Transmuter string
	// Denom is the denom migrated.
	Denom string
	// AlloyedDenom is the denom Transmuter mints. Each holder must receive exactly the amount of
	// Denom it deposited, which requires the two denoms to share a normalization factor.
	AlloyedDenom string
	// Excluded accounts are never migrated.
	Excluded []sdk.AccAddress
	// UnindexedHolders are holders of Denom that the bank denom index does not list.
	UnindexedHolders []sdk.AccAddress
	// MaxHolders bounds the holder list. More holders than this skips the migration.
	MaxHolders int
}

// AlloyMigrationResult summarizes one alloy migration.
type AlloyMigrationResult struct {
	// Migrated is the number of accounts whose denom was deposited into the transmuter.
	Migrated int
	// Skipped is the number of holders that are excluded, are not user accounts, or have no
	// spendable balance of the denom.
	Skipped int
	// Failed is the number of user accounts whose deposit failed and was rolled back.
	Failed int
	// Total is the amount of the denom deposited into the transmuter.
	Total osmomath.Int
}

// AlloyMigrationBuilder builds one osmosis-1 migration. Builders run independently: an error from
// one skips that migration and does not stop the next.
type AlloyMigrationBuilder func(ctx sdk.Context, keepers *keepers.AppKeepers) (AlloyMigration, error)

// alloyMigrationBuilders are the osmosis-1 migrations, in run order.
var alloyMigrationBuilders = []AlloyMigrationBuilder{
	namedAlloyMigration("WBTC to allBTC", mainnetWBTCMigration),
	namedAlloyMigration("Noble USDC to allUSDC", mainnetNobleUSDCMigration),
}

// namedAlloyMigration sets Name even when build fails, so the skip log can name the migration.
func namedAlloyMigration(name string, build AlloyMigrationBuilder) AlloyMigrationBuilder {
	return func(ctx sdk.Context, keepers *keepers.AppKeepers) (AlloyMigration, error) {
		m, err := build(ctx, keepers)
		if m.Name == "" {
			m.Name = name
		}
		return m, err
	}
}

// runMainnetAlloyMigrations runs each configured migration on its own. A pre-flight failure or any
// other error is logged and skipped. It never returns an error, so one migration cannot halt the
// upgrade or skip the others.
func runMainnetAlloyMigrations(ctx sdk.Context, keepers *keepers.AppKeepers) {
	for _, build := range alloyMigrationBuilders {
		runAlloyMigration(ctx, keepers, build)
	}
}

// runAlloyMigration runs one builder and its migration. Failures are logged under the migration name.
func runAlloyMigration(ctx sdk.Context, keepers *keepers.AppKeepers, build AlloyMigrationBuilder) {
	m, err := build(ctx, keepers)
	if err != nil {
		ctx.Logger().Error("v33 upgrade: alloy migration skipped", "name", m.Name, "error", err)
		return
	}
	result, err := MigrateToAlloy(ctx, keepers, m)
	if err != nil {
		ctx.Logger().Error("v33 upgrade: alloy migration skipped", "name", m.Name, "error", err)
		return
	}
	ctx.Logger().Info("v33 upgrade: alloy migration finished",
		"name", m.Name, "migrated", result.Migrated, "skipped", result.Skipped, "failed", result.Failed,
		"total", sdk.NewCoin(m.Denom, result.Total).String())
}

// mainnetWBTCMigration returns the osmosis-1 WBTC migration: the issuer's merchants and custodian
// are excluded, and the holders missing from the denom index are added.
func mainnetWBTCMigration(ctx sdk.Context, keepers *keepers.AppKeepers) (AlloyMigration, error) {
	unindexed, err := parseAddresses(wbtcUnindexedHolders, MaxWBTCHoldersToMigrate)
	if err != nil {
		return AlloyMigration{}, fmt.Errorf("unindexed WBTC holders: %w", err)
	}
	excluded, err := wbtcIssuerMembers(ctx, keepers, WBTCIssuerAddress)
	if err != nil {
		return AlloyMigration{}, err
	}
	return AlloyMigration{
		Name:             "WBTC to allBTC",
		Transmuter:       AllBTCTransmuterAddress,
		Denom:            WBTCDenom,
		AlloyedDenom:     AllBTCDenom,
		Excluded:         excluded,
		UnindexedHolders: unindexed,
		MaxHolders:       MaxWBTCHoldersToMigrate,
	}, nil
}

// mainnetNobleUSDCMigration returns the osmosis-1 Noble USDC migration. Noble USDC is an IBC
// voucher, so there is no issuer to query. Exclusions are the explicit nobleUSDCExcludedBech32 list.
func mainnetNobleUSDCMigration(_ sdk.Context, _ *keepers.AppKeepers) (AlloyMigration, error) {
	unindexed, err := parseAddresses(nobleUSDCUnindexedHolders, MaxNobleUSDCHoldersToMigrate)
	if err != nil {
		return AlloyMigration{}, fmt.Errorf("unindexed Noble USDC holders: %w", err)
	}
	excluded, err := accAddressesFromBech32("excluded Noble USDC account", nobleUSDCExcludedBech32)
	if err != nil {
		return AlloyMigration{}, err
	}
	return AlloyMigration{
		Name:             "Noble USDC to allUSDC",
		Transmuter:       AllUSDCTransmuterAddress,
		Denom:            NobleUSDCDenom,
		AlloyedDenom:     AllUSDCDenom,
		Excluded:         excluded,
		UnindexedHolders: unindexed,
		MaxHolders:       MaxNobleUSDCHoldersToMigrate,
	}, nil
}

// accAddressesFromBech32 parses a configured address list.
func accAddressesFromBech32(label string, list []string) ([]sdk.AccAddress, error) {
	addrs := make([]sdk.AccAddress, 0, len(list))
	for _, bech32 := range list {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", label, bech32, err)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

// MigrateToAlloy deposits the spendable balance of m.Denom of every user account into the alloy
// transmuter with join_pool, executed as the account, so the account receives the alloyed denom in
// return. A user account is one with a public key that is not excluded, a module account or a
// contract (see isUserAccount). Each deposit runs in its own cache context: a failed deposit is
// logged and rolled back without affecting the others.
//
// An error is returned, with no state changed, if the transmuter fails pre-flight or there are more
// than m.MaxHolders holders.
func MigrateToAlloy(ctx sdk.Context, keepers *keepers.AppKeepers, m AlloyMigration) (AlloyMigrationResult, error) {
	result := AlloyMigrationResult{Total: osmomath.ZeroInt()}
	if err := m.validate(); err != nil {
		return result, err
	}
	transmuterAddr, err := sdk.AccAddressFromBech32(m.Transmuter)
	if err != nil {
		return result, fmt.Errorf("transmuter address: %w", err)
	}
	if err := checkTransmuter(ctx, keepers, transmuterAddr, m.Denom, m.AlloyedDenom); err != nil {
		return result, err
	}
	// Collect the holders before changing any balance, since every deposit updates the denom index.
	holders, err := collectHolders(ctx, keepers, m.Denom, m.UnindexedHolders, m.MaxHolders)
	if err != nil {
		return result, err
	}
	return joinHolders(ctx, keepers, m, transmuterAddr, holders), nil
}

// validate checks the migration config before any query or state change.
func (m AlloyMigration) validate() error {
	if m.Name == "" {
		return errors.New("migration name is empty")
	}
	if m.Denom == "" || m.AlloyedDenom == "" {
		return errors.New("migration denom is empty")
	}
	if m.Denom == m.AlloyedDenom {
		return fmt.Errorf("denom and alloyed denom are both %s", m.Denom)
	}
	if m.MaxHolders <= 0 {
		return errors.New("max holders must be positive")
	}
	return nil
}

// joinHolders deposits each eligible holder's spendable balance. A failed deposit is rolled back.
func joinHolders(ctx sdk.Context, keepers *keepers.AppKeepers, m AlloyMigration, transmuter sdk.AccAddress, holders []sdk.AccAddress) AlloyMigrationResult {
	result := AlloyMigrationResult{Total: osmomath.ZeroInt()}
	excluded := make(map[string]struct{}, len(m.Excluded))
	for _, addr := range m.Excluded {
		excluded[string(addr)] = struct{}{}
	}
	for _, holder := range holders {
		if _, isExcluded := excluded[string(holder)]; isExcluded || !isUserAccount(ctx, keepers, holder) {
			result.Skipped++
			continue
		}
		coin := keepers.BankKeeper.SpendableCoin(ctx, holder, m.Denom)
		if !coin.IsPositive() {
			result.Skipped++
			continue
		}
		err := osmoutils.ApplyFuncIfNoErrorLogToDebug(ctx, func(cacheCtx sdk.Context) error {
			return joinAsHolder(cacheCtx, keepers, transmuter, holder, coin, m.AlloyedDenom)
		})
		if err != nil {
			ctx.Logger().Error("v33 upgrade: alloy migration failed for account",
				"name", m.Name, "address", holder.String(), "amount", coin.String(), "error", err)
			result.Failed++
			continue
		}
		result.Migrated++
		result.Total = result.Total.Add(coin.Amount)
	}
	return result
}

// checkTransmuter returns an error unless transmuter is a contract that can take denom 1:1 into
// alloyedDenom: both denoms are listed, their normalization factors are equal, the pool is active,
// and denom is not corrupted. Any failure means the caller must not change state.
func checkTransmuter(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter sdk.AccAddress, denom, alloyedDenom string) error {
	if !keepers.WasmKeeper.HasContractInfo(ctx, transmuter) {
		return fmt.Errorf("transmuter %s is not a contract", transmuter)
	}
	denomFactor, alloyedFactor, err := transmuterNormalization(ctx, keepers, transmuter, denom, alloyedDenom)
	if err != nil {
		return err
	}
	if !denomFactor.Equal(alloyedFactor) {
		return fmt.Errorf("normalization factor of %s (%s) does not equal %s (%s)", denom, denomFactor, alloyedDenom, alloyedFactor)
	}
	if err := transmuterMustBeActive(ctx, keepers, transmuter); err != nil {
		return err
	}
	return denomMustNotBeCorrupted(ctx, keepers, transmuter, denom)
}

// transmuterNormalization reads list_asset_configs and returns the normalization factors of denom
// and alloyedDenom. Both must be listed.
func transmuterNormalization(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter sdk.AccAddress, denom, alloyedDenom string) (osmomath.Int, osmomath.Int, error) {
	var resp v3.ListAssetConfigsResponse
	if err := querySmart(ctx, keepers, transmuter, `{"list_asset_configs": {}}`, &resp); err != nil {
		return osmomath.Int{}, osmomath.Int{}, err
	}
	if len(resp.AssetConfigs) > maxListedAssetConfigs {
		return osmomath.Int{}, osmomath.Int{}, fmt.Errorf("transmuter %s returned %d asset configs", transmuter, len(resp.AssetConfigs))
	}
	denomFactor, ok, err := normalizationFactor(resp.AssetConfigs, denom)
	if err != nil || !ok {
		if err != nil {
			return osmomath.Int{}, osmomath.Int{}, err
		}
		return osmomath.Int{}, osmomath.Int{}, fmt.Errorf("transmuter %s does not list %s", transmuter, denom)
	}
	alloyedFactor, ok, err := normalizationFactor(resp.AssetConfigs, alloyedDenom)
	if err != nil || !ok {
		if err != nil {
			return osmomath.Int{}, osmomath.Int{}, err
		}
		return osmomath.Int{}, osmomath.Int{}, fmt.Errorf("transmuter %s does not list %s", transmuter, alloyedDenom)
	}
	return denomFactor, alloyedFactor, nil
}

// normalizationFactor returns the listed factor for denom. ok is false when denom is absent.
func normalizationFactor(configs []v3.AssetConfig, denom string) (osmomath.Int, bool, error) {
	for _, config := range configs {
		if config.Denom != denom {
			continue
		}
		factor, ok := osmomath.NewIntFromString(config.NormalizationFactor)
		if !ok || !factor.IsPositive() {
			return osmomath.Int{}, false, fmt.Errorf("normalization factor of %s is not a positive integer: %q", denom, config.NormalizationFactor)
		}
		return factor, true, nil
	}
	return osmomath.Int{}, false, nil
}

// transmuterMustBeActive returns an error unless the transmuter reports is_active.
func transmuterMustBeActive(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter sdk.AccAddress) error {
	var resp struct {
		IsActive bool `json:"is_active"`
	}
	if err := querySmart(ctx, keepers, transmuter, `{"is_active": {}}`, &resp); err != nil {
		return err
	}
	if !resp.IsActive {
		return fmt.Errorf("transmuter %s is not active", transmuter)
	}
	return nil
}

// denomMustNotBeCorrupted returns an error when denom is in get_corrupted_denoms.
func denomMustNotBeCorrupted(ctx sdk.Context, keepers *keepers.AppKeepers, transmuter sdk.AccAddress, denom string) error {
	var resp struct {
		CorruptedDenoms []string `json:"corrupted_denoms"`
	}
	if err := querySmart(ctx, keepers, transmuter, `{"get_corrupted_denoms": {}}`, &resp); err != nil {
		return err
	}
	if len(resp.CorruptedDenoms) > maxCorruptedDenoms {
		return fmt.Errorf("transmuter %s returned %d corrupted denoms", transmuter, len(resp.CorruptedDenoms))
	}
	for _, corrupted := range resp.CorruptedDenoms {
		if corrupted == denom {
			return fmt.Errorf("transmuter %s lists %s as corrupted", transmuter, denom)
		}
	}
	return nil
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
// denom. It returns an error if there are more than maxHolders.
func collectHolders(ctx sdk.Context, keepers *keepers.AppKeepers, denom string, unindexed []sdk.AccAddress, maxHolders int) ([]sdk.AccAddress, error) {
	holders := make([]sdk.AccAddress, 0, holderCapHint(maxHolders))
	seen := make(map[string]struct{}, holderCapHint(maxHolders))
	add := func(addr sdk.AccAddress) error {
		if _, ok := seen[string(addr)]; ok {
			return nil
		}
		if len(holders) == maxHolders {
			return fmt.Errorf("more than %d holders of %s", maxHolders, denom)
		}
		seen[string(addr)] = struct{}{}
		holders = append(holders, addr)
		return nil
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
		if err := add(pk.K1()); err != nil {
			return nil, err
		}
	}

	for _, addr := range unindexed {
		if keepers.BankKeeper.GetBalance(ctx, addr, denom).IsZero() {
			continue
		}
		if err := add(addr); err != nil {
			return nil, err
		}
	}
	return holders, nil
}

// holderCapHint is the initial capacity for a holder collection, capped so a large MaxHolders does
// not allocate up front.
func holderCapHint(maxHolders int) int {
	if maxHolders < 2048 {
		return maxHolders
	}
	return 2048
}

// isUserAccount reports whether addr is an account controlled by a key: an existing account with a
// public key that is neither a module account nor a wasm contract.
//
// Accounts without a public key are skipped. That includes protocol accounts (pools, IBC escrows,
// CL spread-reward and incentive addresses, ICA host accounts, contracts) and accounts that have
// never sent a transaction. Including pubkey-less 20-byte accounts would require excluding every
// IBC escrow address. That exclusion list is an open decision, so those accounts stay skipped.
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
	denomBefore := keepers.BankKeeper.GetBalance(ctx, holder, coin.Denom).Amount
	alloyedBefore := keepers.BankKeeper.GetBalance(ctx, holder, alloyedDenom).Amount

	if _, err := keepers.ContractKeeper.Execute(ctx, transmuter, holder, []byte(`{"join_pool": {}}`), sdk.NewCoins(coin)); err != nil {
		return err
	}

	spent := denomBefore.Sub(keepers.BankKeeper.GetBalance(ctx, holder, coin.Denom).Amount)
	if !spent.Equal(coin.Amount) {
		return fmt.Errorf("spent %s %s, expected %s", spent, coin.Denom, coin.Amount)
	}
	received := keepers.BankKeeper.GetBalance(ctx, holder, alloyedDenom).Amount.Sub(alloyedBefore)
	if !received.Equal(coin.Amount) {
		return fmt.Errorf("received %s %s, expected %s", received, alloyedDenom, coin.Amount)
	}
	return nil
}

// parseAddresses parses one bech32 address per line, ignoring blank lines and lines starting with #.
// It returns an error if there are more than maxAddresses.
func parseAddresses(list string, maxAddresses int) ([]sdk.AccAddress, error) {
	addrs := make([]sdk.AccAddress, 0, holderCapHint(maxAddresses))
	for _, line := range strings.Split(list, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(addrs) == maxAddresses {
			return nil, errors.New("too many addresses")
		}
		addr, err := sdk.AccAddressFromBech32(line)
		if err != nil {
			return nil, fmt.Errorf("address %q: %w", line, err)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}
