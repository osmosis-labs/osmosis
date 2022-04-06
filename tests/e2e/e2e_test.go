package e2e

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// func (s *IntegrationTestSuite) TestQueryBalances() {
// 	var (
// 		expectedDenoms    = []string{osmoDenom, stakeDenom}
// 		expectedBalancesA = []uint64{osmoBalanceA, stakeBalanceA - stakeAmountA}
// 		expectedBalancesB = []uint64{osmoBalanceB, stakeBalanceB - stakeAmountB}
// 	)

// 	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
// 	balancesA, err := queryBalances(chainAAPIEndpoint, s.chainA.validators[0].keyInfo.GetAddress().String())
// 	s.Require().NoError(err)
// 	s.Require().NotNil(balancesA)
// 	// s.Require().Equal(2, len(balancesA))

// 	chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))
// 	balancesB, err := queryBalances(chainBAPIEndpoint, s.chainB.validators[0].keyInfo.GetAddress().String())
// 	s.Require().NoError(err)
// 	s.Require().NotNil(balancesB)
// 	// s.Require().Equal(2, len(balancesB))

// 	actualDenomsA := make([]string, 0, 2)
// 	actualBalancesA := make([]uint64, 0, 2)
// 	actualDenomsB := make([]string, 0, 2)
// 	actualBalancesB := make([]uint64, 0, 2)

// 	for _, balanceA := range balancesA {
// 		actualDenomsA = append(actualDenomsA, balanceA.GetDenom())
// 		actualBalancesA = append(actualBalancesA, balanceA.Amount.Uint64())
// 	}

// 	for _, balanceB := range balancesB {
// 		actualDenomsB = append(actualDenomsB, balanceB.GetDenom())
// 		actualBalancesB = append(actualBalancesB, balanceB.Amount.Uint64())
// 	}

// 	s.Require().ElementsMatch(expectedDenoms, actualDenomsA)
// 	s.Require().ElementsMatch(expectedBalancesA, actualBalancesA)
// 	s.Require().ElementsMatch(expectedDenoms, actualDenomsB)
// 	s.Require().ElementsMatch(expectedBalancesB, actualBalancesB)
// }

func queryBalances(endpoint, addr string) (sdk.Coins, error) {
	path := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		endpoint, addr,
	)
	var err error
	var resp *http.Response
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

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	var ibcStakeDenom string

	s.Run("send_uosmo_to_chainB", func() {
		recipient := s.chainB.validators[0].keyInfo.GetAddress().String()
		token := sdk.NewInt64Coin(osmoDenom, 3300000000) // 3,300uosmo
		s.sendIBC(s.chainA.id, s.chainB.id, recipient, token)

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances sdk.Coins
			err      error
		)
		s.Require().Eventually(
			func() bool {
				balances, err = queryBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)

				return balances.Len() == 3
			},
			time.Minute,
			5*time.Second,
		)

		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}
