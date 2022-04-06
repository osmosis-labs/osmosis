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
		expectedDenoms    = []string{osmoDenom, stakeDenom}
		expectedBalancesA = []uint64{osmoBalanceA, stakeBalanceA - stakeAmountA}
		expectedBalancesB = []uint64{osmoBalanceB, stakeBalanceB - stakeAmountB}
	)

	chainAAPIEndpointA := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	balancesA, err := queryBalances(chainAAPIEndpointA, s.chainA.validators[0].keyInfo.GetAddress().String())
	s.Require().NoError(err)
	s.Require().NotNil(balancesA)
	s.Require().Equal(2, len(balancesA))

	chainAAPIEndpointB := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))
	balancesB, err := queryBalances(chainAAPIEndpointB, s.chainB.validators[0].keyInfo.GetAddress().String())
	s.Require().NoError(err)
	s.Require().NotNil(balancesB)
	s.Require().Equal(2, len(balancesB))

	actualDenomsA := make([]string, 0, 2)
	actualBalancesA := make([]uint64, 0, 2)
	actualDenomsB := make([]string, 0, 2)
	actualBalancesB := make([]uint64, 0, 2)

	for _, balanceA := range balancesA {
		actualDenomsA = append(actualDenomsA, balanceA.GetDenom())
		actualBalancesA = append(actualBalancesA, balanceA.Amount.Uint64())
	}

	for _, balanceB := range balancesB {
		actualDenomsB = append(actualDenomsB, balanceB.GetDenom())
		actualBalancesB = append(actualBalancesB, balanceB.Amount.Uint64())
	}

	s.Require().ElementsMatch(expectedDenoms, actualDenomsA)
	s.Require().ElementsMatch(expectedBalancesA, actualBalancesA)
	s.Require().ElementsMatch(expectedDenoms, actualDenomsB)
	s.Require().ElementsMatch(expectedBalancesB, actualBalancesB)
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
