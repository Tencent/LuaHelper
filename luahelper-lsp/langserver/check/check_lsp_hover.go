package check

import (
	"fmt"
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"sort"
	"strings"
)

// GetLspHoverVarStr 提示信息hover
func (a *AllProject) GetLspHoverVarStr(strFile string, varStruct *common.DefineVarStruct) (lableStr, docStr, luaFileStr string) {
	symbol, findList := a.FindVarDefine(strFile, varStruct)
	// 原生的变量没有找到, 直接返回
	if symbol == nil || len(findList) == 0 {
		lableStr = varStruct.Str + " : any"
		return
	}

	lastSymbol := findList[len(findList)-1]
	if lastSymbol == nil {
		return
	}

	strOneComment := ""
	strLastBefore := ""
	strLastType := ""
	strPreFirst := ""
	preFirstFlag := false
	dirManager := common.GConfig.GetDirManager()

	for _, oneSymbol := range findList {
		strType, strLabel1, strDoc1, strPre, flag := a.getVarHoverInfo(oneSymbol, varStruct)
		if !preFirstFlag {
			strPreFirst = strPre
			preFirstFlag = true
		}

		if flag {
			lableStr = strLabel1
			if strPreFirst == "_G." && strings.HasPrefix(strLabel1, "function ") {
				//  修复特殊的 _G.function a() 这样的显示
				lableStr = "[_G] " + strLabel1
			} else {
				if strPreFirst == "local " && strings.HasPrefix(strLabel1, "function _G.") {
					lableStr = strPreFirst + "function " + strings.TrimPrefix(strLabel1, "function _G.")
				} else {
					if strings.HasPrefix(strLabel1, "_G.") {
						lableStr = strLabel1
					} else {
						lableStr = strPreFirst + strLabel1
					}
				}
			}

			docStr = strDoc1
			luaFileStr = dirManager.RemovePathDirPre(oneSymbol.FileName)
			return
		}

		if strLastType == "" || strLastType == "any" {
			strLastType = strType
			strLastBefore = strLabel1
			luaFileStr = dirManager.RemovePathDirPre(oneSymbol.FileName)
		}

		if strOneComment == "" && strDoc1 != "" {
			strOneComment = strDoc1
		}
	}

	lableStr = strLastBefore
	if strPreFirst == "_G." && strings.HasPrefix(strLastBefore, "function ") {
		//  修复特殊的 _G.function a() 这样的显示
		lableStr = "[_G] " + strLastBefore
	} else {
		if strPreFirst == "local " && strings.HasPrefix(strLastBefore, "function _G.") {
			lableStr = strPreFirst + "function " + strings.TrimPrefix(strLastBefore, "function _G.")
		} else {
			if strings.HasPrefix(strLastBefore, "_G.") {
				lableStr = strLastBefore
			} else {
				lableStr = strPreFirst + strLastBefore
			}
		}
	}

	docStr = strOneComment
	return
}

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

	if oneClass.RelateVar != nil {
		for key, value := range oneClass.RelateVar.SubMaps {
			if oldStr, ok := existMap[key]; ok {
				if !needReplaceMapStr(oldStr) {
					continue
				}
			}

			strValueType := value.GetVarTypeDetail()
			if value.ReferFunc != nil {
				strValueType = value.ReferFunc.GetFuncCompleteStr("function", true, false)
			}
			existMap[key] = key + ": " + strValueType + ","
		}
	}
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

// 代码补全时，有注解类型的字段提示
func (a *AllProject) completeAnnotatTypeStr(astType annotateast.Type, fileName string, line int) (str string) {
	str = annotateast.TypeConvertStr(astType)
	classList := a.getAllNormalAnnotateClass(astType, fileName, line)
	if len(classList) == 0 || str == "number" || str == "any" || str == "string" || str == "boolean" {
		// 没有找到相应的class成员，直接返回
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
	return str
}


func needReplaceMapStr(oldStr string) bool {
	if strings.Contains(oldStr, ": number") && !strings.Contains(oldStr, ": number = ") {
		return true
	}

	if strings.Contains(oldStr, ": string") && !strings.Contains(oldStr, ": string = ") {
		return true
	}

	if strings.Contains(oldStr, ": boolean") && !strings.Contains(oldStr, ": boolean = ") {
		return true
	}

	return false
}

// hover 的时候是指向一个table，展开这个table的内容
func (a *AllProject) expandTableHover(symbol *common.Symbol) (str string, existMap map[string]string) {
	// 为已经存在的map，防止重复
	existMap = map[string]string{}

	// 1) 先判断是否有注解类型
	if symbol.AnnotateType != nil {
		strType := annotateast.TypeConvertStr(symbol.AnnotateType)

		classList := a.getAllNormalAnnotateClass(symbol.AnnotateType, symbol.FileName, symbol.GetLine())
		if len(classList) == 0 || strType == "number" || strType == "any" || strType == "string" {
			// 没有找到相应的class成员，直接返回
			return strType, existMap
		}

		str = strType + " = {\n"
		for _, oneClass := range classList {
			a.convertClassInfoToHovers(oneClass, existMap)
		}
	} else {
		str = "table = {\n"
		if symbol.VarInfo == nil || len(symbol.VarInfo.SubMaps) == 0 {
			return "table = { }", existMap
		}
	}

	if symbol.VarInfo != nil {
		for key, value := range symbol.VarInfo.SubMaps {
			if oldStr, ok := existMap[key]; ok {
				if !needReplaceMapStr(oldStr) {
					continue
				}
			}

			strValueType := value.GetVarTypeDetail()
			if value.ReferFunc != nil {
				strValueType = value.ReferFunc.GetFuncCompleteStr("function", true, false)
			}
			existMap[key] = key + ": " + strValueType + ","
		}
	}

	traverseMapInStringOrder(existMap, func(key string, value string) {
		str = str + "\t" + value + "\n"
	})

	str = str + "}"
	return str, existMap
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

func (a *AllProject) getVarHoverInfo(symbol *common.Symbol, varStruct *common.DefineVarStruct) (strType string,
	strLabel, strDoc, strPre string, findFlag bool) {
	// 1) 首先提取注解类型
	if symbol.AnnotateType != nil {
		// 注解类型尝试推导扩展class的field成员信息
		str, _ := a.expandTableHover(symbol)
		strLabel = varStruct.Str + " : " + str

		if symbol.VarInfo != nil {
			if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
				strPre = "local "
			}

			if symbol.VarInfo.ExtraGlobal != nil && symbol.VarInfo.ExtraGlobal.GFlag {
				strPre = "_G."
			}
		}

		strDoc = symbol.AnnotateComment
		strDoc = strings.ReplaceAll(strDoc, "\n", "  \n")
		findFlag = true
		return
	}

	// 2) 获取变量推动的类型，首先判断变量类型是否存在
	if symbol.VarInfo == nil {
		return
	}

	strType = ""
	if symbol.VarInfo.ReferInfo != nil {
		strType = symbol.VarInfo.ReferInfo.GetReferComment()
	} else {
		strType = symbol.VarInfo.GetVarTypeDetail()
		// 判断是否指向的一个table，如果是展开table的具体内容
		if strType == "table" || len(symbol.VarInfo.SubMaps) > 0 {
			strType, _ = a.expandTableHover(symbol)
		}

		referFunc := symbol.VarInfo.ReferFunc
		if referFunc != nil {
			if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
				strPre = "local "
			}

			if symbol.VarInfo.ExtraGlobal != nil && symbol.VarInfo.ExtraGlobal.GFlag {
				strType = "function _G." + referFunc.GetFuncCompleteStr(varStruct.StrVec[len(varStruct.StrVec)-1], true, false)
			} else {
				strType = "function " + referFunc.GetFuncCompleteStr(varStruct.StrVec[len(varStruct.StrVec)-1], true, false)
			}
		}
	}

	strLabel = strings.Join(varStruct.StrVec, ".")
	if symbol.VarInfo.ReferFunc == nil {
		strLabel = strLabel + " : " + strType
		if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
			strPre = "local "
		}

		if symbol.VarInfo.ExtraGlobal != nil && symbol.VarInfo.ExtraGlobal.GFlag {
			strPre = "_G."
		}
	} else {
		strLabel = strType
	}

	strDoc = a.GetLineComment(symbol.FileName, symbol.VarInfo.Loc.EndLine)
	strDoc = GetStrComment(strDoc)
	return
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

	// str = str + "\n\r" + oneStr

	return str
}
