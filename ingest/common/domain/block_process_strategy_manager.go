package commondomain

// BlockProcessStrategyManager is an interface for managing the strategy of pushing the blocks.
// Either all block data or only the block update are the possible options
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
// It is initialized with the strategy of pushing all data.
func NewBlockProcessStrategyManager() BlockProcessStrategyManager {
	return &blockProcessStrategyManager{
		shouldPushAllData: true,
	}
}

func (c *blockProcessStrategyManager) ShouldPushAllData() bool {
	return c.shouldPushAllData
}

func (c *blockProcessStrategyManager) MarkInitialDataIngested() {
	c.shouldPushAllData = false
}

func (c *blockProcessStrategyManager) MarkErrorObserved() {
	c.shouldPushAllData = true
}
