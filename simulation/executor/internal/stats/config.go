package stats

type ExportConfig struct {
	ExportParamsPath   string // custom file path to save the exported params JSON
	ExportParamsHeight int    // height to which export the randomly generated params
	ExportStatePath    string // custom file path to save the exported app state JSON
	ExportStatsPath    string // custom file path to save the exported simulation statistics JSON
	WriteStatsToDB     bool
}
