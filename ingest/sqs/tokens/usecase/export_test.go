package usecase

import "github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"

func GetTokensFromChainRegistry(url string) (map[string]domain.Token, error) {
	return getTokensFromChainRegistry(url)
}
