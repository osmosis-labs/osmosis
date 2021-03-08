package testutil

import (
	"bytes"
	"context"
	"fmt"

	distrcli "github.com/c-osmosis/osmosis/x/distribution/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
)

func MsgWithdrawDelegatorRewardExec(clientCtx client.Context, valAddr fmt.Stringer, extraArgs ...string) ([]byte, error) {
	buf := new(bytes.Buffer)
	clientCtx = clientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	args := []string{valAddr.String()}
	args = append(args, extraArgs...)

	cmd := distrcli.NewWithdrawRewardsCmd()
	cmd.SetErr(buf)
	cmd.SetOut(buf)
	cmd.SetArgs(args)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
