package types

import (
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type CellStateI interface {
	Types() []interface{}         // provided to cel.Types()
	Decls() []*exprpb.Decl        // provided to cel.Declarations()
	Vars() map[string]interface{} // provided to cel.Program.Eval()
}

func NewExpr(expr string, output *exprpb.Type, inputs ...*exprpb.Decl) *Expr {
	return &Expr{
		Expr:   expr,
		Inputs: inputs,
		Output: output,
	}
}
