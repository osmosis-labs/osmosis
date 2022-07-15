package chain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) CreatePool(poolFile, from string) {
	n.t.Logf("creating pool for chain-id: %s", n.chainId)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully created pool from container: %s", n.Name)
}

func (n *NodeConfig) SubmitUpgradeProposal(upgradeVersion string, upgradeHeight int64) {
	n.t.Logf("submitting upgrade proposal on container: %s", n.Name)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "software-upgrade", upgradeVersion, fmt.Sprintf("--title=\"%s upgrade\"", upgradeVersion), "--description=\"upgrade proposal submission\"", fmt.Sprintf("--upgrade-height=%d", upgradeHeight), "--upgrade-info=\"\"", "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Log("successfully submitted upgrade proposal")
}

func (n *NodeConfig) SubmitSuperfluidProposal(asset string) {
	n.t.Logf("submitting superfluid proposal for asset %s on container: %s", asset, n.Name)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "set-superfluid-assets-proposal", fmt.Sprintf("--superfluid-assets=%s", asset), fmt.Sprintf("--title=\"%s superfluid asset\"", asset), fmt.Sprintf("--description=\"%s superfluid asset\"", asset), "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Log("successfully submitted superfluid proposal")
}

func (n *NodeConfig) SubmitTextProposal(text string) {
	n.t.Logf("submitting text proposal on container: %s", n.Name)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Log("successfully submitted text proposal")
}

func (n *NodeConfig) DepositProposal(proposalNumber int) {
	n.t.Logf("depositing to proposal from container: %s", n.Name)
	cmd := []string{"osmosisd", "tx", "gov", "deposit", fmt.Sprintf("%d", proposalNumber), "500000000uosmo", "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Log("successfully deposited to proposal")
}

func (n *NodeConfig) VoteYesProposal(from string, proposalNumber int) {
	n.t.Logf("voting yes on proposal for chain-id: %s", n.chainId)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "yes", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully voted yes on proposal from container: %s", n.Name)
}

func (n *NodeConfig) VoteNoProposal(from string, proposalNumber int) {
	n.t.Logf("voting no on proposal for chain-id: %s", n.chainId)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "no", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully voted no for proposal from container: %s", n.Name)
}

func (n *NodeConfig) LockTokens(tokens string, duration string, from string) {
	n.t.Logf("locking %s for %s on chain-id: %s", tokens, duration, n.chainId)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully created lock from container: %s", n.Name)
}

func (n *NodeConfig) SuperfluidDelegate(lockNumber int, valAddress string, from string) {
	lockStr := strconv.Itoa(lockNumber)
	n.t.Logf("superfluid delegating lock %s to %s on chain-id: %s", lockStr, valAddress, n.chainId)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully superfluid delegated from container: %s", n.Name)
}

func (n *NodeConfig) BankSend(amount string, sendAddress string, receiveAddress string) {
	n.t.Logf("bank sending %s from %s to %s on chain-id: %s", amount, sendAddress, receiveAddress, n.chainId)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully sent tx from container: %s", n.Name)
}

func (n *NodeConfig) CreateWallet(walletName string) string {
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	return walletAddr
}
