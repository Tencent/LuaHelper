package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// 是否给常量赋值 当有a.b.c时，只判断a是否常量 若a是import值，则只判断b是否常量
func (a *Analysis) checkEnum(node *ast.TableConstructorExp) {
	if len(node.KeyExps) == 0 {
		return
	}

	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorEnumValue) {
		return
	}

	fileResult := a.curResult

	// 判断枚举值是否有关联
	if !a.Projects.IsLineFragementEnum(fileResult.Name, node.Loc.StartLine) {
		return
	}

	var tableEnumList common.TableEnumList
	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]
		if valExp == nil {
			continue
		}

		strKeySimple := common.GetExpName(keyExp)
		itemNum, flag := tableEnumList.CheckEnumExp(strKeySimple, valExp)
		if !flag {
			tableEnumList.AddEnumVar(strKeySimple, keyExp, valExp)
			continue
		}

		errStr := fmt.Sprintf("table: enum %s and %s contains duplicate value", itemNum.FieldStr, strKeySimple)
		var relateVec []common.RelateCheckInfo
		oldKeyLoc := common.GetExpLoc(itemNum.KeyExp)
		oldValueLoc := common.GetExpLoc(itemNum.ValueExp)
		oldRangeLoc := lexer.GetRangeLoc(&oldKeyLoc, &oldValueLoc)

		relateVec = append(relateVec, common.RelateCheckInfo{
			LuaFile: fileResult.Name,
			ErrStr:  errStr,
			Loc:     oldRangeLoc,
		})

		expKeyLoc := common.GetExpLoc(keyExp)
		expValueLoc := common.GetExpLoc(valExp)
		rangeLoc := lexer.GetRangeLoc(&expKeyLoc, &expValueLoc)
		fileResult.InsertRelateError(common.CheckErrorEnumValue, errStr, rangeLoc, relateVec)
	}
}
