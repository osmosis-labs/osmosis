package domain

func GetTokensFromChainRegistry(chainRegistryAssetsFileURL string) (map[string]Token, error) {
	return getTokensFromChainRegistry(chainRegistryAssetsFileURL)
}
