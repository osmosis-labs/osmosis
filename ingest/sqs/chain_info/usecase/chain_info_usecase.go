package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
)

type chainInfoUseCase struct {
	contextTimeout         time.Duration
	chainInfoRepository    mvc.ChainInfoRepository
	redisRepositoryManager mvc.TxManager

	// N.B. sometimes the node gets stuck and does not make progress.
	// However, it returns 200 OK for the status endpoint and claims to be not catching up.
	// This has caused the healthcheck to pass with false positives in production.
	// As a result, we need to keep track of the last seen height and time to ensure that the height is
	// updated within a reasonable time frame.
	lastSeenMx            sync.Mutex
	lastSeenUpdatedHeight uint64
	lastSeenUpdatedTime   time.Time
}

// The max number of seconds allowed for there to be no updates
// TODO: epoch???
const MaxAllowedHeightUpdateTimeDeltaSecs = 30

var _ mvc.ChainInfoUsecase = &chainInfoUseCase{}

func NewChainInfoUsecase(timeout time.Duration, chainInfoRepository mvc.ChainInfoRepository, redisRepositoryManager mvc.TxManager) mvc.ChainInfoUsecase {
	return &chainInfoUseCase{
		contextTimeout:         timeout,
		chainInfoRepository:    chainInfoRepository,
		redisRepositoryManager: redisRepositoryManager,

		lastSeenMx: sync.Mutex{},
	}
}

// GetLatestHeight retrieves the latest blockchain height
//
// Despite being a getter, this method also validates that the height is updated within a reasonable time frame.
//
// Sometimes the node gets stuck and does not make progress.
// However, it returns 200 OK for the status endpoint and claims to be not catching up.
// This has caused the healthcheck to pass with false positives in production.
// As a result, we need to keep track of the last seen height and time. Chain ingester pushes
// the latest height into Redis. This method checks that the height is updated within a reasonable time frame.
func (p *chainInfoUseCase) GetLatestHeight(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	latestHeight, err := p.chainInfoRepository.GetLatestHeight(ctx)
	if err != nil {
		return 0, err
	}

	p.lastSeenMx.Lock()
	defer p.lastSeenMx.Unlock()

	currentTimeUTC := time.Now().UTC()

	// Time since last height retrieval
	timeDeltaSecs := int(currentTimeUTC.Sub(p.lastSeenUpdatedTime).Seconds())

	isHeightUpdated := latestHeight > p.lastSeenUpdatedHeight

	// Validate that it does not exceed the max allowed time delta
	if !isHeightUpdated && timeDeltaSecs > MaxAllowedHeightUpdateTimeDeltaSecs {
		return 0, domain.StaleHeightError{
			StoredHeight:            latestHeight,
			TimeSinceLastUpdate:     timeDeltaSecs,
			MaxAllowedTimeDeltaSecs: MaxAllowedHeightUpdateTimeDeltaSecs,
		}
	}

	// Update the last seen height and time
	p.lastSeenUpdatedHeight = latestHeight
	p.lastSeenUpdatedTime = currentTimeUTC

	return latestHeight, nil
}
