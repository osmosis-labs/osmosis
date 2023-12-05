package usecase

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
)

type chainInfoUseCase struct {
	contextTimeout         time.Duration
	chainInfoRepository    mvc.ChainInfoRepository
	redisRepositoryManager mvc.TxManager
}

// The max number of seconds allowed for there to be no updated
// TODO: epoch???
const MaxAllowedHeightUpdateTimeDeltaSecs = 30

var _ mvc.ChainInfoUsecase = &chainInfoUseCase{}

func NewChainInfoUsecase(timeout time.Duration, chainInfoRepository mvc.ChainInfoRepository, redisRepositoryManager mvc.TxManager) mvc.ChainInfoUsecase {
	return &chainInfoUseCase{
		contextTimeout:         timeout,
		chainInfoRepository:    chainInfoRepository,
		redisRepositoryManager: redisRepositoryManager,
	}
}

func (p *chainInfoUseCase) GetLatestHeight(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	latestHeight, err := p.chainInfoRepository.GetLatestHeight(ctx)
	if err != nil {
		return 0, err
	}

	latestHeightRetrievalTime, err := p.chainInfoRepository.GetLatestHeightRetrievalTime(ctx)
	if err != nil {
		return 0, err
	}

	// Current UTC time
	currentTimeUTC := time.Now().UTC()

	// Time since last height retrieval
	timeDeltaSecs := int(currentTimeUTC.Sub(latestHeightRetrievalTime).Seconds())

	// Validate that it does not exceed the max allowed time delta
	if timeDeltaSecs > MaxAllowedHeightUpdateTimeDeltaSecs {
		return 0, domain.StaleHeightError{
			StoredHeight:            latestHeight,
			TimeSinceLastUpdate:     timeDeltaSecs,
			MaxAllowedTimeDeltaSecs: MaxAllowedHeightUpdateTimeDeltaSecs,
		}
	}

	// Store the latest height retrieval time
	if err := p.chainInfoRepository.StoreLatestHeightRetrievalTime(ctx, currentTimeUTC); err != nil {
		return 0, err
	}

	return latestHeight, nil
}
