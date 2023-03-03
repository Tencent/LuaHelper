package analysis

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
)

func (a *Analysis) findVarDefine(varName string, varLoc lexer.Location) (find bool, varInfo *common.VarInfo, varType []string) {
	//先尝试找local变量
	varInfo, find = a.curScope.FindLocVar(varName, varLoc)
	if find {
		if varInfo.IsParam {
			// 如果是函数参数 同时返回参数类型
			varType, ok := a.curFunc.ParamType[varName]
			if ok {
				return find, varInfo, varType
			}

		}
		return find, varInfo, varType
	}

	fromTerm := 0
	find, varInfo, fromTerm = a.findVarDefineGlobal(varName)

	if find && fromTerm == 1 && varInfo.ReferFunc != nil {
		a.loadFuncParamAnnType(varInfo.ReferFunc)
	}

	return find, varInfo, varType
}

// 从全局查找变量定义
func (a *Analysis) findVarDefineGlobal(varName string) (find bool, varInfo *common.VarInfo, fromTerm int) {
	gFlag := false
	strName := varName
	strProPre := ""

	fi := a.curFunc
	fileResult := a.curResult
	// 4) 第三步查找全局中是否有该变量
	firstFile := a.getFirstFileResult(fileResult.Name)
	if a.isSecondTerm() {
		secondFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			if ok, oneVar := secondFileResult.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return ok, oneVar, 2
			}

			if ok, oneVar := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermSecond, strProPre); ok {
				return ok, oneVar, 2
			}
		} else {
			// 非底层的函数，需要查找全局的变量
			if ok, oneVar := firstFile.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return ok, oneVar, 1
			}

			if ok, oneVar := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermFirst, strProPre); ok {
				return ok, oneVar, 1
			}
		}
	} else if a.isThirdTerm() {
		thirdFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			if ok, oneVar := thirdFileResult.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return ok, oneVar, 3
			}
		} else {
			// 非底层的函数，需要查找全局的变量
			if ok, oneVar := firstFile.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return ok, oneVar, 1
			}
		}

		// 查找所有的
		if ok, oneVar := a.AnalysisThird.ThirdStruct.FindThirdGlobalGInfo(gFlag, strName, strProPre); ok {
			return ok, oneVar, 3
		}
	}
	return find, varInfo, 0
}

// findVarDefineWithPre 查找定义 并尝试带出其类型
// 如果preName空，查找varName
// 如果preName是import值 查找varName
// 如果preName非import值 根据findSub 查找preName定义或者preName的成员即varName的定义
func (a *Analysis) findVarDefineWithPre(preName string, varName string, preLoc lexer.Location, varLoc lexer.Location, findSub bool) (find bool, varInfo *common.VarInfo, isPreImport bool, varType []string) {
	find = false

	if preName != "" {
		//有前缀

		ok, preInfo, _ := a.findVarDefine(preName, preLoc)
		if !ok {
			return
		}

		referInfo := preInfo.ReferInfo
		if referInfo == nil {
			// 前缀非import变量

			if findSub && varName != "" {

				subVar, ok := preInfo.SubMaps[varName]
				return ok, subVar, false, nil
			} else {
				return ok, preInfo, false, nil
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
		return find, varInfo, true, nil
	}

	//无前缀 直接找定义
	if len(varName) <= 0 {
		return
	}

	find, varInfo, varType = a.findVarDefine(varName, varLoc)
	return find, varInfo, false, varType
}

//GetAnnTypeByExp 获取表达式类型字符串，如果是引用，则递归查找，(即支持类型传递)
func (a *Analysis) GetAnnTypeByExp(referExp ast.Exp, idx int) (retVec []string) {
	expType := common.GetExpType(referExp)
	argType := common.GetAnnTypeFromLuaType(expType)

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
	isFuncCall := false
	switch exp := referExp.(type) {
	case *ast.NameExp:
		varName = exp.Name
		varLoc = exp.Loc
	case *ast.TableAccessExp:
		preName, varName, keyName, preLoc, varLoc, isTableWhole = common.GetTableNameInfo(exp)
		isTableExp = true
	case *ast.FuncCallExp:
		// 这里最多取两段 如 a.b() 的形式
		if nameExp, ok := exp.PrefixExp.(*ast.NameExp); ok && exp.NameExp == nil {
			varName = nameExp.Name
			varLoc = nameExp.Loc
		} else if preExp, ok := exp.PrefixExp.(*ast.TableAccessExp); ok {
			if nameExp, ok := preExp.PrefixExp.(*ast.NameExp); ok {
				preName = nameExp.Name
				preLoc = nameExp.Loc
			}
			if stringExp, ok := preExp.KeyExp.(*ast.StringExp); ok {
				varName = stringExp.Str
				varLoc = stringExp.Loc
			}
		}
		isFuncCall = true
	}

	if isTableExp {
		//当是表且有截断 不继续推导类型
		if !isTableWhole {
			return
		}
	}

	ok, varInfo, _, varType := a.findVarDefineWithPre(preName, varName, preLoc, varLoc, true)
	if !ok {

		// // 有前缀但找不到成员时候 返回前缀的定义 如c2s.func 直接找c2s的注解
		// if len(preName) > 0 && isFuncCall {
		// 	ok, preInfo, _ := a.findVarDefine(preName, preLoc)
		// 	if !ok {
		// 		return
		// 	}

		// 	defAnnTypeVec := a.Projects.GetAnnotateTypeString(preInfo, preName, varName, 1)
		// 	if len(defAnnTypeVec) > 0 {
		// 		return defAnnTypeVec
		// 	}
		// }

		return
	}

	varIdx := int(varInfo.VarIndex)
	if idx > 0 {
		varIdx = idx
	}

	if varInfo.IsParam {
		// 变量属于函数参数 此处可确认其类型 直接返回
		return varType
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
			return a.GetAnnTypeByExp(varInfo.ReferExp, varIdx)
		}
	}

	if argType == "LuaTypeRefer" && varInfo.ReferExp != nil && !isFuncCall {
		if _, ok := varInfo.ReferExp.(*ast.FuncCallExp); ok {
			// 如local a = func()形式
			// 尝试取func()的返回值类型
			return a.GetAnnTypeByExp(varInfo.ReferExp, varIdx)
		}
	}

	retVec = append(retVec, argType)
	return retVec
}

// 加载函数的参数与返回值的注解类型
func (a *Analysis) loadFuncParamAnnType(referFunc *common.FuncInfo) {
	if referFunc == nil || len(referFunc.ParamType) > 0 || len(referFunc.ReturnVecs) > 0 {
		return
	}

	paramTypeMap := a.Projects.GetFuncParamType(referFunc.FileName, referFunc.Loc.StartLine-1)
	if len(paramTypeMap) > 0 {
		//从函数上方获取到了注解
		for _, paramName := range referFunc.ParamList {
			if annTypeVec, ok := paramTypeMap[paramName]; ok {
				referFunc.ParamType[paramName] = []string{}
				for _, annTypeOne := range annTypeVec {
					typeOne := annotateast.GetAstTypeName(annTypeOne)
					referFunc.ParamType[paramName] = append(referFunc.ParamType[paramName], typeOne)

				}
			}
		}

		//继续获取返回值注解
		referFunc.ReturnType = a.Projects.GetFuncReturnTypeVec(referFunc.FileName, referFunc.Loc.StartLine-1)

		return //从函数上方获取到了注解之后就不再查找类成员函数的注解
	}

	if referFunc.ClassName != "" {
		// 如果是类的成员函数 先找类变量的定义
		find, varDefine, _ := a.findVarDefineGlobal(referFunc.ClassName)
		if !find {
			return
		}

		// 再找类上方注解
		classTypeStr, ok := a.Projects.GetVarAnnType(varDefine.FileName, varDefine.Loc.StartLine-1)
		if !ok {
			return
		}

		// 根据注解找成员函数
		referFunc.ParamType = a.Projects.GetFuncParamTypeByClass(classTypeStr, referFunc.FuncName)
		referFunc.ReturnType = a.Projects.GetFuncReturnTypeByClass(classTypeStr, referFunc.FuncName)

		return
	}

	return
}
