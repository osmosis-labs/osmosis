package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/keepers"
)

const (
	// maxBalanceEntries bounds the balances-store scan. A hit fails the scan. osmosis-1 is far
	// below this. Raise it only after checking that a real export exceeds it.
	maxBalanceEntries = 500_000_000
	// maxDenomHolders bounds how many positive balances of the denom are kept. A hit fails the
	// scan so the printed list is never truncated.
	maxDenomHolders = 1_000_000
	// progressEvery is how often a progress line is written, in balance entries scanned.
	progressEvery = 5_000_000
)

// ScanConfig selects the denom and the scan bounds.
type ScanConfig struct {
	Denom      string
	MaxEntries uint64
	MaxHolders int
	Progress   io.Writer
}

// ScanResult is the unindexed holders of one denom at one version.
type ScanResult struct {
	Holders   int
	Supply    osmomath.Int
	Unindexed []sdk.AccAddress
}

// Scan walks the bank balances store, keeps positive balances of cfg.Denom, and returns the
// holders that the denom index does not list. It returns an error, and no partial list, if a bound
// is hit or the holders do not sum to the denom's supply.
func Scan(ctx sdk.Context, k *keepers.AppKeepers, cfg ScanConfig) (*ScanResult, error) {
	if cfg.Denom == "" {
		return nil, fmt.Errorf("denom is empty")
	}
	if cfg.MaxEntries == 0 || cfg.MaxHolders <= 0 {
		return nil, fmt.Errorf("scan bounds must be positive")
	}
	if cfg.Progress == nil {
		cfg.Progress = io.Discard
	}

	holders, sum, err := scanBalances(ctx, k, cfg)
	if err != nil {
		return nil, err
	}
	indexed, err := indexedHolders(ctx, k, cfg)
	if err != nil {
		return nil, err
	}
	supply := k.BankKeeper.GetSupply(ctx, cfg.Denom).Amount
	if !sum.Equal(supply) {
		return nil, fmt.Errorf("holders of %s sum to %s but supply is %s", cfg.Denom, sum, supply)
	}

	unindexed := make([]sdk.AccAddress, 0)
	for addr := range holders {
		if _, ok := indexed[addr]; ok {
			continue
		}
		unindexed = append(unindexed, sdk.AccAddress(addr))
	}
	sort.Slice(unindexed, func(i, j int) bool {
		return unindexed[i].String() < unindexed[j].String()
	})
	return &ScanResult{Holders: len(holders), Supply: supply, Unindexed: unindexed}, nil
}

// scanBalances returns the positive balances of cfg.Denom, keyed by the raw account bytes, and
// their sum. Iteration stops with an error at cfg.MaxEntries or cfg.MaxHolders.
func scanBalances(ctx sdk.Context, k *keepers.AppKeepers, cfg ScanConfig) (map[string]osmomath.Int, osmomath.Int, error) {
	it, err := k.BankKeeper.Balances.Iterate(ctx, nil)
	if err != nil {
		return nil, osmomath.Int{}, err
	}
	defer it.Close()

	holders := make(map[string]osmomath.Int)
	sum := osmomath.ZeroInt()
	var n uint64
	for ; it.Valid(); it.Next() {
		n++
		if n > cfg.MaxEntries {
			return nil, osmomath.Int{}, fmt.Errorf("more than %d balance entries", cfg.MaxEntries)
		}
		if n%progressEvery == 0 {
			fmt.Fprintf(cfg.Progress, "scanned %d balance entries, %d holders so far\n", n, len(holders))
		}
		key, err := it.Key()
		if err != nil {
			return nil, osmomath.Int{}, err
		}
		if key.K2() != cfg.Denom {
			continue
		}
		amount, err := it.Value()
		if err != nil {
			return nil, osmomath.Int{}, err
		}
		if !amount.IsPositive() {
			continue
		}
		if len(holders) == cfg.MaxHolders {
			return nil, osmomath.Int{}, fmt.Errorf("more than %d holders of %s", cfg.MaxHolders, cfg.Denom)
		}
		holders[string(key.K1())] = amount
		sum = sum.Add(amount)
	}
	return holders, sum, nil
}

// indexedHolders returns the addresses listed under the denom in the bank denom index.
func indexedHolders(ctx sdk.Context, k *keepers.AppKeepers, cfg ScanConfig) (map[string]struct{}, error) {
	it, err := k.BankKeeper.Balances.Indexes.Denom.MatchExact(ctx, cfg.Denom)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	out := make(map[string]struct{})
	n := 0
	for ; it.Valid(); it.Next() {
		n++
		if n > cfg.MaxHolders {
			return nil, fmt.Errorf("more than %d indexed holders of %s", cfg.MaxHolders, cfg.Denom)
		}
		pk, err := it.PrimaryKey()
		if err != nil {
			return nil, err
		}
		out[string(pk.K1())] = struct{}{}
	}
	return out, nil
}

// formatUnindexedList is the holder file: a header comment, then one bech32 address per line.
func formatUnindexedList(chainID, denom string, height int64, appHash string, addrs []sdk.AccAddress) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s holders missing from the bank denom index on %s at height %d\n", denom, chainID, height)
	fmt.Fprintf(&b, "# (app hash %s), found by a full scan of the balances store with cmd/alloy-holder-scan.\n", appHash)
	b.WriteString("# One address per line.\n")
	for _, addr := range addrs {
		b.WriteString(addr.String())
		b.WriteByte('\n')
	}
	return b.String()
}
