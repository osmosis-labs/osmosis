package redis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
)

// chainInfoIngester is an ingester for blockchain information.
// It implements ingest.Ingester.
// It reads the latest blockchain height and writes it to the chainInfo repository.
type chainInfoIngester struct {
	chainInfoRepo     domain.ChainInfoRepository
	repositoryManager domain.TxManager
}

// NewChainInfoIngester returns a new chain information ingester.
func NewChainInfoIngester(chainInfoRepo domain.ChainInfoRepository, repositoryManager domain.TxManager) domain.AtomicIngester {
	return &chainInfoIngester{
		chainInfoRepo:     chainInfoRepo,
		repositoryManager: repositoryManager,
	}
}

// ProcessBlock implements ingest.Ingester.
// It reads the latest blockchain height and stores it in Redis.
func (ci *chainInfoIngester) ProcessBlock(ctx sdk.Context, tx domain.Tx) error {
	height := ctx.BlockHeight()

	ctx.Logger().Info("ingesting latest blockchain height", "height", height)

	err := ci.chainInfoRepo.StoreLatestHeight(sdk.WrapSDKContext(ctx), tx, uint64(height))
	if err != nil {
		ctx.Logger().Error("failed to ingest latest blockchain height", "error", err)
		return err
	}

	return nil
}

var _ domain.AtomicIngester = &chainInfoIngester{}
