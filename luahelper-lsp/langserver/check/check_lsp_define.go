package check

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/annotation/annotatelexer"
	"luahelper-lsp/langserver/check/annotation/annotateparser"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
)

// 查找全局的变量信息
func (a *AllProject) findGlobalVarDefineInfo(comParam *CommonFuncParam, strName string,
	strProPre string, gFlag bool) *common.VarInfo {
	if ok, findVar := comParam.fileResult.FindGlobalVarInfo(strName, gFlag, strProPre); ok {
		return findVar
	}

	// 3)向工程的globalMaps中查找变量
	if comParam.secondProject != nil {
		if ok, findVar := comParam.secondProject.FindGlobalGInfo(strName, results.CheckTermFirst, strProPre); ok {
			return findVar
		}
	}

	if comParam.thirdStruct != nil {
		// 向工程的第一阶段全局_G符号表中查找
		if ok, findVar := comParam.thirdStruct.FindThirdGlobalGInfo(gFlag, strName, strProPre); ok {
			return findVar
		}
	}

	return nil
}

// findFile 表示找到的定义在的lua文件
// findLoc 为定义的地方
func (a *AllProject) findVarDefineInfo(comParam *CommonFuncParam, strName string, strProPre string) (
	findVar *common.VarInfo) {
	// 1) 局部变量找到了该变量
	if locVar, ok := comParam.scope.FindLocVar(strName, comParam.loc); ok {
		return locVar
	}

	// 2) 局部变量没有找到，查找全局变量
	return a.findGlobalVarDefineInfo(comParam, strName, strProPre, false)
}

// 查找所有的协议前缀符号，匹配下
func (a *AllProject) findProtocolDefine(strContent string, secondProject *results.SingleProjectResult) (symbol *common.Symbol) {
	// 查找指定的协议符号
	for _, fileStruct := range a.fileStructMap {
		if !fileStruct.IsCommonFile {
			continue
		}

		if fileStruct.GetFileHandleErr() != results.FileHandleOk {
			continue
		}

		fileResult := fileStruct.FileResult
		if fileResult == nil {
			continue
		}

		proVarInfo, ok := fileResult.ProtocolMaps[strContent]
		if !ok {
			continue
		}

		oneVar := proVarInfo
		for {
			// 项目组特殊的c2s屏蔽掉
			if oneVar.ExtraGlobal.StrProPre != "c2s" {
				symbol = &common.Symbol{
					FileName: fileResult.Name,
					VarInfo:  oneVar,
				}

				return symbol
			}

			if oneVar.ExtraGlobal.Prev == nil {
				break
			}

			oneVar = oneVar.ExtraGlobal.Prev
		}
	}

	return
}

// FindOpenFileDefine 查找打开一个文件，直接跳转到打开的文件
// strFile 为原文件内
// strOpenFile 为直接打开某一文件名
func (a *AllProject) FindOpenFileDefine(strFile string, strOpenFile string) (defineVecs []DefineStruct) {
	// 0) 先查找该文件是否存在
	fileStruct, _ := a.GetCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("FindVarDefine error, not find file=%s", strFile)
		return defineVecs
	}

	// 1) 文件匹配的完整路径
	strOpenFile = common.GetBestMatchReferFile(strFile, strOpenFile, a.allFilesMap)

	// 2) 判断是否为直接打开某一个文件
	if strOpenFile == "" {
		return defineVecs
	}

	fileOpenStruct, _ := a.GetFirstFileStuct(strOpenFile)
	if fileOpenStruct == nil {
		return defineVecs
	}
	defineVecs = append(defineVecs, DefineStruct{
		StrFile: strOpenFile,
		Loc: lexer.Location{
			StartLine:   1,
			StartColumn: 0,
			EndLine:     1,
			EndColumn:   1,
		},
	})

	return defineVecs
}

// 返回文件数最多的第二轮工程
// 如果strFile文件属于多个第二阶段指针，返回第二阶段工程项目最多的文件
func (a *AllProject) findMaxSecondProject(strFile string) (secondProject *results.SingleProjectResult) {
	// 如果该文件属于多个第二阶段指针，返回第二阶段工程项目最多的文件
	maxFileNum := 0
	for strEntry, tmpSecond := range a.analysisSecondMap {
		if _, ok := tmpSecond.AllFiles[strFile]; ok && len(tmpSecond.AllFiles) > maxFileNum {
			secondProject = tmpSecond
			maxFileNum = len(tmpSecond.AllFiles)
			log.Debug("find, belong secondProject file=%s, entryFile=%s, filesNum=%d, entry=%s",
				strFile, tmpSecond.EntryFile, maxFileNum, strEntry)
		}
	}

	return secondProject
}

// 变量查找引用时候，跟踪到变量import或require引入的关系
func (a *AllProject) findLspReferenceVarDefine(comParam *CommonFuncParam, varStruct *common.DefineVarStruct) (string, *common.VarInfo) {
	_, findLocVar := a.findOldDefineInfo(comParam, varStruct)
	if findLocVar == nil {
		// 直接返回
		return "", nil
	}

	findPreFile := findLocVar.FileName
	referInfo := findLocVar.ReferInfo

	// 判断是否有引用其他的信息
	if referInfo == nil || len(varStruct.StrVec) <= 1 {
		return findPreFile, findLocVar
	}

	// import引入一个文件
	strModeMem := varStruct.StrVec[1]
	// 如果是引用，前面的一个切除掉
	varStruct.StrVec = varStruct.StrVec[1:]
	// 查找对应的引用信息
	referFile := a.GetFirstReferFileResult(referInfo)
	if referFile == nil {
		return findPreFile, nil
	}

	referSubType := a.GetReferFrameType(referInfo)
	if referSubType == common.RtypeImport {
		if ok, findVar := referFile.FindGlobalVarInfo(strModeMem, false, ""); ok {
			return findVar.FileName, findVar
		}

		return findPreFile, nil
	}

	if referSubType != common.RtypeRequire {
		log.Error("findLspDefineInfo ReferType err=%s", referInfo.ReferTypeStr)
		return findPreFile, nil
	}

	// 这里为require一个文件，获取这个文件的返回值
	find, returnExp := referFile.MainFunc.GetOneReturnExp()
	if !find {
		return findPreFile, nil
	}

	findExpList := []common.FindExpFile{}
	locInfoFile := a.FindVarReferSymbol(referFile.Name, returnExp, comParam, &findExpList, 1)
	if locInfoFile == nil {
		log.Error("findLspDefineInfo find val err=%s", referInfo.ReferTypeStr)
		return findPreFile, nil
	}

	// todo 后面修改，当查找引用的时候，然后引用变量的信息
	returnExpStr := common.GetExpName(returnExp)
	returnExpStr = common.GetSimpleValue(returnExpStr)
	if returnExpStr != "" {
		varStruct.StrVec = append([]string{returnExpStr}, varStruct.StrVec...)
	}

	subVar := common.GetVarSubGlobalVar(locInfoFile.VarInfo, strModeMem)
	if subVar == nil {
		return findPreFile, nil
	}
	return referFile.Name, locInfoFile.VarInfo
}

// FindReferenceVarDefine 查找引用时候变量的定义
func (a *AllProject) FindReferenceVarDefine(strFile string, varStruct *common.DefineVarStruct) (lastDefine lexer.Location,
	oldSymbol *common.Symbol, isWhole bool) {
	comParam := a.getVarCommonFuncParam(strFile, varStruct)
	if comParam == nil {
		return
	}

	luaInFile, findVar := a.findLspReferenceVarDefine(comParam, varStruct)
	oldSymbol = &common.Symbol{
		FileName: luaInFile,
		VarInfo:  findVar,
	}

	isWhole = true
	if findVar != nil {
		lastDefine = findVar.Loc
		oldVar := findVar
		for i := 1; i < len(varStruct.StrVec); i++ {
			strTemp := varStruct.StrVec[i]
			if oldVar.SubMaps == nil {
				isWhole = false
				break
			}

			if subVarMem, ok := oldVar.SubMaps[strTemp]; ok {
				lastDefine = subVarMem.Loc
				oldVar = subVarMem
			}
		}
	}
	return lastDefine, oldSymbol, isWhole
}

// FindVarDefineInfo 客户端请求变量的定义的接口，尽量返回变量的最深处的引用定义关系
func (a *AllProject) FindVarDefineInfo(strFile string, varStruct *common.DefineVarStruct) (defineVecs []DefineStruct) {
	for {
		oldSymbol, symList := a.FindVarDefine(strFile, varStruct)
		// 原生的变量没有找到, 直接返回
		if oldSymbol == nil {
			return
		}

		if len(symList) == 0 {
			// 没有追踪到，切短下次，继续找
			subLen := len(varStruct.StrVec) - 1
			if subLen == 0 {
				// 子项不能再切分了，退出
				return
			}
			varStruct.StrVec = varStruct.StrVec[0:subLen]
		} else {
			lastSymbol := symList[0]
			// 转换为定义位置信息
			if flag, oneDefine := convertVarInfoFlieToDefine(lastSymbol); flag {
				defineVecs = append(defineVecs, oneDefine)
			}

			return
		}
	}
}

// 客户端请求的信息，转换为通用的CommonFuncParam参数
func (a *AllProject) getVarCommonFuncParam(strFile string, varStruct *common.DefineVarStruct) (
	comParam *CommonFuncParam) {
	// 1）先查找该文件是否存在
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("FindVarDefine error, not find file=%s", strFile)
		return nil
	}

	// 2) 查找该文件属于第哪个第二阶段的指针
	var secondProject *results.SingleProjectResult
	// 3) 文件属于的第三阶段的指针, 代码作用域已经函数域
	var thirdStruct *results.AnalysisThird
	if fileStruct.IsCommonFile {
		secondProject = a.findMaxSecondProject(strFile)
		thirdStruct = a.thirdStruct
	}

	minScope, minFunc := fileStruct.FileResult.FindASTNode(varStruct.PosLine, varStruct.PosCh)
	if minScope == nil || minFunc == nil {
		log.Error("FindVarDefine error, minScope or minFunc is nil file=%s", strFile)
		return nil
	}

	// 5)冒号 函数，self语法进行转换
	common.ChangeFuncSelfToReferVar(minFunc, varStruct)

	// 6) 判断是否找的在table的定义处, 如果是不缺前面的定义
	if len(varStruct.StrVec) == 1 && !varStruct.BracketsFlag {
		strName := varStruct.StrVec[0]
		firstStr, secondStr := minScope.IsExistLocVarTableStrKey(strName, varStruct.PosLine+1, varStruct.PosCh)
		if firstStr == "" {
			// 查找全局变量
			firstStr, secondStr = fileStruct.FileResult.IsExistGlobalVarTableStrKey(strName, varStruct.PosLine+1,
				varStruct.PosCh)
		}

		if firstStr != "" {
			SliceInsert2(&(varStruct.StrVec), 0, firstStr)
			if secondStr != "" {
				SliceInsert2(&(varStruct.StrVec), 1, secondStr)
			}
		}
	}

	loc := lexer.Location{
		StartLine:   varStruct.PosLine + 1,
		StartColumn: varStruct.PosCh,
		EndLine:     varStruct.PosLine + 1,
		EndColumn:   varStruct.PosCh,
	}

	comParam = &CommonFuncParam{
		fileResult:    fileStruct.FileResult,
		fi:            minFunc,
		scope:         minScope,
		loc:           loc,
		secondProject: secondProject,
		thirdStruct:   thirdStruct,
	}
	return comParam
}

// findOldDefineInfo 查找变量最初的定义地方
// findStrName 为查找到的变量的名称
func (a *AllProject) findOldDefineInfo(comParam *CommonFuncParam, varStruct *common.DefineVarStruct) (
	findStrName string, findLocVar *common.VarInfo) {
	// 1) 判断是否为_G的前缀
	if varStruct.StrVec[0] == "_G" {
		// 设置_G的标记，且切分下数据
		varStruct.StrVec = varStruct.StrVec[1:]
		varStruct.IsFuncVec = varStruct.IsFuncVec[1:]
		if len(varStruct.StrVec) <= 0 {
			// 只有_G的，没有其他的内容, 直接返回
			log.Debug("just only _G, return")
			return "", nil
		}

		strName := varStruct.StrVec[0]
		gFlag := !common.GConfig.GetGVarExtendFlag()

		findVar := a.findGlobalVarDefineInfo(comParam, strName, "", gFlag)
		return strName, findVar
	}

	dirManager := common.GConfig.GetDirManager()

	// 2) 有前缀，先找到前缀指向的地方
	// 2) 先判断是否为协议前缀
	strProPre := ""
	if len(varStruct.StrVec) >= 2 && common.GConfig.IsStrProtocol(varStruct.StrVec[0]) &&
		dirManager.IsInDir(comParam.fileResult.Name) && varStruct.StrVec[0] != "" {
		// 如果为协议前缀，要进行切分
		strProPre = varStruct.StrVec[0]
		varStruct.StrVec = varStruct.StrVec[1:]
		varStruct.IsFuncVec = varStruct.IsFuncVec[1:]
	}

	if len(varStruct.StrVec) <= 0 {
		// 内容不够，直接退出
		log.Error("StrVec len error")
		return "", nil
	}

	strName := varStruct.StrVec[0]
	findLocVar = a.findVarDefineInfo(comParam, strName, strProPre)

	// 3) 判断是否查找的为后台协议的前缀内容
	if len(varStruct.StrVec) == 1 && findLocVar == nil && dirManager.IsInDir(comParam.fileResult.Name) {
		symbol := a.findProtocolDefine(strName, comParam.secondProject)
		if symbol != nil {
			findLocVar = symbol.VarInfo
		}
	}

	return strName, findLocVar
}

// FindVarDefine 查找变量的定义, 包括变量之前关联的所有变量
func (a *AllProject) FindVarDefine(strFile string, varStruct *common.DefineVarStruct) (
	oldSymbol *common.Symbol, symList []*common.Symbol) {
	comParam := a.getVarCommonFuncParam(strFile, varStruct)
	if comParam == nil {
		return
	}

	if varStruct.StrVec[0] == "require" && varStruct.IsFuncVec[0] && varStruct.Exp != nil {
		findExpList := []common.FindExpFile{}
		oldSymbol = a.FindVarReferSymbol(comParam.fileResult.Name, varStruct.Exp, comParam, &findExpList, 1)

		// require，已经处理了。上面已经进行了特殊的处理
		if len(varStruct.IsFuncVec) > 0 {
			varStruct.IsFuncVec[0] = false
		}
	} else {
		// 最初始的第一次查找，原始的
		findStrName, findVar := a.findOldDefineInfo(comParam, varStruct)
		if findVar == nil || len(varStruct.StrVec) <= 0 {
			return oldSymbol, symList
		}

		oldSymbol = a.createAnnotateSymbol(findStrName, findVar)
	}
	//调用链中没有函数，走这里
	if oldSymbol != nil {
		symList = a.getDeepVarList(oldSymbol, varStruct, comParam)
		return oldSymbol, symList
	}
	return nil, nil
}

// AnnotateTypeDefine 注解类型代码补全
func (a *AllProject) AnnotateTypeDefine(strFile string, strLine string, line int, col int) (defineVecs []DefineStruct) {
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

	// 2) 遍历位置信息
	typeStr, _, _ := annotateast.GetStateLocInfo(annotateState, col)
	if typeStr == "" {
		return
	}

	strDefineList := a.getStrNameDefineLocVec(typeStr, strFile, line+1)
	for _, oneDefine := range strDefineList {
		one := DefineStruct{
			StrFile: oneDefine.FileName,
			Loc:     oneDefine.Loc,
		}

		defineVecs = append(defineVecs, one)
	}

	return
}
