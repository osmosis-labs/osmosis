package stats

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/cosmos/cosmos-sdk/types/simulation"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

//go:embed schema.sql
var embedFs embed.FS

// TODO: Setup an sql schema file instead of doing it in-line here
type StatsDb struct {
	enabled bool
	db      *sql.DB
}

func SetupStatsDb(config ExportConfig) (StatsDb, error) {
	if !config.WriteStatsToDB {
		return StatsDb{enabled: false}, nil
	}

	setupSqlCmd, err := embedFs.ReadFile("schema.sql")
	if err != nil {
		return StatsDb{}, fmt.Errorf("error in reading schema.sql: %w", err)
	}

	db, err := sql.Open("sqlite3", "./blocks.db")
	if err != nil {
		return StatsDb{}, err
	}

	if _, err := db.Exec(string(setupSqlCmd)); err != nil {
		db.Close()
		return StatsDb{}, fmt.Errorf("error in init from schema.sql init: %w", err)
	}

	return StatsDb{enabled: true, db: db}, nil
}

func (stats StatsDb) Cleanup() {
	if stats.db != nil {
		stats.db.Close()
	}
}

func (stats StatsDb) LogActionResult(header tmproto.Header, opMsg simulation.OperationMsg, resultData []byte) error {
	if !stats.enabled {
		return nil
	}
	appHash := fmt.Sprintf("%X", header.AppHash)
	resData := fmt.Sprintf("%X", resultData)
	sts := "INSERT INTO blocks(height,module,name,comment,passed, gasWanted, gasUsed, msg, resData, appHash) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);"
	_, err := stats.db.Exec(sts, header.Height, opMsg.Route, opMsg.Name, opMsg.Comment, opMsg.OK, opMsg.GasWanted, opMsg.GasUsed, opMsg.Msg, resData, appHash)
	return err
}
