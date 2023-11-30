package usecase

import "github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"

func GetTokensFromChainRegistry(url string) (map[string]domain.Token, error) {
	return getTokensFromChainRegistry(url)
}
