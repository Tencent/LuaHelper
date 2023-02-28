package analysis

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
)

func (a *Analysis) findVarDefine(varName string, varLoc lexer.Location) (find bool, varInfo *common.VarInfo) {
	//先尝试找local变量
	varInfo, find = a.curScope.FindLocVar(varName, varLoc)
	if find {
		return find, varInfo
	}

	return a.findVarDefineGlobal(varName)
}

// 从全局查找变量定义
func (a *Analysis) findVarDefineGlobal(varName string) (find bool, varInfo *common.VarInfo) {

	//没找到就找全局变量
	fi := a.curFunc
	firstFile := a.getFirstFileResult(a.curResult.Name)

	gFlag := false
	strName := varName
	strProPre := ""

	fileResult := a.curResult
	if a.isSecondTerm() {
		secondFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			find, varInfo = secondFileResult.FindGlobalVarInfo(strName, gFlag, strProPre)
			if !find {
				find, varInfo = a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermSecond, strProPre)
			}
		} else {
			// 非底层的函数，需要查找全局的变量
			find, varInfo = firstFile.FindGlobalVarInfo(strName, gFlag, strProPre)
			if !find {
				find, varInfo = a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermFirst, strProPre)
			}
		}
	} else if a.isThirdTerm() {
		thirdFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			find, varInfo = thirdFileResult.FindGlobalVarInfo(strName, gFlag, strProPre)
		} else {
			// 非底层的函数，需要查找全局的变量
			find, varInfo = firstFile.FindGlobalVarInfo(strName, gFlag, strProPre)
		}

		// 查找所有的
		if !find {
			find, varInfo = a.AnalysisThird.ThirdStruct.FindThirdGlobalGInfo(gFlag, strName, strProPre)
		}
	}

	return find, varInfo
}

// findVarDefineWithPre 查找检查
// 如果preName空，查找varName
// 如果preName是import值 查找varName
// 如果preName非import值 根据findSub 查找preName定义或者preName的成员即varName的定义
func (a *Analysis) findVarDefineWithPre(preName string, varName string, preLoc lexer.Location, varLoc lexer.Location, findSub bool) (find bool, varInfo *common.VarInfo, isPreImport bool) {
	find = false

	if preName != "" {
		//有前缀

		ok, preInfo := a.findVarDefine(preName, preLoc)
		if !ok {
			return
		}

		referInfo := preInfo.ReferInfo
		if referInfo == nil {
			// 前缀非import变量

			if findSub && varName != "" {

				subVar, ok := preInfo.SubMaps[varName]
				return ok, subVar, false
			} else {
				return ok, preInfo, false
			}
		}

		// 前缀是模块变量 在模块中再查找
		if len(varName) <= 0 {
			return
		}

		referFile := a.Projects.GetFirstReferFileResult(referInfo)
		if referFile == nil {
			// 文件不存在
			return
		}

		find, varInfo = referFile.FindGlobalVarInfo(varName, false, "")
		return find, varInfo, true
	}

	//无前缀 直接找定义
	if len(varName) <= 0 {
		return
	}

	find, varInfo = a.findVarDefine(varName, varLoc)
	return find, varInfo, false
}
