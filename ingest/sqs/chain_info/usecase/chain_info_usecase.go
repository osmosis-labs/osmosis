package usecase

import (
	"context"
	"time"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
)

type chainInfoUseCase struct {
	contextTimeout         time.Duration
	chainInfoRepository    mvc.ChainInfoRepository
	redisRepositoryManager mvc.TxManager
}

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

	return latestHeight, nil
}
