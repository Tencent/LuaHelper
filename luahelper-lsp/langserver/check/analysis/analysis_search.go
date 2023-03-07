package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"strings"
)

// 获取调用函数指向的定义, 在局部变量或者全局变量中查找
// gFlag 表示是否为_G的全局变量，如果为true表示为_G的，false表示不确定
// strProPre 为协议的前缀，例如c2s. s2s
func (a *Analysis) findStrFuncRefer(loc lexer.Location, strName string, gFlag bool, strProPre string) (f *common.FuncInfo, findTerm int) {
	// 1) 判断该变量是否为lua模块自带或是框架需要屏蔽的变量
	if common.GConfig.IsIgnoreNameVar(strName) {
		return nil, 0
	}

	fileResult := a.curResult

	// 2) 判断是否为需要忽略的文件中的变量
	if common.GConfig.IsIgnoreFileDefineVar(fileResult.Name, strName) {
		return nil, 0
	}

	scope := a.curScope
	fi := a.curFunc

	// 3) 查找局部变量指向的函数信息
	if !gFlag {
		if locVar, ok := scope.FindLocVar(strName, loc); ok {
			return locVar.ReferFunc, 0
		}
	}

	// 4) 第三步查找全局中是否有该变量
	firstFile := a.getFirstFileResult(fileResult.Name)
	if a.isSecondTerm() {
		secondFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			if ok, oneVar := secondFileResult.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar.ReferFunc, 2
			}

			if ok, oneVar := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermSecond, strProPre); ok {
				return oneVar.ReferFunc, 2
			}
		} else {
			// 非底层的函数，需要查找全局的变量
			if ok, oneVar := firstFile.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar.ReferFunc, 1
			}

			if ok, oneVar := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermFirst, strProPre); ok {
				return oneVar.ReferFunc, 1
			}
		}
	} else if a.isThirdTerm() {
		thirdFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找
			if ok, oneVar := thirdFileResult.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar.ReferFunc, 3
			}
		} else {
			// 非底层的函数，需要查找全局的变量
			if ok, oneVar := firstFile.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
				return oneVar.ReferFunc, 1
			}
		}

		// 查找所有的
		if ok, oneVar := a.AnalysisThird.ThirdStruct.FindThirdGlobalGInfo(gFlag, strName, strProPre); ok {
			return oneVar.ReferFunc, 3
		}
	}

	return nil, 0
}

// 获取函数调用的指向的func信息
func (a *Analysis) getFuncCallReferFunc(node *ast.FuncCallStat) (referFunc *common.FuncInfo, strName string, findTerm int) {
	// 1) 直接为 a("1", "2", "3") 函数，检查参数
	if nameExp, ok := node.PrefixExp.(*ast.NameExp); ok {
		strName = nameExp.Name
		referFunc, findTerm = a.findStrFuncRefer(nameExp.Loc, strName, false, "")
		return referFunc, strName, findTerm
	}

	// 2) 判断是否为_G.a() 或是 local a = import("one.lua") a.test()这样的场景
	taExp, ok := node.PrefixExp.(*ast.TableAccessExp)
	if !ok {
		return nil, strName, 0
	}

	fileResult := a.curResult
	scope := a.curScope

	strTabName := common.GetExpName(taExp.PrefixExp)
	strKeyName := common.GetExpName(taExp.KeyExp)

	loc := common.GetTableKeyLoc(taExp)
	// 2) 为_G的全局变量
	if strTabName == "!_G" {
		if !common.JudgeSimpleStr(strKeyName) {
			return nil, strKeyName, 0
		}
		referFunc, findTerm = a.findStrFuncRefer(loc, strKeyName, true, "")
		return referFunc, strKeyName, findTerm
	} else if strings.HasPrefix(strTabName, "!") && !strings.Contains(strTabName, ".") && common.JudgeSimpleStr(strKeyName) {
		// 判断是否为直接协议的调用c2s 或是s2s
		strProPre := common.GConfig.GetStrProtocol(strTabName)
		if strProPre != "" {
			// 为协议的调用
			referFunc, findTerm = a.findStrFuncRefer(loc, strKeyName, false, strProPre)
			return referFunc, strName, findTerm
		}

		// 下面的为模块的调用，先判断最简单的用例
		strName := strTabName[1:]

		// 2.1) 查找局部变量的引用
		var findVar *common.VarInfo

		// 需要先判断该变量是否直接含义key的函数
		loc = common.GetTablePrefixLoc(taExp)
		if locVar, ok := scope.FindLocVar(strName, loc); ok {
			findVar = locVar
		} else {
			// 2.2) 局部变量没有找到，查找全局变量
			if ok, oneVar := fileResult.FindGlobalVarInfo(strName, false, ""); ok {
				findVar = oneVar
			}
		}

		if findVar == nil {
			return nil, strName, 0
		}

		subVar := common.GetVarSubGlobalVar(findVar, strKeyName)
		if subVar != nil {
			return subVar.ReferFunc, strKeyName, 0
		}

		referInfo := findVar.ReferInfo
		if referInfo == nil {
			return nil, strName, 0
		}

		referFile := a.Projects.GetFirstReferFileResult(referInfo)
		if referFile == nil {
			return nil, strName, 0
		}

		// 排查过滤掉的文件变量
		referLuaFile := referInfo.ReferStr
		if common.GConfig.IsIgnoreFileDefineVar(referLuaFile, strKeyName) {
			return nil, strName, 0
		}

		if referInfo.ReferType == common.ReferTypeRequire {
			find, returnExp := referFile.MainFunc.GetLastOneReturnExp()
			if !find {
				return nil, strName, 0
			}

			subExp, ok := returnExp.(*ast.NameExp)
			if !ok {
				return nil, strName, 0
			}

			varInfo, ok := referFile.MainFunc.MainScope.FindLocVar(subExp.Name, subExp.Loc)
			if !ok {
				_, varInfo = referFile.FindGlobalVarInfo(subExp.Name, false, "")
			}

			if varInfo == nil {
				return nil, strName, 0
			}

			subVar := common.GetVarSubGlobalVar(varInfo, strKeyName)
			if subVar == nil {
				return nil, strName, 0
			}

			return subVar.ReferFunc, strKeyName, 0
		}

		if ok, oneVar := referFile.FindGlobalVarInfo(strKeyName, false, ""); ok {
			return oneVar.ReferFunc, strKeyName, 0
		}

		return nil, strKeyName, 0
	}

	return nil, strName, 0
}

// 查找lua文件中所有引用的外表文件，看是否有这样的用法
// 返回true，表示使用了下面的用户，正常
// require("lfs");
// lfs.mkdir("log");
func (a *Analysis) findReferModule(strName string) bool {
	fileResult := a.curResult
	for _, oneRefer := range fileResult.ReferVec {
		strReferName := oneRefer.ReferStr
		if strName != strReferName {
			continue
		}

		if oneRefer.ReferType != common.ReferTypeRequire {
			continue
		}

		referFile := a.GetReferFileResult(oneRefer, results.CheckTermFirst)
		if referFile == nil {
			log.Debug("referFileResult is nil, strName=%s, validstr=%s", strName, oneRefer.ReferValidStr)
			return true
		}
	}

	return false
}

// 判断是否要忽略下列右边的变量没有定义，赋值的问题
// 1) enable_reload = (enable_reload == nil) and true
// 2) enable_reload = (enable_reload ~= nil) and true
// 3) enable_reload = enable_reload or 1
func (a *Analysis) ignoreCircleDefine(strName string, loc lexer.Location, findVar *common.VarInfo,
	binParentExp *ast.BinopExp) bool {
	if binParentExp == nil {
		return false
	}

	// 判断是否同一行的，只考虑同一行的
	if loc.StartLine != findVar.Loc.StartLine {
		return false
	}

	op := binParentExp.Op
	if op == lexer.TkOpEq || op == lexer.TkOpNe || op == lexer.TkOpAnd || op == lexer.TkOpOr {
		subExpLeft := binParentExp.Exp1
		strSubLeft := common.GetExpName(subExpLeft)
		strSubLeftSimple := common.GetSimpleValue(strSubLeft)

		subExpRight := binParentExp.Exp2
		strSubRight := common.GetExpName(subExpRight)
		strSubRightSimple := common.GetSimpleValue(strSubRight)
		if strName == strSubLeftSimple || strName == strSubRightSimple {
			return true
		}
	}

	return false
}

// 查找全局变量的定义，判断是否有找到，没有找到输出错误
// callGflag 表明调用的是否为_G
// gFindGlag 表明查找定义是否为_G
// binParentExp 为二元表达式，父的BinopExp指针， 例如 a = b and c，当对c变量调用cgExp时候，binParentExp为b and c
func (a *Analysis) findGlobalVar(strName string, loc lexer.Location, strProPre string, callGflag bool,
	gFindGlag bool, nameExp ast.Exp, binParentExp *ast.BinopExp) {
	fileResult := a.curResult

	// 0) 如果是在五轮， 判断传人的系统名字是变量还是函数
	if a.isFiveTerm() {
		subExp, ok := nameExp.(*ast.NameExp)
		if ok && subExp.Name == "self" {
			return
		}

		a.ColorResult.InsertOneColorElem(common.CTGlobalVar, &loc)
		return
	}

	// 1）判断该变量是否为lua模块自带或是框架需要屏蔽的变量
	if common.GConfig.IsIgnoreNameVar(strName) {
		return
	}

	// 2) 判断是否为需要忽略的文件中的变量
	if common.GConfig.IsIgnoreFileDefineVar(fileResult.Name, strName) {
		return
	}

	// 3) 判断是否为_G.b = _G.b or 0，忽略这种
	if !a.isFourTerm() && a.ignoreInfo.strName == strName && a.ignoreInfo.line == loc.StartLine {
		return
	}

	fi := a.curFunc
	firstFile := a.getFirstFileResult(fileResult.Name)
	if firstFile == nil {
		return
	}

	preShowStr := ""
	if callGflag {
		preShowStr = "_G."
	}
	if strProPre != "" {
		preShowStr = strProPre + "."
	}

	// 4) 根据不同的轮数查找全局表中是否有该变量
	if a.isSecondTerm() {
		if _, ok := common.GConfig.LuaInMap[strName]; ok {
			return
		}

		if strProPre != "" && common.GConfig.IsIgnoreProtocolPreVar() {
			// 如果协议前缀的告警，忽略告警，不进行查找
			return
		}

		secondFileResult := fileResult
		// 是否因为循环引用或是加载顺序导致的变量未定义
		cycleFlag := false

		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找，只用在checkTerm = 2中的变量查找
			if ok, _ := secondFileResult.FindGlobalVarInfo(strName, gFindGlag, strProPre); ok {
				return
			}

			// 向工程的第二阶段全局_G符号表中查找
			if ok, _ := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermSecond, strProPre); ok {
				return
			}

			// 最底层的函数，调用了顺序导致的变量未定义
			if ok, findVar := firstFile.FindGlobalVarInfo(strName, gFindGlag, strProPre); ok {
				cycleFlag = true

				// 判断是否要忽略指定的循环定义情况
				if a.ignoreCircleDefine(strName, loc, findVar, binParentExp) {
					return
				}
			}

			// 向工程的globalMaps中查找变量
			// 向工程的第一阶段全局_G符号表中查找
			if !cycleFlag {
				if ok, findVar := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermFirst, strProPre); ok {
					cycleFlag = true

					// 判断是否要忽略指定的循环定义情况
					if a.ignoreCircleDefine(strName, loc, findVar, binParentExp) {
						return
					}
				}
			}
		} else {
			// 非顶层的函数，需要查找自己定义的全局的变量
			if ok, _ := firstFile.FindGlobalVarInfo(strName, gFindGlag, strProPre); ok {
				return
			}

			// 向工程的第一阶段全局_G符号表中查找
			if ok, _ := a.SingleProjectResult.FindGlobalGInfo(strName, results.CheckTermFirst, strProPre); ok {
				return
			}
		}

		// 没有找到该变量，输出错误
		if cycleFlag {
			errStr := fmt.Sprintf("crcular reference or load order error, var not define: %s%s", preShowStr, strName)
			secondFileResult.InsertError(common.CheckErrorCycleDefine, errStr, loc)
		} else {
			errStr := fmt.Sprintf("var not define: %s%s", preShowStr, strName)
			secondFileResult.InsertError(common.CheckErrorNoDefine, errStr, loc)
		}
	} else if a.isThirdTerm() {
		if _, ok := common.GConfig.LuaInMap[strName]; ok {
			return
		}

		if strProPre != "" && common.GConfig.IsIgnoreProtocolPreVar() {
			// 如果协议前缀的告警，忽略告警，不进行查找
			return
		}

		thirdFileResult := fileResult
		if fi.FuncLv == 0 {
			// 最顶层的函数，只在前面的定义中查找，只用在checkTerm = 2中的变量查找
			if ok, _ := thirdFileResult.FindGlobalVarInfo(strName, gFindGlag, strProPre); ok {
				return
			}

			// 是否因为循环引用或是加载顺序导致的变量未定义
			if ok, findVar := firstFile.FindGlobalVarInfo(strName, gFindGlag, strProPre); ok {
				// 判断是否要忽略指定的循环定义情况
				if a.ignoreCircleDefine(strName, loc, findVar, binParentExp) {
					return
				}

				errStr := fmt.Sprintf("crcular reference or load order error, var not define: %s%s", preShowStr, strName)
				thirdFileResult.InsertError(common.CheckErrorCycleDefine, errStr, loc)
				return
			}
		} else {
			// 非顶层的函数，需要查找自己定义的全局的变量
			if ok, _ := firstFile.FindGlobalVarInfo(strName, gFindGlag, strProPre); ok {
				return
			}
		}

		// 向第三阶段，散落的全局结构获取全局对象
		if ok, _ := a.AnalysisThird.ThirdStruct.FindThirdGlobalGInfo(gFindGlag, strName, strProPre); ok {
			return
		}
		errStr := fmt.Sprintf("var not define: %s%s", preShowStr, strName)
		thirdFileResult.InsertError(common.CheckErrorNoDefine, errStr, loc)
	} else if a.isFourTerm() {
		// 非顶层的函数，需要查找自己定义的全局的变量
		if ok, findVar := firstFile.FindGlobalVarInfo(strName, gFindGlag, strProPre); ok {
			// 判断是否是自己所要的引用关系
			a.ReferenceResult.MatchVarInfo(a, strName, firstFile.Name, findVar, fi, "", nameExp, false)
			return
		}
		a.ReferenceResult.FindProjectGlobal(a, strName, strProPre, fi, "", nameExp)
	}
}

// 判断局部变量是否定义了未使用
func (a *Analysis) checkLocVarNotUse(node *ast.NameExp) {
	scope := a.curScope
	if locVar, ok := scope.FindLocVar(node.Name, node.Loc); ok {
		locVar.IsUse = true
	}
}

// 查找引用的NameExp 变量是否之前定义过
// binParentExp 为二元表达式，父的BinopExp指针， 例如 a = b and c，当对c变量调用cgExp时候，binParentExp为b and c
func (a *Analysis) findNameStr(node *ast.NameExp, binParentExp *ast.BinopExp) {
	// 第二轮分析引用的变量是否之前有定义
	strName := node.Name
	// 1) self进行替换
	if strName == "self" {
		strName = a.ChangeSelfToReferVar(strName, "")
		index := strings.Index(strName, ".")
		if index >= 0 {
			strName = strName[0:index]
		}
	}

	fileResult := a.curResult
	fi := a.curFunc
	scope := a.curScope

	// 2) 查找局部变量或upvalue中是否有该变量
	if locVar, ok := scope.FindLocVar(strName, node.Loc); ok {
		if a.isFourTerm() {
			// 判断是否是自己所要的引用关系
			a.ReferenceResult.MatchVarInfo(a, strName, fileResult.Name, locVar, fi, "", node, false)
		}
		return
	}

	// 3) 判断是否为，引入的其他模块名，直接调用的模块中的变量
	// 例如 require("lfs");
	// 后面调用 lfs.mkdir("log");
	if a.findReferModule(strName) {
		return
	}

	// 4）查找公共全局变量的函数
	a.findGlobalVar(strName, node.Loc, "", false, false, node, binParentExp)
}

// 直接查找全局变量，判断是否有定义
func (a *Analysis) checkfindGVar(strName string, loc lexer.Location, strProPre string, nameExp ast.Exp) {
	// 查找_G的符号，也会扩大到非_G的
	gFlag := !common.GConfig.GetGVarExtendFlag()
	if strProPre != "" {
		gFlag = false
	}

	// 查找公共全局变量的函数
	a.findGlobalVar(strName, loc, strProPre, true, gFlag, nameExp, nil)
}

// 搜索指定的三层变量调用
func (a *Analysis) findThreeLevelCall(node ast.Exp, nameExp ast.Exp) {
	if !a.isFourTerm() {
		return
	}

	// 1) 判断strKey是否为简单的字符串
	strKey := common.GetExpName(nameExp)

	// 如果不是简单字符，退出
	if !common.JudgeSimpleStr(strKey) {
		return
	}
	keyLoc := common.GetExpLoc(nameExp)

	// 2）第二轮或第三轮判断table取值是否有定义
	strTable := common.GetExpName(node)
	if strings.Contains(strTable, "#") {
		return
	}
	strTableArry := strings.Split(strTable, ".")

	// self进行替换
	if strTableArry[0] == "!self" {
		strTableArry[0] = a.ChangeSelfToReferVar(strTableArry[0], "!")
	}
	_, strTableArry[0] = common.StrRemoveSigh(strTableArry[0])
	for i := 0; i < len(strTableArry); i++ {
		if !common.JudgeSimpleStr(strTableArry[i]) {
			return
		}
	}

	fileResult := a.curResult
	fi := a.curFunc
	scope := a.curScope
	strOne := strTableArry[0]
	if strOne == "_G" {
		if len(strTableArry) <= 1 {
			return
		}

		strTwo := strTableArry[1]
		strTableArry = strTableArry[1:]
		// 为协议的前缀
		strProPre := ""
		if common.GConfig.IsStrProtocol(strTwo) {
			strProPre = strTwo
		}

		strPreExp := strings.Join(strTableArry, ".")

		// 查找全局中是否有该变量
		firstFile := a.getFirstFileResult(fileResult.Name)
		if firstFile != nil {
			if ok, findVar := firstFile.FindGlobalVarInfo(strTwo, true, strProPre); ok {
				// 判断是否是自己所要的引用关系
				a.ReferenceResult.MatchVarInfo(a, strTwo, firstFile.Name, findVar, fi, strPreExp, nameExp, false)
				return
			}
		}

		// 所有工程的文件
		a.ReferenceResult.FindProjectGlobal(a, strTwo, strProPre, fi, strPreExp, nameExp)
	} else {
		// 只能是引用
		var referInfo *common.ReferInfo
		if locVar, ok := scope.FindLocVar(strOne, keyLoc); ok {
			referInfo = locVar.ReferInfo
			if referInfo == nil {
				a.ReferenceResult.MatchVarInfo(a, strOne, fileResult.Name, locVar, fi, strTable, nameExp, false)
				return
			}
		} else {
			// 2.2) 局部变量没有找到，查找全局变量
			if ok, findVar := fileResult.FindGlobalVarInfo(strOne, false, ""); ok {
				referInfo = findVar.ReferInfo
				if referInfo == nil {
					a.ReferenceResult.MatchVarInfo(a, strOne, fileResult.Name, findVar, fi, strTable, nameExp, false)
					return
				}
			} else {
				// 所有工程的文件
				if a.ReferenceResult.FindProjectGlobal(a, strOne, "", fi, strTable, nameExp) {
					return
				}
			}
		}

		// 没有找到引用信息，退出，目前只处理import引用的方式
		if referInfo == nil {
			return
		}

		// 切除第一个引用的值
		strTableArry = strTableArry[1:]
		strTwo := strKey
		strPreExp := strings.Join(strTableArry, ".")
		if len(strTableArry) >= 1 {
			strTwo = strTableArry[0]
		} else {
			strPreExp = ""
		}

		// 引用的lua文件
		referValidStr := referInfo.ReferValidStr
		// 排查过滤掉的文件变量
		if common.GConfig.IsIgnoreFileDefineVar(referValidStr, strTwo) {
			return
		}

		referFileResult := a.GetReferFileResult(referInfo, results.CheckTermFour)
		if referFileResult == nil {
			return
		}

		referFrameSubType := a.Projects.GetReferFrameType(referInfo)
		if referFrameSubType == common.RtypeImport {
			if ok, findVar := referFileResult.FindGlobalVarInfo(strTwo, false, ""); ok {
				// 所有工程的文件
				a.ReferenceResult.MatchVarInfo(a, strTwo, referFileResult.Name, findVar, fi, strPreExp, nameExp, false)
				return
			}
		} else if referFrameSubType == common.RtypeRequire {
			// 这里为require一个文件，获取这个文件的返回值
			locVar := referFileResult.MainFunc.GetOneReturnVar()
			if locVar == nil {
				return
			}

			subGlobal := common.GetVarSubGlobalVar(locVar, strTwo)
			if subGlobal != nil {
				// 所有工程的文件
				a.ReferenceResult.MatchVarInfo(a, strTwo, referFileResult.Name, locVar, fi, strPreExp, nameExp, true)
			}
		}
	}
}

// 对变量的调用进行展开，例如:
// local a = {}
// print(a.b.c)
// a变量的身上挂了一个字符串属性：b.c
func (a *Analysis) expandVarStrMap(node *ast.TableAccessExp) {
	// 下面的判断只在第一轮，且是非实时检查时才触发
	if !a.isFirstTerm() {
		return
	}

	strKey := common.GetExpName(node.KeyExp)
	if !common.JudgeSimpleStr(strKey) {
		return
	}

	strPre := common.GetExpName1(node.PrefixExp)
	preVec := strings.Split(strPre, ".")
	if len(preVec) == 0 {
		return
	}

	for i := 1; i < len(preVec); i++ {
		if strings.HasPrefix(preVec[i], "!") {
			// 如果是以!开头，表示为变量，进行替换处理
			preVec[i] = "!var"
		}
	}

	strOne := preVec[0]
	strName := common.GetSimpleValue(strOne)
	if strName == "" {
		return
	}
	if strName == "self" {
		strName = a.ChangeSelfToReferVar(strName, "")
	}

	// 2.1) 查找局部变量的引用
	loc := common.GetTablePrefixLoc(node)
	varInfo := a.findFileVar(strName, loc)
	if varInfo == nil {
		return
	}

	vecExpand := preVec[1:]
	vecExpand = append(vecExpand, strKey)

	strExpand := strings.Join(vecExpand, ".")
	if strExpand == "" {
		return
	}

	if varInfo.ExpandStrMap == nil {
		varInfo.ExpandStrMap = map[string]struct{}{}
	}
	varInfo.ExpandStrMap[strExpand] = struct{}{}
}

func (a *Analysis) findFileVar(strName string, loc lexer.Location) *common.VarInfo {
	scope := a.curScope
	// 首先查找当前局部变量下有没有当前表名
	if locVar, ok := scope.FindLocVar(strName, loc); ok {
		return locVar
	}

	// 再次查找当前文件下全局变量有没有当前表名
	globalMaps := a.curResult.GlobalMaps
	if gVar, ok := globalMaps[strName]; ok {
		return gVar
	}

	// 判断当前文件是否是第一次分析table
	// 比如a.b.c.d 会先从 a.b.c.d 再从 a.b.c 最后 a.b 进入此函数
	nodefineMaps := a.curResult.NodefineMaps
	if noVar, ok := nodefineMaps[strName]; ok {
		return noVar
	}

	return nil
}

// 第一轮， if not a then ，这样的赋值语句a.b = 1，检查
// binParentExp 为二元表达式，父的BinopExp指针， 例如 a = b and c，当对c变量调用cgExp时候，binParentExp为b and c
func (a *Analysis) checkIfNotTableAccess(node *ast.TableAccessExp, binParentExp *ast.BinopExp) {
	// 下面的判断只在第一轮，且是非实时检查时才触发
	if !a.isFirstTerm() || a.realTimeFlag {
		return
	}

	strTable := common.GetExpName(node.PrefixExp)
	strName := common.GetSimpleValue(strTable)
	if strName == "" {
		return
	}

	fi := a.curFunc
	scope := a.curScope

	// 2.1) 查找局部变量的引用
	loc := common.GetTablePrefixLoc(node)
	varInfo, flag := scope.FindLocVar(strName, loc)
	if !flag {
		return
	}
	// if not a
	// a.b  这样的告警查看

	// 当前的 common.FuncInfo 为 fi *FuncInfo，
	// 向上的scope中查找 strName, 并且返回 locInfo 和 funcInfo
	findNotVal, findNotScop := scope.CheckNotVarInfo(strName)
	if findNotVal == nil {
		return
	}

	fileResult := a.curResult
	findNotScopFi := findNotScop.FindMinFunc()
	// 需要是在同一个fi里面
	if findNotScopFi == fi && findNotVal == varInfo {
		// 需要屏蔽掉下面的例子
		/*
			 if not a then
				 print(a and a.b)		 -- 前面有and语句，且and语句的第一个表达式为a
				 print(a ~= nil and a.b) -- 前面有and语句，且and语句的第一个表达式为a ~= nil
				 print(a == nil and a.b) -- 前面有and语句，且and语句的第一个表达式为a == nil
			 end
		*/
		if binParentExp != nil && binParentExp.Op == lexer.TkOpAnd {
			// 1) 判断左边的变量，是否直接为strName; 即为这样的表达式 a and a.b 忽略这类告警
			expLeft := binParentExp.Exp1
			strLeft := common.GetExpName(expLeft)
			strLeftSimple := common.GetSimpleValue(strLeft)
			if strName == strLeftSimple {
				return
			}

			// 2) 判断左边的变量，是否为 strName ~= nil; 即为这样的表达式 a ~= nil and a.b 忽略这类告警
			// 或是 a == nil and a.b 忽略这类告警
			subBinop, flag := expLeft.(*ast.BinopExp)
			if flag && (subBinop.Op == lexer.TkOpNe || subBinop.Op == lexer.TkOpEq) {
				subExpLeft := subBinop.Exp1
				strSubLeft := common.GetExpName(subExpLeft)
				strSubLeftSimple := common.GetSimpleValue(strSubLeft)

				subExpRight := subBinop.Exp2
				strSubRight := common.GetExpName(subExpRight)
				strSubRightSimple := common.GetSimpleValue(strSubRight)
				if strName == strSubLeftSimple || strName == strSubRightSimple {
					return
				}
			}
		}

		errStr := fmt.Sprintf("before if not %s, %s may be is nil", strName, strName)
		fileResult.InsertError(common.CheckErrorNotIfVar, errStr, loc)
	}
}

// 查找冒号的调用, 判断是否要找的引用其他的
// a = {}
// function a:test() end
// a:test() -- 冒号调用时候，判断是否要查找
func (a *Analysis) findFuncColon(prefixExp ast.Exp, nameExp ast.Exp, nodeLoc lexer.Location) {
	// 第一轮跳过
	if a.isFirstTerm() {
		return
	}

	fileResult := a.curResult
	fi := a.curFunc
	scope := a.curScope

	strKey := common.GetExpName(nameExp)
	keyLoc := common.GetExpLoc(nameExp)
	// 如果不是简单字符，退出
	if !common.JudgeSimpleStr(strKey) {
		return
	}

	// 第二轮或第三轮判断table取值是否有定义
	strTable := common.GetExpName(prefixExp)
	strProPre := common.GConfig.GetStrProtocol(strTable)

	// self进行转换
	if strTable == "!self" {
		strTable = a.ChangeSelfToReferVar(strTable, "!")
	}

	// 1) _G.aa这样的用例，判断全局变量aa是否有定义
	if strTable == "!_G" {
		// 直接查找全局变量strKey，判断是否有定义
		a.checkfindGVar(strKey, keyLoc, "", nameExp)
		return
	}

	// 2) c2s.One() 协议前缀的调用
	if strProPre != "" && !strings.Contains(strTable, ".") {
		// 直接查找协议strKey，判断是否有定义
		a.checkfindGVar(strKey, keyLoc, strProPre, nameExp)
		return
	}

	if !a.isFourTerm() {
		return
	}

	// 3.1) 在第四轮中查找三层次的调用关系
	// 例如 local a = import("one.lua") ，调用 a.bb.cc 其中bb为定义的table， cc为table中的字符串key值
	// 或是 _G.bb.cc 其中bb为定义的全局table， cc为table中的字符串key值
	if a.isFourTerm() && strings.Contains(strTable, ".") {
		a.findThreeLevelCall(prefixExp, nameExp)
		return
	}

	// 3) 判断引用import里面的成员是有定义
	// local a = import("one.lua")
	// b = a.dd, 判断a.dd是否有定义（需要跟进one.lua里面，判断dd是否为全局变量）
	// 判断是否为两层的调用
	strName := common.GetSimpleValue(strTable)
	if strName == "" {
		return
	}

	// 2.1) 查找局部变量的引用
	// 这里loc = keyLoc 问题不大
	loc := keyLoc
	var referInfo *common.ReferInfo
	if locVar, ok := scope.FindLocVar(strName, loc); ok {
		referInfo = locVar.ReferInfo
		// 第四轮查找引用的时候，判断是否为返回table的字符串key
		if a.isFourTerm() && referInfo == nil {
			a.ReferenceResult.MatchVarInfo(a, strName, fileResult.Name, locVar, fi, strTable, nameExp, false)
			return
		}
	} else {
		// 2.2) 局部变量没有找到，查找全局变量
		if ok, findVar := fileResult.FindGlobalVarInfo(strName, false, ""); ok {
			referInfo = findVar.ReferInfo
		}

		if a.isFourTerm() && referInfo == nil {
			// 查找全局中是否有该变量
			firstFile := a.getFirstFileResult(fileResult.Name)
			if firstFile != nil {
				// find self globalInfo
				if ok, findVar := firstFile.FindGlobalVarInfo(strName, false, ""); ok {
					// 判断是否是自己所要的引用关系
					a.ReferenceResult.MatchVarInfo(a, strName, firstFile.Name, findVar, fi, strTable, nameExp, false)
					return
				}
			}

			// 所有工程的文件
			if a.ReferenceResult.FindProjectGlobal(a, strName, "", fi, strTable, nameExp) {
				return
			}
		}
	}

	if referInfo == nil {
		return
	}

	referFrameSubType := a.Projects.GetReferFrameType(referInfo)
	if referFrameSubType == common.RtypeRequire && a.isFourTerm() {
		// 第四轮，查找应用
		checkTerm := a.getChangeCheckTerm()
		// 非顶层的function 查找第一次判断的
		referFile := a.GetReferFileResult(referInfo, checkTerm)
		if referFile == nil {
			// 没有加载引用的文件，可能是so文件
			return
		}

		locVar := referFile.MainFunc.GetOneReturnVar()
		if locVar == nil {
			return
		}

		subGlobal := common.GetVarSubGlobalVar(locVar, strKey)
		if subGlobal != nil {
			// 所有工程的文件
			// analysis.analysisFour.matchVarInfo(strKey, referFile.name, subGlobal, fi, "", nameExp, true)
			a.ReferenceResult.MatchVarInfo(a, strKey, referFile.Name, locVar, fi, "", nameExp, true)
		}

		return
	}

	if referFrameSubType != common.RtypeImport {
		return
	}

	// 引用的lua文件
	referValidStr := referInfo.ReferValidStr
	// 排查过滤掉的文件变量
	if common.GConfig.IsIgnoreFileDefineVar(referValidStr, strKey) {
		return
	}

	referFile := a.GetReferFileResult(referInfo, results.CheckTermFour)
	if referFile == nil {
		return
	}

	if ok, oneVar := referFile.FindGlobalVarInfo(strKey, false, ""); ok {
		// 所有工程的文件
		a.ReferenceResult.MatchVarInfo(a, strKey, referFile.Name, oneVar, fi, "", nameExp, false)
		return
	}
	// todo  这里可能需要增加判断校验
}

// 获取矫正后的checkTerm
func (a *Analysis) getChangeCheckTerm() (checkTerm results.CheckTerm) {
	fi := a.curFunc
	checkTerm = results.CheckTermFirst
	if fi.FuncLv == 0 {
		checkTerm = results.CheckTermSecond
	}
	if a.isThirdTerm() {
		checkTerm = results.CheckTermThird
	}
	if a.isFourTerm() {
		checkTerm = results.CheckTermFour
	}
	if a.isFiveTerm() {
		checkTerm = results.CheckTermFive
	}

	return checkTerm
}

// 变量根的block，查询赋值的self
func (a *Analysis) getAsignSelfReferName() string {
	fi := a.curFunc
	fileResult := a.curResult
	for _, stat := range fileResult.Block.Stats {
		assignStat, ok := stat.(*ast.AssignStat)
		if !ok {
			continue
		}

		if assignStat.Loc.EndLine < fi.Loc.StartLine {
			continue
		}

		// 范围超过了
		if assignStat.Loc.StartLine > fi.Loc.EndLine {
			break
		}

		if len(assignStat.VarList) != 1 {
			break
		}

		taExp, ok := assignStat.VarList[0].(*ast.TableAccessExp)
		if !ok {
			continue
		}

		keyFuncExp, ok1 := assignStat.ExpList[0].(*ast.FuncDefExp)
		if !ok1 {
			continue
		}

		if !keyFuncExp.IsColon {
			continue
		}

		tabName := common.GetExpName(taExp.PrefixExp)
		strKeyName := common.GetExpName(taExp.KeyExp)
		if !common.JudgeSimpleStr(strKeyName) {
			continue
		}

		tabOne, tabTwo := common.GetTableStrTwoStr(tabName)
		if tabOne != "" && tabTwo != "" {
			if tabOne == "_G" {
				return tabTwo
			}
		} else {
			strName := common.GetSimpleValue(tabName)
			if strName != "" {
				return strName
			}
		}
	}
	return ""
}

// ChangeSelfToReferVar 增加前缀
// strTable 值为self，转换为对应的变量
func (a *Analysis) ChangeSelfToReferVar(strTable string, prefixStr string) (str string) {
	fi := a.curFunc
	str = strTable
	var firstColonFunc *common.FuncInfo
	fileResult := a.curResult
	if a.isFirstTerm() {
		firstColonFunc = fi
	} else {
		// 获取第一轮的结构，因为此时RelateVar的结构还没有构造，只能查找第一轮的数据
		fileStruct, _ := a.Projects.GetFirstFileStuct(fileResult.Name)
		if fileStruct == nil || fileStruct.FileResult == nil {
			return
		}

		var firstFunc *common.FuncInfo
		if fi.FuncID >= 0 && len(fileStruct.FileResult.FuncIDVec) > fi.FuncID {
			firstFunc = fileStruct.FileResult.FuncIDVec[fi.FuncID]
		}

		if firstFunc == nil {
			return
		}
		firstColonFunc = firstFunc.FindFirstColonFunc()
	}

	if firstColonFunc == nil {
		return
	}
	if !firstColonFunc.IsColon {
		return
	}

	if firstColonFunc.RelateVar == nil {
		if !a.isFirstTerm() {
			return
		}

		// 第一轮，获取
		// 获取第一轮的结构，因为此时RelateVar的结构还没有构造，只能查找第一轮的数据
		strName := a.getAsignSelfReferName()
		if strName == "" {
			return
		}

		str = prefixStr + strName
		return
	}

	str = prefixStr + firstColonFunc.RelateVar.StrName
	return
}

// 查找引用的table名是否有定义过
func (a *Analysis) findTableDefine(node *ast.TableAccessExp) {
	// 1) 第一轮跳过
	if a.isFirstTerm() {
		return
	}

	fileResult := a.curResult
	fi := a.curFunc
	scope := a.curScope

	// 3) 判断key是否为普通的
	strKey := common.GetExpName(node.KeyExp)
	// 如果不是简单字符，退出
	if !common.JudgeSimpleStr(strKey) {
		return
	}
	keyLoc := common.GetTableKeyLoc(node)

	// 4) 第二轮或第三轮判断table取值是否有定义
	strTable := common.GetExpName(node.PrefixExp)
	// self进行替换
	if strTable == "!self" {
		strTable = a.ChangeSelfToReferVar(strTable, "!")
	}
	strProPre := common.GConfig.GetStrProtocol(strTable)

	// 5) _G.aa这样的用例，判断全局变量aa是否有定义
	if strTable == "!_G" {
		// 直接查找全局变量strKey，判断是否有定义
		a.checkfindGVar(strKey, keyLoc, "", node.KeyExp)
		return
	}

	// 第五轮，不进行展开分析, 只分析到_G.a ,获取a的变量
	if a.isFiveTerm() {
		return
	}

	// 6) c2s.One() 协议前缀的调用
	if strProPre != "" && !strings.Contains(strTable, ".") {
		// 直接查找协议strKey，判断是否有定义
		a.checkfindGVar(strKey, keyLoc, strProPre, node.KeyExp)
		return
	}

	// 7.1) 在第四轮中查找三层次的调用关系
	// 例如 local a = import("one.lua") ，调用 a.bb.cc 其中bb为定义的table， cc为table中的字符串key值
	// 或是 _G.bb.cc 其中bb为定义的全局table， cc为table中的字符串key值
	if a.isFourTerm() && strings.Contains(strTable, ".") {
		a.findThreeLevelCall(node.PrefixExp, node.KeyExp)
		return
	}

	// 8) 判断引用import里面的成员是有定义
	// local a = import("one.lua")
	// b = a.dd, 判断a.dd是否有定义（需要跟进one.lua里面，判断dd是否为全局变量）
	// 判断是否为两层的调用
	strName := common.GetSimpleValue(strTable)
	if strName == "" {
		return
	}

	// 9.1) 查找局部变量的引用
	loc := common.GetTablePrefixLoc(node)
	var referInfo *common.ReferInfo
	if locVar, ok := scope.FindLocVar(strName, loc); ok {
		referInfo = locVar.ReferInfo
		// 第四轮查找引用的时候，判断是否为返回table的字符串key
		if a.isFourTerm() && referInfo == nil {
			a.ReferenceResult.MatchVarInfo(a, strName, fileResult.Name, locVar, fi, strTable, node.KeyExp, false)
			return
		}
	} else {
		// 2.2) 局部变量没有找到，查找全局变量
		if ok, oneVar := fileResult.FindGlobalVarInfo(strName, false, ""); ok {
			referInfo = oneVar.ReferInfo
		}

		if a.isFourTerm() && referInfo == nil {
			// 查找全局中是否有该变量
			firstFile := a.getFirstFileResult(fileResult.Name)
			if firstFile != nil {
				// find self globalInfo
				if ok, oneVar := firstFile.FindGlobalVarInfo(strName, false, ""); ok {
					// 判断是否是自己所要的引用关系
					a.ReferenceResult.MatchVarInfo(a, strName, firstFile.Name, oneVar, fi, strName, node.KeyExp, false)
					return
				}
			}

			// 所有工程的文件
			if a.ReferenceResult.FindProjectGlobal(a, strName, "", fi, strName, node.KeyExp) {
				return
			}
			return
		}
	}

	// 没有找到引用信息，退出，目前只处理import引用的方式
	if referInfo == nil {
		return
	}

	// 引用的lua文件
	referValidStr := referInfo.ReferValidStr
	// 排查过滤掉的文件变量
	if common.GConfig.IsIgnoreFileDefineVar(referValidStr, strKey) {
		return
	}

	referFrameSubType := a.Projects.GetReferFrameType(referInfo)
	if referFrameSubType == common.RtypeRequire && a.isFourTerm() {
		// 第四轮，查找应用
		checkTerm := a.getChangeCheckTerm()
		// 非顶层的function 查找第一次判断的
		referFile := a.GetReferFileResult(referInfo, checkTerm)
		if referFile == nil {
			// 没有加载引用的文件，可能是so文件
			return
		}

		locVar := referFile.MainFunc.GetOneReturnVar()
		if locVar == nil {
			return
		}

		subGlobal := common.GetVarSubGlobalVar(locVar, strKey)
		if subGlobal != nil {
			// 所有工程的文件
			a.ReferenceResult.MatchVarInfo(a, strKey, referFile.Name, locVar, fi, "", node.KeyExp, true)
		}
	}

	if referFrameSubType != common.RtypeImport {
		return
	}

	// 定位到引用的AnalysisFileResult
	checkTerm := a.getChangeCheckTerm()
	// 非顶层的function 查找第一次判断的
	referFile := a.GetReferFileResult(referInfo, checkTerm)
	if referFile == nil {
		// 没有加载引用的文件，可能是so文件
		return
	}

	if ok, oneVar := referFile.FindGlobalVarInfo(strKey, false, ""); ok {
		if a.isFourTerm() {
			// 所有工程的文件
			a.ReferenceResult.MatchVarInfo(a, strKey, referFile.Name, oneVar, fi, "", node.KeyExp, false)
		}
		return
	}

	// 如果查找import的对象的全局函数，也是在屏蔽对象里面，返回，不进行告警
	if _, ok := common.GConfig.LuaInMap[strKey]; ok {
		return
	}

	cycleFlag := false
	if fi.FuncLv == 0 && a.isSecondTerm() {
		referFile = a.GetReferFileResult(referInfo, results.CheckTermFirst)
		if referFile != nil {
			if ok, _ := referFile.FindGlobalVarInfo(strKey, false, ""); ok {
				cycleFlag = true
			}
		}
	}

	// 判断被引用的文件，是否忽略特定的变量为定义
	if common.GConfig.IsIgnoreErrorFile(referFile.Name, common.CheckErrorNoDefine) {
		return
	}

	// not find
	if cycleFlag {
		errStr := fmt.Sprintf("crcular reference or load order error, mode has no var: %s.%s", strName, strKey)
		fileResult.InsertError(common.CheckErrorImportVar, errStr, node.Loc)
	} else {
		errStr := fmt.Sprintf("mode has no var: %s.%s", strName, strKey)
		fileResult.InsertError(common.CheckErrorImportVar, errStr, node.Loc)
	}
}

// GetImportRefer 根据传人的exp，判断是否为导入的函数调用
func (a *Analysis) GetImportRefer(node *ast.FuncCallExp) *common.ReferInfo {
	return a.GetImportReferByCallExp(node)
}

// 获取引入其他库和文件是否错误，需要判断参数是否正常
// strFlag表示是否正确或失败，ok表示成功，其他表示失败
// strFirst表示获取的第一个参数名，字符串
func (a *Analysis) getImportReferError(funcName string, funcExp *ast.FuncCallExp) (strFlag string, strFirst string) {
	fileResult := a.curResult
	if len(funcExp.Args) != 1 {
		strErr := fmt.Sprintf("%s other file need only one param, now param num(%d)", funcName, len(funcExp.Args))
		strFlag = "param num error"
		//loc :=
		fileResult.InsertError(common.CheckErrorCallParam, strErr, funcExp.Loc)
		return
	}

	firstExp := funcExp.Args[0]
	switch expValue := firstExp.(type) {
	case *ast.StringExp:
		strFlag = "ok"
		strFirst = expValue.Str
		return
	default:
		strFlag = "func param error"
		return
	}
}

// GetImportReferByCallExp 根据传人的exp，判断是否为导入的函数调用
func (a *Analysis) GetImportReferByCallExp(funcExp *ast.FuncCallExp) *common.ReferInfo {
	if funcExp.NameExp != nil {
		return nil
	}

	parensExp, ok := funcExp.PrefixExp.(*ast.ParensExp)
	if ok {
		funcExp, ok = parensExp.Exp.(*ast.FuncCallExp)

		if !ok {
			return nil
		}
	}

	fileResult := a.curResult
	callExp, ok := funcExp.PrefixExp.(*ast.NameExp)
	if !ok {
		return nil
	}

	if !common.GConfig.ReferOtherFileMap[callExp.Name] {
		return nil
	}

	// 为主动引入了其他的模块和lua文件
	strFlag, strFirst := a.getImportReferError(callExp.Name, funcExp)
	if strFlag != "ok" {
		return nil
	}

	oneRefer := common.CreateOneReferInfo(callExp.Name, strFirst, funcExp.Loc)
	if oneRefer == nil {
		return nil
	}

	// 先查找该引用是否有效
	fileResult.CheckReferFile(oneRefer, a.Projects.GetAllFilesMap(), a.Projects.GetFileIndexInfo())

	// 全局的引用文件里面增加一个引用对象
	// if fileResult.ReferVec == nil {
	// 	fileResult.ReferVec = []*common.ReferInfo{}
	// }
	fileResult.ReferVec = append(fileResult.ReferVec, oneRefer)
	// 只有在第二阶段，才进行跟引用进入分析
	if !a.isSecondTerm() {
		return oneRefer
	}

	a.deepHanleReferFile(oneRefer)

	// 导入引入的符号表到当前lua文件中判断是否为require("one")
	// 第二轮中，加载require导入的_G符号
	a.InsertRequireInfoGlobalVars(oneRefer, results.CheckTermSecond)
	return oneRefer
}

// 对当前的TableAccessExp 进行分析，若之前有变量的定义，则不插入，否则插入到globalNoDefineMaps中
func (a *Analysis) analysisNoDefineName(node *ast.NameExp) {
	name, loc := node.Name, common.GetExpLoc(node)
	if name == " " {
		return
	}

	if !a.isNeedAnalysisNameExp(name, loc) {
		return
	}

	if _, ok := a.curResult.NodefineMaps[node.Name]; !ok {
		newVar := common.CreateVarInfo(a.curResult.Name, common.LuaTypeAll, nil, node.Loc, 1)
		a.insertGlobalNoDefineMap(node.Name, newVar)
	}
}

func (a *Analysis) analysisNoDefineStr(node *ast.StringExp) {
	name, loc := node.Str, common.GetExpLoc(node)
	if name == " " {
		return
	}

	if !a.isNeedAnalysisNameExp(name, loc) {
		return
	}

	if _, ok := a.curResult.NodefineMaps[name]; !ok {
		newVar := common.CreateVarInfo(a.curResult.Name, common.LuaTypeAll, nil, node.Loc, 1)
		a.insertGlobalNoDefineMap(name, newVar)
	}
}

// 判断是否需要分析当前的nameExp
func (a *Analysis) isNeedAnalysisNameExp(strName string, loc lexer.Location) bool {
	fileResult := a.curResult
	scope := a.curScope
	// 首先查找当前局部变量下有没有当前表名
	if _, ok := scope.FindLocVar(strName, loc); ok {
		return false
	}

	// 再次查找当前文件下全局变量有没有当前表名
	globalMaps := fileResult.GlobalMaps
	if _, ok := globalMaps[strName]; ok {
		return false
	}

	// 判断当前文件是否是第一次分析table
	// 比如a.b.c.d 会先从 a.b.c.d 再从 a.b.c 最后 a.b 进入此函数
	nodefineMaps := fileResult.NodefineMaps
	if _, ok := nodefineMaps[strName]; ok {
		return false
	}

	// 判断是否为协议变量，如果是则返回
	if common.GConfig.IsStrProtocol(strName) {
		return false
	}

	return true
}

// 将varInfo 插入到globalNoDefineMaps
func (a *Analysis) insertGlobalNoDefineMap(strName string, varInfo *common.VarInfo) {
	globalNoDefineMaps := a.curResult.NodefineMaps
	if _, ok := globalNoDefineMaps[strName]; !ok {
		globalNoDefineMaps[strName] = varInfo
	}
}
