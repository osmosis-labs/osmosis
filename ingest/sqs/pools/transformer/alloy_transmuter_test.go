package poolstransformer_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
	"github.com/osmosis-labs/sqs/sqsdomain"
)

func (s *PoolTransformerTestSuite) TestUpdateAlloyedTransmuterPool() {
	s.Setup()

	s.FundAcc(s.TestAccs[0], sdk.NewCoins(
		sdk.NewCoin(apptesting.DefaultTransmuterDenomA, osmomath.NewInt(100000000)),
		sdk.NewCoin(apptesting.DefaultTransmuterDenomB, osmomath.NewInt(100000000)),
	))

	pool := s.PrepareAlloyTransmuterPool(s.TestAccs[0], apptesting.AlloyTransmuterInstantiateMsg{
		PoolAssetConfigs:                []apptesting.AssetConfig{{Denom: apptesting.DefaultTransmuterDenomA, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomANormFactor)}, {Denom: apptesting.DefaultTransmuterDenomB, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomBNormFactor)}},
		AlloyedAssetSubdenom:            apptesting.DefaultAlloyedSubDenom,
		AlloyedAssetNormalizationFactor: osmomath.NewInt(apptesting.DefaultAlloyedDenomNormFactor),
		Admin:                           s.TestAccs[0].String(),
		Moderator:                       s.TestAccs[1].String(),
	})

	// Create OSMO / USDC pool
	// Note that spot price is 1 OSMO = 2 USDC
	usdcOsmoPoolID := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(USDC, defaultAmount), sdk.NewCoin(UOSMO, halfDefaultAmount))

	// Initialize the pool ingester
	poolIngester := s.initializePoolIngester(usdcOsmoPoolID)

	cosmWasmPoolModel := sqsdomain.CosmWasmPoolModel{}
	poolDenoms := []string{apptesting.DefaultTransmuterDenomA, apptesting.DefaultTransmuterDenomB}

	poolIngester.UpdateAlloyTransmuterInfo(s.Ctx, pool.GetId(), pool.GetAddress(), &cosmWasmPoolModel, &poolDenoms)

	alloyedDenom := fmt.Sprintf("factory/%s/alloyed/%s", pool.GetAddress(), apptesting.DefaultAlloyedSubDenom)

	// Check if the pool has been updated
	s.Equal(sqsdomain.CWPoolData{
		AlloyTransmuter: &sqsdomain.AlloyTransmuterData{
			AlloyedDenom: alloyedDenom,
			AssetConfigs: []sqsdomain.TransmuterAssetConfig{
				{Denom: apptesting.DefaultTransmuterDenomA, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomANormFactor)},
				{Denom: apptesting.DefaultTransmuterDenomB, NormalizationFactor: osmomath.NewInt(apptesting.DefaultTransmuterDenomBNormFactor)},
				{Denom: alloyedDenom, NormalizationFactor: osmomath.NewInt(apptesting.DefaultAlloyedDenomNormFactor)}},
		},
	}, cosmWasmPoolModel.Data)

	s.Equal([]string{
		apptesting.DefaultTransmuterDenomA,
		apptesting.DefaultTransmuterDenomB,
		alloyedDenom,
	}, poolDenoms)
}
