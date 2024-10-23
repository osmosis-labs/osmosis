package domain

// BlockProcessStrategyManager is an interface for managing the strategy of pushing the blocks.
// Either all block data or only the block update are the possible options
// It is initialized with the strategy of pushing all data.
// If it observes an error, it will switch to pushing all data.
// If it ingested initial data and observed no error, it will switch to pushing only changed data.
type BlockProcessStrategyManager interface {
	// ShouldPushAllData returns true if all data should be pushed.
	ShouldPushAllData() bool

	// MarkInitialDataIngested marks the initial data as ingested.
	// After calling this function, ShouldPushAllData should return false.
	MarkInitialDataIngested()

	// MarkErrorObserved marks that an error has been observed.
	MarkErrorObserved()
}

type blockProcessStrategyManager struct {
	shouldPushAllData bool
}

var _ BlockProcessStrategyManager = &blockProcessStrategyManager{}

// NewBlockProcessStrategyManager creates a new push strategy manager.
func NewBlockProcessStrategyManager() BlockProcessStrategyManager {
	return &blockProcessStrategyManager{
		shouldPushAllData: true,
	}
}

// ShouldPushAllData returns true if all data should be pushed.
func (c *blockProcessStrategyManager) ShouldPushAllData() bool {
	return c.shouldPushAllData
}

// MarkInitialDataIngested marks the initial data as ingested.
func (c *blockProcessStrategyManager) MarkInitialDataIngested() {
	c.shouldPushAllData = false
}

// MarkErrorObserved marks that an error has been observed.
func (c *blockProcessStrategyManager) MarkErrorObserved() {
	c.shouldPushAllData = true
}
