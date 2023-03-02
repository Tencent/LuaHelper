package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"strings"
)

// 返回第一个参数为为产生的funcInfo
// parentVar 为父的变量指针，如果不为nil，这个子的Exp可能也会创建子的Var，需要关联起来
// binParentExp 为二元表达式，父的BinopExp指针， 例如 a = b and c，当对c变量调用cgExp时候，binParentExp为b and c
func (a *Analysis) cgExp(node ast.Exp, parentVar *common.VarInfo,
	binParentExp *ast.BinopExp) (newFunc *common.FuncInfo, newRefer *common.ReferInfo) {
	switch exp := node.(type) {
	case *ast.ParensExp:
		a.cgExp(exp.Exp, parentVar, binParentExp)
	case *ast.VarargExp:
		a.cgVarargExp(exp)
	case *ast.NameExp:
		a.cgNameExp(exp, binParentExp)
	case *ast.FuncDefExp:
		newFunc = a.cgFuncDefExp(exp)
	case *ast.TableConstructorExp:
		a.cgTableConstructorExp(exp, parentVar)
	case *ast.UnopExp:
		a.cgUnopExp(exp)
	case *ast.BinopExp:
		a.cgBinopExp(exp, parentVar)
	case *ast.TableAccessExp:
		a.cgTableAccessExp(exp, binParentExp)
	case *ast.FuncCallExp:
		newRefer = a.cgFuncCallExp(exp)
	}

	return newFunc, newRefer
}

// ...
func (a *Analysis) cgVarargExp(node *ast.VarargExp) {
	if !a.isFirstTerm() {
		return
	}
	/*if fi.IsVararg {
		return
	}
	fileResult.InsertError(self, CHECK_VARARG_PARAM_ERROR, "cannot use '...' outside a vararg function", node.Loc)
	*/
}

// r[a] := name
// binParentExp 为二元表达式，父的BinopExp指针， 例如 a = b and c，当对c变量调用cgExp时候，binParentExp为b and c
func (a *Analysis) cgNameExp(node *ast.NameExp, binParentExp *ast.BinopExp) {
	// 第一轮
	if a.isFirstTerm() {
		// 查下局部变量是否有定义了未使用
		a.checkLocVarNotUse(node)
		// 查下当前之前有没有定义，若没有定义需要插入到相应的globalNoDefineMaps中
		a.analysisNoDefineName(node)
		return
	}

	a.findNameStr(node, binParentExp)
}

// 检查函数的是否包含同名的参数
func (a *Analysis) checkDuplicateFunParam(node *ast.FuncDefExp) {
	// 下面的判断只在第一轮，且是非实时检查时才触发
	if !a.isFirstTerm() || a.realTimeFlag {
		return
	}

	fileResult := a.curResult
	parLen := len(node.ParList)
	if parLen <= 1 {
		return
	}

	// 如果是第一轮，校验函数的参数是否有同名的
	for i := 0; i < parLen-1; i++ {
		for j := i + 1; j < parLen; j++ {
			if node.ParList[j] == "_" {
				continue
			}

			if node.ParList[j] == node.ParList[i] {
				errStr := fmt.Sprintf("func param has duplicate var:'%s'", node.ParList[j])
				fileResult.InsertError(common.CheckErrorDuplicateParam, errStr, node.ParLocList[j])
			}
		}
	}
}

// f[a] := function(args) body end
func (a *Analysis) cgFuncDefExp(node *ast.FuncDefExp) *common.FuncInfo {
	fi := a.curFunc
	scope := a.curScope
	subFi := common.CreateFuncInfo(fi, fi.FuncLv+1, node.Loc, node.IsVararg, scope, a.curResult.Name)
	subFi.IsColon = node.IsColon
	subFi.ClassName = node.ClassName
	subFi.FuncName = node.FuncName

	fileResult := a.curResult
	fileResult.InertNewFunc(subFi)

	// 在第一轮中，检查函数的是否包含同名的参数
	a.checkDuplicateFunParam(node)

	// 所有的参数放入到局部变量中
	for index, param := range node.ParList {
		varIndex := uint8(index + 1)

		locVar := subFi.MainScope.AddLocVar(fileResult.Name, param, common.LuaTypeAll, nil, node.ParLocList[index],
			varIndex)
		locVar.IsParam = true
		locVar.IsUse = true

		// 函数所有的参数放入数组进去，函数代码提示的时候有用
		subFi.ParamList = append(subFi.ParamList, param)
	}

	if a.isSecondTerm() {
		//获取参数与返回值的注解类型
		a.loadFuncParamAnnType(subFi)
	}

	// 备份
	backupFunc := a.curFunc
	backupScope := a.curScope

	a.curFunc = subFi
	a.curScope = subFi.MainScope
	a.cgBlock(node.Block)
	a.exitScope()

	// 还原
	a.curFunc = backupFunc
	a.curScope = backupScope
	return subFi
}

func (a *Analysis) cgTableConstructorExp(node *ast.TableConstructorExp, parentVar *common.VarInfo) {
	fileResult := a.curResult

	var tabKeyMap map[string]lexer.Location
	if a.isFirstTerm() && !a.realTimeFlag {
		tabKeyMap = make(map[string]lexer.Location, len(node.KeyExps))
	}

	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]
		if keyExp == nil {
			a.cgExp(valExp, nil, nil)
			continue
		}

		strKeySimple := ""
		strExp, ok := keyExp.(*ast.StringExp)
		if ok {
			strKeySimple = strExp.Str
		}

		var subVar *common.VarInfo
		if parentVar != nil && strKeySimple != "" {
			// 创建子的subKey
			subVar = common.CreateVarInfo(fileResult.Name, common.GetExpType(valExp), nil, strExp.Loc, 1)
		}

		a.cgExp(keyExp, nil, nil)
		// key值为nil
		oneFunc, oneRefer := a.cgExp(valExp, subVar, nil)
		if subVar != nil {
			subVar.ReferFunc = oneFunc
			subVar.ReferInfo = oneRefer
			subVar.ReferExp = valExp

			if !parentVar.IsExistMember(strKeySimple) {
				parentVar.InsertSubMember(strKeySimple, subVar)
			}
		}

		// 只在第一轮检查，且是非实时时，才对table里面的key是否包含重复的值进行校验
		if !a.isFirstTerm() || a.realTimeFlag {
			continue
		}

		strKey, strShow, loc := common.GetTableConstuctorKeyStr(keyExp, node.Loc)
		if strKey == "" {
			continue
		}

		// 第一轮检查table是否有重复的key
		if beforeLoc, ok := tabKeyMap[strKey]; ok {
			errStr := fmt.Sprintf("the table contains duplicate keys: %s", strShow)
			var relateVec []common.RelateCheckInfo
			relateVec = append(relateVec, common.RelateCheckInfo{
				LuaFile: fileResult.Name,
				ErrStr:  errStr,
				Loc:     beforeLoc,
			})
			fileResult.InsertRelateError(common.CheckErrorTableDuplicateKey, errStr, loc, relateVec)
		} else {
			tabKeyMap[strKey] = loc
		}
	}
}

// r[a] := op exp
func (a *Analysis) cgUnopExp(node *ast.UnopExp) {
	if node.Op == lexer.TkOpNot && !a.isFirstTerm() && !a.isFourTerm() && !a.isFiveTerm() &&
		a.ignoreInfo.inif {
		strExpName := common.GetExpName(node.Exp)
		strSubKey := common.GetExpSubKey(strExpName)
		if strSubKey != "" {
			a.ignoreInfo.strName = strSubKey
			a.ignoreInfo.line = node.Loc.EndLine
		}
	}

	a.cgExp(node.Exp, nil, nil)
	a.ignoreInfo.strName = ""
}

// r[a] := exp1 op exp2
func (a *Analysis) cgBinopExp(node *ast.BinopExp, parentVar *common.VarInfo) {
	if node.Op == lexer.TkOpEq && a.ignoreInfo.inif && !a.isFirstTerm() && !a.isFourTerm() &&
		!a.isFiveTerm() {
		if _, ok := node.Exp2.(*ast.NilExp); ok {
			strExpName := common.GetExpName(node.Exp1)
			strSubKey := common.GetExpSubKey(strExpName)
			if strSubKey != "" {
				a.ignoreInfo.strName = strSubKey
				a.ignoreInfo.line = node.Loc.StartLine
			}
		}
	}

	if node.Op == lexer.TkOpOr && !a.ignoreInfo.inif && !a.isFirstTerm() && !a.isFourTerm() &&
		!a.isFiveTerm() &&
		a.ignoreInfo.assignName != "" {
		strExpName := common.GetExpName(node.Exp1)
		strSubKey := common.GetExpSubKey(strExpName)
		if strSubKey == a.ignoreInfo.assignName {
			a.ignoreInfo.strName = strSubKey
			a.ignoreInfo.line = node.Loc.StartLine
		}
	}

	a.cgExp(node.Exp1, parentVar, node)
	a.cgExp(node.Exp2, parentVar, node)
	a.ignoreInfo.strName = ""

	// 下面的判断只在第一轮，且是非实时检查时才触发
	if !a.isFirstTerm() || a.realTimeFlag {
		return
	}

	fileResult := a.curResult
	// 判断二元表达式 or、and 两边的值写法有问题，例如 a = a or true ，始终为true; b = b and false, 始终为false
	if node.Op == lexer.TkOpOr || node.Op == lexer.TkOpAnd {
		if node.Op == lexer.TkOpOr {
			// 判断 or 二元表达式两边是否有 true， 例如 a = a or true，结果始终为true
			_, flag1 := node.Exp1.(*ast.TrueExp)
			_, flag2 := node.Exp2.(*ast.TrueExp)
			if flag1 || flag2 {
				oneLoc := common.GetExpLoc(node.Exp1)
				twoLoc := common.GetExpLoc(node.Exp2)
				if !oneLoc.IsInitialLoc() && !twoLoc.IsInitialLoc() {
					errLoc := lexer.GetRangeLoc(&oneLoc, &twoLoc)
					errStr := "or expression is always true"
					fileResult.InsertError(common.CheckErrorOrAlwaysTrue, errStr, errLoc)
				}
			}
		}

		if node.Op == lexer.TkOpAnd {
			// 判断 and 二元表达式两边是否有false， 例如 b = b and false，结果始终为false
			_, flag1 := node.Exp1.(*ast.FalseExp)
			_, flag2 := node.Exp2.(*ast.FalseExp)
			if flag1 || flag2 {
				oneLoc := common.GetExpLoc(node.Exp1)
				twoLoc := common.GetExpLoc(node.Exp2)
				if !oneLoc.IsInitialLoc() && !twoLoc.IsInitialLoc() {
					errLoc := lexer.GetRangeLoc(&oneLoc, &twoLoc)
					errStr := "and expression is always false"
					fileResult.InsertError(common.CheckErrorAndAlwaysFalse, errStr, errLoc)
				}
			}
		}
	}

	//浮点数做等于或不等于判断 告警
	if !common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorFloatEq) {
		if node.Op == lexer.TkOpEq || node.Op == lexer.TkOpNe {
			_, ok1 := node.Exp1.(*ast.FloatExp)
			_, ok2 := node.Exp2.(*ast.FloatExp)

			if ok1 || ok2 {
				errStr := "float compare error"
				fileResult.InsertError(common.CheckErrorFloatEq, errStr, node.Loc)
			}
		}
	}

	// 第一阶段，判断两边的是否一样
	//lexer.TkOpOr                        // or
	//lexer.TkOpAnd                       // and
	//lexer.TkOpLt                        // <
	//lexer.TkOpLe                        // <=
	//lexer.TkOpGt                        // >
	//lexer.TkOpGe                        // >=
	//lexer.TkOpEq                        // ==
	//lexer.TkOpNe                        // ~=
	if node.Op == lexer.TkOpOr || node.Op == lexer.TkOpAnd || node.Op == lexer.TkOpLt ||
		node.Op == lexer.TkOpLe || node.Op == lexer.TkOpGt || node.Op == lexer.TkOpGe ||
		node.Op == lexer.TkOpEq || node.Op == lexer.TkOpNe {
		leftExpStr := common.GetExpName(node.Exp1)
		if strings.Contains(leftExpStr, "#") {
			return
		}

		rightExpStr := common.GetExpName(node.Exp2)
		if strings.Contains(rightExpStr, "#") {
			return
		}

		if leftExpStr == rightExpStr {
			oneLoc := common.GetExpLoc(node.Exp1)
			twoLoc := common.GetExpLoc(node.Exp2)
			if !oneLoc.IsInitialLoc() && !twoLoc.IsInitialLoc() {
				errLoc := lexer.GetRangeLoc(&oneLoc, &twoLoc)
				rightExpStr = strings.TrimPrefix(rightExpStr, "!")
				errStr := fmt.Sprintf("binary expression right var:'%s' is the same to left var:'%s'", rightExpStr,
					rightExpStr)
				fileResult.InsertError(common.CheckErrorDuplicateExp, errStr, errLoc)
			}
		}
	}

}

// r[a] := f(args)
func (a *Analysis) cgFuncCallExp(node *ast.FuncCallExp) *common.ReferInfo {
	// 在第二阶段，里面可能深度加载依赖的文件
	newRefer := a.GetImportRefer(node)

	//nArgs := len(node.Args)
	a.cgExp(node.PrefixExp, nil, nil)

	if node.NameExp != nil {
		//analysis.cgExp(node.ast.NameExp)

		// todo，例如有这样的定义
		// a = {}
		// function a:test() end
		// a:test() -- 冒号调用时候，判断是否要查找
		a.findFuncColon(node.PrefixExp, node.NameExp, node.Loc)
	}

	for _, arg := range node.Args {
		a.cgExp(arg, nil, nil)
	}

	// 第二轮或第三轮函数参数check
	a.cgFuncCallParamCheck(node)

	return newRefer
}

// r[a] := prefix[key]
// binParentExp 为二元表达式，父的BinopExp指针， 例如 a = b and c，当对c变量调用cgExp时候，binParentExp为b and c
func (a *Analysis) cgTableAccessExp(node *ast.TableAccessExp, binParentExp *ast.BinopExp) {
	a.cgExp(node.PrefixExp, nil, nil)
	a.cgExp(node.KeyExp, nil, nil)

	if a.isFirstTerm() {
		if oneName, ok := node.PrefixExp.(*ast.NameExp); ok {
			if oneName.Name == "_G" {
				if strExp, ok1 := node.KeyExp.(*ast.StringExp); ok1 {
					a.analysisNoDefineStr(strExp)
				}
			}
		}

		a.checkIfNotTableAccess(node, binParentExp)

		a.expandVarStrMap(node)
		return
	}

	// 查找引用的table之前是否有定义过
	// 例如 one = import("one.lua")
	//      one.test_one() 是否有定义
	a.findTableDefine(node)
	a.checkTableAccess(node)
}
