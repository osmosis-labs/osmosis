package usecase

import (
	"context"
	"time"

	"github.com/go-redis/redis"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
)

type chainInfoUseCase struct {
	contextTimeout         time.Duration
	chainInfoRepository    mvc.ChainInfoRepository
	redisRepositoryManager mvc.TxManager
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
	}
}

func (p *chainInfoUseCase) GetLatestHeight(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	latestHeight, err := p.chainInfoRepository.GetLatestHeight(ctx)
	if err != nil {
		return 0, err
	}

	// Current UTC time
	currentTimeUTC := time.Now().UTC()

	latestHeightRetrievalTime, err := p.chainInfoRepository.GetLatestHeightRetrievalTime(ctx)
	if err != nil {
		// If there is no entry, then we can assume that the height has never been retrieved,
		// so we store the current time.
		// TODO: clean up this error handling
		if err.Error() == redis.Nil.Error() {
			// Store the latest height retrieval time
			if err := p.chainInfoRepository.StoreLatestHeightRetrievalTime(ctx, currentTimeUTC); err != nil {
				return 0, err
			}

			return latestHeight, nil
		}

		return 0, err
	}

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
