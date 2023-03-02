package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"strings"
)

func (a *Analysis) funcCallParamTypeCheck(node *ast.FuncCallStat, referFunc *common.FuncInfo) {

	// 第二轮或第三轮函数参数check
	if !a.isNeedCheck() {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorCallParamType) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorCallParamType]; !ok {
		return
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
		argCallTypeVec := a.GetAnnTypeStrForRefer(argExp, -1)
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
