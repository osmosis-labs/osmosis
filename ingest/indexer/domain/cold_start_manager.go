package domain

// ColdStartManager is an interface for managing the cold start state of the indexer.
type ColdStartManager interface {
	// HasIngestedInitialData returns true if the indexer has ingested the initial data.
	HasIngestedInitialData() bool

	// MarkInitialDataIngested marks the initial data as ingested.
	MarkInitialDataIngested()
}

type coldStartManager struct {
	hasIngestedInitialData bool
}

var _ ColdStartManager = &coldStartManager{}

// NewColdStartManager creates a new cold start manager.
func NewColdStartManager() ColdStartManager {
	return &coldStartManager{}
}

func (c *coldStartManager) HasIngestedInitialData() bool {
	return c.hasIngestedInitialData
}

func (c *coldStartManager) MarkInitialDataIngested() {
	c.hasIngestedInitialData = true
}
