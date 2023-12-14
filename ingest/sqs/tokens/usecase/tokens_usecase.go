package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/json"
)

type tokensUseCase struct {
	contextTimeout time.Duration
}

// Struct to represent the JSON structure
type AssetList struct {
	ChainName string `json:"chain_name"`
	Assets    []struct {
		Description string `json:"description"`
		DenomUnits  []struct {
			Denom    string `json:"denom"`
			Exponent int    `json:"exponent"`
		} `json:"denom_units"`
		Base     string        `json:"base"`
		Name     string        `json:"name"`
		Display  string        `json:"display"`
		Symbol   string        `json:"symbol"`
		Traces   []interface{} `json:"traces"`
		LogoURIs struct {
			PNG string `json:"png"`
			SVG string `json:"svg"`
		} `json:"logo_URIs"`
		CoingeckoID string   `json:"coingecko_id"`
		Keywords    []string `json:"keywords"`
	} `json:"assets"`
}

const assetListFileURL = "https://raw.githubusercontent.com/osmosis-labs/assetlists/main/osmosis-1/osmosis-1.assetlist.json"

var _ domain.TokensUsecase = &tokensUseCase{}

// NewTokensUsecase will create a new tokens use case object
func NewTokensUsecase(timeout time.Duration) domain.TokensUsecase {
	return &tokensUseCase{
		contextTimeout: timeout,
	}
}

// GetDenomPrecisions implements domain.TokensUsecase.
func (tu *tokensUseCase) GetDenomPrecisions(ctx context.Context) (map[string]int, error) {
	tokensByDenomMap, err := getTokensFromChainRegistry(assetListFileURL)
	if err != nil {
		return nil, err
	}

	denomPrecisions := make(map[string]int, len(tokensByDenomMap))
	for _, token := range tokensByDenomMap {
		denomPrecisions[token.ChainDenom] = token.Precision
	}

	return denomPrecisions, nil
}

// getTokensFromChainRegistry fetches the tokens from the chain registry.
// It returns a map of tokens by chain denom.
func getTokensFromChainRegistry(chainRegistryAssetsFileURL string) (map[string]domain.Token, error) {
	// Fetch the JSON data from the URL
	response, err := http.Get(chainRegistryAssetsFileURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Decode the JSON data
	var assetList AssetList
	err = json.NewDecoder(response.Body).Decode(&assetList)
	if err != nil {
		return nil, err
	}

	tokensByChainDenom := make(map[string]domain.Token)

	// Iterate through each asset and its denom units to print exponents
	for _, asset := range assetList.Assets {
		token := domain.Token{}

		if len(asset.DenomUnits) == 1 {
			// At time of script creation, only the following tokens have 1 denom unit with zero exponent:
			// one ibc/FE2CD1E6828EC0FAB8AF39BAC45BC25B965BA67CCBC50C13A14BD610B0D1E2C4 0
			// one ibc/52E12CF5CA2BB903D84F5298B4BFD725D66CAB95E09AA4FC75B2904CA5485FEB 0
			// one ibc/E27CD305D33F150369AB526AEB6646A76EC3FFB1A6CA58A663B5DE657A89D55D 0
			//
			// These seem as tokens that are not useful in routing so we silently skip them.
			continue
		}

		for _, denom := range asset.DenomUnits {
			if denom.Exponent == 0 {
				token.ChainDenom = denom.Denom
			}

			if denom.Exponent > 0 {
				// There are edge cases where we have 3 denom exponents for a token.
				// We filter out the intermediate denom exponents and only use the human readable denom.
				if denom.Denom == "mluna" || denom.Denom == "musd" || denom.Denom == "msomm" || denom.Denom == "mkrw" || denom.Denom == "uarch" {
					continue
				}

				token.HumanDenom = denom.Denom
				token.Precision = denom.Exponent
			}
		}

		tokensByChainDenom[token.ChainDenom] = token
	}

	return tokensByChainDenom, nil
}
