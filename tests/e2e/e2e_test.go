package e2e

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *IntegrationTestSuite) TestQueryBalances() {
	var (
		expectedDenoms   = []string{osmoDenom, stakeDenom}
		expectedBalances = []uint64{osmoBalance, stakeBalance - stakeAmount}
	)

	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chain.id][0].GetHostPort("1317/tcp"))
	balances, err := queryBalances(chainAAPIEndpoint, s.chain.validators[0].keyInfo.GetAddress().String())
	s.Require().NoError(err)
	s.Require().NotNil(balances)
	s.Require().Equal(2, len(balances))

	actualDenoms := make([]string, 0, 2)
	actualBalances := make([]uint64, 0, 2)

	for _, balance := range balances {
		actualDenoms = append(actualDenoms, balance.GetDenom())
		actualBalances = append(actualBalances, balance.Amount.Uint64())
	}

	s.Require().ElementsMatch(expectedDenoms, actualDenoms)
	s.Require().ElementsMatch(expectedBalances, actualBalances)
}

func queryBalances(endpoint, addr string) (sdk.Coins, error) {
	path := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		endpoint, addr,
	)
	resp, err := http.Get(path)
	retriesLeft := 5
	for {
		resp, err = http.Get(path)

		if resp.StatusCode == http.StatusServiceUnavailable {
			retriesLeft--
			if retriesLeft == 0 {
				return nil, errors.New(fmt.Sprintf("exceeded retry limit of %d with %d", retriesLeft, http.StatusServiceUnavailable))
			}
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.GetBalances(), nil
}
