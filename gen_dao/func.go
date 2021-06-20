package gen_dao

import (
	"go/ast"
	"log"
	"strings"
)

type AnalysisContext interface {
}

type FunctionAnalysis struct {
	targetFunc FunctionDescription
	ctx        AnalysisContext
	usedVar    []*UsedTypeDesc
}

func (m *FunctionAnalysis) Analysis(desc FunctionDescription) {
	m.targetFunc = desc
}

func isCameFromGorm(value UsedTypeDesc) bool {
	if strings.Contains(value.TypeDescription.Package, "gorm") {
		return true
	}

	return false
}

func (m *FunctionAnalysis) parseExpr(expr ast.Expr) {

}

func (m *FunctionAnalysis) listAllCall() (functions []FunctionDescription) {
	for _, stmt := range m.targetFunc.FuncDecl.Body.List {
		switch realStmt := stmt.(type) {
		case *ast.AssignStmt:
			//check right

		case *ast.ExprStmt:
			m.parseExpr(realStmt.X)

		default:
			log.Println("not concerned")
		}

	}

	return nil
}

func (m *FunctionAnalysis) HasUsePackage(packID string) bool {
	for _, param := range m.targetFunc.Params {
		if isCameFromGorm(param.UsedTypeDesc) {
			return true
		}
	}

	for _, ret := range m.targetFunc.Ret {
		if isCameFromGorm(ret.UsedTypeDesc) {
			return true
		}
	}

	//check used function

	return false
}

/*

a object used to analysis function implement

variable declare ,
variable set

call other function


return
*/
type VariableDescription struct {
	Name string
	Type *TypeDescription
}

type FunctionStackAnalysis struct {
	desc       *FunctionDescription
	declareVar []*VariableDescription
}

func (f *FunctionStackAnalysis) Analysis() {

}
