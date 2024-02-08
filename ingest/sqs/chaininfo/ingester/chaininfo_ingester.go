package chaininfoingester

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain/repository"
	chaininforedisrepo "github.com/osmosis-labs/sqs/sqsdomain/repository/redis/chaininfo"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
)

// chainInfoIngester is an ingester for blockchain information.
// It implements ingest.Ingester.
// It reads the latest blockchain height and writes it to the chainInfo repository.
type chainInfoIngester struct {
	chainInfoRepo     chaininforedisrepo.ChainInfoRepository
	repositoryManager repository.TxManager
}

// New returns a new chain information ingester.
func New(chainInfoRepo chaininforedisrepo.ChainInfoRepository, repositoryManager repository.TxManager) domain.AtomicIngester {
	return &chainInfoIngester{
		chainInfoRepo:     chainInfoRepo,
		repositoryManager: repositoryManager,
	}
}

// ProcessBlock implements ingest.Ingester.
// It reads the latest blockchain height and stores it in Redis.
func (ci *chainInfoIngester) ProcessBlock(ctx sdk.Context, tx repository.Tx) error {
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
