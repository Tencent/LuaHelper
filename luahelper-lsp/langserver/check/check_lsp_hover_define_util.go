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
		existMap[strName] = strName + ": " + strFiled + ","
	}

	getVarInfoMapStr(oneClass.RelateVar, existMap)
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
			if needReplaceMapStr(oldStr, strType) {
				existMap[oneStr] = newStr
			}
		} else {
			existMap[oneStr] = newStr
		}
	}
	if len(existMap) == 0 {
		return
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

func getVarInfoMapStr(varInfo *common.VarInfo, existMap map[string]string) {
	if varInfo == nil {
		return
	}

	for key, value := range varInfo.SubMaps {
		strValueType := value.GetVarTypeDetail()
		if value.ReferFunc != nil {
			strValueType = value.ReferFunc.GetFuncCompleteStr("function", true, false)
		}
		newStr := key + ": " + strValueType + ","
		if oldStr, ok := existMap[key]; ok {
			if needReplaceMapStr(oldStr, strValueType) {
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
			if needReplaceMapStr(oldStr, strType) {
				existMap[oneStr] = newStr
			}
		} else {
			existMap[oneStr] = newStr
		}
	}
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

	getVarInfoMapStr(symbol.VarInfo, existMap)
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
			return
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
