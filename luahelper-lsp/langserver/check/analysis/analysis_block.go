package analysis

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
)

func (a *Analysis) cgBlock(node *ast.Block) {
	for _, stat := range node.Stats {
		a.cgStat(stat)
	}

	if node.RetExps != nil {
		a.cgRetStat(node.RetExps)
	}
}

func (a *Analysis) cgRetStat(exps []ast.Exp) {
	nExps := len(exps)
	if nExps == 0 {
		return
	}

	returnInfo := &common.ReturnInfo{}
	for _, exp := range exps {
		a.cgExp(exp, nil, nil)

		// 只在第一轮中才进行存储返回值
		// if !a.isFirstTerm() {
		// 	continue
		// }

		returnItem := common.ReturnItem{}
		returnItem.VarType = common.GetExpType(exp)
		returnItem.VarInfo = a.getExpReferVarInfo(exp)
		returnItem.ReturnExp = exp

		returnInfo.ReturnVarVec = append(returnInfo.ReturnVarVec, returnItem)

		a.funcReturnCheck(returnInfo)
	}

	fi := a.curFunc
	if a.isFirstTerm() {
		fi.InsertReturn(returnInfo)
	}
}
