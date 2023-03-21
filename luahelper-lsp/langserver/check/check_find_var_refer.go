package check

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"strings"
)

// FindDeepSymbolList 递归查找一个变量的关联其他的变量
// containTableFlag 为是否包含LuaTypeTable 这个类型, 默认为false
// varIndex 为推导的变量index, 默认从1开始。例如local a, b 其中a的varIndex为1，b的varIndex为2
func (a *AllProject) FindDeepSymbolList(luaInFile string, exp ast.Exp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile, containTableFlag bool, varIndex uint8) (symList []*common.Symbol) {
	for {
		luaType := common.GetExpType(exp)
		if !containTableFlag && luaType == common.LuaTypeTable {
			// table定义的，理论上不存在引用其他的变量
			return
		}

		// varIndex的取值，取值最新的
		if len(symList) > 0 {
			varIndex = symList[len(symList)-1].VarInfo.VarIndex
		}

		symbol := a.FindVarReferSymbol(luaInFile, exp, comParam, findExpList, varIndex)
		if symbol == nil {
			break
		}

		if isHasSymbol(symbol, symList) {
			break
		}

		// 找到了，保存下
		symList = append(symList, symbol)

		if symbol.VarFlag == common.FirstAnnotateFlag || symbol.VarInfo == nil {
			break
		}

		// 递归进行查找
		exp = symbol.VarInfo.ReferExp
	}

	return
}

// FindVarReferSymbol 查找一个变量的关联其他的变量。
// findExpList 为已经跟踪到的引用表达式信息列表，防止重复（死循环)
func (a *AllProject) FindVarReferSymbol(luaInFile string, node ast.Exp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile, varIndex uint8) (symbol *common.Symbol) {
	if isHasFindExpFile(findExpList, luaInFile, node) {
		log.Debug("isHasFindExpFile true, isHasFindExpFile=%s", luaInFile)
		return nil
	}

	insertFindExpFile(findExpList, luaInFile, node)

	switch exp := node.(type) {
	case *ast.ParensExp:
		return a.FindVarReferSymbol(luaInFile, exp.Exp, comParam, findExpList, varIndex)
	case *ast.NameExp:
		// 获取这个引用的变量名，关联的VarInfo
		symbol = a.findStrReferSymbol(luaInFile, exp.Name, exp.Loc, false, comParam, findExpList)
		return symbol
	case *ast.BinopExp:
		if exp.Op == lexer.TkOpOr {
			// 如果是or 二元表达式，关联第一个表达式
			return a.FindVarReferSymbol(luaInFile, exp.Exp1, comParam, findExpList, varIndex)
		} else if exp.Op == lexer.TkOpAdd {
			// 如果是and 二元表达式，关联第二个表达式
			return a.FindVarReferSymbol(luaInFile, exp.Exp2, comParam, findExpList, varIndex)
		}
		return nil
	case *ast.TableAccessExp:
		return a.getTableAccessRelateSymbol(luaInFile, exp, comParam, findExpList)
	case *ast.TableConstructorExp:
		// 如果函数返回的是table的构造，那么table的构造是一个匿名的变量，先构造下匿名变量
		newVar := common.CreateVarInfo(luaInFile, common.LuaTypeTable, exp, exp.Loc, 1)

		// 构造这个变量的table 构造的成员，都构造出来
		newVar.MakeVarMemTable(node, luaInFile, exp.Loc)
		symbol := a.createAnnotateSymbol(luaInFile, newVar)
		return symbol
	case *ast.FuncCallExp:
		return a.getFuncRelateSymbol(luaInFile, exp, comParam, findExpList, varIndex)
	case *ast.FuncDefExp:
		return a.getFuncDefSymbol(luaInFile, exp, comParam, findExpList, varIndex)
	}

	return
}

// 创建一个变量的时候，创建完整的信息, 里面会设置变量的注解类型
// fileName 为变量出现的文件名
// strName 为变量的名称
func (a *AllProject) createAnnotateSymbol(strName string, varInfo *common.VarInfo) (symbol *common.Symbol) {
	symbol = common.GetDefaultSymbol(varInfo.FileName, varInfo)

	a.setInfoFileAnnotateTypes(strName, symbol)
	return symbol
}

// 传入位置和变量名，进一步查找对应的变量
func (a *AllProject) findLocReferSymbol(fileResult *results.FileResult, posLine int, posChar int, luaInFile string,
	strName string, loc lexer.Location, gFlag bool, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile) (symbol *common.Symbol) {

	// 0) 获取到scope
	minScope, minFunc := fileResult.FindASTNode(posLine, posChar)
	if minScope == nil || minFunc == nil {
		return nil
	}

	// 1) 查找局部变量指向的函数信息
	if !gFlag {
		if locVar, ok := minScope.FindLocVar(strName, loc); ok {
			return a.createAnnotateSymbol(strName, locVar)
		}
	}

	if comParam.secondProject != nil {
		// 非底层的函数，需要查找全局的变量
		if ok, oneVar := fileResult.FindGlobalVarInfo(strName, gFlag, ""); ok {
			return a.createAnnotateSymbol(strName, oneVar)
		}

		if ok, oneVar := comParam.secondProject.FindGlobalGInfo(strName, results.CheckTermFirst, ""); ok {
			return a.createAnnotateSymbol(strName, oneVar)
		}
	}

	if comParam.thirdStruct != nil {
		// 非底层的函数，需要查找全局的变量
		if ok, oneVar := fileResult.FindGlobalVarInfo(strName, gFlag, ""); ok {
			return a.createAnnotateSymbol(strName, oneVar)
		}

		// 查找所有的
		if ok, oneVar := comParam.thirdStruct.FindThirdGlobalGInfo(gFlag, strName, ""); ok {
			return a.createAnnotateSymbol(strName, oneVar)
		}
	}

	return nil
}

// 查找一个变量的直接引用, 另外一个变量VarInfo
func (a *AllProject) findStrReferSymbol(luaInFile string, strName string, loc lexer.Location, gFlag bool,
	comParam *CommonFuncParam, findExpList *[]common.FindExpFile) (symbol *common.Symbol) {
	fileStruct, ok := a.GetCacheFileStruct(luaInFile)
	if !ok {
		return nil
	}
	if fileStruct.HandleResult != results.FileHandleOk {
		return nil
	}

	fileResult := fileStruct.FileResult
	if fileResult == nil {
		return nil
	}

	// 2) 判断是否为需要忽略的文件中的变量
	if common.GConfig.IsIgnoreFileDefineVar(luaInFile, strName) {
		return nil
	}

	// 查找到对应位置的scope
	// 特殊的处理，例如下面的例子
	// b = 1
	// local b = b
	// print(b) -- 这里对b找定义的时候，会找到， local b = b，此时对b的exp找引用
	// 由于坐标问题，会一直找到local b = b，而找不到前面的一行b = 1
	// 解决的方法，先找前一行，如果找不多，再找本行
	loc.StartColumn = 0
	loc.EndColumn = 0
	referSymbol := a.findLocReferSymbol(fileResult, loc.StartLine-1, 0, luaInFile, strName,
		loc, gFlag, comParam, findExpList)
	if referSymbol != nil {
		return referSymbol
	}

	// 4) 判断是否为系统的函数，或是是全局变量模块
	sysVar := common.GConfig.GetSysVar(strName)
	if sysVar == nil {
		return nil
	}

	// 系统的变量，没有注解系统
	symbol = common.GetDefaultSymbol("", sysVar)
	return symbol
}

// 判断是否为简单类型类别的子成员
// classList 为所有的列表
// strKey 为key
func (a *AllProject) getClassListSubMem(classList []*common.OneClassInfo, strKey string) (symbol *common.Symbol) {
	for _, oneClass := range classList {
		symbol = a.getClassInfoSubMem(oneClass, strKey)
		// 只要其中一个classInfo 的子成员包含了strKey表示找到了
		if symbol != nil {
			return symbol
		}
	}

	return symbol
}

// 获取注解系统的子成员信息
// todo 这个函数的位置后面会搬迁
func (a *AllProject) getClassInfoSubMem(classInfo *common.OneClassInfo, strKey string) (symbol *common.Symbol) {
	if fieldState, ok := classInfo.FieldMap[strKey]; ok {
		// 匹配到了
		symbol = &common.Symbol{
			FileName:     classInfo.LuaFile,
			VarInfo:      nil,
			AnnotateType: fieldState.FiledType,
			VarFlag:      common.FirstAnnotateFlag, // 先获取的注解类型
			AnnotateLine: classInfo.LastLine,
			AnnotateLoc:  fieldState.NameLoc,
		}

		symbol.AnnotateComment = fieldState.Comment
		symbol.StrPreComment = "field"
		symbol.StrPreClassName = classInfo.ClassState.Name

		// 附加判断下，这个变量类型是否关联变量，获取对应的VarInfo
		if classInfo.RelateVar != nil {
			// 自己的子项中，含有
			if subVar, ok := classInfo.RelateVar.SubMaps[strKey]; ok {
				symbol.VarInfo = subVar
			}
		}
		return
	}

	// 注解里面没有直接含有，判断子成员是否有
	if classInfo.RelateVar != nil {
		// 自己的子项中，含有
		if subVar, ok := classInfo.RelateVar.SubMaps[strKey]; ok {
			symbol = a.createAnnotateSymbol(strKey, subVar)
			return symbol
		}
	}

	// todo 这里是否要考虑到引用的信息，参考函数：varInfoHasSubKey

	return
}

// 获取对应这个type对应的ArrayType
func (a *AllProject) getArrayTypeMemKey(astType annotateast.Type, fileName string, line int) (symbol *common.Symbol) {
	subType := a.GetAllArrayType(fileName, astType)
	if subType == nil {
		return
	}

	symbol = &common.Symbol{
		FileName:     fileName,
		VarInfo:      nil,
		AnnotateType: subType,
		VarFlag:      common.FirstAnnotateFlag,
		AnnotateLine: line,
		AnnotateLoc:  annotateast.GetAstTypeLoc(subType),
	}

	return symbol
}

// 获取这个type对应的TableType
func (a *AllProject) getTableTypeMemKey(astType annotateast.Type, fileName string, line int) (symbol *common.Symbol) {
	subType := a.GetAllTableType(fileName, astType)
	if subType == nil {
		return
	}

	symbol = &common.Symbol{
		FileName:     fileName,
		VarInfo:      nil,
		AnnotateType: subType,
		VarFlag:      common.FirstAnnotateFlag,
		AnnotateLine: line,
		AnnotateLoc:  annotateast.GetAstTypeLoc(subType),
	}

	return symbol
}

// 根据table的调用，获取到对应的变量
func (a *AllProject) getTableAccessRelateSymbol(luaInFile string, node *ast.TableAccessExp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile) (symbol *common.Symbol) {
	strKey := common.GetExpName(node.KeyExp)
	keyLoc := common.GetTableKeyLoc(node)
	defineStruct := ExpToDefineVarStruct(node.PrefixExp)
	defineLen := len(defineStruct.StrVec)
	if !defineStruct.ValidFlag || defineLen < 1 {
		return
	}

	// 1）先处理_G.a 这样的情况
	if defineLen == 1 && defineStruct.StrVec[0] == "_G" {
		// 先只处理简单的情况
		if !common.JudgeSimpleStr(strKey) {
			return
		}

		// 这里先指定_G
		symbol = a.findStrReferSymbol(luaInFile, strKey, keyLoc, true, comParam, findExpList)
		return symbol
	}

	// 2）判断是否为简单的取值a.b
	var lastSymbol *common.Symbol
	_, sampleFlag := (node.PrefixExp).(*ast.NameExp)
	if sampleFlag {
		strOne := defineStruct.StrVec[0]

		if strOne == "self" {
			strOne = a.selfChangeStrName(luaInFile, node.Loc)
			if strOne == "" {
				strOne = "self"
			}
		}

		// 首先查找父的模块变量
		lastSymbol = a.findStrReferSymbol(luaInFile, strOne, keyLoc, false, comParam, findExpList)
	} else {
		lastSymbol = a.FindVarReferSymbol(luaInFile, node.PrefixExp, comParam, findExpList, 1)
	}

	if lastSymbol == nil {
		return nil
	}

	var lastExp ast.Exp    // 最终查找关联到的表达式
	var lastLuaFile = ""   // 最终查找关联到的表达式所在的文件
	var varIndex uint8 = 1 // 多次申明变量时候，变量的index

	subSymbol := a.symbolHasSubKey(lastSymbol, strKey, comParam, findExpList)
	if subSymbol != nil {
		lastSymbol = subSymbol
		return lastSymbol
	}

	if lastSymbol.VarInfo == nil {
		return nil
	}

	lastExp = lastSymbol.VarInfo.ReferExp
	lastLuaFile = lastSymbol.FileName
	varIndex = lastSymbol.VarInfo.VarIndex

	// 没有找到，那么这个变量，它关联的上一层变量呢
	// 递归进行查找
	subFindFlag := false
	tmpList := a.FindDeepSymbolList(lastLuaFile, lastExp, comParam, findExpList, false, varIndex)
	for _, oneSymbol := range tmpList {
		subSymbol := a.symbolHasSubKey(oneSymbol, strKey, comParam, findExpList)
		if subSymbol != nil {
			lastSymbol = subSymbol
			subFindFlag = true
			break
		}
	}

	// 最终也是没有找到
	if !subFindFlag {
		return nil
	}

	return lastSymbol
}

func (a *AllProject) getReferReferInfoSymbol(referFile *results.FileResult, referInfo *common.ReferInfo, strKey string,
	comParam *CommonFuncParam, findExpList *[]common.FindExpFile) (symbol *common.Symbol) {
	referSubType := a.GetReferFrameType(referInfo)
	if referSubType == common.RtypeImport {
		if subVar, ok := referFile.GlobalMaps[strKey]; ok {
			symbol := a.createAnnotateSymbol(strKey, subVar)
			return symbol
		}
		return
	}

	if referSubType == common.RtypeRequire {
		find, returnExp := referFile.MainFunc.GetLastOneReturnExp()
		if !find {
			return
		}

		// 这里也需要做判断，函数返回的变量逐层跟踪，目前只跟踪了一层
		symList := a.FindDeepSymbolList(referFile.Name, returnExp, comParam, findExpList, true, 1)
		// 所有关联的Var，都查找一边
		for _, oneSymbol := range symList {
			symbol = a.symbolHasSubKey(oneSymbol, strKey, comParam, findExpList)
			if symbol != nil {
				return symbol
			}
		}
	}

	return
}

// 判断变量是否含有子的strKey子项
func (a *AllProject) symbolHasSubKey(oldSymbol *common.Symbol, strKey string,
	comParam *CommonFuncParam, findExpList *[]common.FindExpFile) (symbol *common.Symbol) {
	simpleStrFlag := common.JudgeSimpleStr(strKey)
	line := oldSymbol.GetLine()

	// 关联变量的注解类型
	// 简单的类型，先判断是否为子成员
	if simpleStrFlag {
		classList := a.getAllNormalAnnotateClass(oldSymbol.AnnotateType, oldSymbol.FileName, line)
		if subSombol := a.getClassListSubMem(classList, strKey); subSombol != nil {
			// 表示通过注解类型找到了子成员
			symbol = subSombol
			return
		}
	}

	// 判断是否为 ArrayType
	arraySubFile := a.getArrayTypeMemKey(oldSymbol.AnnotateType, oldSymbol.FileName, line)
	if arraySubFile != nil {
		symbol = arraySubFile
		return
	}

	// 判断是否为 TableType
	tableSubFile := a.getTableTypeMemKey(oldSymbol.AnnotateType, oldSymbol.FileName, line)
	if tableSubFile != nil {
		symbol = tableSubFile
		return
	}

	// 如果是非简单的字符串或没有变量信息，直接返回
	if !simpleStrFlag || oldSymbol.VarInfo == nil {
		return
	}

	// 判断这个注解类型是否包含子key
	// 递归查找子成员
	symbol = a.varInfoHasSubKey(oldSymbol.VarInfo, strKey, comParam, findExpList)
	return symbol
}

// 判断变量是否含有子的strKey子项
func (a *AllProject) varInfoHasSubKey(varInfo *common.VarInfo, strKey string,
	comParam *CommonFuncParam, findExpList *[]common.FindExpFile) (symbol *common.Symbol) {
	// 自己的子项中，含有
	subVar, ok := varInfo.SubMaps[strKey]
	if ok {
		symbol := a.createAnnotateSymbol(strKey, subVar)
		return symbol
	}

	// 判断引用的是否含义
	referInfo := varInfo.ReferInfo
	if referInfo == nil {
		// 没有引用信息
		return
	}

	referFile := a.GetFirstReferFileResult(referInfo)
	if referFile == nil {
		// 文件不存在
		return
	}

	// 引用关系的查找
	return a.getReferReferInfoSymbol(referFile, referInfo, strKey, comParam, findExpList)
}

// 获取原表右侧的简单元素, 目前支持strKey为__index与__call
func (a *AllProject) getSimpleMetatableRight(luaInFile string, rightExp ast.Exp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile, strKey string) (symbol *common.Symbol) {
	expTable, ok := rightExp.(*ast.TableConstructorExp)
	if !ok {
		// 查询变量是否有关联到子成员 __index
		rerferVarFile := a.FindVarReferSymbol(luaInFile, rightExp, comParam, findExpList, 1)
		if rerferVarFile != nil {
			subSymbol := a.symbolHasSubKey(rerferVarFile, strKey, comParam, findExpList)
			if subSymbol != nil && subSymbol.VarInfo != nil {
				findVarFile := a.FindVarReferSymbol(subSymbol.FileName, subSymbol.VarInfo.ReferExp, comParam,
					findExpList, 1)
				return findVarFile
			}

			return nil
		}
		return nil
	}

	for i, keyExp := range expTable.KeyExps {
		valExp := expTable.ValExps[i]
		if keyExp == nil {
			continue
		}

		oneKeyStr := common.GetExpName(keyExp)
		if oneKeyStr != strKey {
			continue
		}

		// 递归查找
		return a.FindVarReferSymbol(luaInFile, valExp, comParam, findExpList, 1)
	}

	return nil
}

// 判断是否为原表简单函数的构造
// 完整的表达式为 a = setmetatable({}, {__index = b})
// 第二个表达式为 {__index = b}
func (a *AllProject) getSimpleMetatable(luaInFile string, node *ast.FuncCallExp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile) (symbol *common.Symbol) {
	// 先获取左侧的元素
	leftSymbol := a.FindVarReferSymbol(luaInFile, node.Args[0], comParam, findExpList, 1)

	// 获取原表右侧的元素__call元素
	callRightSymbol := a.getSimpleMetatableRight(luaInFile, node.Args[1], comParam, findExpList, "__call")
	if callRightSymbol != nil {
		return callRightSymbol
	}

	// 获取原表右侧的元素__index元素
	rightSymbol := a.getSimpleMetatableRight(luaInFile, node.Args[1], comParam, findExpList, "__index")
	if leftSymbol == nil {
		return rightSymbol
	}
	if rightSymbol == nil {
		return leftSymbol
	}

	// 优先级高与低的symbol
	var highSymbol *common.Symbol
	var lowSymbol *common.Symbol
	// 1) 先判断两边是否有注解，左边有注解，右边没有，右边的往左边合并
	if leftSymbol.AnnotateType != nil && rightSymbol.AnnotateType == nil {
		highSymbol = leftSymbol
		lowSymbol = rightSymbol
	} else if leftSymbol.AnnotateType == nil && rightSymbol.AnnotateType != nil {
		// 2) 左边没有注解，右边含义注解，左边向右边移动
		highSymbol = rightSymbol
		lowSymbol = leftSymbol
	} else {
		highSymbol = leftSymbol
		lowSymbol = rightSymbol
	}

	// 优先级高的没有变量，低的含有
	if highSymbol.VarInfo == nil && lowSymbol.VarInfo != nil {
		highSymbol.VarInfo = lowSymbol.VarInfo
		return highSymbol
	}

	// 优先级高的没有变量, 低的也没有
	if highSymbol.VarInfo == nil && lowSymbol.VarInfo == nil {
		return highSymbol
	}

	//  两边都含有变量，变量进行合并
	if len(highSymbol.VarInfo.SubMaps) == 0 && len(lowSymbol.VarInfo.SubMaps) == 0 {
		return highSymbol
	}

	if len(highSymbol.VarInfo.SubMaps) == 0 {
		highSymbol.VarInfo.SubMaps = map[string]*common.VarInfo{}
	}
	for key, oneVar := range lowSymbol.VarInfo.SubMaps {
		highSymbol.VarInfo.SubMaps[key] = oneVar
	}

	return highSymbol
}

// 判断是否绑定了特定的注解推导类型
func (a *AllProject) getBindAnnotateSetType(luaInFile string, strFuncName string,
	node *ast.FuncCallExp, findExpList *[]common.FindExpFile) (matchFlag bool, symbol *common.Symbol) {
	matchFlag = false

	// 如果没有配置自动的类型推导直接返回
	flag, oneSet := common.GConfig.MatchAnnotateSet(strFuncName)
	if !flag {
		return
	}

	matchFlag = true

	// 只有一层的推导
	argLen := len(node.Args)
	if argLen < oneSet.ParamIndex || oneSet.ParamIndex < 1 {
		// 参数个数不匹配
		return
	}

	paramExp := node.Args[oneSet.ParamIndex-1]
	strExp, ok := paramExp.(*ast.StringExp)
	strName := ""
	if ok {
		strName = strExp.Str
	} else {
		comParam := a.getCommFunc(luaInFile, node.Loc.EndLine, node.Loc.StartColumn)
		if comParam == nil {
			return
		}

		nameVarFile := a.FindVarReferSymbol(luaInFile, paramExp, comParam, findExpList, 1)
		if nameVarFile == nil || nameVarFile.VarInfo == nil {
			return
		}

		strType := nameVarFile.VarInfo.GetVarTypeDetail()
		if !strings.HasPrefix(strType, "string: \"") {
			return
		}

		strType = strings.TrimPrefix(strType, "string: \"")
		strType = strings.TrimSuffix(strType, "\"")
		strName = strType
	}

	if oneSet.SplitFlag == 1 {
		index := strings.Index(strName, ".")
		if index >= 0 {
			strName = strName[index+1:]
		}

		// 查找/
		index = strings.Index(strName, "/")
		if index >= 0 {
			strName = strName[index+1:]
		}
	}

	strAnnotateName := ""
	if oneSet.PrefixStr == "" {
		// 判断是否有多个选择的
		for _, oneStr := range oneSet.PrefixStrList {
			strTmp := oneStr + strName + oneSet.SuffixStr
			if !a.judgeExistAnnoteTypeStr(strTmp) {
				continue
			}

			strAnnotateName = strTmp
			break
		}

		if strAnnotateName == "" {
			strAnnotateName = strName + oneSet.SuffixStr
		}
	} else {
		strAnnotateName = oneSet.PrefixStr + strName + oneSet.SuffixStr
	}

	if !a.judgeExistAnnoteTypeStr(strAnnotateName) {
		return
	}

	normalAst := &annotateast.NormalType{
		StrName: strAnnotateName,
		NameLoc: node.Loc,
	}

	symbol = &common.Symbol{
		FileName:     luaInFile,
		VarInfo:      nil,
		AnnotateType: normalAst,
		VarFlag:      common.FirstAnnotateFlag,
		AnnotateLine: node.Loc.StartLine,
	}

	return
}

func (a *AllProject) selfChangeStrName(luaInFile string, loc lexer.Location) (strName string) {
	strName = ""

	// 1）先查找该文件是否存在
	fileStruct := a.getVailidCacheFileStruct(luaInFile)
	if fileStruct == nil {
		log.Error("FindVarDefine error, not find file=%s", luaInFile)
		return
	}

	minScope, minFunc := fileStruct.FileResult.FindASTNode(loc.StartLine-1, loc.StartColumn)
	if minScope == nil || minFunc == nil {
		log.Error("FindVarDefine error, minScope or minFunc is nil file=%s", luaInFile)
		return
	}

	firstColonFunc := minFunc.FindFirstColonFunc()
	if firstColonFunc == nil {
		return
	}
	if !firstColonFunc.IsColon {
		return
	}

	if firstColonFunc.RelateVar == nil {
		return
	}

	return firstColonFunc.RelateVar.StrName
}

// GetImportReferByCallExp 根据传入的exp，判断是否为导入的函数调用
func (a *AllProject) getImportReferSymbol(luaInFile string, funcExp *ast.FuncCallExp,
	comParam *CommonFuncParam, findExpList *[]common.FindExpFile) (symbol *common.Symbol) {
	symbol = nil
	fileStruct := a.getVailidCacheFileStruct(luaInFile)

	if fileStruct == nil {
		log.Error("FindVarDefine error, not find file=%s", luaInFile)
		return nil
	}

	if funcExp.NameExp != nil {
		return nil
	}

	callExp, ok := funcExp.PrefixExp.(*ast.NameExp)
	if !ok {
		return nil
	}

	if !common.GConfig.ReferOtherFileMap[callExp.Name] {
		return nil
	}

	if len(funcExp.Args) != 1 {
		return nil
	}

	firstExp := funcExp.Args[0]
	if _, ok := firstExp.(*ast.StringExp); !ok {
		return nil
	}

	strFirst := firstExp.(*ast.StringExp).Str
	oneRefer := common.CreateOneReferInfo(callExp.Name, strFirst, funcExp.Loc)
	if oneRefer == nil {
		return nil
	}

	// 先查找该引用是否有效
	fileStruct.FileResult.CheckReferFile(oneRefer, a.allFilesMap, a.fileIndexInfo)
	if !oneRefer.Valid {
		return nil
	}

	referFile := a.GetFirstReferFileResult(oneRefer)
	if referFile == nil {
		// 文件不存在
		return nil
	}

	if oneRefer.ReferType == common.ReferTypeRequire {
		find, returnExp := referFile.MainFunc.GetLastOneReturnExp()
		if !find {
			return nil
		}

		symbol = a.FindVarReferSymbol(referFile.Name, returnExp, comParam, findExpList, 1)
	}

	return symbol
}

// 获取函数定义，例如这样的 a = function()
func (a *AllProject) getFuncDefSymbol(luaInFile string, node *ast.FuncDefExp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile, varIndex uint8) (symbol *common.Symbol) {
	fileStruct := a.fileStructMap[luaInFile]
	if fileStruct == nil {
		log.Debug("fileStruct error, strFile=%s", luaInFile)
		return
	}

	// 文件加载失败的忽略
	if fileStruct.HandleResult != results.FileHandleOk {
		return
	}

	// 如果函数返回的是table的构造，那么table的构造是一个匿名的变量，先构造下匿名变量
	newVar := common.CreateVarInfo(luaInFile, common.LuaTypeRefer, nil, node.Loc, 1)
	for _, oneFunc := range fileStruct.FileResult.FuncIDVec {
		if oneFunc.Loc == node.Loc {
			newVar.ReferFunc = oneFunc
			break
		}
	}
	if newVar.ReferFunc == nil {
		return
	}

	symbol = common.GetDefaultSymbol(newVar.FileName, newVar)
	return symbol
}

// 根据table的调用，获取到对应的变量
func (a *AllProject) getFuncRelateSymbol(luaInFile string, node *ast.FuncCallExp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile, varIndex uint8) (symbol *common.Symbol) {

	// 如果是简单的函数调用，直接是这种a("b")
	if nameExp, ok := node.PrefixExp.(*ast.NameExp); ok && node.NameExp == nil {
		strName := nameExp.Name
		// 这两个函数，在变量的referInfo里面已经存在了
		if strName == "require" || common.GConfig.IsFrameReferOtherFile(strName) {
			//return nil
			referSymbol := a.getImportReferSymbol(luaInFile, node, comParam, findExpList)
			return referSymbol
		}

		// 考虑是下面简单原表的调用
		// LogicJediKnight = setmetatable({}, {__index = AbstractLogicTwoStageDraw})
		if strName == "setmetatable" && len(node.Args) == 2 {
			// 判断第一个是否为简单的table构造，先处理这种情况
			return a.getSimpleMetatable(luaInFile, node, comParam, findExpList)
		}
	}

	strTable := common.GetExpName(node.PrefixExp)
	_, tableSigh := common.StrRemoveSigh(strTable)
	keyLoc := common.GetExpLoc(node.PrefixExp)

	selfFlag := false
	if tableSigh == "self" {
		tableSigh = a.selfChangeStrName(luaInFile, node.Loc)
		if tableSigh == "" {
			tableSigh = "self"
		} else {
			selfFlag = true
		}
	}

	strFuncName := ""
	if node.NameExp != nil {
		strFuncName = node.NameExp.Str
	} else {
		strFuncName = tableSigh
	}

	// 客户端的特殊的函数，绑定特定的注解，例如下面的代码段，绑定的注解类型为：UActImageDownloader
	// local ActImageDownloader = KismetLibrary.New("/Script/Client.ActImageDownloader");
	// 客户端的特殊的函数，绑定特定的注解, 例如下面的代码段，绑定的注解类型为ULobby_Food_Collect_UIBP
	// ActivityFoodCollect.uiObj = GetUIObject(bp_activity_food_collect, "Lobby_Food_Collect_UIBP");
	if ok, matchVarFile := a.getBindAnnotateSetType(luaInFile, strFuncName, node, findExpList); ok {
		return matchVarFile
	}

	var beforeSymbol *common.Symbol
	if selfFlag {
		beforeSymbol = a.findStrReferSymbol(luaInFile, tableSigh, keyLoc, false, comParam, findExpList)
	} else {
		// 递归查找，例如： uiButton.new():setX(10):setY(10)
		// 修复 https://github.com/Tencent/LuaHelper/issues/42
		beforeSymbol = a.FindVarReferSymbol(luaInFile, node.PrefixExp, comParam, findExpList, 1)
	}

	if beforeSymbol == nil {
		return
	}
	funcSymbol := beforeSymbol
	if node.NameExp != nil {
		strAfter := node.NameExp.Str
		varSubFile := a.symbolHasSubKey(beforeSymbol, strAfter, comParam, findExpList)
		if varSubFile != nil {
			funcSymbol = varSubFile
		} else {
			if beforeSymbol.VarInfo == nil {
				return
			}

			// 没有找到，那么这个变量，它关联的上一层变量呢
			// 递归进行查找
			exp := beforeSymbol.VarInfo.ReferExp
			subFindFlag := false
			symList := a.FindDeepSymbolList(beforeSymbol.FileName, exp, comParam, findExpList, false,
				beforeSymbol.VarInfo.VarIndex)
			for _, oneSymbol := range symList {
				if subSym := a.symbolHasSubKey(oneSymbol, strAfter, comParam, findExpList); subSym != nil {
					funcSymbol = subSym
					subFindFlag = true
					break
				}
			}

			if !subFindFlag {
				return nil
			}
		}
	}

	// 如果没有找到，直接返回
	if funcSymbol == nil {
		return nil
	}

	// 获取对应的函数注释返回值
	if ok, funcFile := a.getFuncIndexReturnSymbol(funcSymbol, varIndex, node, comParam, findExpList); ok {
		return funcFile
	}

	if funcSymbol.VarInfo == nil {
		return
	}

	if funcSymbol.VarInfo.ReferFunc != nil {
		// -- 这里获取第一个返回值有时推导不出来，因为函数可能有多个返回值，是否要获取最后一个返回值
		flag, expReturn := funcSymbol.VarInfo.ReferFunc.GetReturnIndexExp(varIndex)
		if !flag {
			// 没有找到
			return
		}

		if symTmp := extChangeSymbol(expReturn, funcSymbol.FileName); symTmp != nil {
			return symTmp
		}

		// 递归查找
		tmpList := a.FindDeepSymbolList(funcSymbol.FileName, expReturn, comParam, findExpList, true, varIndex)
		if len(tmpList) == 1 {
			return tmpList[0]
		}

		if len(tmpList) == 0 {
			return nil
		}

		firstSmbol := tmpList[0]
		lastSmbol := tmpList[len(tmpList)-1]
		if firstSmbol.AnnotateType != nil && lastSmbol.AnnotateType == nil {
			return firstSmbol
		}
		if firstSmbol.AnnotateType == nil && lastSmbol.AnnotateType != nil {
			return lastSmbol
		}

		if firstSmbol.VarInfo == nil {
			return lastSmbol
		}

		if lastSmbol.VarInfo == nil {
			return firstSmbol
		}

		if len(lastSmbol.VarInfo.SubMaps) > len(firstSmbol.VarInfo.SubMaps) {
			return lastSmbol
		}

		return firstSmbol
	}

	// 没有找到，那么这个变量，它关联的上一层变量呢
	// 递归进行查找
	exp := funcSymbol.VarInfo.ReferExp
	// 这里可能会引起一样的递归
	symList := a.FindDeepSymbolList(funcSymbol.FileName, exp, comParam, findExpList, false, funcSymbol.VarInfo.VarIndex)
	for _, oneSymbol := range symList {
		// 如果找到了，判断是否有注解类型
		// 获取对应的函数注释返回值
		if ok, funcFile := a.getFuncIndexReturnSymbol(oneSymbol, varIndex, node, comParam, findExpList); ok {
			return funcFile
		}

		if oneSymbol.VarInfo == nil || oneSymbol.VarInfo.ReferFunc == nil {
			continue
		}

		ok, expReturn := oneSymbol.VarInfo.ReferFunc.GetReturnIndexExp(varIndex)
		if !ok {
			// 没有找到
			return
		}

		// 递归查找
		tmpList := a.FindDeepSymbolList(oneSymbol.FileName, expReturn, comParam, findExpList, true, varIndex)
		if len(tmpList) > 0 {
			return tmpList[len(tmpList)-1]
		}

		return nil
	}
	return nil
}

// 判断是否重复了，递归可能会包含重复的
func isHasSymbol(symbol *common.Symbol, symList []*common.Symbol) bool {
	if len(symList) == 0 {
		return false
	}

	// todo 为空的都认为不重复
	if symbol.VarFlag == common.FirstAnnotateFlag || symbol.VarInfo == nil {
		return false
	}

	for _, oneSymbol := range symList {
		if symbol.FileName != oneSymbol.FileName {
			continue
		}

		if !lexer.CompareTwoLoc(&(symbol.VarInfo.Loc), &(oneSymbol.VarInfo.Loc)) {
			continue
		}

		oneExp := symbol.VarInfo.ReferExp
		twoExp := oneSymbol.VarInfo.ReferExp

		if common.IsSameExp(oneExp, twoExp) {
			return true
		}
	}

	return false
}

// isHasFindExpFile 判断跟踪的引用Exp之前是否跟踪到过
func isHasFindExpFile(findExpList *[]common.FindExpFile, luaInFile string, node ast.Exp) bool {
	for _, oneFindExp := range *findExpList {
		if oneFindExp.FileName != luaInFile {
			continue
		}

		oneLoc := common.GetExpLoc(oneFindExp.FindExp)
		twoLoc := common.GetExpLoc(node)

		if !lexer.CompareTwoLoc(&(oneLoc), &(twoLoc)) {
			continue
		}

		if common.IsSameExp(oneFindExp, node) {
			return true
		}
	}

	return false
}

// 插入一个根据到的中间引用结果
func insertFindExpFile(findExpList *[]common.FindExpFile, luaInFile string, node ast.Exp) {
	oneFindExp := common.FindExpFile{
		FileName: luaInFile,
		FindExp:  node,
	}
	*findExpList = append(*findExpList, oneFindExp)
}
