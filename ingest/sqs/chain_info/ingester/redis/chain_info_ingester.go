package redis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
)

// chainInfoIngester is an ingester for blockchain information.
// It implements ingest.Ingester.
// It reads the latest blockchain height and writes it to the chainInfo repository.
type chainInfoIngester struct {
	chainInfoRepo     mvc.ChainInfoRepository
	repositoryManager mvc.TxManager
	logger            log.Logger
}

// NewChainInfoIngester returns a new chain information ingester.
func NewChainInfoIngester(chainInfoRepo mvc.ChainInfoRepository, repositoryManager mvc.TxManager) mvc.AtomicIngester {
	return &chainInfoIngester{
		chainInfoRepo:     chainInfoRepo,
		repositoryManager: repositoryManager,
	}
}

// ProcessBlock implements ingest.Ingester.
// It reads the latest blockchain height and stores it in Redis.
func (ci *chainInfoIngester) ProcessBlock(ctx sdk.Context, tx mvc.Tx) error {
	height := ctx.BlockHeight()

	ci.logger.Info("ingesting latest blockchain height", zap.Int64("height", height))

	err := ci.chainInfoRepo.StoreLatestHeight(sdk.WrapSDKContext(ctx), tx, uint64(height))
	if err != nil {
		ci.logger.Error("failed to ingest latest blockchain height", zap.Error(err))
		return err
	}

	return nil
}

// SetLogger implements ingest.AtomicIngester.
func (ci *chainInfoIngester) SetLogger(logger log.Logger) {
	ci.logger = logger
}

var _ mvc.AtomicIngester = &chainInfoIngester{}
