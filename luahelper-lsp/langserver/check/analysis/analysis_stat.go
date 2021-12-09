package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"strings"
)

func (a *Analysis) cgStat(node ast.Stat) {
	switch stat := node.(type) {
	case *ast.FuncCallStat:
		a.cgFuncCallStat(stat)
	case *ast.DoStat:
		a.cgDoStat(stat)
	case *ast.WhileStat:
		a.cgWhileStat(stat)
	case *ast.RepeatStat:
		a.cgRepeatStat(stat)
	case *ast.IfStat:
		a.cgIfStat(stat)
	case *ast.ForNumStat:
		a.cgForNumStat(stat)
	case *ast.ForInStat:
		a.cgForInStat(stat)
	case *ast.AssignStat:
		a.cgAssignStat(stat)
	case *ast.LocalVarDeclStat:
		a.cgLocalVarDeclStat(stat)
	case *ast.LocalFuncDefStat:
		a.cgLocalFuncDefStat(stat)
	case *ast.LabelStat:
		a.cgLabelStat(stat)
	case *ast.GotoStat:
		a.cgGotoStat(stat)
	case *ast.BreakStat:
		a.cgBreakStat(stat)
	}
}

// 当调用这样的函数时 aaa:bb("1", "2")
// 其中aaa 为 PrefixExp， bb 为NameExp，括号内的为参数
func (a *Analysis) cgFuncCallStat(node *ast.FuncCallStat) {
	// 在第二阶段，里面可能深度加载依赖的文件，直接函数调用例如import("one.two")
	a.GetImportReferByCallExp(node)
	a.cgExp(node.PrefixExp, nil, nil)

	if node.NameExp != nil {
		//analysis.cgExp(node.NameExp)

		// todo，例如有这样的定义
		// a = {}
		// function a:test() end
		// a:test() -- 冒号调用时候，判断是否要查找
		a.findFuncColon(node.PrefixExp, node.NameExp, node.Loc)
	}

	for _, argExp := range node.Args {
		a.cgExp(argExp, nil, nil)
	}

	// 第二轮或第三轮函数参数check
	a.cgFuncCallParamCheck(node)
}

// 检查调用函数匹配的参数
// 第二轮或第三轮检查参数个数是否匹配
func (a *Analysis) cgFuncCallParamCheck(node *ast.FuncCallStat) {
	// 第二轮或第三轮函数参数check
	if !a.isNeedCheck() {
		return
	}

	// 判断是否开启了函数调用参数个数不匹配的校验
	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorCallParam) {
		return
	}

	fileResult := a.curResult
	nArgs := len(node.Args)

	var referFunc *common.FuncInfo
	var referStr string
	referFunc, referStr = a.getFuncCallReferFunc(node)
	if referFunc == nil {
		return
	}

	if referFunc.IsVararg {
		return
	}

	paramLen := len(referFunc.ParamList)

	if nArgs > paramLen {
		//调用处参数个数大于定义参数个数的，直接告警
		errorStr := fmt.Sprintf("%s call func param num(%d) > func define param num(%d)", referStr, nArgs, paramLen)
		fileResult.InsertError(common.CheckErrorCallParam, errorStr, node.Loc)
		return
	} else if nArgs < paramLen {
		// 函数调用处参数个数小于定义参数个数的，支持注解辅助检查

		// 如果未获取过
		if !referFunc.ParamDefaultInit {
			referFunc.ParamDefaultInit = true
			referFunc.ParamDefaultNum = a.Projects.GetFuncDefaultParamInfo(referFunc.FileName, referFunc.Loc.StartLine-1,
				referFunc.ParamList)
		}

		// 如果没有注解 不告警
		if referFunc.ParamDefaultNum == -1 {
			return
		}

		// 默认参数个数+调用填参个数 < 定义参数个数
		if nArgs+referFunc.ParamDefaultNum < paramLen {
			errorStr := fmt.Sprintf("%s call func param num(%d) < func define param num(%d)", referStr, nArgs, paramLen)
			fileResult.InsertError(common.CheckErrorCallParam, errorStr, node.Loc)
			return
		}
	}
}

func (a *Analysis) cgBreakStat(node *ast.BreakStat) {
}

func (a *Analysis) cgDoStat(node *ast.DoStat) {
	a.enterScope()

	backupScope := a.curScope

	subScope := common.CreateScopeInfo(backupScope, nil, node.Loc)
	backupScope.AppendSubScope(subScope)
	a.curScope = subScope

	a.cgBlock(node.Block)
	a.exitScope()

	a.curScope = backupScope
}

/*
           ______________
          /  false? jmp  |
         /               |
while exp do block end <-'
      ^           \
      |___________/
           jmp
*/
func (a *Analysis) cgWhileStat(node *ast.WhileStat) {
	a.cgExp(node.Exp, nil, nil)
	a.enterScope()

	backupScope := a.curScope

	subScope := common.CreateScopeInfo(backupScope, nil, node.Loc)
	backupScope.AppendSubScope(subScope)

	a.curScope = subScope
	a.cgBlock(node.Block)
	a.exitScope()
	a.curScope = backupScope
}

/*
        ______________
       |  false? jmp  |
       V              /
repeat block until exp
*/
func (a *Analysis) cgRepeatStat(node *ast.RepeatStat) {
	a.enterScope()

	backupScope := a.curScope

	subScope := common.CreateScopeInfo(backupScope, nil, node.Loc)
	backupScope.AppendSubScope(subScope)

	a.curScope = subScope
	a.cgBlock(node.Block)
	a.cgExp(node.Exp, nil, nil)

	a.exitScope()
	a.curScope = backupScope
}

/*
         _________________       _________________       _____________
        / false? jmp      |     / false? jmp      |     / false? jmp  |
       /                  V    /                  V    /              V
if exp1 then block1 elseif exp2 then block2 elseif true then block3 end <-.
                   \                       \                       \      |
                    \_______________________\_______________________\_____|
                    jmp                     jmp                     jmp
*/
func (a *Analysis) cgIfStat(node *ast.IfStat) {
	// 在第一轮工程的分析, 新增下面的检测规则, 前面认为ss可能为空，为空的话对他调用函数有bug
	// if not ss then
	//     ss.test()
	// end

	// 检查重复条件
	if a.isFirstTerm() && !common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorDuplicateIf) {
		for i, _ := range node.Exps {
			for j := i + 1; j < len(node.Exps); j++ {
				if common.CompExp(node.Exps[i], node.Exps[j]) {
					errStr := fmt.Sprintf("same if condition one")
					a.curResult.InsertError(common.CheckErrorDuplicateIf, errStr, common.GetExpLoc(node.Exps[i]))

					errStr = fmt.Sprintf("same if condition another")
					a.curResult.InsertError(common.CheckErrorDuplicateIf, errStr, common.GetExpLoc(node.Exps[j]))
				}
			}
		}
	}

	scope := a.curScope
	// 临时的这样的结果
	var midNotVarMap map[string]common.NotValStruct
	for i, exp := range node.Exps {
		a.ignoreInfo.inif = true

		if a.isFirstTerm() && !a.realTimeFlag && midNotVarMap != nil {
			for strKey, oneNotValStruct := range midNotVarMap {
				scope.SetMidNotValStruct(strKey, oneNotValStruct)
			}
		}

		a.cgExp(exp, nil, nil)
		a.ignoreInfo.inif = false

		backupScope := a.curScope
		subScope := common.CreateScopeInfo(scope, nil, node.Blocks[i].Loc)
		scope.AppendSubScope(subScope)

		a.curScope = subScope

		//  第一轮中的处理
		if a.isFirstTerm() && !a.realTimeFlag {
			simpleNotValueArr := make([]string, 0)
			common.GetAllExpSimpleNotValue(exp, 0, &simpleNotValueArr)

			// if not a then 需要判断a是否为局部变量
			for _, strName := range simpleNotValueArr {
				flag, locVarInfo := scope.FindLocVar(strName, node.Loc)
				if !flag {
					continue
				}

				// 这样的case需要特殊处理下
				//if not a or not b then
				//	if not a then
				//	 else
				//		 print(a.ss), 这一行应该不报错
				//	 end
				//end
				// 需要在前面找下，判断是否有也有if not 这样的
				findNotVal, _ := scope.FindNotVarInfo(strName)
				if findNotVal != nil {
					if a.isFirstTerm() && midNotVarMap == nil {
						midNotVarMap = map[string]common.NotValStruct{}
					}

					// 之前也存在，现在的存放中在中间里面
					midNotVarMap[strName] = common.NotValStruct{
						Var:     findNotVal,
						SetFlag: true,
					}
				}

				notVal := common.NotValStruct{
					Var:     locVarInfo,
					SetFlag: false,
				}

				subScope.SetNotVarMapStruct(strName, notVal)
			}

			// 前面的逻辑为 if not a then
			// 现在又新加 if a then，则这个if里面a是有定义的
			simpleValueArr := make([]string, 0)
			common.GetAllExpSimpleValue(exp, &simpleValueArr)
			for _, strName := range simpleValueArr {
				findNotVal, _ := scope.FindNotVarInfo(strName)
				if findNotVal == nil {
					continue
				}

				notVal := common.NotValStruct{
					Var:     findNotVal,
					SetFlag: true,
				}

				subScope.SetMidNotValStruct(strName, notVal)
			}
		}

		a.enterScope()
		a.cgBlock(node.Blocks[i])
		a.exitScope()

		a.curScope = backupScope
	}

}

// for var=exp1,exp2,exp3 do
// var 从 exp1 变化到 exp2，每次变化以 exp3 为步长递增 var，并执行一次 "执行体"
// exp3 是可选的，如果不指定，默认为1。
func (a *Analysis) cgForNumStat(node *ast.ForNumStat) {
	a.enterScope()

	backupScope := a.curScope

	subScope := common.CreateScopeInfo(backupScope, nil, node.Loc)
	backupScope.AppendSubScope(subScope)

	a.curScope = subScope

	a.cgExp(node.InitExp, nil, nil)
	a.cgExp(node.StepExp, nil, nil)
	a.cgExp(node.LimitExp, nil, nil)

	locVar := subScope.AddLocVar(node.VarName, common.LuaTypeInter, nil, node.VarLoc, 1)
	locVar.IsUse = true

	a.cgBlock(node.Block)
	a.exitScope()

	a.curScope = backupScope
}

// 判断是否为特殊的pairs或ipairs迭代函数，如果是获取pairs或ipairs里面的exp
func getForCycleData(expList []ast.Exp) (referExp ast.Exp, ipairsFlag bool) {
	if len(expList) != 1 {
		return
	}

	oneExp := expList[0]
	funcExp, flag := oneExp.(*ast.FuncCallExp)
	if !flag {
		return
	}

	callExp, ok := funcExp.PrefixExp.(*ast.NameExp)
	if !ok {
		return
	}

	funcName := callExp.Name
	if funcName != "pairs" && funcName != "ipairs" {
		return
	}

	if len(funcExp.Args) != 1 {
		return
	}

	referExp = funcExp.Args[0]
	if funcName == "ipairs" {
		ipairsFlag = true
	}

	return
}

//泛型for
//a = {"one", "two", "three"}
//for i, v in ipairs(a) do
func (a *Analysis) cgForInStat(node *ast.ForInStat) {
	a.enterScope()
	backupScope := a.curScope

	subScope := common.CreateScopeInfo(backupScope, nil, node.Loc)
	backupScope.AppendSubScope(subScope)

	a.curScope = subScope
	for _, oneExp := range node.ExpList {
		a.cgExp(oneExp, nil, nil)
	}

	referExp, ipairsFlag := getForCycleData(node.ExpList)
	for index, name := range node.NameList {
		varIndex := uint8(index + 1)
		locVar := subScope.AddLocVar(name, common.LuaTypeRefer, nil, node.NameLocList[index], varIndex)
		locVar.IsForParam = true
		locVar.IsUse = true
		if referExp != nil {
			locVar.ForCycle = &common.ForCycleInfo{
				Exp:        referExp,
				IpairsFlag: ipairsFlag,
			}
		}
	}

	a.cgBlock(node.Block)
	a.exitScope()

	a.curScope = backupScope
}

//::continue::
func (a *Analysis) cgLabelStat(node *ast.LabelStat) {
	a.curFunc.InsertLabel(node.Name, node.Loc)
}

//goto label
func (a *Analysis) cgGotoStat(node *ast.GotoStat) {
	// 只在第二轮或第三轮校验
	if !a.isNeedCheck() {
		return
	}

	fileResult := a.curResult
	fi := a.curFunc
	curScopeLv := fi.ScopeLv
	// 在第二轮进行检查
	funcID := fi.FuncID

	firstFile := a.getFirstFileResult(fileResult.Name)
	if firstFile == nil {
		return
	}

	if funcID >= len(firstFile.FuncIDVec) {
		return
	}

	firstFunc := firstFile.FuncIDVec[funcID]
	if firstFunc == nil {
		return
	}

	if firstFunc.FindLabel(node.Name, curScopeLv) {
		return
	}

	errStr := fmt.Sprintf("not find valid:'%s' for <goto> ", node.Name)
	fileResult.InsertError(common.CheckErrorGotoLabel, errStr, node.Loc)
}

func (a *Analysis) cgLocalFuncDefStat(node *ast.LocalFuncDefStat) {
	scope := a.curScope
	locVar := scope.AddLocVar(node.Name, common.LuaTypeFunc, node.Exp, node.NameLoc, 1)

	subFi := a.cgFuncDefExp(node.Exp)
	locVar.ReferFunc = subFi
}

func (a *Analysis) cgLocalVarDeclStat(node *ast.LocalVarDeclStat) {
	nNames := len(node.NameList)
	nExps := len(node.ExpList)
	fileResult := a.curResult
	if a.isFirstTerm() {
		if nNames < nExps {
			errStr := fmt.Sprintf("local define param(%d) < param num(%d) error", nExps, nNames)
			fileResult.InsertError(common.CheckErrorLocalParamNum, errStr, node.Loc)
		} else if nNames > nExps && nExps > 0 {
			canWarning := true
			for i := 0; i < nExps; i++ {
				expNode := node.ExpList[i]

				if common.IsOneValueType(expNode) {
					//当右边表达式类型全部是单值时 才能报警
				} else {
					canWarning = false
					break
				}
			}

			if canWarning {
				errStr := fmt.Sprintf("local define param num(%d) > assign param num(%d) error", nNames, nExps)
				fileResult.InsertError(common.CheckErrorLocalParamNum, errStr, node.Loc)
			}
		}
	}

	// 最后一个表达式，是否为函数调用
	lastExpFuncFlag := false
	scope := a.curScope

	// 1) 首先给匹配的变量赋明确的值
	for i, exp := range node.ExpList {
		varIndex := uint8(i + 1)
		tempVar := common.CreateVarInfo(common.LuaTypeAll, nil, lexer.Location{}, varIndex)
		oneFunc, oneRefer := a.cgExp(exp, tempVar, nil)
		if i >= nNames {
			break
		}

		strName := node.NameList[i]
		if oneRefer != nil {
			oneRefer.ReferVarLocal = true
		}

		nowLoc := node.VarLocList[i]
		varInfo := scope.AddLocVar(strName, common.GetExpType(exp), exp, nowLoc, varIndex)
		oneAttr := node.AttrList[i]
		if oneAttr == ast.RDKTOCLOSE {
			varInfo.IsClose = true
		}

		switch exp.(type) {
		case *ast.FuncDefExp:
			// 定义为local abcd1 = function ()
			varInfo.ReferFunc = oneFunc
		case *ast.FuncCallExp:
			varInfo.ReferInfo = oneRefer
			// 如果局部变量是获取的函数返回值，且引用了其他的文件
			if oneRefer != nil {
				varInfo.IsUse = true
			}
			// 最后一个表达式是函数调用
			if i == nExps-1 {
				lastExpFuncFlag = true
			}
		}

		// 构造这个变量的table 构造的成员
		varInfo.SubMaps = tempVar.SubMaps

		// 关联这个变量，引用其他的变量
		// 当前是局部变量赋值的时候，关联这个变量指向其他的变量
		varInfo.ReferExp = exp

		// 判断指向的ReferExp是否为有效的, 如果为empty，设置对应的标记
		if common.IsLocalReferExpEmpty(strName, exp) {
			varInfo.IsExpEmpty = true
		}
	}

	// 2) 给不匹配的变量，赋值nil
	for i := nExps; i < nNames; i++ {
		varIndex := uint8(i + 1)
		nowLoc := node.VarLocList[i]
		oneAttr := node.AttrList[i]
		if lastExpFuncFlag {
			locVar := scope.AddLocVar(node.NameList[i], common.LuaTypeRefer, nil, nowLoc, varIndex)
			if oneAttr == ast.RDKTOCLOSE {
				locVar.IsClose = true
			}
			// 关联到函数的表达式
			locVar.ReferExp = node.ExpList[nExps-1]
		} else {
			locVar := scope.AddLocVar(node.NameList[i], common.LuaTypeNil, nil, nowLoc, varIndex)
			if oneAttr == ast.RDKTOCLOSE {
				locVar.IsClose = true
			}
			locVar.IsExpEmpty = true
		}
	}
}

// 解析赋值表达式，左侧的变量，判断是否要定义变量
// 返回值 needDefine 表示是否要定义变量
// 返回值 flagG 表示是否为_G的变量
// 返回值 strName 表示赋值的前半部分名字, 去掉了前缀_G.
// 返回值 strProPre 表示是否为协议的前缀 为c2s 或是s2s
// 返回值 lexer.LocInfo 表示全局变量的位置信息
// strVec 表示切割处理的字符串列表
// localVarInfo 表明是对local变量进行赋值
func (a *Analysis) checkLeftAssign(valExp ast.Exp) (needDefine bool, flagG bool, strName, strProPre string,
	loc lexer.Location, strVec []string, locList []lexer.Location, varInfo *common.VarInfo) {
	needDefine = true // 是否要定义变量
	flagG = false     // 否为_G的变量
	strName = ""      // 赋值变量，前半部分名字
	strProPre = ""    // 协议的前缀 为c2s 或是s2s
	varInfo = nil

	fileResult := a.curResult
	scope := a.curScope
	fi := a.curFunc

	// 1）判断是否为直接名称的赋值
	if nameExp, ok := valExp.(*ast.NameExp); ok {
		strName = nameExp.Name
		loc = nameExp.Loc
		// 先在局部变量中查找
		if flag, varTemp := scope.FindLocVar(strName, loc); flag == ok {
			needDefine = false
			varInfo = varTemp
			return
		}

		// 全局变量中查找
		findFlag, findVar := fileResult.FindGlobalVar(strName, fi.FuncLv, fi.ScopeLv, loc, "", false)
		if findFlag {
			needDefine = false
			varInfo = findVar
		}

		// 如果都没有找到，表示要定义变量
		return
	}

	// 2) 判断是否为_G table变量的赋值， 例如_G.a = 1
	taExp, ok := valExp.(*ast.TableAccessExp)
	if !ok {
		needDefine = false
		return
	}

	// table的名称
	a.cgExp(taExp.PrefixExp, nil, nil)

	// table的key
	a.cgExp(taExp.KeyExp, nil, nil)

	// key值不为简单字符串，直接退出
	strKeyName := common.GetExpName(taExp.KeyExp)
	if !common.JudgeSimpleStr(strKeyName) {
		needDefine = false
		return
	}

	tabName := common.GetExpName(taExp.PrefixExp)
	strProPre = common.GConfig.GetStrProtocol(tabName)

	loc = common.GetExpLoc(taExp.KeyExp)
	if loc.IsInitialLoc() {
		loc = taExp.Loc
	}

	// 这里为_G.b = d ，这样的赋值表达式
	if tabName == "!_G" {
		strName = strKeyName
		flagG = true
		findFlag, findVar := fileResult.FindGlobalVar(strKeyName, fi.FuncLv, fi.ScopeLv, loc, "", false)
		// 找到了，b的定义
		if findFlag {
			needDefine = false
			varInfo = findVar
		}
		return
	}

	// 如果为协议的前缀
	if strProPre != "" {
		strName = strKeyName
		flagG = false
		return
	}

	needDefine = false

	splitArray := strings.Split(tabName, ".")
	splitArray[0] = strings.TrimPrefix(splitArray[0], "!")
	for i := 1; i < len(splitArray); i++ {
		if !common.JudgeSimpleStr(splitArray[i]) {
			// 不是简单的字符串，直接返回
			return
		}
	}

	// 获取table 每一个成员和子key的loc list
	locList = common.GetTableLocList(valExp)
	if splitArray[0] == "_G" {
		strName = splitArray[1]
		flagG = true
		findFlag, findVar := fileResult.FindGlobalVar(strName, fi.FuncLv, fi.ScopeLv, loc, "", false)
		// 找到了，b的定义
		if findFlag {
			varInfo = findVar
		}

		strVec = append(strVec, splitArray[1:]...)
		strVec = append(strVec, strKeyName)

		if len(locList) > 1 {
			locList = locList[1:]
		}
	} else {
		if splitArray[0] == "self" {
			splitArray[0] = a.ChangeSelfToReferVar(splitArray[0], "")
		}

		strVec = append(strVec, splitArray...)
		strVec = append(strVec, strKeyName)

		strName = splitArray[0]
		if flag, varTemp := scope.FindLocVar(strName, loc); flag == ok {
			varInfo = varTemp
			return
		}

		if flag, findVar := fileResult.FindGlobalVar(strName, fi.FuncLv, fi.ScopeLv, loc, "", false); flag {
			varInfo = findVar
		}
	}
	return
}

// 判断是否对if not a then 对a变量的赋值
func (a *Analysis) handleIfNotValAssign(valExp ast.Exp) {
	strName := ""
	if nameExp, ok := valExp.(*ast.NameExp); ok {
		strName = nameExp.Name
	}

	if strName == "" {
		return
	}

	scope := a.curScope

	findNotVal, findNotScope := scope.FindNotVarInfo(strName)
	if findNotVal == nil {
		return
	}

	if notValStruct, ok := findNotScope.NotVarMap[strName]; ok {
		notValStruct.SetFlag = true
		findNotScope.SetNotVarMapStruct(strName, notValStruct)
	}

	return
}

//  处理table成员变量的构造
func (a *Analysis) handleTableMemConstruct(strMemName string, strMemValue string, oneFunc *common.FuncInfo,
	newRefer *common.ReferInfo, loc lexer.Location, strProPre string, nExps int, i int, node *ast.AssignStat) {

	scope := a.curScope
	fi := a.curFunc

	fileResult := a.curResult
	// 1) 先在局部变量中查找
	findOk, locVarInfo := scope.FindLocVar(strMemName, loc)
	if findOk && !locVarInfo.IsExistMember(strMemValue) {
		newVar := common.CreateOneVarInfo(loc, newRefer, oneFunc, 1)
		if oneFunc != nil && oneFunc.IsColon {
			//目前只有冒号函数才进行反向关联函数的变量
			oneFunc.RelateVar = common.CreateFuncRelateVar(strMemName, locVarInfo)
		}

		if nExps >= (i + 1) {
			expNode := node.ExpList[i]
			newVar.VarType = common.GetExpType(expNode)
			newVar.ReferExp = expNode
		}

		locVarInfo.InsertSubMember(strMemValue, newVar)
		return
	}

	// 2)判断是否为当前文件的全局变量赋值, 例如a.b = 3, 其中strMemName赋值为a, strMemValue赋值为b
	// 一定为全局变量，在全局变量中查找
	findMemFlag, memVar := fileResult.FindGlobalVar(strMemName, fi.FuncLv, fi.ScopeLv, loc,
		strProPre, false)

	// 定义在本文件中的global 变量，尝试构建所有的成员信息
	if findMemFlag && !memVar.IsExistMember(strMemValue) {
		newVar := common.CreateOneVarInfo(loc, newRefer, oneFunc, 1)
		if oneFunc != nil && oneFunc.IsColon {
			//目前只有冒号函数才进行反向关联函数的变量
			oneFunc.RelateVar = common.CreateFuncRelateVar(strMemName, memVar)
		}

		if nExps >= (i + 1) {
			expNode := node.ExpList[i]
			newVar.VarType = common.GetExpType(expNode)
			newVar.ReferExp = expNode
		}

		memVar.InsertSubMember(strMemValue, newVar)
	}
}

// 处理assign语句没有直接定义变量的时候
func (a *Analysis) handleNotNeedDefine(node *ast.AssignStat, findVar *common.VarInfo, strVec []string,
	locList []lexer.Location, tmpVar *common.VarInfo, loc lexer.Location, newRefer *common.ReferInfo,
	newFunc *common.FuncInfo, i int) {
	nExps := len(node.ExpList)

	parentVar := findVar
	strVecLen := len(strVec)
	for j := 1; j < strVecLen; j++ {
		strKeyValue := strVec[j]
		if subVar, ok := parentVar.SubMaps[strKeyValue]; ok {
			// 如果已经存在这个key
			parentVar = subVar

			// 判断这个subkey，之前是否关联的ReferExp为nil
			// 判断指向的ReferExp是否为有效的, 如果为empty，设置对应的标记
			if !(j == strVecLen-1 && nExps >= (i+1)) {
				continue
			}
			expNode := node.ExpList[i]
			if !common.IsReferExpEmpty(expNode, subVar.ReferExp, true) {
				continue
			}
			subVar.IsExpEmpty = false

			if newFunc != nil && newFunc.IsColon {
				//目前只有冒号函数才进行反向关联函数的变量
				newFunc.RelateVar = common.CreateFuncRelateVar(strVec[j-1], subVar)
				subVar.ReferFunc = newFunc
			}

			subVar.VarType = common.GetExpType(expNode)
			subVar.ReferExp = expNode
			continue
		}

		// 下面的逻辑为parentVar.SubMaps 不包含 strKeyValue
		subLoc := loc
		if j < len(locList) {
			subLoc = locList[j]
		}

		// 不存在这个key，需要创建
		if j != strVecLen-1 {
			// 中间的key
			newSubVar := common.CreateOneVarInfo(subLoc, nil, nil, 1)
			newSubVar.VarType = common.LuaTypeTable
			parentVar.InsertSubMember(strKeyValue, newSubVar)
			parentVar = newSubVar
			continue
		}

		// 最后一次，需要绑定详细的信息
		newSubVar := common.CreateOneVarInfo(subLoc, newRefer, newFunc, 1)
		newSubVar.SubMaps = tmpVar.SubMaps
		strMemName := strVec[j-1]
		if (j - 1) != 0 {
			strMemName = strings.Join(strVec[0:j], ".")
		}

		if newFunc != nil && newFunc.IsColon {
			//目前只有冒号函数才进行反向关联函数的变量
			newFunc.RelateVar = common.CreateFuncRelateVar(strMemName, parentVar)
		}

		if nExps >= (i + 1) {
			expNode := node.ExpList[i]
			newSubVar.VarType = common.GetExpType(expNode)
			newSubVar.ReferExp = expNode
		}

		parentVar.InsertSubMember(strKeyValue, newSubVar)
	}
}

// varlist ‘=’ explist
// varlist ::= var {‘,’ var}
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
/*
1) 下面这种全局函数的定义会走进来
function abcd2()
	local ssd
end
2) 下面这种全局函数的定义也会走进来
abcd3 = function ()
	local ddddd
end
3) 这种函数定义也会走进来
function ddd.ccc(a, b)
    -- body
end
*/
func (a *Analysis) cgAssignStat(node *ast.AssignStat) {
	nVars := len(node.VarList)
	nExps := len(node.ExpList)

	fileResult := a.curResult
	fi := a.curFunc

	lastExp := node.ExpList[nExps-1]

	// 最后一个表达式是否为指向函数
	_, lastExpFuncFlag := lastExp.(*ast.FuncCallExp)

	for i, valExp := range node.VarList {
		var newVar *common.VarInfo
		var newRefer *common.ReferInfo
		var newFunc *common.FuncInfo
		varIndex := uint8(i + 1)
		tmpVar := common.CreateVarInfo(common.LuaTypeAll, nil, lexer.Location{}, varIndex)
		if nExps >= (i + 1) {
			expNode := node.ExpList[i]

			// 第二轮扫描和第三轮中为这样的赋值语句 a = a 0r 0, 忽略a没有定义
			if a.isNeedCheck() {
				ignoreStr, ignoreLine := common.JudgeIgnoreAssinVale(valExp, expNode)
				if ignoreStr != "" {
					a.ignoreInfo.strName = ignoreStr
					a.ignoreInfo.line = ignoreLine
				} else {
					assignSubKey := common.GetExpSubKey(common.GetExpName(valExp))
					a.ignoreInfo.assignName = assignSubKey
				}
			}

			// 递归调用表达式的值
			newFunc, newRefer = a.cgExp(expNode, tmpVar, nil)
			a.ignoreInfo.strName = ""
			a.ignoreInfo.assignName = ""
		}

		// 解析赋值表达式，左侧的变量，判断是否要定义变量
		// needDefineFlag 判断是否要定义全局变量
		// gGlag 是否为_G的变量
		// strProPre 表示是否为协议的前缀 为c2s 或是s2s
		// strName 赋值变量，前半部分名字, 去掉了_G.
		needDefineFlag, gGlag, strName, strProPre, loc, strVec, locList, findVar := a.checkLeftAssign(valExp)

		// 需要定义变量
		if needDefineFlag {
			newVar = common.CreateOneGlobal(fi.FuncLv, fi.ScopeLv, loc, gGlag, newRefer, newFunc, fileResult.Name)
			newVar.VarIndex = varIndex
			newVar.ExtraGlobal.StrProPre = strProPre

			if nExps >= (i + 1) {
				expNode := node.ExpList[i]
				newVar.VarType = common.GetExpType(expNode)
				newVar.ReferExp = expNode

				// 判断指向的ReferExp是否为有效的, 如果为empty，设置对应的标记
				if common.IsReferExpEmpty(valExp, expNode, false) {
					newVar.IsExpEmpty = true
				}

			} else if lastExpFuncFlag {
				newVar.VarType = common.GetExpType(lastExp)
				newVar.ReferExp = lastExp
			}

			// 拷贝过来
			newVar.SubMaps = tmpVar.SubMaps

			// 插入全局变量
			a.insertAnalysisGlobalVar(strName, newVar)
		} else {
			strVecLen := len(strVec)
			if findVar != nil && strVecLen > 0 {
				a.handleNotNeedDefine(node, findVar, strVec, locList, tmpVar, loc, newRefer, newFunc, i)
			}

			// 如果为不需要定义的变量，判断变量之前是否指向的exp 为nil
			// 例如语句块:
			// a = nil  -- 之前这里a定义的时候，指向块为exp 为nil没有含义
			// a = b    -- 现在又进行了赋值，重新指向a的exp为b，这样能进行有效的跟踪

			if findVar != nil && findVar.IsExpEmpty && strVecLen == 0 &&
				(findVar.SubMaps == nil || len(findVar.SubMaps) == 0) {
				// 标志重置
				findVar.IsExpEmpty = false

				// 之前的nil，这次指向有效的exp
				if nExps >= (i + 1) {
					expNode := node.ExpList[i]
					findVar.VarType = common.GetExpType(expNode)
					findVar.ReferExp = expNode

					// 判断指向的ReferExp是否为有效的, 如果为empty，设置对应的标记
					if common.IsReferExpEmpty(valExp, expNode, false) {
						findVar.IsExpEmpty = true
					}
				}

				// 拷贝过来
				findVar.SubMaps = tmpVar.SubMaps
			}

			// 第一轮检查中，如果前面定义了局部变量，后面未直接使用
			// local a = 1
			// a = 2			-- 这行也是未直接使用，也应该告警，这些为赋值的
			// 这些未直接使用的，记录到一个列表中，告警需要使用
			if findVar != nil && !findVar.IsGlobal() && a.isFirstTerm() && !a.realTimeFlag &&
				!findVar.IsUse {
				noUseLoc := common.GetExpLoc(valExp)
				if !noUseLoc.IsInitialLoc() {
					findVar.NoUseAssignLocs = append(findVar.NoUseAssignLocs, noUseLoc)
				}
			}
		}

		// 是否定义了变量
		defineVarFlag := needDefineFlag
		if !defineVarFlag && (a.isFourTerm() || a.isSixTerm()) {
			if nameExp, ok := valExp.(*ast.NameExp); ok {
				// 第四轮，变量赋值的引用查找
				a.findNameStr(nameExp, nil)
			}

			if taExp, ok := valExp.(*ast.TableAccessExp); ok {
				// 第四轮，table的关键key值的赋值，查找引用, 定义出需要去重
				a.findTableDefine(taExp)
			}
		}

		if !defineVarFlag && a.isFiveTerm() {
			if taExp, ok := valExp.(*ast.TableAccessExp); ok {
				// 第五轮，table的关键key值的赋值，查找table的调用
				a.findTableDefine(taExp)
			}
		}

		// 一一对应的赋值
		if nExps >= (i + 1) {
			expNode := node.ExpList[i]
			if newVar != nil && common.IsVarargOrFuncCall(expNode) {
				// 如果为函数调用
				newVar.VarType = common.LuaTypeRefer
			}
		}

		if i+1 > nExps {
			// 返回值赋值给多个参数
			if common.IsVarargOrFuncCall(lastExp) && newVar != nil {
				newVar.VarType = common.LuaTypeRefer
			}
		}

		// 第一轮中判断是否有这样的语句 if not a then
		// 对a进行赋值 a = 1
		if a.isFirstTerm() {
			if nExps >= (i + 1) {
				expNode := node.ExpList[i]
				if _, ok := expNode.(*ast.NilExp); !ok {
					a.handleIfNotValAssign(valExp)
				}
			} else {
				a.handleIfNotValAssign(valExp)
			}
		}
		// if not a
		// a.b  这样的告警查看

		// 当前的 FuncInfo 为 fi *FuncInfo，
		// 向上的scope中查找 strName, 并且返回 locInfo 和 funcInfo

		// 第一轮中检查是否有 a.b = 1 其中a在if not a里面
		if a.isFirstTerm() {
			switch exp := valExp.(type) {
			case *ast.TableAccessExp:
				a.checkIfNotTableAccess(exp, nil)
			}
		}
	}

	if a.isFirstTerm() {
		if nVars < nExps {
			errStr := fmt.Sprintf("define param num(%d) < assign param num(%d) error", nVars, nExps)
			fileResult.InsertError(common.CheckErrorAssignParamNum, errStr, node.Loc)

		} else if nVars > nExps {
			canWarning := true
			for i := 0; i < nExps; i++ {
				expNode := node.ExpList[i]

				if common.IsOneValueType(expNode) {
					//当右边表达式类型全部是单值时 才能报警
				} else {
					canWarning = false
					break
				}
			}

			if canWarning {
				errStr := fmt.Sprintf("define param num(%d) > assign param num(%d) error", nVars, nExps)
				fileResult.InsertError(common.CheckErrorAssignParamNum, errStr, node.Loc)
			}
		} else {
			if !common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorSelfAssign) {
				isSame := true
				for i := 0; i < nExps; i++ {
					if !common.CompExp(node.VarList[i], node.ExpList[i]) {
						isSame = false
					}
				}

				if isSame {
					errStr := fmt.Sprintf("self assign error")
					fileResult.InsertError(common.CheckErrorSelfAssign, errStr, node.Loc)
				}
			}
		}
	}

	if nVars < nExps {
		for i := nVars; i < nExps; i++ {
			expNode := node.ExpList[i]
			// 递归调用表达式的值
			a.cgExp(expNode, nil, nil)
		}
	}
}
