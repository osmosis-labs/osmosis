package cmd

// DONTCOVER

import (
  "github.com/spf13/cobra"

  "fmt"
  "strconv"

  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"
  "github.com/syndtr/goleveldb/leveldb/util"
  tmstore "github.com/tendermint/tendermint/store"
  tmdb "github.com/tendermint/tm-db"

  "os/exec"

  "github.com/cosmos/cosmos-sdk/client"
  "github.com/tendermint/tendermint/config"
)

var fheight = "full_height"
var mheight = "min_height"

// get cmd to convert any bech32 address to an osmo prefix.
func forceprune() *cobra.Command {
  cmd := &cobra.Command{
    Use:   "forceprune",
    Short: "forceprune",
    Long: `forceprune
Example:
  osmosisd forceprune -f 188000 -m 1000
  `,
    RunE: func(cmd *cobra.Command, args []string) error {

      full_height_flag, err := cmd.Flags().GetString(fheight)
      if err != nil {
        return err
      }

      min_height_flag, err := cmd.Flags().GetString(mheight)
      if err != nil {
        return err
      }

      clientCtx := client.GetClientContextFromCmd(cmd)
      conf := config.DefaultConfig()
      db_path := string(clientCtx.HomeDir) + "/" + string(conf.DBPath)

      cmdr := exec.Command("osmosisd", "status")
      err = cmdr.Run()

      if err == nil {
        // continue only if throws errror
        return nil
      }

      full_height, err := strconv.ParseInt(full_height_flag, 10, 64)
      if err != nil {
        panic(err)
      }
      min_height, err := strconv.ParseInt(min_height_flag, 10, 64)
      if err != nil {
        panic(err)
      }

      o := opt.Options{
        DisableSeeksCompaction: true,
      }

      db_bs, err := tmdb.NewGoLevelDBWithOpts("blockstore", db_path, &o)
      if err != nil {
        panic(err)
      }

      bs := tmstore.NewBlockStore(db_bs)
      start_height := bs.Base()
      current_height := bs.Height()

      fmt.Println("Pruning Block Store ...")
      bs.PruneBlocks(current_height - int64(full_height))
      fmt.Println("Compacting Block Store ...")
      db_bs.Close()

      db, err := leveldb.OpenFile(db_path+"/blockstore.db", &o)
      if err != nil {
        panic(err)
      }
      if err = db.CompactRange(*util.BytesPrefix([]byte{})); err != nil {
        panic(err)
      }

      db, err = leveldb.OpenFile(db_path+"/state.db", &o)
      if err != nil {
        panic(err)
      }
      a := []string{"validatorsKey:", "consensusParamsKey:", "abciResponsesKey:"}
      fmt.Println("Pruning State Store ...")
      for i, s := range a {
        fmt.Println(i, s)

        retain_height := int64(0)
        if s == "abciResponsesKey:" {
          retain_height = current_height - int64(min_height)
        } else {
          retain_height = current_height - int64(full_height)
        }

        batch := new(leveldb.Batch)
        pruned := uint64(0)

        fmt.Println(start_height, current_height, retain_height)
        for c := start_height; c < retain_height; c++ {
          batch.Delete([]byte(s + strconv.FormatInt(c, 10)))
          pruned++

          if pruned%1000 == 0 && pruned > 0 {
            err := db.Write(batch, nil)
            if err != nil {
              panic(err)
            }
            batch.Reset()
            batch = new(leveldb.Batch)
          }
        }

        err := db.Write(batch, nil)
        if err != nil {
          panic(err)
        }
        batch.Reset()
      }
      fmt.Println("Compacting State Store ...")
      if err = db.CompactRange(*util.BytesPrefix([]byte{})); err != nil {
        panic(err)
      }
      fmt.Println("Done ...")

      return nil
    },
  }

  cmd.Flags().StringP(fheight, "f", "188000", "Full height to chop to")
  cmd.Flags().StringP(mheight, "m", "1000", "Min height for ABCI to chop to")
  return cmd
}

