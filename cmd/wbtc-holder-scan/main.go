// Command wbtc-holder-scan lists every holder of native WBTC on an Osmosis node's state, read
// directly from the node's application database. It exists because the bank denom index on
// osmosis-1 is incomplete (it was never backfilled, see #6923), so the DenomOwners query and the
// v33 upgrade's index-based holder list both miss accounts.
//
// The command only reads state. Run it against a STOPPED node, or better, a copy of one: opening
// the application database and wasm cache takes their locks.
//
// Usage:
//
//	go run ./cmd/wbtc-holder-scan --home ~/.osmosisd --out ./wbtc-scan
//
// It writes, into --out:
//   - summary.json: height, app hash, supply reconciliation, per-category totals
//   - holders.csv: every holder with its account facts and classification
//   - unindexed.txt: holders missing from the bank denom index ("address amount" per line)
//
// It exits non-zero if the holders do not sum to the denom's total supply.
package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"

	osmosis "github.com/osmosis-labs/osmosis/v32/app"
)

const (
	defaultDenom      = "factory/osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743/wbtc"
	defaultIssuer     = "osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743"
	defaultTransmuter = "osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	home := flag.String("home", "", "node home directory containing data/application.db (required)")
	out := flag.String("out", "wbtc-scan", "output directory")
	backend := flag.String("db-backend", "", "goleveldb or pebbledb; default: app-db-backend from app.toml, else goleveldb")
	height := flag.Int64("height", 0, "state version to read; 0 reads the latest committed height")
	chainID := flag.String("chain-id", "osmosis-1", "chain ID set on the read context")
	cfg := Config{}
	flag.StringVar(&cfg.Denom, "denom", defaultDenom, "denom to scan")
	flag.StringVar(&cfg.Issuer, "issuer", defaultIssuer, "WBTC controller contract (merchants and custodian)")
	flag.StringVar(&cfg.Transmuter, "transmuter", defaultTransmuter, "allBTC alloy transmuter contract")
	flag.Parse()
	if *home == "" {
		return errors.New("--home is required")
	}

	st, err := openState(*home, *backend, *height, *chainID)
	if err != nil {
		return err
	}
	defer st.close()

	fmt.Fprintf(os.Stderr, "scanning %s at height %d\n", cfg.Denom, st.ctx.BlockHeight())
	report, err := Scan(st.ctx, &st.app.AppKeepers, cfg, os.Stderr)
	if err != nil {
		return err
	}
	report.AppHash = st.appHash
	if err := writeOutputs(*out, report); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "holders=%d indexed=%d unindexed=%d (%s) would_migrate=%d (%s) supply_matches=%v\n",
		report.Holders, report.IndexedHolders, report.UnindexedHolders, report.UnindexedAmount,
		report.WouldMigrate.Accounts, report.WouldMigrate.Amount, report.SupplyMatches)
	if !report.SupplyMatches {
		return fmt.Errorf("holders sum to %s but supply is %s", report.HoldersTotal, report.Supply)
	}
	return nil
}

// state is a read-only view of one committed version of the node state.
type state struct {
	ctx     sdk.Context
	app     *osmosis.OsmosisApp
	appHash string
	close   func()
}

// openState loads the app from the node's database and returns a context over a cache of the
// requested version. The cache is never written back, so the database is not modified. Node
// options from app.toml are deliberately not loaded, so no SQS, indexer or telemetry service
// starts; only the database backend is read from it.
func openState(home, backend string, height int64, chainID string) (*state, error) {
	opts := viper.New()
	opts.Set(flags.FlagHome, home)
	if backend == "" {
		backend = appTomlBackend(home)
	}
	opts.Set("app-db-backend", backend)

	// dbm.NewDB creates a database that does not exist, which would make a wrong --home look like
	// a chain with no holders.
	if _, err := os.Stat(filepath.Join(home, "data", "application.db")); err != nil {
		return nil, fmt.Errorf("no application database under %s: %w", home, err)
	}

	db, err := dbm.NewDB("application", server.GetAppDBBackend(opts), filepath.Join(home, "data"))
	if err != nil {
		return nil, fmt.Errorf("opening application db: %w", err)
	}
	app := osmosis.NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, home, 0, opts, osmosis.EmptyWasmOpts)
	closeApp := func() { _ = app.Close() }

	if app.LastBlockHeight() == 0 {
		closeApp()
		return nil, errors.New("application database has no committed state")
	}
	cms := app.CommitMultiStore()
	if height == 0 {
		height = app.LastBlockHeight()
	}
	var ms storetypes.CacheMultiStore
	if height == app.LastBlockHeight() {
		ms = cms.CacheMultiStore()
	} else if ms, err = cms.CacheMultiStoreWithVersion(height); err != nil {
		closeApp()
		return nil, fmt.Errorf("loading version %d: %w", height, err)
	}
	appHash := ""
	if height == app.LastBlockHeight() {
		appHash = strings.ToUpper(hex.EncodeToString(cms.LastCommitID().Hash))
	}

	ctx := sdk.NewContext(ms, cmtproto.Header{Height: height, ChainID: chainID}, false, log.NewNopLogger())
	return &state{ctx: ctx, app: app, appHash: appHash, close: closeApp}, nil
}

// appTomlBackend returns app-db-backend from the node's app.toml, or goleveldb if unset.
func appTomlBackend(home string) string {
	v := viper.New()
	v.SetConfigFile(filepath.Join(home, "config", "app.toml"))
	if err := v.ReadInConfig(); err == nil && v.GetString("app-db-backend") != "" {
		return v.GetString("app-db-backend")
	}
	return string(dbm.GoLevelDBBackend)
}

func writeOutputs(dir string, r *Report) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	summary, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), append(summary, '\n'), 0o644); err != nil {
		return err
	}
	if err := writeHoldersCSV(filepath.Join(dir, "holders.csv"), r.holders); err != nil {
		return err
	}
	var b strings.Builder
	for _, h := range r.holders {
		if !h.InDenomIndex {
			fmt.Fprintf(&b, "%s %s\n", h.Address, h.Amount)
		}
	}
	return os.WriteFile(filepath.Join(dir, "unindexed.txt"), []byte(b.String()), 0o644)
}

func writeHoldersCSV(path string, holders []Holder) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	header := []string{"address", "amount", "spendable", "in_denom_index", "category", "would_migrate",
		"account_type", "has_pubkey", "is_module", "is_contract", "is_ica", "is_vesting", "tags"}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, h := range holders {
		row := []string{h.Address, h.Amount.String(), h.Spendable.String(), fmt.Sprint(h.InDenomIndex), h.Category,
			fmt.Sprint(h.WouldMigrate), h.AccountType, fmt.Sprint(h.HasPubKey), fmt.Sprint(h.IsModule),
			fmt.Sprint(h.IsContract), fmt.Sprint(h.IsICA), fmt.Sprint(h.IsVesting), strings.Join(h.Tags, ";")}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
