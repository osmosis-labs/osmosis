package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
	cltypes "github.com/osmosis-labs/osmosis/v32/x/concentrated-liquidity/types"
)

// maxIssuerMembers bounds the merchant list read from the WBTC controller. The query is not
// paginated here, so a full page means the list may be truncated and the scan fails instead.
const maxIssuerMembers = 100

// progressEvery is how often (in scanned balance entries) a progress line is written.
const progressEvery = 5_000_000

// Config selects the denom to scan and the contracts that define its special accounts.
type Config struct {
	Denom      string
	Issuer     string
	Transmuter string
}

// Holder is one account holding Config.Denom, found by a full scan of the bank balances store.
type Holder struct {
	Address      string       `json:"address"`
	Amount       osmomath.Int `json:"amount"`
	Spendable    osmomath.Int `json:"spendable"`
	InDenomIndex bool         `json:"in_denom_index"`
	AccountType  string       `json:"account_type"`
	HasPubKey    bool         `json:"has_pubkey"`
	IsModule     bool         `json:"is_module"`
	IsContract   bool         `json:"is_contract"`
	IsICA        bool         `json:"is_ica"`
	IsVesting    bool         `json:"is_vesting"`
	Tags         []string     `json:"tags"`
	Category     string       `json:"category"`
	WouldMigrate bool         `json:"would_migrate"`
}

// CategoryTotal sums the holders of one category.
type CategoryTotal struct {
	Accounts int          `json:"accounts"`
	Amount   osmomath.Int `json:"amount"`
}

// Report is the result of a scan.
type Report struct {
	ChainID           string                   `json:"chain_id"`
	Height            int64                    `json:"height"`
	AppHash           string                   `json:"app_hash"`
	Denom             string                   `json:"denom"`
	Supply            osmomath.Int             `json:"supply"`
	HoldersTotal      osmomath.Int             `json:"holders_total"`
	SupplyMatches     bool                     `json:"supply_matches"`
	BalanceEntries    uint64                   `json:"balance_entries_scanned"`
	Holders           int                      `json:"holders"`
	IndexedHolders    int                      `json:"indexed_holders"`
	UnindexedHolders  int                      `json:"unindexed_holders"`
	UnindexedAmount   osmomath.Int             `json:"unindexed_amount"`
	StaleIndexEntries []string                 `json:"stale_index_entries"`
	WouldMigrate      CategoryTotal            `json:"would_migrate"`
	Categories        map[string]CategoryTotal `json:"categories"`
	Merchants         []string                 `json:"issuer_merchants"`
	Custodian         string                   `json:"issuer_custodian"`
	holders           []Holder
}

// Scan reads every bank balance entry, keeps the holders of cfg.Denom, compares them with the
// bank denom index, and classifies each holder. It only reads state.
func Scan(ctx sdk.Context, k *keepers.AppKeepers, cfg Config, progress io.Writer) (*Report, error) {
	balances, entries, err := scanBalances(ctx, k, cfg.Denom, progress)
	if err != nil {
		return nil, fmt.Errorf("scanning balances: %w", err)
	}
	indexed, err := indexedHolders(ctx, k, cfg.Denom)
	if err != nil {
		return nil, fmt.Errorf("reading denom index: %w", err)
	}
	tags, merchants, custodian, err := buildTags(ctx, k, cfg)
	if err != nil {
		return nil, err
	}

	r := &Report{
		ChainID: ctx.ChainID(), Height: ctx.BlockHeight(), Denom: cfg.Denom,
		Supply: k.BankKeeper.GetSupply(ctx, cfg.Denom).Amount, HoldersTotal: osmomath.ZeroInt(),
		BalanceEntries: entries, UnindexedAmount: osmomath.ZeroInt(),
		WouldMigrate: CategoryTotal{Amount: osmomath.ZeroInt()}, Categories: map[string]CategoryTotal{},
		Merchants: merchants, Custodian: custodian, StaleIndexEntries: []string{},
	}
	for addr, amount := range balances {
		h := classify(ctx, k, addr, amount, cfg.Denom, tags[addr])
		h.InDenomIndex = indexed[addr]
		r.add(h)
	}
	for addr := range indexed {
		if _, ok := balances[addr]; !ok {
			r.StaleIndexEntries = append(r.StaleIndexEntries, addr)
		}
	}
	sort.Strings(r.StaleIndexEntries)
	sort.Slice(r.holders, func(i, j int) bool {
		if !r.holders[i].Amount.Equal(r.holders[j].Amount) {
			return r.holders[i].Amount.GT(r.holders[j].Amount)
		}
		return r.holders[i].Address < r.holders[j].Address
	})
	r.SupplyMatches = r.HoldersTotal.Equal(r.Supply)
	return r, nil
}

func (r *Report) add(h Holder) {
	r.holders = append(r.holders, h)
	r.Holders++
	r.HoldersTotal = r.HoldersTotal.Add(h.Amount)
	if h.InDenomIndex {
		r.IndexedHolders++
	} else {
		r.UnindexedHolders++
		r.UnindexedAmount = r.UnindexedAmount.Add(h.Amount)
	}
	c, ok := r.Categories[h.Category]
	if !ok {
		c.Amount = osmomath.ZeroInt()
	}
	c.Accounts++
	c.Amount = c.Amount.Add(h.Amount)
	r.Categories[h.Category] = c
	if h.WouldMigrate {
		r.WouldMigrate.Accounts++
		r.WouldMigrate.Amount = r.WouldMigrate.Amount.Add(h.Spendable)
	}
}

// scanBalances iterates the primary balances store. The denom index cannot be used here: on
// osmosis-1 it was never backfilled (#6923), so balances untouched since are missing from it.
func scanBalances(ctx sdk.Context, k *keepers.AppKeepers, denom string, progress io.Writer) (map[string]osmomath.Int, uint64, error) {
	it, err := k.BankKeeper.Balances.Iterate(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer it.Close()

	out := make(map[string]osmomath.Int, 4096)
	var n uint64
	for ; it.Valid(); it.Next() {
		n++
		if n%progressEvery == 0 {
			fmt.Fprintf(progress, "scanned %d balance entries, %d holders so far\n", n, len(out))
		}
		key, err := it.Key()
		if err != nil {
			return nil, n, err
		}
		if key.K2() != denom {
			continue
		}
		amount, err := it.Value()
		if err != nil {
			return nil, n, err
		}
		if amount.IsPositive() {
			out[key.K1().String()] = amount
		}
	}
	return out, n, nil
}

// indexedHolders returns the addresses listed under denom in the bank denom index.
func indexedHolders(ctx sdk.Context, k *keepers.AppKeepers, denom string) (map[string]bool, error) {
	it, err := k.BankKeeper.Balances.Indexes.Denom.MatchExact(ctx, denom)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	out := make(map[string]bool, 4096)
	for ; it.Valid(); it.Next() {
		pk, err := it.PrimaryKey()
		if err != nil {
			return nil, err
		}
		out[pk.K1().String()] = true
	}
	return out, nil
}

// buildTags labels the protocol-owned addresses that are plain accounts and so cannot be told
// apart by account type: IBC transfer escrows and CL pool sub-addresses. It also tags the WBTC
// issuer's merchants and custodian, and the transmuter.
func buildTags(ctx sdk.Context, k *keepers.AppKeepers, cfg Config) (map[string][]string, []string, string, error) {
	tags := map[string][]string{}
	tag := func(addr sdk.AccAddress, t string) { tags[addr.String()] = append(tags[addr.String()], t) }

	for _, ch := range k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, transfertypes.PortID) {
		if ch.PortId == transfertypes.PortID {
			tag(transfertypes.GetEscrowAddress(ch.PortId, ch.ChannelId), "ibc_escrow:"+ch.ChannelId)
		}
	}
	pools, err := k.ConcentratedLiquidityKeeper.GetPools(ctx)
	if err != nil {
		return nil, nil, "", fmt.Errorf("listing CL pools: %w", err)
	}
	for _, p := range pools {
		cl, ok := p.(cltypes.ConcentratedPoolExtension)
		if !ok {
			return nil, nil, "", fmt.Errorf("pool %d is not a CL pool", p.GetId())
		}
		id := fmt.Sprint(cl.GetId())
		tag(cl.GetAddress(), "cl_pool:"+id)
		tag(cl.GetSpreadRewardsAddress(), "cl_spread_rewards:"+id)
		tag(cl.GetIncentivesAddress(), "cl_incentives:"+id)
	}

	// An empty issuer skips the merchant and custodian lookup; only tests do this.
	var merchants []string
	var custodian string
	if cfg.Issuer != "" {
		var err error
		if merchants, custodian, err = issuerMembers(ctx, k, cfg.Issuer); err != nil {
			return nil, nil, "", err
		}
		for _, m := range merchants {
			tags[m] = append(tags[m], "wbtc_merchant")
		}
		tags[custodian] = append(tags[custodian], "wbtc_custodian")
	}
	tags[cfg.Transmuter] = append(tags[cfg.Transmuter], "allbtc_transmuter")
	return tags, merchants, custodian, nil
}

// issuerMembers reads the merchants and custodian from the WBTC controller contract.
func issuerMembers(ctx sdk.Context, k *keepers.AppKeepers, issuer string) ([]string, string, error) {
	addr, err := sdk.AccAddressFromBech32(issuer)
	if err != nil {
		return nil, "", fmt.Errorf("issuer address: %w", err)
	}
	var merchants struct {
		Merchants []string `json:"merchants"`
	}
	q := fmt.Sprintf(`{"list_merchants":{"limit":%d}}`, maxIssuerMembers)
	if err := querySmart(ctx, k, addr, q, &merchants); err != nil {
		return nil, "", err
	}
	if len(merchants.Merchants) >= maxIssuerMembers {
		return nil, "", errors.New("issuer merchant list may be truncated; add pagination")
	}
	var custodian struct {
		Address string `json:"address"`
	}
	if err := querySmart(ctx, k, addr, `{"get_custodian":{}}`, &custodian); err != nil {
		return nil, "", err
	}
	if custodian.Address == "" {
		return nil, "", errors.New("issuer returned an empty custodian")
	}
	return merchants.Merchants, custodian.Address, nil
}

func querySmart(ctx sdk.Context, k *keepers.AppKeepers, contract sdk.AccAddress, msg string, out any) error {
	bz, err := k.WasmKeeper.QuerySmart(ctx, contract, []byte(msg))
	if err != nil {
		return fmt.Errorf("query %s on %s: %w", msg, contract, err)
	}
	if err := json.Unmarshal(bz, out); err != nil {
		return fmt.Errorf("decode %s response: %w", msg, err)
	}
	return nil
}

// classify fills in the account facts for one holder and decides whether the v33 migration, with
// the pubkey filter and issuer exclusions, would move its balance.
func classify(ctx sdk.Context, k *keepers.AppKeepers, addr string, amount osmomath.Int, denom string, tags []string) Holder {
	acc := sdk.MustAccAddressFromBech32(addr)
	h := Holder{Address: addr, Amount: amount, Spendable: osmomath.ZeroInt(), Tags: tags, AccountType: "none"}
	if tags == nil {
		h.Tags = []string{}
	}
	h.IsContract = k.WasmKeeper.HasContractInfo(ctx, acc)
	if a := k.AccountKeeper.GetAccount(ctx, acc); a != nil {
		h.AccountType = fmt.Sprintf("%T", a)
		h.HasPubKey = a.GetPubKey() != nil
		_, h.IsModule = a.(sdk.ModuleAccountI)
		_, h.IsICA = a.(*icatypes.InterchainAccount)
		_, h.IsVesting = a.(vestingexported.VestingAccount)
		h.Spendable = k.BankKeeper.SpendableCoin(ctx, acc, denom).Amount
	}
	h.Category = category(h)
	h.WouldMigrate = h.Category == "user_pubkey" && h.Spendable.IsPositive()
	return h
}

// category assigns exactly one bucket per holder; the first match wins.
func category(h Holder) string {
	for _, t := range h.Tags {
		switch {
		case t == "allbtc_transmuter":
			return "allbtc_transmuter"
		case t == "wbtc_merchant", t == "wbtc_custodian":
			return "wbtc_issuer_member"
		}
	}
	switch {
	case h.IsModule:
		return "module_account"
	case h.IsContract:
		return "contract"
	case h.IsICA:
		return "ica_host_account"
	case hasTagPrefix(h.Tags, "ibc_escrow:"):
		return "ibc_escrow"
	case hasTagPrefix(h.Tags, "cl_spread_rewards:"):
		return "cl_spread_rewards"
	case hasTagPrefix(h.Tags, "cl_incentives:"):
		return "cl_incentives"
	case h.AccountType == "none":
		return "no_account"
	case h.HasPubKey:
		return "user_pubkey"
	default:
		return "no_pubkey_other"
	}
}

func hasTagPrefix(tags []string, prefix string) bool {
	for _, t := range tags {
		if len(t) >= len(prefix) && t[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
