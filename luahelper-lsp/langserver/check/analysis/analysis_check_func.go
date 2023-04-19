package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"strings"
)

//函数调用处的参数类型检查
func (a *Analysis) funcCallParamTypeCheck(node *ast.FuncCallStat, referFunc *common.FuncInfo, findTerm int) {

	// 第二轮或第三轮函数参数check
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorCallParamType) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorCallParamType]; !ok {
		return
	}

	if findTerm == 1 {
		a.loadFuncParamAnnType(referFunc)
	}

	for i, argExp := range node.Args {
		if i >= len(referFunc.ParamList) {
			//可能是可变参数导致
			break
		}

		allAnnTypeVec, ok := referFunc.ParamType[referFunc.ParamList[i]]
		if !ok {
			//该参数没写注解
			continue
		}

		//函数调用处的参数类型
		argCallTypeVec := a.GetAnnTypeByExp(argExp, -1)
		if len(argCallTypeVec) == 0 {
			// 取不到参数类型
			continue
		}

		hasMatch := false
		for _, argCallTypeOne := range argCallTypeVec {
			for _, argAnnTypeOne := range allAnnTypeVec {
				if a.CompAnnTypeAndCodeType(argAnnTypeOne, argCallTypeOne) {
					hasMatch = true
					break
				}
			}
		}

		if hasMatch {
			continue
		}

		loc := common.GetExpLoc(argExp)
		allAnnTypeStr := strings.Join(allAnnTypeVec, "|")
		argCallTypeStr := strings.Join(argCallTypeVec, "|")

		//类型不一致，报警
		errorStr := fmt.Sprintf("Expected parameter of type '%s', '%s' provided", allAnnTypeStr, argCallTypeStr)
		a.curResult.InsertError(common.CheckErrorCallParamType, errorStr, loc)
	}
}

// 函数体内的返回值类型检查 检查函数的返回值类型与注解类型是否匹配 一次检查一个return语句
func (a *Analysis) funcReturnCheck(retInfo *common.ReturnInfo) {
	// 第二轮或第三轮函数参数check
	if !a.isNeedCheck() || a.realTimeFlag {
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

	if len(a.curFunc.ReturnType) == 0 {
		//无注解不检查
		return
	}

	//注解返回值可以有多个类型 如---@return number, string|number 第二个返回值可以是字符串或者数
	//retInfo.ReturnVarVec 是一个return语句 oneReturn是一个返回值
	//例如 "return 0,true" 0是一个oneReturn
	for i, oneReturn := range retInfo.ReturnVarVec {

		if i >= len(a.curFunc.ReturnType) {
			//实际返回值个数超过注解中的返回值数量
			continue
		}

		//获取return处表达式的返回值类型
		returnTypeVec := a.GetAnnTypeByExp(oneReturn.ReturnExp, -1)
		if len(returnTypeVec) == 0 {
			// 无法识别类型
			continue
		}

		allReturnTypeStr := strings.Join(returnTypeVec, "|")
		allAnnTypeStr := strings.Join(a.curFunc.ReturnType[i], "|")

		loc := common.GetExpLoc(oneReturn.ReturnExp)
		hasMatch := false
		for _, codeType := range returnTypeVec {
			for _, annType := range a.curFunc.ReturnType[i] {
				if a.CompAnnTypeAndCodeType(annType, codeType) {
					hasMatch = true
					break
				}
			}
		}

		if !hasMatch {

			//类型不一致，报警
			errorStr := fmt.Sprintf("Return value is expected to be '%s', '%s' returned", allAnnTypeStr, allReturnTypeStr)

			a.curResult.InsertError(common.CheckErrorFuncRetErr, errorStr, loc)
		}
	}
}

func (a *Analysis) checkLocFuncCall() {

	// 如果是一轮校验，判断是否要校验局部变量是否定义未使用
	if !a.isFirstTerm() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorLocFuncNotCall) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorLocFuncNotCall]; !ok {
		return
	}

	scope := a.curScope
	if scope == nil {
		return
	}

	fileResult := a.curResult

	// 扫描当前scope，判断哪些局部函数定义了未使用
	for varName, varInfoList := range scope.LocVarMap {
		// _ 局部变量忽略, _G也忽略
		if varName == "_" || varName == "_G" {
			continue
		}

		for _, oneVar := range varInfoList.VarVec {
			if oneVar.IsUse || oneVar.IsClose {
				continue
			}

			//只看局部函数
			if oneVar.ReferFunc == nil {
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

			errorStr := fmt.Sprintf("Unused local function '%s' ", varName)
			fileResult.InsertError(common.CheckErrorLocFuncNotCall, errorStr, oneVar.Loc)

			// 清除掉
			oneVar.NoUseAssignLocs = nil
		}
	}
}
