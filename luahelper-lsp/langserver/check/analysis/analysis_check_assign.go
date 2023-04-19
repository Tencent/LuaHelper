package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"strings"
)

// 判断复制表达式是否改变了类型
func (a *Analysis) checkAssignTypeSame(leftNode ast.Exp, rightNode ast.Exp) {
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorAssignType) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorAssignType]; !ok {
		return
	}

	//当左值是表成员时，不判断，如a.b, a["b"]
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
			if a.CompAnnTypeForAssign(rightType, leftType) {
				hasMatch = true
				break
			}
		}
	}

	if !hasMatch {
		loc := common.GetExpLoc(leftNode)
		leftTypeVecStr := strings.Join(leftTypeVec, "|")
		rightTypeStr := strings.Join(rightTypeVec, "|")

		errStr := fmt.Sprintf("Type '%s' can not be assigned '%s'", leftTypeVecStr, rightTypeStr)
		a.curResult.InsertError(common.CheckErrorAssignType, errStr, loc)
	}
}
