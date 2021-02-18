package types

var _ CellI = ExampleCell{}

func (cell ExampleCell) CellState() CellState {
	return CellState{}
}
