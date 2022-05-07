package check

import (
	"fmt"
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"sort"
	"strings"
)

type mapEntryHandler func(string, string)

// 按字母顺序遍历map
func traverseMapInStringOrder(params map[string]string, handler mapEntryHandler) {
	keys := make([]string, 0)
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for index, k := range keys {
		if index == common.GConfig.PreviewFieldsNum {
			strMore := fmt.Sprintf("...(+%d)", len(keys)-common.GConfig.PreviewFieldsNum)
			handler(k, strMore)
			break
		} else {
			handler(k, params[k])
		}
	}
}

func (a *AllProject) convertClassInfoToHovers(oneClass *common.OneClassInfo, existMap map[string]string) {
	// 1) oneClass所有的成员
	for strName, fieldState := range oneClass.FieldMap {
		if _, ok := existMap[strName]; ok {
			continue
		}

		strFiled := annotateast.TypeConvertStr(fieldState.FiledType)

		if fieldState.Comment != "" {
			existMap[strName] = strName + ": " + strFiled + ",  -- " + fieldState.Comment
		} else {
			existMap[strName] = strName + ": " + strFiled + ","
		}
	}

	a.getVarInfoMapStr(oneClass.RelateVar, existMap)
}

// 代码补全时，有注解类型的字段提示
func (a *AllProject) completeAnnotatTypeStr(astType annotateast.Type, fileName string, line int) (str string) {
	str = annotateast.TypeConvertStr(astType)
	if isDefaultType(str) {
		return str
	}

	classList := a.getAllNormalAnnotateClass(astType, fileName, line)
	if len(classList) == 0 {
		// 没有找到相应的class成员，直接返回
		str = str + a.getSymbolAliasMultiCandidate(astType, fileName, line)
		return str
	}

	// 为已经存在的map，防止重复
	var existMap map[string]string = map[string]string{}
	strType := annotateast.TypeConvertStr(astType)
	str = strType + " = {\n"
	for _, oneClass := range classList {
		a.convertClassInfoToHovers(oneClass, existMap)
	}

	traverseMapInStringOrder(existMap, func(key string, value string) {
		str = str + "\t" + value + "\n"
	})

	str = str + "}"
	str = str + a.getSymbolAliasMultiCandidate(astType, fileName, line)
	return str
}

func (a *AllProject) getClassFieldStr(classInfo *common.OneClassInfo) (str string) {
	var existMap map[string]string = map[string]string{}
	str = " = {\n"
	a.convertClassInfoToHovers(classInfo, existMap)

	traverseMapInStringOrder(existMap, func(key string, value string) {
		str = str + "\t" + value + "\n"
	})

	str = str + "}"
	return str
}

// funcFlag 表示是否为函数名匹配，增加（）后能够匹配
func matchVecsExpandStrMap(inputVec []string, inputFuncVec []bool, expandStr string) (remainVec []string) {
	expandVec := strings.Split(expandStr, ".")
	if len(expandVec) <= len(inputVec) {
		return
	}

	for i, inputStr := range inputVec {
		oneStr := expandVec[i]

		if len(inputFuncVec) > i && inputFuncVec[i] {
			inputStr = inputStr + "()"
		}

		// 直接相等，匹配上来
		if inputStr == oneStr {
			continue
		}

		// #或是！开头都认为是相同的
		if (strings.HasPrefix(inputStr, "!") || strings.HasPrefix(inputStr, "#")) &&
			(strings.HasPrefix(oneStr, "!") || strings.HasPrefix(oneStr, "#")) {
			continue
		}

		// 没有匹配到，提前返回
		return
	}

	remainVec = expandVec[len(inputVec):]
	return
}

func isDefaultType(str string) bool {
	if str == "number" || str == "any" || str == "string" || str == "boolean" || str == "nil" || str == "thread" ||
		str == "userdata" || str == "lightuserdata" || str == "integer" || str == "void" {
		return true
	}

	return false
}

func needReplaceMapStr(oldStr string, newType string, newStr string) bool {
	if newType == "any" {
		return false
	}

	if strings.Contains(oldStr, " -- ") {
		// 如果旧的含有 -- 注释，不用替换
		return false
	}

	if strings.Contains(oldStr, ": number") && !strings.Contains(oldStr, ": number = ") {
		return true
	}

	if strings.Contains(oldStr, ": string") && !strings.Contains(oldStr, ": string = ") {
		return true
	}

	if strings.Contains(oldStr, ": boolean") && !strings.Contains(oldStr, ": boolean = ") {
		return true
	}

	if strings.Contains(oldStr, ": any") {
		return true
	}

	return false
}

// 补全的为expand 的变量，获取展开的信息
func (a *AllProject) getCompleteExpandInfo(item *common.OneCompleteData) (luaFileStr string) {
	cache := a.completeCache
	compStruct := cache.GetCompleteVar()
	tempVec := compStruct.StrVec
	tempVec = append(tempVec, item.Label)
	str := getVarInfoExpandStrHover(item.ExpandVarInfo, tempVec, compStruct.IsFuncVec, "")
	if str != "" {
		item.Detail = str
	}

	return luaFileStr
}

// 只获取expandStrMap的hover
func getVarInfoExpandStrHover(varInfo *common.VarInfo, inputVec []string, inputFuncVec []bool, strPre string) (str string) {
	if varInfo == nil || varInfo.ExpandStrMap == nil {
		return
	}

	beforeStrPre := inputVec[0]

	var existMap map[string]string = map[string]string{}
	for str := range varInfo.ExpandStrMap {
		str = beforeStrPre + "." + str
		remainVec := matchVecsExpandStrMap(inputVec, inputFuncVec, str)
		if len(remainVec) == 0 {
			continue
		}

		oneStr := remainVec[0]
		strType := "any"
		if len(remainVec) > 1 {
			strType = "table"
		}
		newStr := oneStr + ": " + strType + ","
		if oldStr, ok := existMap[oneStr]; ok {
			if needReplaceMapStr(oldStr, strType, newStr) {
				existMap[oneStr] = newStr
			}
		} else {
			existMap[oneStr] = newStr
		}
	}
	if len(existMap) == 0 {
		return "any"
	}

	if strPre != "" {
		str = strPre + " : table = {\n"
	} else {
		str = " table = {\n"
	}

	traverseMapInStringOrder(existMap, func(key string, value string) {
		str = str + "\t" + value + "\n"
	})

	str = str + "}"
	return
}

// 查看varInfo详细的类型，递归查找
func (a *AllProject) GetVarRelateTypeStr(varInfo *common.VarInfo) (strType string) {
	strType = varInfo.GetVarTypeDetail()

	if varInfo.ReferFunc != nil {
		//strType = varInfo.ReferFunc.GetFuncCompleteStr("function", true, false)
		strType = a.getFuncShowStr(varInfo, "function", true, false, true, false)
		return
	}

	if strType != "any" {
		return
	}

	if varInfo.ReferExp == nil {
		return
	}

	loc := varInfo.Loc
	comParam := a.getCommFunc(varInfo.FileName, loc.StartLine, loc.StartColumn)
	findExpList := &[]common.FindExpFile{}
	symbol := a.FindVarReferSymbol(varInfo.FileName, varInfo.ReferExp, comParam, findExpList, 1)
	if symbol == nil {
		return
	}

	if symbol.AnnotateType != nil {
		strType = annotateast.TypeConvertStr(symbol.AnnotateType)
	} else {
		if symbol.VarInfo != nil {
			strTmp := getParamVarinfoType(symbol.VarInfo)
			if strTmp != "any" {
				strType = strTmp
			}
		}
	}

	return strType
}

func (a *AllProject) getVarInfoMapStr(varInfo *common.VarInfo, existMap map[string]string) {
	if varInfo == nil {
		return
	}

	for key, value := range varInfo.SubMaps {
		// strValueType := value.GetVarTypeDetail()
		// if value.ReferFunc != nil {
		// 	strValueType = value.ReferFunc.GetFuncCompleteStr("function", true, false)
		// }
		strValueType := a.GetVarRelateTypeStr(value)
		newStr := key + ": " + strValueType + ","

		strComment := a.GetLineComment(value.FileName, value.Loc.StartLine)
		strComment = strings.TrimPrefix(strComment, " ")
		if strComment != "" {
			// 只提取一行注释
			strVec := strings.Split(strComment, "\n")
			if len(strVec) > 0 {
				strComment = strVec[0]
			}
		}
		if strComment != "" {
			if strings.HasPrefix(strComment, "-") {
				newStr = newStr + "  --" + strComment
			} else {
				newStr = newStr + "  -- " + strComment
			}

		}
		if oldStr, ok := existMap[key]; ok {
			if needReplaceMapStr(oldStr, strValueType, newStr) {
				existMap[key] = newStr
			}
		} else {
			existMap[key] = newStr
		}
	}

	for key := range varInfo.ExpandStrMap {
		strVec := strings.Split(key, ".")
		if len(strVec) == 0 {
			continue
		}

		oneStr := strVec[0]
		if !common.JudgeSimpleStr(oneStr) {
			continue
		}

		strType := "any"
		if len(strVec) > 1 {
			strType = "table"
		}

		newStr := oneStr + ": " + strType + ","
		if oldStr, ok := existMap[oneStr]; ok {
			if needReplaceMapStr(oldStr, strType, newStr) {
				existMap[oneStr] = newStr
			}
		} else {
			existMap[oneStr] = newStr
		}
	}
}

func isAllClassDefault(classList []*common.OneClassInfo) bool {
	for _, oneClass := range classList {
		if oneClass.ClassState != nil && !isDefaultType(oneClass.ClassState.Name) {
			return false
		}
	}

	return true
}

// hover 的时候是指向一个table，展开这个table的内容
func (a *AllProject) expandTableHover(symbol *common.Symbol) (str string, existMap map[string]string) {
	// 为已经存在的map，防止重复
	existMap = map[string]string{}

	// 1) 先判断是否有注解类型
	if symbol.AnnotateType != nil {
		strType := annotateast.TypeConvertStr(symbol.AnnotateType)
		// if isDefaultType(strType) {
		// 	return strType, existMap
		// }

		classList := a.getAllNormalAnnotateClass(symbol.AnnotateType, symbol.FileName, symbol.GetLine())
		if isAllClassDefault(classList) {
			// 没有找到相应的class成员，直接返回
			strCandidate := a.getSymbolAliasMultiCandidate(symbol.AnnotateType, symbol.FileName, symbol.GetLine())
			if strCandidate == "" {
				return strType, existMap
			}
			if strType == "" {
				strType = strCandidate
				return strType, existMap
			}

			if strings.HasPrefix(strCandidate, "\n") {
				strType = strType + strCandidate
			} else {
				strType = strType + " | " + strCandidate
			}

			return strType, existMap
		}

		str = strType + " = {\n"
		for _, oneClass := range classList {
			a.convertClassInfoToHovers(oneClass, existMap)
		}
	} else {
		str = "table = {\n"
		if symbol.VarInfo == nil || (len(symbol.VarInfo.SubMaps) == 0 && len(symbol.VarInfo.ExpandStrMap) == 0) {
			return "table = { }", existMap
		}
	}

	a.getVarInfoMapStr(symbol.VarInfo, existMap)
	traverseMapInStringOrder(existMap, func(key string, value string) {
		str = str + "\t" + value + "\n"
	})

	str = str + "}"

	if symbol.AnnotateType != nil {
		str = str + a.getSymbolAliasMultiCandidate(symbol.AnnotateType, symbol.FileName, symbol.GetLine())
	}
	return str, existMap
}

func (a *AllProject) getSymbolAliasMultiCandidateMap(annotateType annotateast.Type, fileName string, line int) (strMap map[string]string) {
	strMap = map[string]string{}

	if annotateType == nil {
		return
	}

	multiType, ok := annotateType.(*annotateast.MultiType)
	if !ok {
		return
	}

	for _, oneType := range multiType.TypeList {
		// 判断是否在几个候选词中
		if constType, ok := oneType.(*annotateast.ConstType); ok {
			oneStr := ""
			if constType.QuotesFlag {
				oneStr = "\"" + constType.Name + "\""
			} else {
				oneStr = constType.Name
			}

			strMap[oneStr] = ""
			continue
		}

		simpleType, ok := oneType.(*annotateast.NormalType)
		if !ok {
			continue
		}

		if isDefaultType(simpleType.StrName) {
			continue
		}

		a.getAliasMultiCandidateMap(simpleType.StrName, fileName, line, strMap)
	}

	return strMap
}

// 判断是否为函数参数的候选词代码补全, 函数的参数为alias类型，补全参数常量
func (a *AllProject) paramCandidateComplete(comParam *CommonFuncParam, completeVar *common.CompleteVarStruct) bool {
	if completeVar.ParamCandidateType == nil {
		return false
	}

	strMap := a.getSymbolAliasMultiCandidateMap(completeVar.ParamCandidateType, comParam.fi.FileName, comParam.loc.StartLine)
	if len(strMap) > 0 {
		if completeVar.SplitByte == '\'' || completeVar.SplitByte == '"' {
			a.completeCache.SetClearParamQuotes(true)
		}

		for strKey, strComment := range strMap {
			if completeVar.SplitByte == '\'' || completeVar.SplitByte == '"' {
				// 如果分割的是为单引号或是双引号，strKey需要为引号的字符串
				if !strings.HasPrefix(strKey, "\"") {
					continue
				}

				if completeVar.SplitByte == '\'' {
					strKey = strings.ReplaceAll(strKey, "\"", "'")
				}
			} else {
				if completeVar.SplitByte != ' ' && strings.HasPrefix(strKey, "\"") {
					continue
				}
			}

			a.completeCache.InsertCompleteNormal(strKey, strComment, "", common.IKConstant)
		}
		return true
	}

	if completeVar.OnelyParamQuotesFlag {
		return true
	}

	return false
}

func (a *AllProject) mergeTwoExistMap(symbol *common.Symbol, fristStr string, fristMap map[string]string,
	secondStr string, secondMap map[string]string) (mergeStr string) {
	if len(fristMap) == 0 {
		return secondStr
	}

	if len(secondMap) == 0 {
		return fristStr
	}

	// 1) 先判断是否有注解类型
	if symbol.AnnotateType != nil {
		strType := annotateast.TypeConvertStr(symbol.AnnotateType)
		mergeStr = strType + " = {\n"
	} else {
		mergeStr = "table = {\n"
	}

	for oneStr, oneValue := range fristMap {
		_, ok := secondMap[oneStr]
		if !ok {
			secondMap[oneStr] = oneValue
		}
	}

	traverseMapInStringOrder(secondMap, func(key string, value string) {
		mergeStr = mergeStr + "\t" + value + "\n"
	})

	mergeStr = mergeStr + "}"
	return mergeStr
}

// GetStrComment 对注释进行一些处理, 提取出注解的特殊的markdown格式
func GetStrComment(strComment string) (str string) {
	if strComment == "" {
		return strComment
	}

	// 之前lua格式的doc
	preLuaStr := ""

	// 删除注释一些空格和其他的多余的格式
	splitStrArr := strings.Split(strComment, "\n")
	annotatFlag := false
	for index, oneStr := range splitStrArr {
		oneStr = strings.TrimLeft(oneStr, " ")
		oneStr = strings.TrimPrefix(oneStr, "-*")
		//oneStr = strings.TrimPrefix(oneStr, "*")
		oneStr = strings.TrimPrefix(oneStr, "-")
		oneStr = strings.TrimLeft(oneStr, " ")

		annotatFlag = false
		if strings.HasPrefix(oneStr, "@class") || strings.HasPrefix(oneStr, "@alias") || strings.HasPrefix(oneStr, "@param") ||
			strings.HasPrefix(oneStr, "@return") || strings.HasPrefix(oneStr, "@type") || strings.HasPrefix(oneStr, "@overload") ||
			strings.HasPrefix(oneStr, "@generic") || strings.HasPrefix(oneStr, "@vararg") || strings.HasPrefix(oneStr, "@version") {
			annotatFlag = true
		}

		if annotatFlag {
			if preLuaStr == "" {
				// 如果之前没有注解，拼接lua
				preLuaStr = fmt.Sprintf("\n```%s\n%s\n", "lua", oneStr)
			} else {
				// 如果之前有注解，拼接这次
				preLuaStr = preLuaStr + oneStr + "\n"
			}

			if index == len(splitStrArr)-1 {
				if preLuaStr != "" {
					str = str + preLuaStr + "```"
				}
			}
		} else {
			if preLuaStr == "" {
				str = str + "  \n" + oneStr
			} else {
				str = str + preLuaStr + "```" + "  \n" + oneStr
				preLuaStr = ""
			}
		}
	}

	return str
}

func (a *AllProject) expandNodefineMapComplete(luaInFile string, strFind string, comParam *CommonFuncParam,
	completeVar *common.CompleteVarStruct, varInfo *common.VarInfo) {
	for key := range varInfo.ExpandStrMap {
		key = strFind + "." + key
		remainVec := matchVecsExpandStrMap(completeVar.StrVec, completeVar.IsFuncVec, key)
		if len(remainVec) == 0 {
			continue
		}

		strOne := remainVec[0]
		if strOne == "" || !common.JudgeSimpleStr(strOne) {
			continue
		}

		if a.completeCache.ExistStr(strOne) || a.completeCache.IsExcludeStr(strOne) {
			continue
		}

		a.completeCache.InsertCompleteExpand(strOne, "", "", common.IKVariable, varInfo)
	}
}

// 判断是否为注解table中的key值补全，例如:
// ---@type oneTable
// local one = {
//	  b-- 此时输入b时候，代码补全one的成员
//}
func (a *AllProject) needAnnotateTableFieldRepair(comParam *CommonFuncParam, completeVar *common.CompleteVarStruct) {
	if len(completeVar.StrVec) != 1 || completeVar.LastEmptyFlag {
		return
	}

	strName := completeVar.StrVec[0]
	// 1) 优先判断局部变量
	firstStr, onVar := comParam.scope.GetTableKeyVar(strName, completeVar.PosLine+1, completeVar.PosCh)
	if firstStr == "" {
		// 2) 局部变量没有找到，查找全局变量
		firstStr, onVar = comParam.fileResult.GetGlobalVarTableStrKey(strName, completeVar.PosLine+1, completeVar.PosCh)
	}

	if firstStr != "" {
		symbol := a.createAnnotateSymbol(comParam.fileResult.Name, onVar)
		if symbol.AnnotateType == nil {
			// 变量没有关联注解类型，返回
			return
		}

		completeVar.StrVec[0] = firstStr
		completeVar.LastEmptyFlag = true
		return
	}
}
func getParamVarinfoType(oneVar *common.VarInfo) string {
	if oneVar == nil {
		return "any"
	}

	str := oneVar.GetVarTypeDetail()
	strSplit := strings.Split(str, " ")
	oneStr := strSplit[0]
	if oneStr != "any" {
		return oneStr
	}

	if len(oneVar.SubMaps) > 0 || len(oneVar.ExpandStrMap) > 0 {
		return "table"
	}

	return "any"
}

// 获取函数一个return值的返回类型
func (a *AllProject) getOneFuncReturnStr(fileName string, oneReturn common.ReturnItem) (strType string) {
	strType = common.GetLuaTypeString(common.GetExpType(oneReturn.ReturnExp), oneReturn.ReturnExp)
	if oneReturn.ReturnExp == nil {
		return
	}

	loc := common.GetExpLoc(oneReturn.ReturnExp)
	comParam := a.getCommFunc(fileName, loc.StartLine, loc.StartColumn)
	findExpList := &[]common.FindExpFile{}
	symbol := a.FindVarReferSymbol(fileName, oneReturn.ReturnExp, comParam, findExpList, 1)
	if symbol == nil {
		return
	}

	if symbol.AnnotateType != nil {
		strType = annotateast.TypeConvertStr(symbol.AnnotateType)
	} else {
		if symbol.VarInfo != nil {
			strTmp := getParamVarinfoType(symbol.VarInfo)
			if strTmp != "any" {
				strType = strTmp
			}
		}
	}

	return strType
}

// 获取函数的完整展示信息
// paramTipFlag 表示是否提示函数的参数
// colonFlag 如果是冒号语法，有时候需要忽略掉self
// returnFlag 是否需要获取函数的返回值类型
// returnMultiline 多个返回值时，是否需要多行显示
func (a *AllProject) getFuncShowStr(varInfo *common.VarInfo, funcName string, paramTipFlag, colonFlag, returnFlag, returnMultiline bool) (str string) {
	if varInfo == nil || varInfo.ReferFunc == nil {
		return
	}

	if !paramTipFlag {
		return funcName
	}

	fun := varInfo.ReferFunc
	inLuaFile := fun.FileName
	lastLine := fun.Loc.StartLine
	annotateParamInfo := a.GetFuncParamInfo(inLuaFile, lastLine-1)

	// 1) 获取函数的参数
	funcName += "("
	preFlag := false
	for index, oneParam := range fun.ParamList {
		if colonFlag && fun.IsColon && index == 0 {
			continue
		}

		if preFlag {
			funcName += ", "
		}

		funcName += oneParam

		paramShortStr, annType := a.getAnnotateFuncParamDocument(oneParam, annotateParamInfo, inLuaFile, lastLine-1)
		if paramShortStr != "" {
			funcName += ": " + annotateast.TypeConvertStr(annType)
		} else {
			// 获取参数变量关联的varinfo
			oneVar := fun.MainScope.GetParamVarInfo(oneParam)
			funcName += ": " + getParamVarinfoType(oneVar)
		}

		preFlag = true
	}

	if fun.IsVararg {
		if preFlag {
			funcName += ", "
		}
		funcName += "..."
	}

	funcName += ")"

	// 如果不需要返回类型，直接返回
	if !returnFlag {
		return funcName
	}

	// 2 获取返回值
	// 2.1) 先获取注解的参数
	oldSymbol := common.GetDefaultSymbol(varInfo.FileName, varInfo)
	flag, _, typeList, commentList := a.getFuncReturnAnnotateTypeList(oldSymbol)
	var resultList []string = []string{}

	if flag {
		for i, oneType := range typeList {
			oneStr := annotateast.TypeConvertStr(oneType)
			if returnMultiline && len(commentList) > i && commentList[i] != "" {
				oneStr += "  -- " + commentList[i]
			}
			resultList = append(resultList, oneStr)
		}
	}

	if len(fun.ReturnVecs) == 0 {

		for i, oneStr := range resultList {
			if returnMultiline {
				funcName = fmt.Sprintf("%s\n  ->%d. %s", funcName, i+1, oneStr)
			} else {
				if i == 0 {
					funcName = funcName + ": " + oneStr
				} else {
					funcName = funcName + ", " + oneStr
				}
			}
		}

		if returnMultiline && len(resultList) > 0 {
			funcName += "\n"
		}

		return funcName
	}

	oldLen := len(resultList)

	for _, oneReturnInfo := range fun.ReturnVecs {
		for i, oneReturn := range oneReturnInfo.ReturnVarVec {
			if i < oldLen {
				continue
			}

			strType := a.getOneFuncReturnStr(varInfo.FileName, oneReturn)
			if len(resultList) > i {
				if strType != "any" {
					resultList[i] = strType
				}
			} else {
				resultList = append(resultList, strType)
			}
		}
	}

	for i, oneStr := range resultList {
		if returnMultiline {
			funcName = fmt.Sprintf("%s\n  ->%d. %s", funcName, i+1, oneStr)
		} else {
			if i == 0 {
				funcName = funcName + ": " + oneStr
			} else {
				funcName = funcName + ", " + oneStr
			}
		}
	}
	if returnMultiline && len(resultList) > 0 {
		funcName += "\n"
	}
	return funcName
}

// getVarCompleteExt 获取变量关联的所有子成员信息，用于代码补全
// excludeMap 表示这次因为冒号语法剔除掉的字符串
// colonFlag 表示是否获取冒号成员
func (a *AllProject) getVarCompleteExt(fileName string, varInfo *common.VarInfo, colonFlag bool) {
	if varInfo == nil || varInfo.SubMaps == nil {
		return
	}

	for strName, oneVar := range varInfo.SubMaps {
		if colonFlag && (oneVar.ReferFunc == nil || !oneVar.ReferFunc.IsColon) {
			a.completeCache.InsertExcludeStr(strName)
			// 只获取冒号成员
			continue
		}

		// 对.的用法特殊提，能够提示：函数
		getJugdeColon := common.GConfig.GetJudgeColonFlag()
		if getJugdeColon != 0 && !colonFlag && (oneVar.ReferFunc != nil && oneVar.ReferFunc.IsColon) {
			a.completeCache.InsertExcludeStr(strName)
			continue
		}

		if a.completeCache.ExistStr(strName) {
			continue
		}

		if getJugdeColon == 0 && !colonFlag && (oneVar.ReferFunc != nil && oneVar.ReferFunc.IsColon) {
			a.completeCache.InsertCompleteVarInclude(fileName, strName, oneVar)
		} else {
			a.completeCache.InsertCompleteVar(fileName, strName, oneVar)
		}
	}

	// 获取扩展的成员
	for key := range varInfo.ExpandStrMap {
		strRemainList := strings.Split(key, ".")
		strOne := strRemainList[0]
		if !common.JudgeSimpleStr(strOne) {
			continue
		}

		if a.completeCache.ExistStr(strOne) || a.completeCache.IsExcludeStr(strOne) {
			continue
		}

		a.completeCache.InsertCompleteExpand(strOne, "", "", common.IKVariable, varInfo)
	}
}
