package analysis

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
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

//GetAnnTypeStrForRefer 获取表达式类型字符串，如果是引用，则递归查找，(即支持类型传递)
func (a *Analysis) GetAnnTypeStrForRefer(referExp ast.Exp, idx int) (retVec []string) {

	argType := common.GetAnnTypeFromExp(referExp)
	if argType != "LuaTypeRefer" {
		retVec = append(retVec, argType)
		return
	}

	//若是引用，则继续查找定义
	preName := ""
	varName := ""
	keyName := ""
	preLoc := lexer.Location{}
	varLoc := lexer.Location{}
	isTableExp := false
	isTableWhole := false
	switch exp := referExp.(type) {
	case *ast.NameExp:
		varName = exp.Name
		varLoc = exp.Loc
	case *ast.TableAccessExp:
		preName, varName, keyName, preLoc, varLoc, isTableWhole = common.GetTableNameInfo(exp)
		isTableExp = true
	case *ast.FuncCallExp:
		if nameExp, ok := exp.PrefixExp.(*ast.NameExp); ok && exp.NameExp == nil {
			varName = nameExp.Name
			varLoc = nameExp.Loc
		}
	}

	if isTableExp {
		//当是表且有截断 不继续推导类型
		if !isTableWhole {
			return
		}
	}

	ok, varInfo, _ := a.findVarDefineWithPre(preName, varName, preLoc, varLoc, true)
	if !ok {
		return
	}

	varIdx := int(varInfo.VarIndex)
	if idx > 0 {
		varIdx = idx
	}

	//优先取变量定义处的注解类型
	defAnnTypeVec := a.Projects.GetAnnotateTypeString(varInfo, varName, keyName, varIdx)
	if len(defAnnTypeVec) > 0 {
		return defAnnTypeVec
	}

	//若无注解，则取变量定义处表达式推导的类型
	//如果是表exp，到这里已经是：表exp完整，如果有keyname就取keyname的类型
	if isTableExp && len(keyName) > 0 {
		//必须取到keyName的类型 否则退出
		if keyVarInfo, ok := varInfo.SubMaps[keyName]; ok {
			argType = common.GetAnnTypeFromLuaType(keyVarInfo.VarType)
		} else {
			return
		}
	} else {
		argType = common.GetAnnTypeFromLuaType(varInfo.VarType)
	}

	//例如：
	//---@type classA
	//local tableA = {}
	//argType是table, defAnnType是classA,
	//当tableA作为参数时，table或者classA都可以匹配

	if argType == "LuaTypeRefer" && isTableWhole {
		//若仍是LuaTypeRefer 且完整解析了table 可以递归
		//table的递归会导致栈溢出，先屏蔽
		if !isTableExp {
			return a.GetAnnTypeStrForRefer(varInfo.ReferExp, varIdx)
		}
	}

	retVec = append(retVec, argType)
	return retVec
}

//
func (a *Analysis) loadFuncParamAnnType(referFunc *common.FuncInfo) {
	if referFunc == nil {
		return
	}

	// 在此获取函数参数的注解类型 并保存
	paramTypeMap := a.Projects.GetFuncParamType(referFunc.FileName, referFunc.Loc.StartLine-1)
	if len(paramTypeMap) > 0 {
		//从函数上方获取到了注解
		for _, paramName := range referFunc.ParamList {
			if annTypeVec, ok := paramTypeMap[paramName]; ok {
				for _, annTypeOne := range annTypeVec {
					typeOne := annotateast.GetAstTypeName(annTypeOne)

					referFunc.ParamType[paramName] = append(referFunc.ParamType[paramName], typeOne)
				}
			}
		}

		return //从函数上方获取到了注解之后就不再查找类成员函数的注解
	}

	if referFunc.ClassName != "" {
		// 如果是类的成员函数 先找类变量的定义
		find, varDefine := a.findVarDefineGlobal(referFunc.ClassName)
		if !find {
			return
		}

		// 再找类上方注解
		classTypeStr, ok := a.Projects.GetVarAnnType(varDefine.FileName, varDefine.Loc.StartLine-1)
		if !ok {
			return
		}

		// 根据注解找成员函数
		paramTypeMapByClass := a.Projects.GetFuncParamTypeByClass(classTypeStr, referFunc.FuncName)
		if len(paramTypeMapByClass) == 0 {
			return
		}

		for _, paramName := range referFunc.ParamList {
			if annTypeStr, ok := paramTypeMapByClass[paramName]; ok {
				referFunc.ParamType[paramName] = annTypeStr
			}
		}

		return
	}

	return
}
