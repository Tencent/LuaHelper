package analysis

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
)

// 专门处理变量引用另外一个变量
// 例如 local a = b;  -- a 变量是关联指向b变量的，那么a变量要有b变量所有的方法，代码提示的时候
// 变量追踪的特性

// getExpReferVarInfo 获取exp相关联的其他变量，局部变量或是全局变量
// 第一轮，遍历中可能不太准确， 目前有第二轮、第三轮中来获取该值
func (a *Analysis) getExpReferVarInfo(node ast.Exp) (locVar *common.VarInfo) {
	// 目前只有在第一轮中构造，在第四轮查找变量引用时候用到
	if !a.isFirstTerm() {
		return
	}

	switch exp := node.(type) {
	case *ast.ParensExp:
		return a.getExpReferVarInfo(exp.Exp)
	case *ast.NameExp:
		// 获取这个引用的变量名，关联的VarInfo
		return a.findStrReferVarInfo(exp.Name, exp.Loc, false, "")
	case *ast.BinopExp:
		if exp.Op == lexer.TkOpOr {
			// 如果是or 二元表达式，关联第一个表达式
			return a.getExpReferVarInfo(exp.Exp1)
		} else if exp.Op == lexer.TkOpAdd {
			// 如果是and 二元表达式，关联第二个表达式
			return a.getExpReferVarInfo(exp.Exp2)
		}
		return nil
	case *ast.TableAccessExp:
		return a.getTableAccessRelateVar(exp)
	case *ast.TableConstructorExp:
		// 如果函数返回的是table的构造，那么table的构造是一个匿名的变量，先构造下匿名变量
		newVar := common.CreateVarInfo(a.curResult.Name, common.LuaTypeTable, exp, exp.Loc, 1)

		// 构造这个变量的table 构造的成员，都构造出来
		newVar.MakeVarMemTable(node, a.curResult.Name, exp.Loc)
		return newVar
	case *ast.FuncCallExp:
		strName := ""

		// 如果是简单的函数调用，直接是这种a("b")
		if nameExp, ok := exp.PrefixExp.(*ast.NameExp); ok && exp.NameExp == nil {
			strName = nameExp.Name

			// 这个为函数调用，关联到函数调用的返回值，关联到第一个返回值
			// 先获取这个变量
			referVarInfo := a.findStrReferVarInfo(strName, exp.Loc, false, "")
			if referVarInfo != nil && referVarInfo.ReferFunc != nil {
				// 如果关联的变量不为空，先找到关联的函数的返回值
				return referVarInfo.ReferFunc.GetOneReturnVar()
			}
		}
	}

	return
}

// 查找一个变量的直接引用, 另外一个变量VarInfo
func (a *Analysis) findStrReferVarInfo(strName string, loc lexer.Location, gFlag bool, strProPre string) *common.VarInfo {
	// 1) 判断该变量是否为lua模块自带或是框架需要屏蔽的变量
	if common.GConfig.IsIgnoreNameVar(strName) {
		return nil
	}

	fileResult := a.curResult

	// 2) 判断是否为需要忽略的文件中的变量
	if common.GConfig.IsIgnoreFileDefineVar(fileResult.Name, strName) {
		return nil
	}

	scope := a.curScope
	fi := a.curFunc

	// 3) 查找局部变量指向的函数信息
	if !gFlag {
		if locVar, ok := scope.FindLocVar(strName, loc); ok {
			return locVar
		}
	}

	if a.isFirstTerm() {
		// 1) 查找这个文件的全局变量是否有该变量
		_, findSymbol := fileResult.FindGlobalVarInfo(strName, false, "")
		return findSymbol
	}

	// 4) 第三步查找全局中是否有该变量
	firstFile := a.getFirstFileResult(fileResult.Name)
	if a.isSecondTerm() {
		secondResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			if ok, oneVar := secondResult.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar
			}

			if ok, oneVar := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermSecond, strProPre); ok {
				return oneVar
			}
		} else {
			// 非底层的函数，需要查找全局的变量
			if ok, oneVar := firstFile.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar
			}

			if ok, oneVar := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermFirst, strProPre); ok {
				return oneVar
			}
		}
	} else if a.isThirdTerm() {
		thirdFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			if ok, oneVar := thirdFileResult.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar
			}
		} else {
			// 非底层的函数，需要查找全局的变量
			if ok, oneVar := firstFile.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar
			}
		}

		// 查找所有的
		if ok, oneVar := a.AnalysisThird.ThirdStruct.FindThirdGlobalGInfo(gFlag, strName, strProPre); ok {
			return oneVar
		}
	}

	return nil
}

// 根据table的调用，获取到对应的变量
func (a *Analysis) getTableAccessRelateVar(node *ast.TableAccessExp) (locVar *common.VarInfo) {
	// 3) 判断key是否为普通的
	strKey := common.GetExpName(node.KeyExp)
	// 如果不是简单字符，退出
	if !common.JudgeSimpleStr(strKey) {
		return
	}
	keyLoc := common.GetTableKeyLoc(node)
	strTable := common.GetExpName(node.PrefixExp)
	if strTable == "!_G" {
		// 这里先指定_G
		locVar = a.findStrReferVarInfo(strKey, keyLoc, true, "")
		return locVar
	}

	// strTable 不为_G
	strPre := common.GetSimpleValue(strTable)
	if strPre == "" {
		return
	}

	preVarInfo := a.findStrReferVarInfo(strPre, keyLoc, false, "")
	if preVarInfo == nil {
		return
	}

	subMaps := preVarInfo.SubMaps
	if subMaps == nil {
		return
	}

	locVar = subMaps[strKey]
	return
}
