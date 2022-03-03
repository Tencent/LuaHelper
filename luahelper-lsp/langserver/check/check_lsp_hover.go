package check

import (
	"fmt"
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/annotation/annotatelexer"
	"luahelper-lsp/langserver/check/annotation/annotateparser"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"strings"
)

// AnnotateTypeHover 注解类型代码补全
func (a *AllProject) AnnotateTypeHover(strFile, strLine, strWord string, line, col int) (strLabel, strHover, strLuaFile string) {
	l := annotatelexer.CreateAnnotateLexer(&strLine, 0, 0)

	// 判断这行内容是否以-@开头，是否合法
	if !l.CheckHeardValid() {
		return
	}

	// 后面的内容进行词法解析
	annotateState, parseErr := annotateparser.ParserLine(l)
	_, flag := annotateState.(*annotateast.AnnotateNotValidState)
	// 1) 判断是否为解析有效
	if flag || parseErr.ErrType != annotatelexer.AErrorOk {
		return
	}

	dirManager := common.GConfig.GetDirManager()

	// 2) 遍历位置信息
	typeStr, noticeStr, commentStr := annotateast.GetStateLocInfo(annotateState, col)

	if typeStr == "" && noticeStr == "alias name" {
		// 判断是否alias多个候选词
		// ---@alias exitcode2 '"exit"' | '"signal"'
		strCandidate := a.getAliasMultiCandidate(strWord, strFile, line)
		noticeStr = noticeStr + strCandidate
		strLabel = strWord + " : " + noticeStr
		strHover = commentStr
		return
	}

	if typeStr == "" && noticeStr == "class name" {
		typeStr = strWord
	}

	if typeStr != "" {
		createType := a.getAnnotateStrTypeInfo(typeStr, strFile, line)
		if createType == nil {
			return "", "not find annotate type", ""
		}

		if createType.AliasInfo != nil {
			strComment := createType.AliasInfo.AliasState.Comment
			strLabel = "alias " + createType.AliasInfo.AliasState.Name
			strCandidate := a.getAliasMultiCandidate(typeStr, strFile, line)
			strLabel = strLabel + strCandidate
			if strComment != "" {
				strHover = strComment
			}

			strLuaFile = dirManager.RemovePathDirPre(createType.AliasInfo.LuaFile)
			return
		}

		if createType.ClassInfo != nil {
			//str := a.expandTableHover(symbol)
			strLabel = "class " + typeStr

			if len(createType.ClassInfo.ClassState.ParentNameList) > 0 {
				strLabel = strLabel + " : " + strings.Join(createType.ClassInfo.ClassState.ParentNameList, " , ")
			} else {
				strLabel = "class " + typeStr + a.getClassFieldStr(createType.ClassInfo)
			}

			strComment := createType.ClassInfo.ClassState.Comment
			if strComment != "" {
				strHover = strComment
			}

			strLuaFile = dirManager.RemovePathDirPre(createType.ClassInfo.LuaFile)
			return
		}
	}

	if noticeStr == "comment info" {
		strHover = commentStr + " : " + noticeStr
		return
	}

	if noticeStr != "" {
		strLabel = strWord + " : " + noticeStr
		strHover = commentStr
		return
	}

	return
}

// GetLspHoverVarStr 提示信息hover
func (a *AllProject) GetLspHoverVarStr(strFile string, varStruct *common.DefineVarStruct) (lableStr, docStr, luaFileStr string) {
	symbol, findList := a.FindVarDefine(strFile, varStruct)

	if symbol == nil && len(varStruct.StrVec) == 1 {
		// 1) 判断是否为系统的函数提示
		if flag, str1, str2 := judgetSystemModuleOrFuncHover(varStruct.StrVec[0]); flag {
			lableStr = str1
			docStr = str2
			docStr = strings.ReplaceAll(docStr, "\n", "  \n")
			return
		}
	}

	if symbol == nil && len(varStruct.StrVec) == 2 {
		// 2) 判断是否为系统的模块函数提示
		if flag, str1, str2 := judgetSystemModuleMemHover(varStruct.StrVec[0], varStruct.StrVec[1]); flag {
			lableStr = str1
			docStr = str2
			docStr = strings.ReplaceAll(docStr, "\n", "  \n")
			return
		}
	}

	if symbol == nil && len(findList) == 0 {
		// 没有找到变量的定义，查找当前文件的：NodefineMaps
		symbol = a.getNodefineMapVar(strFile, varStruct)
	}

	if symbol != nil && len(findList) == 0 {
		lableStr = getVarInfoExpandStrHover(symbol.VarInfo, varStruct.StrVec, varStruct.IsFuncVec, varStruct.Str)
		if lableStr != "" {
			return
		}
	}

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
		strType, strLabel1, strDoc1, strPre, flag := a.getVarHoverInfo(strFile, oneSymbol, varStruct)
		if !preFirstFlag {
			strPreFirst = strPre
			preFirstFlag = true
		}

		if flag {
			lableStr = strLabel1
			lableStr = strPreFirst + strLabel1
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
	lableStr = strPreFirst + strLastBefore
	docStr = strOneComment
	return
}

func (a *AllProject) getNodefineMapVar(strFile string, varStruct *common.DefineVarStruct) (symbol *common.Symbol) {
	// 1）先查找该文件是否存在
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("getNodefineMapVar error, not find file=%s", strFile)
		return nil
	}
	fileResult := fileStruct.FileResult
	if fileResult == nil {
		log.Error("getNodefineMapVar error, not find file=%s", strFile)
		return nil
	}

	splitArray := varStruct.StrVec

	if splitArray[0] == "_G" {
		splitArray = splitArray[1:]
	}
	if len(splitArray) == 0 {
		return
	}
	strName := splitArray[0]

	if findVar, ok := fileResult.NodefineMaps[strName]; ok {
		symbol = common.GetDefaultSymbol(findVar.FileName, findVar)
	}
	return
}

func isDefaultType(str string) bool {
	if str == "number" || str == "any" || str == "string" || str == "boolean" || str == "nil" || str == "thread" ||
		str == "userdata" || str == "lightuserdata" || str == "integer" || str == "void" {
		return true
	}

	return false
}

func needReplaceMapStr(oldStr string, strValueType string) bool {
	if strValueType == "any" {
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
	if len(compStruct.StrVec) <= 1 {
		return
	}

	tempVec := compStruct.StrVec
	tempVec = append(tempVec, item.Label)
	str := getVarInfoExpandStrHover(item.ExpandVarInfo, tempVec, compStruct.IsFuncVec, "")
	if str != "" {
		item.Detail = str
	}

	return luaFileStr
}

func isAllClassDefault(classList []*common.OneClassInfo) bool {
	for _, oneClass := range classList {
		if oneClass.ClassState != nil && !isDefaultType(oneClass.ClassState.Name) {
			return false
		}
	}

	return true
}

// 获取函数的完整展示信息
// paramTipFlag 表示是否提示函数的参数
// colonFlag 如果是冒号语法，有时候需要忽略掉self
// returnFlag 是否需要获取函数的返回值类型
func (a *AllProject) getFuncShowStr(varInfo *common.VarInfo, funcName string, paramTipFlag, colonFlag, returnFlag bool) (str string) {
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
	flag, _, typeList := a.getFuncReturnAnnotateTypeList(oldSymbol)
	var resultList []string = []string{}

	if flag {
		for _, oneType := range typeList {
			oneStr := annotateast.TypeConvertStr(oneType)
			resultList = append(resultList, oneStr)
		}
	}

	if len(fun.ReturnVecs) == 0 {
		for i, oneStr := range resultList {
			funcName = fmt.Sprintf("%s\n  ->%d. %s", funcName, i+1, oneStr)
		}

		if len(resultList) > 0 {
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
		funcName = fmt.Sprintf("%s\n  ->%d. %s", funcName, i+1, oneStr)
	}
	if len(resultList) > 0 {
		funcName += "\n"
	}
	return funcName
}

// 获取函数一个return值的返回类型
func (a *AllProject) getOneFuncReturnStr(fileName string, oneReturn common.ReturnItem) (strType string) {
	strType = common.GetLuaTypeString(common.GetExpType(oneReturn.ReturnExp), oneReturn.ReturnExp)
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

func (a *AllProject) getVarHoverInfo(strFile string, symbol *common.Symbol, varStruct *common.DefineVarStruct) (strType string,
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
		if strType == "table" || len(symbol.VarInfo.SubMaps) > 0 || len(symbol.VarInfo.ExpandStrMap) > 0 {
			strType, _ = a.expandTableHover(symbol)
		}

		referFunc := symbol.VarInfo.ReferFunc
		if referFunc != nil {
			if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
				strPre = "local "
			}
			strFunc := a.getFuncShowStr(symbol.VarInfo, varStruct.StrVec[len(varStruct.StrVec)-1], true, false, true)
			strType = "function " + strFunc
		}
	}

	strLabel = strings.Join(varStruct.StrVec, ".")
	if symbol.VarInfo.ReferFunc == nil {
		strLabel = strLabel + " : " + strType
		if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
			strPre = "local "
		}
	} else {
		strLabel = strType
	}

	strDoc = a.GetLineComment(symbol.FileName, symbol.VarInfo.Loc.EndLine)
	strDoc = GetStrComment(strDoc)
	return
}

// 判断变量是否直接为系统模块或函数的hover
func judgetSystemModuleOrFuncHover(strName string) (flag bool, labStr, docStr string) {
	if oneSystemTip, ok := common.GConfig.SystemTipsMap[strName]; ok {
		flag = true
		labStr = oneSystemTip.Detail
		docStr = oneSystemTip.Documentation
		return
	}

	if oneMouleInfo, ok := common.GConfig.SystemModuleTipsMap[strName]; ok {
		flag = true
		labStr = oneMouleInfo.Detail
		docStr = oneMouleInfo.Documentation
		return
	}

	return
}

// 判断是否为系统的模块中的成员hover
func judgetSystemModuleMemHover(strName string, strKey string) (flag bool, lableStr, docStr string) {
	if oneMouleInfo, ok := common.GConfig.SystemModuleTipsMap[strName]; ok {
		flag = true
		if oneSystemTip, ok1 := oneMouleInfo.ModuleFuncMap[strKey]; ok1 {
			lableStr = oneSystemTip.Detail
			docStr = oneSystemTip.Documentation
		}
		return
	}

	return
}
