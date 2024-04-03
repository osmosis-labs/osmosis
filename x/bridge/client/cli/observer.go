package cli

import (
	"bufio"
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer"
	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer/bitcoin"
	"github.com/osmosis-labs/osmosis/v24/x/bridge/observer/osmosis"
)

func BridgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "Bridging-related operations",
	}
	cmd.AddCommand(
		StartObserverCmd(),
	)
	return cmd
}

func StartObserverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-observer",
		Short: "Start observer",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)

			inBuf := bufio.NewReader(cmd.InOrStdin())
			keyringBackend, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
			if err != nil {
				return err
			}

			// attempt to lookup address from Keybase if no address was provided
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf, clientCtx.Codec)
			if err != nil {
				return err
			}

			ctx := context.Background()

			grpcConn, err := grpcConnection("127.0.0.1:9090", true)
			if err != nil {
				return fmt.Errorf("can't obtain grpc conn to osmosis %s", err)
			}

			// Osmosis observer
			osmoClient := osmosis.NewClient(
				"my-test-chain",
				grpcConn,
				kb,
				clientCtx.TxConfig,
			)

			osmoRpcClient, err := rpchttp.New("tcp://127.0.0.1:26657", "/websocket") // tcp conn
			if err != nil {
				return fmt.Errorf("failed to initialize osmo ws connection %s", err)
			}

			osmoChainClient := osmosis.NewChainClient(
				serverCtx.Logger,
				osmoClient,
				osmoRpcClient,
				clientCtx.TxConfig,
				"osmo1xfjq5avug5g9579j5txyrqyft3gjh82wn5rg5a",
			)

			// Bitcoin observer
			btcRpcClient, err := rpcclient.New(&rpcclient.ConnConfig{
				Host:         "go.getblock.io/049662d399444608887621279811222c",
				DisableTLS:   false,
				HTTPPostMode: true,
				User:         "test",
				Pass:         "test",
				Params:       chaincfg.TestNet3Params.Name,
			}, nil)
			if err != nil {
				return fmt.Errorf("failed to initialize btc rpc client %s", err)
			}

			btcChainClient, err := bitcoin.NewChainClient(
				serverCtx.Logger,
				btcRpcClient,
				"2N4qEFwruq3zznQs78twskBrNTc6kpq87j1",
				time.Second,
				2584782,
				chaincfg.TestNet3Params,
			)
			if err != nil {
				return fmt.Errorf("failed to initialize btc observer %s", err)
			}

			// Observer
			obs := observer.NewObserver(serverCtx.Logger, map[observer.ChainId]observer.Client{
				observer.ChainIdBitcoin: btcChainClient,
				observer.ChainIdOsmosis: osmoChainClient,
			}, time.Second)

			// Start observer
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			err = obs.Start(ctx)
			if err != nil {
				return fmt.Errorf("failed start observer %s", err)
			}

			serverCtx.Logger.Info("Observer started")

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
			<-stop

			_ = obs.Stop(ctx) // returns no err

			serverCtx.Logger.Info("Observer stopped")

			return nil
		},
	}

	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")

	return cmd
}

func grpcConnection(url string, disableTls bool) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if disableTls {
		creds = insecure.NewCredentials()
	} else {
		certs, err := x509.SystemCertPool()
		if err != nil {
			return nil, errorsmod.Wrapf(err, "Failed to load system certificates")
		}
		creds = credentials.NewClientTLSFromCert(certs, "")
	}
	return grpc.Dial(url, grpc.WithTransportCredentials(creds))
}
