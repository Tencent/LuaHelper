package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"strings"
)

func (a *Analysis) checkBinopExpTypeSame(node *ast.BinopExp) {
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorBinopType) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorBinopType]; !ok {
		return
	}

	if node.Op == lexer.TkOpMinus ||
		node.Op == lexer.TkOpAdd ||
		node.Op == lexer.TkOpMul ||
		node.Op == lexer.TkOpDiv {
		//go on
	} else {
		return
	}

	leftNode := node.Exp1
	rightNode := node.Exp2

	//当是表成员时，不判断，如a.b, a["b"]
	if _, ok := leftNode.(*ast.TableAccessExp); ok {
		return
	}
	if _, ok := rightNode.(*ast.TableAccessExp); ok {
		return
	}

	leftTypeVec := a.GetAnnTypeByExp(leftNode, -1)
	if len(leftTypeVec) == 0 {
		return
	}

	rightTypeVec := a.GetAnnTypeByExp(rightNode, -1)
	if len(rightTypeVec) == 0 {
		return
	}

	hasMatch := false
	for _, leftType := range leftTypeVec {
		for _, rightType := range rightTypeVec {
			if a.CompAnnTypeForBinop(rightType, leftType) {
				hasMatch = true
				break
			}
		}
	}

	if !hasMatch {
		loc := common.GetExpLoc(node)
		leftTypeVecStr := strings.Join(leftTypeVec, "|")
		rightTypeStr := strings.Join(rightTypeVec, "|")

		errStr := fmt.Sprintf("Binary expression has different type, '%s' with '%s'", leftTypeVecStr, rightTypeStr)
		a.curResult.InsertError(common.CheckErrorBinopType, errStr, loc)
	}
}
