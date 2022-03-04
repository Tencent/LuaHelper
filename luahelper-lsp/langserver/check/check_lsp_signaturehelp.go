package check

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"strings"
)

// SignaturehelpFunc 定位到函数的参数位置
// 查找变量的定义
func (a *AllProject) SignaturehelpFunc(strFile string, varStruct *common.DefineVarStruct) (flag bool,
	sinatureInfo common.SignatureHelpInfo, paramInfo []common.SignatureHelpInfo) {
	flag = false

	// 3) 判断是否为项目中定义的函数
	oldSymbol, symList := a.FindVarDefine(strFile, varStruct)
	if oldSymbol == nil || oldSymbol.VarInfo == nil || len(symList) == 0 {
		// // 1) 判断是否为系统函数的参数提示
		strVecLen := len(varStruct.StrVec)
		if strVecLen == 1 {
			flag, sinatureInfo, paramInfo = a.judgetSystemFuncSignature(varStruct.StrVec[0])
			if flag {
				return
			}
		}

		// 2) 判断是否为系统的模块函数提示
		if strVecLen == 2 {
			flag, sinatureInfo, paramInfo = a.judgetSystemModuleFuncSigatrue(varStruct.StrVec[0], varStruct.StrVec[1])
			if flag {
				return
			}
		}

		return
	}

	// 找到第一个信息
	lastSymbol := symList[len(symList)-1]
	if lastSymbol.VarInfo == nil || lastSymbol.VarFlag == common.FirstAnnotateFlag {
		return a.getAnnotateFuncSignature(lastSymbol, varStruct.ColonFlag)
	}

	referFunc := lastSymbol.VarInfo.ReferFunc
	if referFunc == nil {
		// 如果指向的函数为空，判断是否为注解的函数的类型
		return a.getAnnotateFuncSignature(lastSymbol, varStruct.ColonFlag)
	}

	if varStruct.ColonFlag && !referFunc.IsColon {
		// 传人的是带冒号的，如果语法不是带冒号的，退出
		return
	}

	// 变化包含在的lua文件
	inLuaFile := lastSymbol.FileName
	lastLine := referFunc.Loc.StartLine

	strName := varStruct.StrVec[len(varStruct.StrVec)-1]
	funAllStr := a.getFuncShowStr(lastSymbol.VarInfo, strName, true, varStruct.ColonFlag, false)
	sinatureInfo.Label = funAllStr
	strDocumentation := getFinalStrComment(a.GetLineComment(inLuaFile, lastLine), false)

	annotateParamInfo := a.GetFuncParamInfo(inLuaFile, lastLine-1)

	// 表示是否获取到一个参数的有效标记
	for index, strOneParam := range referFunc.ParamList {
		if varStruct.ColonFlag && index == 0 {
			// 如果是带冒号的语法，忽略第一个self参数
			continue
		}

		annotateFlag := false
		paramShortStr, annType := a.getAnnotateFuncParamDocument(strOneParam, annotateParamInfo, inLuaFile, lastLine-1)
		if paramShortStr != "" {
			annotateFlag = true
		} else {
			paramShortStr = getFuncParamShortDocument(strDocumentation, strOneParam)
		}

		if paramShortStr == "" {
			// 如果没有找到匹配的参数，获取所有的
			paramShortStr = strDocumentation
		} else {
			paramShortStr = strOneParam + " : " + paramShortStr
		}

		oneSignatureParam := common.SignatureHelpInfo{
			Label:         strOneParam,
			Documentation: paramShortStr,
			AnnotateFlag:  annotateFlag,
			AnnType:       annType,
		}

		paramInfo = append(paramInfo, oneSignatureParam)
	}

	if referFunc.IsVararg {
		oneSignatureParam := common.SignatureHelpInfo{
			Label:         "...",
			Documentation: strDocumentation,
		}
		paramInfo = append(paramInfo, oneSignatureParam)
	}

	flag = true
	return
}

// 获取完全为注解的函数的类型
func (a *AllProject) getAnnotateFuncSignature(symbol *common.Symbol, colonFlag bool) (flag bool,
	sinatureInfo common.SignatureHelpInfo, paramInfo []common.SignatureHelpInfo) {
	// 判断是否有注解类型
	if symbol.AnnotateType == nil {
		return
	}

	funcType := a.GetAstTypeFuncType(symbol.AnnotateType, symbol.FileName, symbol.AnnotateLine)
	if funcType == nil {
		return
	}

	flag = true
	var astType annotateast.Type = funcType

	className := symbol.StrPreClassName
	funcColonFlag := false

	// 检查下面的逻辑是否类似下面的，下面的含义表示ClassA有一个：的FunctionC 函数
	// 当为这样的片段，且是：补全时候，忽略到第一个self参数
	// ---@class ClassA
	// ---@field FunctionC fun(self:ClassA):void
	if colonFlag && len(funcType.ParamNameList) > 0 && len(funcType.ParamTypeList) > 0 &&
		funcType.ParamNameList[0] == "self" && annotateast.TypeConvertStr(funcType.ParamTypeList[0]) == className {
		funcColonFlag = true
		sinatureInfo.Label = annotateast.FuncTypeConvertStr(funcType, 1)
	} else {
		sinatureInfo.Label = annotateast.TypeConvertStr(astType)
	}

	// 表示是否获取到一个参数的有效标记
	for index, strOneParam := range funcType.ParamNameList {
		if index == 0 && funcColonFlag {
			continue
		}

		strType := ""
		if len(funcType.ParamTypeList) > index {
			strType = annotateast.TypeConvertStr(funcType.ParamTypeList[index])
		}

		paramShortStr := strOneParam + " : " + strType
		oneSignatureParam := common.SignatureHelpInfo{
			Label:         strOneParam,
			Documentation: paramShortStr,
		}
		paramInfo = append(paramInfo, oneSignatureParam)
	}

	return
}

// 获取函数参数的注解信息
func (a *AllProject) getAnnotateFuncParamDocument(strOneParam string, paramInfo *common.FragementParamInfo,
	fileName string, line int) (strShort string, annType annotateast.Type) {
	// 先判断是否有注解信息
	if paramInfo == nil {
		return "", annType
	}

	for _, oneParam := range paramInfo.ParamList {
		if oneParam.Name != strOneParam {
			continue
		}

		strShort = a.getSymbolAliasMultiCandidate(oneParam.ParamType, fileName, line)
		if strShort == "" {
			strShort = annotateast.TypeConvertStr(oneParam.ParamType)
		}
		annType = oneParam.ParamType

		if oneParam.Comment != "" {
			strShort = strShort + " -- " + oneParam.Comment
		}

		break
	}

	return strShort, annType
}

// 获取函数中每一个参数的简短提示
// strAllDocment 为函数的所有的注释
// strOneParam 为其中的一个参数
// paramInfo 为注解的所有参数信息
func getFuncParamShortDocument(strAllDocment string, strOneParam string) (strShort string) {

	// 1) 先把所有的注释切分成多行的数组
	splitStrArr := strings.Split(strAllDocment, "\n")

	strModuleArray := []string{}
	// a) 首先判断是否为emmylua的注释类型
	// @param strParam
	strModuleArray = append(strModuleArray, "@param "+strOneParam+" ")

	// b-1) 判断是否为luaide的注释类型
	// @ss:
	strModuleArray = append(strModuleArray, "@"+strOneParam+":")

	// b-2) 判断是否为luaide的注释类型
	// ss :
	strModuleArray = append(strModuleArray, strOneParam+" :")

	// b-3) 判断是否为luaide的注释类型
	// ss :
	strModuleArray = append(strModuleArray, strOneParam+":")

	// c-1) 判断是否为最简单的注释
	//  ss 表示是否要
	strModuleArray = append(strModuleArray, " "+strOneParam+" ")

	// c-2) 判断是否为最简单的注释
	//  ss表示是否要
	strModuleArray = append(strModuleArray, " "+strOneParam)

	// c-3) 判断是否为最简单的注释
	// ss 表示是否要
	strModuleArray = append(strModuleArray, strOneParam+" ")

	// c-4) 判断是否为最简单的注释
	// ss表示是否要
	strModuleArray = append(strModuleArray, strOneParam)

	findIndex := 10000
	findShort := ""
	for _, oneLineStr := range splitStrArr {
		for index, strModule := range strModuleArray {
			searchIndex := strings.Index(oneLineStr, strModule)
			if searchIndex < 0 {
				continue
			}

			partComment := oneLineStr[len(strModule)+searchIndex:]
			// 如果首行有空格，删除掉
			partComment = strings.TrimPrefix(partComment, " ")

			if index < findIndex {
				findIndex = index
				findShort = partComment
			}
			break
		}
	}

	return findShort
}

// 系统函数转换为对应的结构
func (a *AllProject) systemFuncConver(oneSystemTips *common.SystemNoticeInfo) (sinatureInfo common.SignatureHelpInfo,
	paramInfo []common.SignatureHelpInfo) {
	sinatureInfo.Label = oneSystemTips.Detail
	sinatureInfo.Documentation = oneSystemTips.Documentation

	for _, oneParamInfo := range oneSystemTips.FuncParamVec {
		oneSignatureParam := common.SignatureHelpInfo{
			Label:         oneParamInfo.Label,
			Documentation: oneParamInfo.Documentation,
		}
		paramInfo = append(paramInfo, oneSignatureParam)
	}
	return
}

// 判断是否为系统函数sigatrueHelp
func (a *AllProject) judgetSystemFuncSignature(strName string) (flag bool,
	sinatureInfo common.SignatureHelpInfo, paramInfo []common.SignatureHelpInfo) {
	if oneSystemTips, ok := common.GConfig.SystemTipsMap[strName]; ok {
		flag = true
		sinatureInfo, paramInfo = a.systemFuncConver(&oneSystemTips)
	}

	return
}

// 判断是否为系统模块中的成员函数sigatrueHelp
func (a *AllProject) judgetSystemModuleFuncSigatrue(strName string, strKey string) (flag bool,
	sinatureInfo common.SignatureHelpInfo, paramInfo []common.SignatureHelpInfo) {
	oneMouleInfo, ok := common.GConfig.SystemModuleTipsMap[strName]
	if !ok {
		return
	}

	if oneSystemTips, ok := oneMouleInfo.ModuleFuncMap[strKey]; ok {
		flag = true
		sinatureInfo, paramInfo = a.systemFuncConver(oneSystemTips)
	}
	return
}
