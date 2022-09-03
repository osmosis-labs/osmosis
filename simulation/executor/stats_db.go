package simulation

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type statsDb struct {
	enabled bool
	db      *sql.DB
}

func setupStatsDb(config ExportConfig) (statsDb, error) {
	if !config.WriteStatsToDB {
		return statsDb{enabled: false}, nil
	}
	db, err := sql.Open("sqlite3", "./blocks.db")
	if err != nil {
		return statsDb{}, err
	}

	sts := `
	DROP TABLE IF EXISTS blocks;
	CREATE TABLE blocks (id INTEGER PRIMARY KEY, height INT,module TEXT, name TEXT, comment TEXT, passed BOOL, gasWanted INT, gasUsed INT, msg STRING, resData STRING, appHash STRING);
	`
	_, err = db.Exec(sts)

	if err != nil {
		db.Close()
		return statsDb{}, err
	}
	return statsDb{enabled: true, db: db}, nil
}

func (stats statsDb) cleanup() {
	stats.db.Close()
}

func (stats statsDb) logActionResult(header tmproto.Header, opMsg simulation.OperationMsg, resultData []byte) error {
	if !stats.enabled {
		return nil
	}
	appHash := fmt.Sprintf("%X", header.AppHash)
	resData := fmt.Sprintf("%X", resultData)
	sts := "INSERT INTO blocks(height,module,name,comment,passed, gasWanted, gasUsed, msg, resData, appHash) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);"
	_, err := stats.db.Exec(sts, header.Height, opMsg.Route, opMsg.Name, opMsg.Comment, opMsg.OK, opMsg.GasWanted, opMsg.GasUsed, opMsg.Msg, resData, appHash)
	return err
}
