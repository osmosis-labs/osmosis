package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestInitAppConfigTemplate(t *testing.T) {
	// This test validates templates uses existing fields in config
	appTemplate, appConfig := initAppConfig()
	tmpl := template.New("appDefaultValues")
	tmpl, err := tmpl.Parse(appTemplate)
	require.NoError(t, err)
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, appConfig)
	require.NoError(t, err)
}

func TestBuildAssetMaps(t *testing.T) {
	assets := []Asset{
		{Symbol: "USDC", CoinMinimalDenom: "factory/osmo1alloy/alloyed/allUSDC", SourceDenom: "factory/osmo1alloy/alloyed/allUSDC", Decimals: 6},
		{Symbol: "ETH.axl", CoinMinimalDenom: "ibc/AXLETH", SourceDenom: "weth-wei", Decimals: 18},
		{Symbol: "ATOM", CoinMinimalDenom: "ibc/ATOM", SourceDenom: "uatom", Decimals: 6},
		{Symbol: "DUP", CoinMinimalDenom: "ibc/DUP1", SourceDenom: "uusdc", Decimals: 6},
		{Symbol: "dup", CoinMinimalDenom: "ibc/DUP2", SourceDenom: "uusdc", Decimals: 18},
		{Symbol: "OLD", Decimals: 6},
		{Symbol: "PSTAKE", CoinMinimalDenom: "ibc/PSTAKE", SourceDenom: "ibc/A6E3AF63", Decimals: 18},
		{Symbol: "INJ", CoinMinimalDenom: "ibc/INJ", SourceDenom: "inj", Decimals: 18},
	}
	baseMap, baseMapRev := buildAssetMaps(assets, true, true)

	coin, err := transformCoinValueToBaseInt("1.5", "usdc", baseMap)
	require.NoError(t, err)
	require.Equal(t, "1500000factory/osmo1alloy/alloyed/allUSDC", coin)

	coin, err = transformCoinValueToBaseInt("2", "eth.axl", baseMap)
	require.NoError(t, err)
	require.Equal(t, "2000000000000000000ibc/AXLETH", coin)

	// A symbol shared by two assets and an entry without a denom are never converted.
	_, err = transformCoinValueToBaseInt("1", "dup", baseMap)
	require.Error(t, err)
	_, err = transformCoinValueToBaseInt("1", "old", baseMap)
	require.Error(t, err)

	// Output names stay in base units. A sourceDenom shared by two assets, shaped like an Osmosis
	// denom, or equal to a symbol is not used.
	require.Equal(t, map[string]string{"ibc/AXLETH": "weth-wei", "ibc/ATOM": "uatom"}, baseMapRev)

	// The embedded lists must be in the frontend format, or every conversion is silently off.
	for _, fileName := range []string{"osmosis-1-assetlist.json", "osmo-test-5-assetlist.json"} {
		byteValue, err := assetFS.ReadFile(fileName)
		require.NoError(t, err)
		var assetList AssetList
		require.NoError(t, json.Unmarshal(byteValue, &assetList))
		require.True(t, hasUsableAsset(assetList.Assets), fileName)
	}
}
