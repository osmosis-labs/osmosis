package e2e

import (
	"fmt"
	"io"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *IntegrationTestSuite) TestQueryDenomBalance() {

	s.Run("TestQueryDenomBalance", func() {
		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))
		_, err := queryDenomBalance(chainBAPIEndpoint, s.chainA.validators[0].keyInfo.GetAddress().String(), "osmo") 
		s.Require().NoError(err)
	})
}

func queryDenomBalance(endpoint, addr, denom string) (sdk.Coin, error) {
	var zeroCoin sdk.Coin

	path := fmt.Sprintf(
		"%s/osmosis/bank/v1beta1/balances/%s/by_denom?denom=%s",
		endpoint, addr, denom,
	)
	resp, err := http.Get(path)
	if err != nil {
		return zeroCoin, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroCoin, err
	}

	var balanceResp banktypes.QueryBalanceResponse
	if err := cdc.UnmarshalJSON(bz, &balanceResp); err != nil {
		return zeroCoin, err
	}

	return *balanceResp.Balance, nil
}
