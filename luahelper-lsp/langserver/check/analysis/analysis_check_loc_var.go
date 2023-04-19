package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
)

func (a *Analysis) checkLocVarCall() {

	// 如果是一轮校验，判断是否要校验局部变量是否定义未使用
	if !a.isFirstTerm() || a.realTimeFlag {
		return
	}

	// 判断是否开启了局部变量定义了是否未使用的告警
	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorLocalNoUse) {
		return
	}

	scope := a.curScope
	if scope == nil {
		return
	}

	fileResult := a.curResult

	// 扫描当前scope，判断哪些局部变量定义了未使用
	for varName, varInfoList := range scope.LocVarMap {
		// _ 局部变量忽略, _G也忽略
		if varName == "_" || varName == "_G" {
			continue
		}

		if common.GConfig.IsIgnoreLocNotUseVar(varName) {
			// 如果为系统忽略的局部变量定义了，未使用的，忽略掉
			continue
		}

		for _, oneVar := range varInfoList.VarVec {
			if oneVar.IsUse || oneVar.IsClose {
				continue
			}

			//定义的局部函数忽略
			if oneVar.ReferFunc != nil {
				continue
			}

			// 判断指向的关联变量，是否为系统的函数或模块
			// 例如 local math = math 这样的忽略掉
			expName := common.GetExpName(oneVar.ReferExp)

			// 1) 判断是否直接关联到的系统模块或函数
			oneStr := common.GetExpSubKey(expName)
			if oneStr != "" {
				if common.GConfig.IsInSysNotUseMap(oneStr) {
					// 为系统的模块或函数名，忽略掉
					continue
				}
			}

			// 2) 判断是否关联到系统模块的成员， 例如：local concat = table.concat
			flagG, strRet := common.StrRemovePreG(expName)
			if flagG {
				expName = "!" + strRet
			}
			moduleName, keyName := common.GetTableStrTwoStr(expName)
			if moduleName != "" && keyName != "" {
				if common.GConfig.IsInSysNotUseMap(moduleName) {
					// 为系统的模块或函数名，忽略掉
					continue
				}
			}

			errorStr := fmt.Sprintf("%s declared and not used", varName)
			fileResult.InsertError(common.CheckErrorLocalNoUse, errorStr, oneVar.Loc)

			// 遍历所有的定义了未使用，只是简单的赋值
			for _, subVar := range oneVar.NoUseAssignLocs {
				errorStr := fmt.Sprintf("%s declared and not used, this just assign", varName)
				fileResult.InsertError(common.CheckErrorNoUseAssign, errorStr, subVar)
			}

			// 清除掉
			oneVar.NoUseAssignLocs = nil
		}
	}
}
