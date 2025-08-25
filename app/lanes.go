package app

import (
	"context"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/hashicorp/go-metrics"

	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/v2/block"
	"github.com/skip-mev/block-sdk/v2/block/base"
	defaultlane "github.com/skip-mev/block-sdk/v2/lanes/base"
	mevlane "github.com/skip-mev/block-sdk/v2/lanes/mev"
)

const (
	maxTxPerTopOfBlockAuctionLane = 500  // this is the maximum # of bids that will be held in the app-side in-memory mempool
	maxTxPerDefaultLane           = 3000 // all other txs
)

var (
	defaultLaneBlockspacePercentage           = math.LegacyMustNewDecFromStr("0.90")
	topOfBlockAuctionLaneBlockspacePercentage = math.LegacyMustNewDecFromStr("0.10")
)

// CreateLanes walks through the process of creating the lanes for the block sdk. In this function
// we create three separate lanes - MEV, Free, and Default - and then return them.
func CreateLanes(app *OsmosisApp, txConfig client.TxConfig) (*mevlane.MEVLane, *base.BaseLane) {
	// Create the signer extractor. This is used to extract the expected signers from
	// a transaction. Each lane can have a different signer extractor if needed.
	signerAdapter := signerextraction.NewDefaultAdapter()

	// Create the configurations for each lane. These configurations determine how many
	// transactions the lane can store, the maximum block space the lane can consume, and
	// the signer extractor used to extract the expected signers from a transaction.

	// Create a mev configuration that accepts maxTxPerTopOfBlockAuctionLane transactions and consumes topOfBlockAuctionLaneBlockspacePercentage of the
	// block space.
	mevConfig := base.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       txConfig.TxEncoder(),
		TxDecoder:       txConfig.TxDecoder(),
		MaxBlockSpace:   topOfBlockAuctionLaneBlockspacePercentage,
		SignerExtractor: signerAdapter,
		MaxTxs:          maxTxPerTopOfBlockAuctionLane,
	}

	// Create a default configuration that accepts maxTxPerDefaultLane transactions and consumes defaultLaneBlockspacePercentage of the
	// block space.
	defaultConfig := base.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       txConfig.TxEncoder(),
		TxDecoder:       txConfig.TxDecoder(),
		MaxBlockSpace:   defaultLaneBlockspacePercentage,
		SignerExtractor: signerAdapter,
		MaxTxs:          maxTxPerDefaultLane,
	}

	// Create the match handlers for each lane. These match handlers determine whether or not
	// a transaction belongs in the lane.

	// Create the final match handler for the mev lane.
	factory := mevlane.NewDefaultAuctionFactory(txConfig.TxDecoder(), signerAdapter)
	mevMatchHandler := factory.MatchHandler()

	// Create the final match handler for the default lane. I.e this will direct all txs that are
	// not free nor mev to this lane
	defaultMatchHandler := base.DefaultMatchHandler()

	// Create the lanes.
	mevLane := mevlane.NewMEVLane(
		mevConfig,
		factory,
		mevMatchHandler,
	)

	defaultMempool := base.NewMempoolWithDefaultOrdering(
		base.DefaultTxPriority(),
		defaultConfig.SignerExtractor,
		defaultConfig.MaxTxs,
	)
	defaultLane := defaultlane.NewDefaultLaneWithMempool(
		defaultConfig,
		defaultMatchHandler,
		defaultMempool,
	)

	return mevLane, defaultLane
}

var _ block.Mempool = (*LanedMempoolWithTelemetry)(nil)

const AppMempoolLabel = "app_mempool"

type LanedMempoolWithTelemetry struct {
	lanedMempool block.LanedMempool
}

// NewLanedMempoolWithTelemetry creates a new telemetry-enabled mempool wrapper
func NewLanedMempoolWithTelemetry(lanedMempool block.LanedMempool) *LanedMempoolWithTelemetry {
	return &LanedMempoolWithTelemetry{
		lanedMempool: lanedMempool,
	}
}

// Insert attempts to insert a Tx into the app-side mempool returning
// an error upon failure.
func (m *LanedMempoolWithTelemetry) Insert(ctx context.Context, tx sdk.Tx) error {
	m.emitTxDistributionMetric()
	err := m.lanedMempool.Insert(ctx, tx)
	if err != nil {
		telemetry.IncrCounter(1, AppMempoolLabel, "insert_error")
	}
	return err
}

// Select returns an Iterator over the app-side mempool. If txs are specified,
// then they shall be incorporated into the Iterator. The Iterator is not thread-safe to use.
func (m *LanedMempoolWithTelemetry) Select(ctx context.Context, txs [][]byte) mempool.Iterator {
	return m.lanedMempool.Select(ctx, txs)
}

// CountTx returns the number of transactions currently in the mempool.
func (m *LanedMempoolWithTelemetry) CountTx() int {
	return m.lanedMempool.CountTx()
}

// Remove attempts to remove a transaction from the mempool, returning an error
// upon failure.
func (m *LanedMempoolWithTelemetry) Remove(tx sdk.Tx) error {
	m.emitTxDistributionMetric()
	err := m.lanedMempool.Remove(tx)
	if err != nil {
		telemetry.IncrCounter(1, AppMempoolLabel, "remove_error")
	}
	return err
}

// Registry returns the mempool's lane registry.
func (m *LanedMempoolWithTelemetry) Registry() []block.Lane {
	return m.lanedMempool.Registry()
}

// Contains returns the any of the lanes currently contain the transaction.
func (m *LanedMempoolWithTelemetry) Contains(tx sdk.Tx) bool {
	return m.lanedMempool.Contains(tx)
}

// GetTxDistribution returns the number of transactions in each lane.
func (m *LanedMempoolWithTelemetry) GetTxDistribution() map[string]uint64 {
	return m.lanedMempool.GetTxDistribution()
}

func (m *LanedMempoolWithTelemetry) emitTxDistributionMetric() {
	txDistribution := m.GetTxDistribution()
	for lane, count := range txDistribution {
		telemetry.SetGaugeWithLabels([]string{AppMempoolLabel, "tx_count_by_lane"}, float32(count), []metrics.Label{
			{
				Name:  "lane_name",
				Value: lane,
			},
		})
	}
}
