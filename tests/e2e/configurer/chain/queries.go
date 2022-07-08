package chain

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (c *Config) QueryRPC(path string) ([]byte, error) {
	var err error
	var resp *http.Response
	retriesLeft := 5
	for {
		resp, err = http.Get(path)

		if resp.StatusCode == http.StatusServiceUnavailable {
			retriesLeft--
			if retriesLeft == 0 {
				return nil, err
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
	return bz, nil
}

func (c *Config) QueryChainStatus(validatorIdx int) []byte {
	cmd := []string{"osmosisd", "status"}
	_, errBuf, err := c.containerManager.ExecCmd(c.t, c.Id, validatorIdx, cmd, "")
	require.NoError(c.t, err)
	return errBuf.Bytes()
}

func (c *Config) QueryCurrentChainHeightFromValidator(validatorIdx int) int {
	var block syncInfo
	require.Eventually(
		c.t,
		func() bool {
			out := c.QueryChainStatus(validatorIdx)
			err := json.Unmarshal(out, &block)
			return err == nil
		},
		1*time.Minute,
		time.Second,
		"Osmosis node failed to retrieve height info",
	)
	currentHeight, err := strconv.Atoi(block.SyncInfo.LatestHeight)
	require.NoError(c.t, err)
	return currentHeight
}

func (c *Config) QueryBalances(validatorIndex int, addr string) (sdk.Coins, error) {
	cmd := []string{"osmosisd", "query", "bank", "balances", addr, "--output=json"}
	outBuf, _, err := c.containerManager.ExecCmd(c.t, c.Id, validatorIndex, cmd, "")
	require.NoError(c.t, err)

	var balancesResp banktypes.QueryAllBalancesResponse
	err = util.Cdc.UnmarshalJSON(outBuf.Bytes(), &balancesResp)
	require.NoError(c.t, err)

	return balancesResp.GetBalances(), nil
}

func (c *Config) QueryPropTally(validatorIdx int, addr string) (sdk.Int, sdk.Int, sdk.Int, sdk.Int, error) {
	hostPort, err := c.containerManager.GetValidatorHostPort(c.Id, validatorIdx, "1317/tcp")
	require.NoError(c.t, err)

	endpoint := fmt.Sprintf("http://%s", hostPort)

	path := fmt.Sprintf(
		"%s/cosmos/gov/v1beta1/proposals/%s/tally",
		endpoint, addr,
	)
	bz, err := c.QueryRPC(path)
	require.NoError(c.t, err)

	var balancesResp govtypes.QueryTallyResultResponse
	if err := util.Cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt(), err
	}
	noTotal := balancesResp.Tally.No
	yesTotal := balancesResp.Tally.Yes
	noWithVetoTotal := balancesResp.Tally.NoWithVeto
	abstainTotal := balancesResp.Tally.Abstain

	return noTotal, yesTotal, noWithVetoTotal, abstainTotal, nil
}

func (c *Config) QueryValidatorOperatorAddresses() {
	for i, val := range c.NodeConfigs {
		cmd := []string{"osmosisd", "debug", "addr", val.PublicKey}
		c.t.Logf("extracting validator operator addresses for chain-id: %s", c.Id)
		_, errBuf, err := c.containerManager.ExecCmd(c.t, c.Id, i, cmd, "")
		require.NoError(c.t, err)
		re := regexp.MustCompile("osmovaloper(.{39})")
		operAddr := fmt.Sprintf("%s\n", re.FindString(errBuf.String()))
		c.NodeConfigs[i].OperatorAddress = strings.TrimSuffix(operAddr, "\n")
	}
}

func (c *Config) QueryIntermediaryAccount(validatorIdx int, denom string, valAddr string) (int, error) {
	hostPort, err := c.containerManager.GetValidatorHostPort(c.Id, validatorIdx, "1317/tcp")
	require.NoError(c.t, err)

	endpoint := fmt.Sprintf("http://%s", hostPort)

	intAccount := superfluidtypes.GetSuperfluidIntermediaryAccountAddr(denom, valAddr)
	path := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations/%s",
		endpoint, valAddr, intAccount,
	)
	bz, err := c.QueryRPC(path)
	require.NoError(c.t, err)

	var stakingResp stakingtypes.QueryDelegationResponse
	err = util.Cdc.UnmarshalJSON(bz, &stakingResp)
	require.NoError(c.t, err)

	intAccBalance := stakingResp.DelegationResponse.Balance.Amount.String()
	intAccountBalance, err := strconv.Atoi(intAccBalance)
	require.NoError(c.t, err)
	return intAccountBalance, err
}
