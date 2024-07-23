package service

import sdk "github.com/cosmos/cosmos-sdk/types"

type SQSStreamingService = sqsStreamingService

func (s *sqsStreamingService) ProcessBlockRecoverError(ctx sdk.Context) error {
	return s.processBlockRecoverError(ctx)
}
