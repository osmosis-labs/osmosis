// Command alloy-holder-scan prints the holders of one denom that the bank denom index does not
// list. The index on osmosis-1 was never backfilled, so accounts whose balance has not changed
// since are missing from it. The v33 alloy migration reads this list together with the index.
//
// It opens an osmosisd application database read-only in effect: state is loaded into a cache
// multistore that is never written back. Run it against a stopped node, or a copy of one. Opening
// the database takes its lock, and constructing the app may touch the wasm cache. It does not commit.
//
// Usage:
//
//	go run ./cmd/alloy-holder-scan --home ~/.osmosisd --denom ibc/498A0751...
//
// Stdout is the holder file: a header comment with the chain, height and app hash, then one bech32
// address per line. The scan fails instead of printing a partial list if the balances iteration,
// the holder set, or the supply check exceeds its bound or does not reconcile.
package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/log"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"

	osmosis "github.com/osmosis-labs/osmosis/v32/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	home := flag.String("home", "", "node home directory containing data/application.db (required)")
	denom := flag.String("denom", "", "denom to scan (required)")
	backend := flag.String("db-backend", "", "goleveldb or pebbledb; default: app-db-backend from app.toml, else goleveldb")
	height := flag.Int64("height", 0, "state version to read; 0 reads the latest committed height")
	chainID := flag.String("chain-id", "osmosis-1", "chain id written into the header comment")
	flag.Parse()
	if *home == "" {
		return errors.New("--home is required")
	}
	if *denom == "" {
		return errors.New("--denom is required")
	}

	st, err := openState(*home, *backend, *height, *chainID)
	if err != nil {
		return err
	}
	defer st.close()

	fmt.Fprintf(os.Stderr, "scanning %s at height %d\n", *denom, st.ctx.BlockHeight())
	result, err := Scan(st.ctx, &st.app.AppKeepers, ScanConfig{
		Denom:      *denom,
		MaxEntries: maxBalanceEntries,
		MaxHolders: maxDenomHolders,
		Progress:   os.Stderr,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "holders=%d unindexed=%d supply=%s\n", result.Holders, len(result.Unindexed), result.Supply)
	fmt.Print(formatUnindexedList(*chainID, *denom, st.ctx.BlockHeight(), st.appHash, result.Unindexed))
	return nil
}

// state is a cache over one committed version of the node. close releases the database.
type state struct {
	ctx     sdk.Context
	app     *osmosis.OsmosisApp
	appHash string
	close   func()
}

// openState loads the app from the node's database and returns a context over a cache of the
// requested version. The cache is never written back.
func openState(home, backend string, height int64, chainID string) (*state, error) {
	dbPath := filepath.Join(home, "data", "application.db")
	// dbm.NewDB creates a database that does not exist, which would make a wrong --home look empty.
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("no application database under %s: %w", home, err)
	}
	dbBackend := appDBBackend(home, backend)
	db, err := dbm.NewDB("application", dbBackend, filepath.Join(home, "data"))
	if err != nil {
		return nil, fmt.Errorf("opening application db: %w", err)
	}

	opts := viper.New()
	opts.Set(flags.FlagHome, home)
	opts.Set("app-db-backend", string(dbBackend))
	app := osmosis.NewOsmosisApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, home, 0, opts, osmosis.EmptyWasmOpts)
	closeApp := func() { _ = app.Close() }
	if app.LastBlockHeight() == 0 {
		closeApp()
		return nil, errors.New("application database has no committed state")
	}
	if height == 0 {
		height = app.LastBlockHeight()
	}

	ms, appHash, err := cacheAt(app, height)
	if err != nil {
		closeApp()
		return nil, err
	}
	ctx := sdk.NewContext(ms, cmtproto.Header{Height: height, ChainID: chainID}, false, log.NewNopLogger()).WithChainID(chainID)
	return &state{ctx: ctx, app: app, appHash: appHash, close: closeApp}, nil
}

// cacheAt returns a cache of the committed version and the app hash recorded for it.
func cacheAt(app *osmosis.OsmosisApp, height int64) (storetypes.CacheMultiStore, string, error) {
	cms := app.CommitMultiStore()
	rms, ok := cms.(*rootmulti.Store)
	if !ok {
		return nil, "", fmt.Errorf("commit multistore is %T", cms)
	}
	info, err := rms.GetCommitInfo(height)
	if err != nil {
		return nil, "", fmt.Errorf("commit info at height %d: %w", height, err)
	}
	appHash := strings.ToUpper(hex.EncodeToString(info.CommitID().Hash))

	if height == app.LastBlockHeight() {
		return cms.CacheMultiStore(), appHash, nil
	}
	ms, err := cms.CacheMultiStoreWithVersion(height)
	if err != nil {
		return nil, "", fmt.Errorf("loading version %d: %w", height, err)
	}
	return ms, appHash, nil
}

// appDBBackend returns the application database backend: the flag, else app.toml, else goleveldb.
func appDBBackend(home, backend string) dbm.BackendType {
	opts := viper.New()
	opts.Set(flags.FlagHome, home)
	if backend != "" {
		opts.Set("app-db-backend", backend)
		return server.GetAppDBBackend(opts)
	}
	cfg := viper.New()
	cfg.SetConfigFile(filepath.Join(home, "config", "app.toml"))
	if err := cfg.ReadInConfig(); err == nil && cfg.GetString("app-db-backend") != "" {
		opts.Set("app-db-backend", cfg.GetString("app-db-backend"))
	}
	return server.GetAppDBBackend(opts)
}
