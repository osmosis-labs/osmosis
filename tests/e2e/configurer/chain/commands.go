package chain

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	appparams "github.com/osmosis-labs/osmosis/v11/app/params"
	"github.com/osmosis-labs/osmosis/v11/tests/e2e/configurer/config"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) CreatePool(poolFile, from string) {
	n.LogActionF("creating pool from file %s", poolFile)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created pool")
}

func (n *NodeConfig) StoreWasmCode(wasmFile, from string) {
	n.LogActionF("storing wasm code from file %s", wasmFile)
	cmd := []string{"osmosisd", "tx", "wasm", "store", wasmFile, fmt.Sprintf("--from=%s", from), "--gas=auto", "--gas-prices=0.1uosmo", "--gas-adjustment=1.3"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully stored")
}

func (n *NodeConfig) InstantiateWasmContract(codeId, initMsg, from string) {
	n.LogActionF("instantiating wasm contract %s with %s", codeId, initMsg)
	cmd := []string{"osmosisd", "tx", "wasm", "instantiate", codeId, initMsg, fmt.Sprintf("--from=%s", from), "--no-admin", "--label=ratelimit"}
	n.LogActionF(strings.Join(cmd, " "))
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully initialized")
}

// QueryParams extracts the params for a given subspace and key. This is done generically via json to avoid having to
// specify the QueryParamResponse type (which may not exist for all params).
func (n *NodeConfig) QueryParams(subspace, key string, result any) {
	cmd := []string{"osmosisd", "query", "params", "subspace", subspace, key, "--output=json"}

	out, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)

	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(n.t, err)
}

func (n *NodeConfig) SubmitParamChangeProposal(proposalJson, from string) {
	n.LogActionF("submitting param change proposal %s", proposalJson)
	// ToDo: Is there a better way to do this?
	wd, err := os.Getwd()
	require.NoError(n.t, err)
	localProposalFile := wd + "/scripts/param_change_proposal.json"
	f, err := os.Create(localProposalFile)
	require.NoError(n.t, err)
	_, err = f.WriteString(proposalJson)
	require.NoError(n.t, err)
	err = f.Close()
	require.NoError(n.t, err)

	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "param-change", "/osmosis/param_change_proposal.json", "--is-expedited=true", fmt.Sprintf("--from=%s", from)}

	_, _, err = n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)

	err = os.Remove(localProposalFile)
	require.NoError(n.t, err)

	n.LogActionF("successfully submitted param change proposal")
}

func (n *NodeConfig) FailIBCTransfer(from, recipient, amount string) {
	n.LogActionF("IBC sending %s from %s to %s", amount, from, recipient)

	cmd := []string{"osmosisd", "tx", "ibc-transfer", "transfer", "transfer", "channel-0", recipient, amount, fmt.Sprintf("--from=%s", from)}

	_, _, err := n.containerManager.ExecTxCmdWithSuccessString(n.t, n.chainId, n.Name, cmd, "rate limit exceeded")
	require.NoError(n.t, err)

	n.LogActionF("Failed to send IBC transfer (as expected)")
}

func (n *NodeConfig) SubmitUpgradeProposal(upgradeVersion string, upgradeHeight int64, initialDeposit sdk.Coin) {
	n.LogActionF("submitting upgrade proposal %s for height %d", upgradeVersion, upgradeHeight)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%d", upgradeHeight), "--upgrade-info=\"\"", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted upgrade proposal")
}

func (n *NodeConfig) SubmitSuperfluidProposal(asset string, initialDeposit sdk.Coin) {
	n.LogActionF("submitting superfluid proposal for asset %s", asset)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted superfluid proposal for asset %s", asset)
}

func (n *NodeConfig) SubmitTextProposal(text string, initialDeposit sdk.Coin, isExpedited bool) {
	n.LogActionF("submitting text gov proposal")
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val", fmt.Sprintf("--deposit=%s", initialDeposit)}
	if isExpedited {
		cmd = append(cmd, "--is-expedited=true")
	}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully submitted text gov proposal")
}

func (n *NodeConfig) DepositProposal(proposalNumber int, isExpedited bool) {
	n.LogActionF("depositing on proposal: %d", proposalNumber)
	deposit := sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.MinDepositValue)).String()
	if isExpedited {
		deposit = sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.MinExpeditedDepositValue)).String()
	}
	cmd := []string{"osmosisd", "tx", "gov", "deposit", fmt.Sprintf("%d", proposalNumber), deposit, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully deposited on proposal %d", proposalNumber)
}

func (n *NodeConfig) VoteYesProposal(from string, proposalNumber int) {
	n.LogActionF("voting yes on proposal: %d", proposalNumber)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "yes", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully voted yes on proposal %d", proposalNumber)
}

func (n *NodeConfig) VoteNoProposal(from string, proposalNumber int) {
	n.LogActionF("voting no on proposal: %d", proposalNumber)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "no", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully voted no on proposal: %d", proposalNumber)
}

func (n *NodeConfig) LockTokens(tokens string, duration string, from string) {
	n.LogActionF("locking %s for %s", tokens, duration)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully created lock")
}

func (n *NodeConfig) SuperfluidDelegate(lockNumber int, valAddress string, from string) {
	lockStr := strconv.Itoa(lockNumber)
	n.LogActionF("superfluid delegating lock %s to %s", lockStr, valAddress)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully superfluid delegated lock %s to %s", lockStr, valAddress)
}

func (n *NodeConfig) BankSend(amount string, sendAddress string, receiveAddress string) {
	n.LogActionF("bank sending %s from address %s to %s", amount, sendAddress, receiveAddress)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.LogActionF("successfully sent bank sent %s from address %s to %s", amount, sendAddress, receiveAddress)
}

func (n *NodeConfig) CreateWallet(walletName string) string {
	n.LogActionF("creating wallet %s", walletName)
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.LogActionF("created wallet %s, waller address - %s", walletName, walletAddr)
	return walletAddr
}

func (n *NodeConfig) QueryPropStatusTimed(proposalNumber int, desiredStatus string, totalTime chan time.Duration) {
	start := time.Now()
	require.Eventually(
		n.t,
		func() bool {
			status, err := n.QueryPropStatus(proposalNumber)
			if err != nil {
				return false
			}

			return status == desiredStatus
		},
		1*time.Minute,
		10*time.Millisecond,
		"Osmosis node failed to retrieve prop tally",
	)
	elapsed := time.Since(start)
	totalTime <- elapsed
}
