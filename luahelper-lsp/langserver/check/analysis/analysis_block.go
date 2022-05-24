package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/annotation/annotateast"
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

		a.cgFuncReturnCheck(returnInfo)
	}

	fi := a.curFunc
	if a.isFirstTerm() {
		fi.InsertReturn(returnInfo)
	}
}

// 检查函数的返回值类型与注解类型是否匹配 一次检查一个return语句
func (a *Analysis) cgFuncReturnCheck(retInfo *common.ReturnInfo) {
	// 第二轮或第三轮函数参数check
	if !a.isNeedCheck() {
		return
	}

	// 判断是否开启了函数调用参数个数不匹配的校验
	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorFuncRetErr) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorFuncRetErr]; !ok {
		return
	}

	if retInfo == nil || a.curFunc == nil {
		return
	}

	//获取注解的返回值类型
	annRetVec := a.Projects.GetFuncReturnType(a.curFunc.FileName, a.curFunc.Loc.StartLine-1)

	if len(annRetVec) == 0 {
		//无注解不检查
		return
	}

	for i, oneReturn := range retInfo.ReturnVarVec {

		if i >= len(annRetVec) {
			//实际返回值个数超过注解中的返回值数量
			continue
		}

		//合并该返回值的注解类型 合成一个字符串
		allAnnTypeStr := ""
		for _, annRetTypeOne := range annRetVec[i] {
			typeOne := annotateast.GetAstTypeName(annRetTypeOne)
			if len(allAnnTypeStr) > 0 {
				allAnnTypeStr = fmt.Sprintf("%s|", allAnnTypeStr)
			}
			allAnnTypeStr = fmt.Sprintf("%s%s", allAnnTypeStr, typeOne)
		}

		loc := common.GetExpLoc(oneReturn.ReturnExp)
		hasMatch := false
		allTypeStr := ""

		//获取return处表达式的返回值类型
		retType := common.GetAnnTypeFromLuaType(oneReturn.VarType)
		if retType == "LuaTypeRefer" {

			// 取return处的引用
			refRetTypeVec := a.GetAnnTypeStrForRefer(oneReturn.ReturnExp, -1)
			if len(refRetTypeVec) == 0 {
				// 无法识别类型
				continue
			}

			for _, retRetType := range refRetTypeVec {
				if len(allTypeStr) > 0 {
					allTypeStr = fmt.Sprintf("%s|", allTypeStr)
				}
				allTypeStr = fmt.Sprintf("%s%s", allTypeStr, retRetType)
			}

			//注解返回值可以有多个类型 如---@return number, string|number 第二个返回值可以是字符串或者数
			for _, annRetTypeOne := range annRetVec[i] {
				annType := annotateast.GetAstTypeName(annRetTypeOne)
				for _, retRetType := range refRetTypeVec {
					if a.CompAnnTypeAndCodeType(annType, retRetType) {
						hasMatch = true
						break
					}
				}

				if hasMatch {
					break
				}
			}

		} else {
			allTypeStr = retType
			//注解返回值可以有多个类型 如---@return number, string|number 第二个返回值可以是字符串或者数
			for _, annRetTypeOne := range annRetVec[i] {
				annType := annotateast.GetAstTypeName(annRetTypeOne)

				if a.CompAnnTypeAndCodeType(annType, retType) {
					hasMatch = true
					break
				}
			}
		}

		if !hasMatch {

			//类型不一致，报警
			errorStr := fmt.Sprintf("Return value is expected to be '%s', '%s' returned", allAnnTypeStr, allTypeStr)

			a.curResult.InsertError(common.CheckErrorFuncRetErr, errorStr, loc)
		}
	}
}
