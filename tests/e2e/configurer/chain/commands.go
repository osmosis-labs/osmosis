package chain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/stretchr/testify/require"
)

func (n *NodeConfig) CreatePool(poolFile, from string) {
	n.t.Logf("creating pool from file %s from container %s", poolFile, n.Name)
	cmd := []string{"osmosisd", "tx", "gamm", "create-pool", fmt.Sprintf("--pool-file=/osmosis/%s", poolFile), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully created pool from container: %s", n.Name)
}

func (n *NodeConfig) SubmitUpgradeProposal(upgradeVersion string, upgradeHeight int64) {
	n.t.Logf("submitting upgrade proposal %s for height %d, from container: %s", upgradeVersion, upgradeHeight, n.Name)
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
	n.t.Logf("successfully submitted superfluid proposal for asset %s on container: %s", asset, n.Name)
}

func (n *NodeConfig) SubmitTextProposal(text string) {
	n.t.Logf("submitting text proposal from container: %s", n.Name)
	cmd := []string{"osmosisd", "tx", "gov", "submit-proposal", "--type=text", fmt.Sprintf("--title=\"%s\"", text), "--description=\"test text proposal\"", "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully submitted from container: %s", n.Name)
}

func (n *NodeConfig) DepositProposal(proposalNumber int) {
	n.t.Logf("depositing to proposal from container %s, on proposal: %d", n.Name, proposalNumber)
	cmd := []string{"osmosisd", "tx", "gov", "deposit", fmt.Sprintf("%d", proposalNumber), "500000000uosmo", "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully deposited from container %s, on proposal: %d", n.Name, proposalNumber)
}

func (n *NodeConfig) VoteYesProposal(from string, proposalNumber int) {
	n.t.Logf("voting yes on proposal from node container: %s, on proposal: %d", n.Name, proposalNumber)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "yes", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully voted yes from node container: %s, on proposal: %d", n.Name, proposalNumber)
}

func (n *NodeConfig) VoteNoProposal(from string, proposalNumber int) {
	n.t.Logf("voting no on proposal from node container: %s, on proposal: %d", n.Name, proposalNumber)
	cmd := []string{"osmosisd", "tx", "gov", "vote", fmt.Sprintf("%d", proposalNumber), "no", fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully voted no from node container: %s, on proposal: %d", n.Name, proposalNumber)
}

func (n *NodeConfig) LockTokens(tokens string, duration string, from string) {
	n.t.Logf("locking %s for %s on from container %s", tokens, duration, n.Name)
	cmd := []string{"osmosisd", "tx", "lockup", "lock-tokens", tokens, fmt.Sprintf("--duration=%s", duration), fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully created lock from container: %s", n.Name)
}

func (n *NodeConfig) SuperfluidDelegate(lockNumber int, valAddress string, from string) {
	lockStr := strconv.Itoa(lockNumber)
	n.t.Logf("superfluid delegating lock %s to %s from container %s", lockStr, valAddress, n.Name)
	cmd := []string{"osmosisd", "tx", "superfluid", "delegate", lockStr, valAddress, fmt.Sprintf("--from=%s", from)}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully superfluid delegated lock %s to %s from container: %s", lockStr, valAddress, n.Name)
}

func (n *NodeConfig) BankSend(amount string, sendAddress string, receiveAddress string) {
	n.t.Logf("bank sending %s from address %s to %s, from container %s", amount, sendAddress, receiveAddress, n.Name)
	cmd := []string{"osmosisd", "tx", "bank", "send", sendAddress, receiveAddress, amount, "--from=val"}
	_, _, err := n.containerManager.ExecTxCmd(n.t, n.chainId, n.Name, cmd)
	require.NoError(n.t, err)
	n.t.Logf("successfully sent bank sent %s from address %s to %s, from container %s", amount, sendAddress, receiveAddress, n.Name)
}

func (n *NodeConfig) CreateWallet(walletName string) string {
	n.t.Logf("creating wallet %s, from container %s", walletName, n.Name)
	cmd := []string{"osmosisd", "keys", "add", walletName, "--keyring-backend=test"}
	outBuf, _, err := n.containerManager.ExecCmd(n.t, n.Name, cmd, "")
	require.NoError(n.t, err)
	re := regexp.MustCompile("osmo1(.{38})")
	walletAddr := fmt.Sprintf("%s\n", re.FindString(outBuf.String()))
	walletAddr = strings.TrimSuffix(walletAddr, "\n")
	n.t.Logf("created wallet %s, waller address - %s from container %s", walletName, walletAddr, n.Name)
	return walletAddr
}
