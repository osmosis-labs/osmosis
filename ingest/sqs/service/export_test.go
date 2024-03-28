package service

import sdk "github.com/cosmos/cosmos-sdk/types"

type SQSStreamingService = sqsStreamingService

func (s *sqsStreamingService) ProcessBlock(ctx sdk.Context) error {
	return s.processBlock(ctx)
}

func (s *sqsStreamingService) SetShouldProcessAllBlockData(shouldProcessAllBlockData bool) {
	s.shouldProcessAllBlockData = shouldProcessAllBlockData
}

func (s *sqsStreamingService) GetShouldProcessAllBlockData() bool {
	return s.shouldProcessAllBlockData
}

func (s *sqsStreamingService) ProcessBlockRecoverError(ctx sdk.Context) error {
	return s.processBlockRecoverError(ctx)
}
