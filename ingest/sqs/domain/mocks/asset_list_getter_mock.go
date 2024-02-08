package mocks

import (
	"context"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
)

type AssetListGetterMock struct {
	tokenPrecisionMap map[string]int
}

// GetDenomPrecisions implements domain.TokensUsecase.
func (tu *AssetListGetterMock) GetDenomPrecisions(ctx context.Context) (map[string]int, error) {
	return tu.tokenPrecisionMap, nil
}

var _ domain.AssetListGetter = &AssetListGetterMock{}
