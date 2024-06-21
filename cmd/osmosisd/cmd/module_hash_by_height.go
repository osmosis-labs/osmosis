package cmd

// DONTCOVER

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// forceprune gets cmd to convert any bech32 address to an osmo prefix.
func moduleHashByHeightQuery(appCreator servertypes.AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-hash-by-height [height]",
		Short: "Get module hashes at a given height",
		Long: `Get module hashes at a given height. This command is useful for debugging and verifying the state of the application at a given height.
Example:
	osmosisd module-hash-by-height 16841115,
`,
		Args: cobra.ExactArgs(1), // Ensure exactly one argument is provided
		RunE: func(cmd *cobra.Command, args []string) error {
			heightToRetrieveString := args[0]

			serverCtx := server.GetServerContextFromCmd(cmd)

			height, err := strconv.ParseInt(heightToRetrieveString, 10, 64)
			if err != nil {
				return err
			}

			commitInfoForHeight, err := getModuleHashesAtHeight(serverCtx, appCreator, height)
			if err != nil {
				return err
			}

			// Print the CommitInfo to the console.
			fmt.Println(commitInfoForHeight.String())

			return nil
		},
	}

	return cmd
}

func getModuleHashesAtHeight(svrCtx *server.Context, appCreator servertypes.AppCreator, height int64) (*storetypes.CommitInfo, error) {
	home := svrCtx.Config.RootDir
	db, err := openDB(home, server.GetAppDBBackend(svrCtx.Viper))
	if err != nil {
		return nil, fmt.Errorf("error opening DB, make sure osmosisd is not running when calling this query: %w", err)
	}
	app := appCreator(svrCtx.Logger, db, nil, svrCtx.Viper)

	commitInfoForHeight, err := app.CommitMultiStore().GetCommitInfo(height)
	if err != nil {
		return nil, err
	}

	// Create a new slice of StoreInfos for storing the modified hashes.
	storeInfos := make([]storetypes.StoreInfo, len(commitInfoForHeight.StoreInfos))

	for i, storeInfo := range commitInfoForHeight.StoreInfos {
		// Convert the hash to a hexadecimal string.
		hash := strings.ToUpper(hex.EncodeToString(storeInfo.CommitId.Hash))

		// Create a new StoreInfo with the modified hash.
		storeInfos[i] = storetypes.StoreInfo{
			Name: storeInfo.Name,
			CommitId: storetypes.CommitID{
				Version: storeInfo.CommitId.Version,
				Hash:    []byte(hash),
			},
		}
	}

	// Sort the storeInfos slice based on the module name.
	sort.Slice(storeInfos, func(i, j int) bool {
		return storeInfos[i].Name < storeInfos[j].Name
	})

	// Create a new CommitInfo with the modified StoreInfos.
	commitInfoForHeight = &storetypes.CommitInfo{
		Version:    commitInfoForHeight.Version,
		StoreInfos: storeInfos,
	}

	return commitInfoForHeight, nil
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}
