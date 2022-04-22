package cmd

// DONTCOVER

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	tmstore "github.com/tendermint/tendermint/store"
	tmdb "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/tendermint/tendermint/config"
)

const (
	batchMaxSize      = 1000
	kValidators       = "validatorsKey:"
	kConsensusParams  = "consensusParamsKey:"
	kABCIResponses    = "abciResponsesKey:"
	fullHeight        = "full_height"
	minHeight         = "min_height"
	defaultFullHeight = "188000"
	defaultMinHeight  = "1000"
)

// get cmd to convert any bech32 address to an osmo prefix.
func forceprune() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forceprune",
		Short: "Example osmosisd forceprune -f 188000 -m 1000, which would keep blockchain and state data of last 188000 blocks (approximately 2 weeks) and ABCI responses of last 1000 blocks.",
		Long:  "Forceprune options prunes and compacts blockstore.db and state.db. One needs to shut down chain before running forceprune. By default it keeps last 188000 blocks (approximately 2 weeks of data) blockstore and state db (validator and consensus information) and 1000 blocks of abci responses from state.db. Everything beyond these heights in blockstore and state.db is pruned. ABCI Responses are stored in index db and so redundant especially if one is running pruned nodes. As a result we are removing ABCI data from state.db aggressively by default. One can override height for blockstore.db and state.db by using -f option and for abci response by using -m option. Example osmosisd forceprune -f 188000 -m 1000.",
		RunE: func(cmd *cobra.Command, args []string) error {
			full_height_flag, err := cmd.Flags().GetString(fullHeight)
			if err != nil {
				return err
			}

			min_height_flag, err := cmd.Flags().GetString(minHeight)
			if err != nil {
				return err
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			conf := config.DefaultConfig()
			db_path := clientCtx.HomeDir + "/" + conf.DBPath

			cmdr := exec.Command("osmosisd", "status")
			err = cmdr.Run()

			if err == nil {
				// continue only if throws errror
				return nil
			}

			full_height, err := strconv.ParseInt(full_height_flag, 10, 64)
			if err != nil {
				return err
			}

			min_height, err := strconv.ParseInt(min_height_flag, 10, 64)
			if err != nil {
				return err
			}

			opts := opt.Options{
				DisableSeeksCompaction: true,
			}

			db_bs, err := tmdb.NewGoLevelDBWithOpts("blockstore", db_path, &opts)
			if err != nil {
				return err
			}

			bs := tmstore.NewBlockStore(db_bs)
			start_height := bs.Base()
			current_height := bs.Height()

			fmt.Println("Pruning Block Store ...")
			prunedBlocks, err := bs.PruneBlocks(current_height - full_height)
			defer db_bs.Close()
			if err != nil {
				return err
			}
			fmt.Println("Pruned Block Store ...", prunedBlocks)
			db_bs.Close()

			fmt.Println("Compacting Block Store ...")

			db, err := leveldb.OpenFile(db_path+"/blockstore.db", &opts)
			defer db.Close()
			if err != nil {
				return err
			}
			if err = db.CompactRange(*util.BytesPrefix([]byte{})); err != nil {
				return err
			}

			db, err = leveldb.OpenFile(db_path+"/state.db", &opts)
			if err != nil {
				return err
			}
			stateDBKeys := []string{kValidators, kConsensusParams, kABCIResponses}
			fmt.Println("Pruning State Store ...")
			for i, s := range stateDBKeys {
				fmt.Println(i, s)

				retain_height := int64(0)
				if s == kABCIResponses {
					retain_height = current_height - min_height
				} else {
					retain_height = current_height - full_height
				}

				batch := new(leveldb.Batch)
				curBatchSize := uint64(0)

				fmt.Println(start_height, current_height, retain_height)

				for c := start_height; c < retain_height; c++ {
					batch.Delete([]byte(s + strconv.FormatInt(c, 10)))
					curBatchSize++

					if curBatchSize%batchMaxSize == 0 && curBatchSize > 0 {
						err := db.Write(batch, nil)
						if err != nil {
							return err
						}
						batch.Reset()
						batch = new(leveldb.Batch)
					}
				}

				err := db.Write(batch, nil)
				if err != nil {
					return err
				}
				batch.Reset()
			}

			fmt.Println("Compacting State Store ...")
			if err = db.CompactRange(*util.BytesPrefix([]byte{})); err != nil {
				return err
			}
			fmt.Println("Done ...")

			return nil
		},
	}

	cmd.Flags().StringP(fullHeight, "f", defaultFullHeight, "Full height to chop to")
	cmd.Flags().StringP(minHeight, "m", defaultMinHeight, "Min height for ABCI to chop to")
	return cmd
}
