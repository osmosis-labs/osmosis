package types

import exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

var _ CellStateI = (*ExampleCellState)(nil)

func (cell *ExampleCellState) Types() []interface{} {
	return nil
}

func (cell *ExampleCellState) Decls() []*exprpb.Decl {
	return nil
}

func (cell *ExampleCellState) Vars() map[string]interface{} {
	return nil
}
