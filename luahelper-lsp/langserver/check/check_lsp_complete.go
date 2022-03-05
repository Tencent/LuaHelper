package check

import (
	"fmt"
	"luahelper-lsp/langserver/check/analysis"
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	"strings"
	"time"
)

// 代码提示高级功能，例如前面出现了ss.data , 当输入ss.时候，自动代码提示data
// valFileName 为变量所在的文件
// fileName 为对哪个文件进行展开查找
func (a *AllProject) completeExtension(valFileName string, fileName string,
	varInfo *common.VarInfo, strName string, sufVec []string, posLine int,
	posCh int) (strMap map[string]bool) {

	a.setCheckTerm(results.CheckTermFive)

	// 高级功能，会再次完整的遍历下AST
	analysisFive := results.CreateAnalysisFiveFile(fileName)
	analysisFive.FileName = valFileName
	analysisFive.FindVar = varInfo
	analysisFive.StrName = strName
	analysisFive.SufVec = sufVec
	analysisFive.PosLine = posLine
	analysisFive.PosCh = posCh
	a.handleCompleteName(analysisFive)

	strSuf := strings.Join(sufVec, ".")
	if strSuf == "." {
		strSuf = ""
	}

	// 所有提示的字符串
	tipMap := map[string]bool{}
	// 遍历所有的对table的访问
	for oneStr := range analysisFive.FindContentMap {
		if strSuf == "" {
			// 如果为空，获取第一个字符串
			oneVec := strings.Split(oneStr, ".")
			tipMap[oneVec[0]] = true
			continue
		}

		// 不为空
		strPre := strSuf + "."
		if !strings.HasPrefix(oneStr, strPre) {
			continue
		}

		oneStr = strings.TrimPrefix(oneStr, strPre)
		oneVec := strings.Split(oneStr, ".")
		tipMap[oneVec[0]] = true
	}

	return tipMap
}

func (a *AllProject) handleCompleteName(c *results.CompleteFileResult) {
	log.Debug("fivefile=%s", c.StrFile)
	strFile := pathpre.GetRemovePreStr(c.StrFile)
	// 获取缓存中的文件
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("handleOneFile luafile:%s file not valid", strFile)
		return
	}

	fileResult := fileStruct.FileResult

	// 创建第五轮遍历的包裹对象
	time1 := time.Now()
	analysis := analysis.CreateAnalysis(results.CheckTermFive, c.StrFile)
	analysis.CompleteResult = c
	analysis.Projects = a
	analysis.HandleTermTraverseAST(results.CheckTermFive, fileResult, nil)

	ftime := time.Since(time1).Milliseconds()
	log.Debug("handleCompleteName handleOneFile %s, cost time=%d(ms)", strFile, ftime)
}

// CodeCompleteFile 提示输入所有的lua文件或是库
func (a *AllProject) CodeCompleteFile(strFile string, referNameStr string, referType common.ReferType,
	preFileStr string) {
	separateLen := strings.LastIndex(preFileStr, ".")
	if separateLen == -1 {
		separateLen = strings.LastIndex(preFileStr, "/")
	}
	preStr := ""
	if separateLen != -1 {
		preStr = preFileStr[0 : separateLen+1]
	}

	// 如果为框架中引入其他的文件，判断是否要包含文件的后缀名
	suffixFlag := common.JudgeReferSuffixFlag(referType, referNameStr)

	dirManager := common.GConfig.GetDirManager()
	mainDir := dirManager.GetMainDir()

	var tipList []common.OneTipFile
	dirManager.GetAllCompleFile(mainDir, referType, suffixFlag, &tipList)
	for _, oneTip := range tipList {
		var strName string
		if common.GConfig.ReferMatchPathFlag {
			if !strings.HasPrefix(oneTip.StrName, preStr) {
				continue
			}
			strName = oneTip.StrName
		} else {
			index := strings.Index(oneTip.StrName, preStr)
			if index < 0 {
				continue
			}

			strName = oneTip.StrName[index:]
		}

		a.completeCache.InsertCompleteNormal(strName, oneTip.StrDesc, "", common.IKVariable)
	}
}

// CodeComplete 代码进行补全
// sufThreeStrVec 切分之后的从第三个开始数组
// colonFlag 表示是否为冒号的语法
func (a *AllProject) CodeComplete(strFile string, completeVar common.CompleteVarStruct) {
	// 1）先查找该文件是否存在
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("CodeComplete error, file not valid file=%s", strFile)
		return
	}

	var secondProject *results.SingleProjectResult
	var thirdStruct *results.AnalysisThird
	if fileStruct.IsCommonFile {
		// 2) 查找该文件属于第哪个第二阶段的指针
		secondProject = a.findMaxSecondProject(strFile)

		// 2) 文件属于的第三阶段的指针
		thirdStruct = a.thirdStruct
	}
	minScope, minFunc := fileStruct.FileResult.FindASTNode(completeVar.PosLine, completeVar.PosCh)
	if minScope == nil || minFunc == nil {
		log.Error("CodeComplete error, minScope or minFunc is nil file=%s", strFile)
		return
	}

	// 5)冒号 函数，self语法进行转换
	common.ChangeSelfToVarComplete(minFunc, &completeVar)

	loc := lexer.Location{
		StartLine:   completeVar.PosLine + 1,
		StartColumn: completeVar.PosCh,
		EndLine:     completeVar.PosLine + 1,
		EndColumn:   completeVar.PosCh,
	}

	// 5) 开始真正的代码补全
	comParam := &CommonFuncParam{
		fileResult:    fileStruct.FileResult,
		fi:            minFunc,
		scope:         minScope,
		loc:           loc,
		secondProject: secondProject,
		thirdStruct:   thirdStruct,
	}

	a.completeCache.SetColonFlag(completeVar.ColonFlag)
	a.lspCodeComplete(comParam, &completeVar)
}

// FuncCommentComplete 生成函数的注释提示
func (a *AllProject) FuncCommentComplete(strFile string, posLine int) (completeVecs []common.CompletionItemStruct) {
	// 1）先查找该文件是否存在
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("CodeComplete error, not valid file=%s", strFile)
		return
	}

	// 3) 获取对应行是否有函数的定义
	funcInfo := fileStruct.FileResult.GetLineFuncInfo(posLine + 2)
	if funcInfo == nil {
		return
	}

	// 4) 组装想要的数据
	insertText := "--- ${1:func desc}"
	// 变量所有的参数
	index := 2
	for i, oneParam := range funcInfo.ParamList {
		// 忽略掉第一个参数
		if funcInfo.IsColon && i == 0 {
			continue
		}

		oneParamFormatStr := fmt.Sprintf("\n---@param %s ${%d:any}", oneParam, index)
		index++
		insertText = insertText + oneParamFormatStr
	}

	oneFuncComment := common.CompletionItemStruct{
		Label:         "generate func doc comments",
		Kind:          common.IKSnippet,
		InsetText:     insertText,
		Detail:        "generate document comments for the function",
		Documentation: "",
	}

	completeVecs = append(completeVecs, oneFuncComment)
	return
}

// 查找所有的协议前缀
func (a *AllProject) protocolCodeComplete(strProPre string, secondProject *results.SingleProjectResult, thirdStruct *results.AnalysisThird) {
	var globalGmaps map[string]*common.VarInfoList

	// 向工程的globalMaps中查找变量
	if secondProject != nil {
		globalGmaps = secondProject.FirstProtcolMaps
	} else if thirdStruct.GlobalVarMaps != nil {
		globalGmaps = thirdStruct.ProtcolVarMaps
	}
	if globalGmaps == nil {
		return
	}

	// 向工程的globalMaps中查找变量
	for strName, varInfoList := range globalGmaps {
		if len(varInfoList.VarVec) == 0 {
			continue
		}

		for i := len(varInfoList.VarVec) - 1; i >= 0; i-- {
			oneVar := varInfoList.VarVec[i]
			if oneVar.ExtraGlobal.StrProPre != strProPre {
				continue
			}

			a.completeCache.InsertCompleteVar(oneVar.FileName, strName, oneVar)
		}
	}
}

// _G所有变量的代码补全
// funcFlag 表示是否只获取函数
// gFlag 表示是否只获取_G前缀的
// ignoreFile 表示忽略指定文件的变量，前面已经加入过了
func (a *AllProject) gValueComplete(comParam *CommonFuncParam, completeVar *common.CompleteVarStruct,
	gFlag bool, ignoreFile string) {
	var globalGmaps map[string]*common.VarInfoList

	// 向工程的globalMaps中查找变量
	if comParam.secondProject != nil {
		globalGmaps = comParam.secondProject.FirstGlobalGMaps
	} else if comParam.thirdStruct != nil {
		globalGmaps = comParam.thirdStruct.GlobalVarMaps
	}

	if globalGmaps == nil {
		return
	}

	// 也把当前文件_G显示出来
	if ignoreFile == "" {
		for strName, oneVar := range comParam.fileResult.GlobalMaps {
			// todo，这里处理了所有的符号包含_G和非_G的

			if !common.IsCompleteNeedShow(strName, completeVar) {
				continue
			}
			if a.completeCache.ExistStr(strName) {
				continue
			}

			a.completeCache.InsertCompleteVar(comParam.fileResult.Name, strName, oneVar)
		}
	}

	for strName, varInfoList := range globalGmaps {
		// 判断是否重复了
		if !common.IsCompleteNeedShow(strName, completeVar) {
			continue
		}

		if a.completeCache.ExistStr(strName) {
			continue
		}

		oneVar := varInfoList.GetLastOneVar()
		if oneVar == nil {
			continue
		}

		fileName := oneVar.FileName
		if fileName == ignoreFile {
			continue
		}

		a.completeCache.InsertCompleteVar(fileName, strName, oneVar)
	}
}

// 前缀是_G符号
func (a *AllProject) gPreComplete(comParam *CommonFuncParam, completeVar *common.CompleteVarStruct) {
	// 1) 单纯_G符号的代码补全
	lenStrVec := len(completeVar.StrVec)
	// 最后的进行转换
	if lenStrVec >= 2 && !completeVar.LastEmptyFlag {
		completeVar.StrVec = completeVar.StrVec[0 : lenStrVec-1]
		lenStrVec = len(completeVar.StrVec)
		completeVar.LastEmptyFlag = true
	}

	if lenStrVec == 1 && completeVar.LastEmptyFlag {
		gFlag := !common.GConfig.GetGVarExtendFlag()
		a.gValueComplete(comParam, completeVar, gFlag, "")
		return
	}

	if lenStrVec <= 1 {
		return
	}

	strName := completeVar.StrVec[1]
	completeVar.StrVec = completeVar.StrVec[1:]
	if len(completeVar.IsFuncVec) >= 1 {
		completeVar.IsFuncVec = completeVar.IsFuncVec[1:]
	}
	//  获取到全局的信息
	findVar := a.findGlobalVarDefineInfo(comParam, strName, "", false)

	// 没有找到
	if findVar == nil {
		return
	}

	oldSymbol := a.createAnnotateSymbol(strName, findVar)
	varStruct := &common.DefineVarStruct{
		StrVec:    completeVar.StrVec,
		IsFuncVec: completeVar.IsFuncVec,
		PosLine:   completeVar.PosLine,
		PosCh:     completeVar.PosCh,
		ColonFlag: completeVar.ColonFlag,
	}
	symList := a.getDeepVarList(oldSymbol, varStruct, comParam)
	a.varInfoDeepComplete(oldSymbol, symList, varStruct, completeVar, comParam)
}

// CompleteResultCh 结构的协程封装
type CompleteResultCh struct {
	strMap map[string]bool
}

// CompleteGoParam 告警展开，协程的参数
type CompleteGoParam struct {
	a            *AllProject
	luaInFile    string // 变量所在的文件
	fileName     string // fileName 为对哪个文件进行展开查找
	strFind      string
	comParam     *CommonFuncParam
	completeVar  *common.CompleteVarStruct
	localVarInfo *common.VarInfo
	sufStrVec    []string
}

// 其中一个展开的协程
func goCompleteExtension(copleteParam CompleteGoParam, retCompleteResultCh chan<- CompleteResultCh) {
	strMap := copleteParam.a.completeExtension(copleteParam.luaInFile, copleteParam.fileName,
		copleteParam.localVarInfo, copleteParam.strFind, copleteParam.sufStrVec,
		copleteParam.completeVar.PosLine+1, copleteParam.completeVar.PosCh)
	completeResultCh := CompleteResultCh{
		strMap: strMap,
	}

	retCompleteResultCh <- completeResultCh
}

// 对一个文件协程分析处理的结果，插入到补全缓存中
func (a *AllProject) insertFileCacheStrMap(tipMap map[string]bool) {
	for strOne := range tipMap {
		if a.completeCache.ExistStr(strOne) || a.completeCache.IsExcludeStr(strOne) {
			continue
		}

		a.completeCache.InsertCompleteNormal(strOne, "", "", common.IKVariable)
	}
}

// 对指定的文件，开始进行展开功能，获取代码的提示
// luaInFile 为变量所在的文件
// strFind 为对变量进行引用的文件
func (a *AllProject) getFileCompleteExt(luaInFile string, strFind string, comParam *CommonFuncParam,
	completeVar *common.CompleteVarStruct, varInfo *common.VarInfo, sufStrVec []string) {
	// 1）如果变量所在的文件，和变量的定义在同一个文件，只需要对一个文件进行展开
	if luaInFile == comParam.fileResult.Name {
		tipMap1 := a.completeExtension(luaInFile, comParam.fileResult.Name, varInfo, strFind, sufStrVec,
			completeVar.PosLine+1, completeVar.PosCh)
		a.insertFileCacheStrMap(tipMap1)
		return
	}

	// 2) 如果两个不相等，需要启动两个协程来解决
	retCompleteResultCh := make(chan CompleteResultCh, 2)
	copleteParam := CompleteGoParam{
		a:            a,
		luaInFile:    luaInFile,
		fileName:     luaInFile,
		strFind:      strFind,
		comParam:     comParam,
		completeVar:  completeVar,
		localVarInfo: varInfo,
		sufStrVec:    sufStrVec,
	}

	// 1) 开启第一个协程
	go goCompleteExtension(copleteParam, retCompleteResultCh)

	// 2) 开启第一个协程
	copleteParam.fileName = comParam.fileResult.Name
	go goCompleteExtension(copleteParam, retCompleteResultCh)

	// 3) 接收一个协程的数据
	resultch1 := <-retCompleteResultCh
	a.insertFileCacheStrMap(resultch1.strMap)

	// 4) 接收另外一个协程的数据
	resultch2 := <-retCompleteResultCh
	a.insertFileCacheStrMap(resultch2.strMap)
}

// import文件的所有成员放入进来
func (a *AllProject) getImportFileComlete(referFile *results.FileResult) {
	// 补全引用的那个文件所有的全局符号
	for strName, subVar := range referFile.GlobalMaps {
		// 所有全局符号的信息， _G的去掉,有引用信息的也去掉
		if subVar.ExtraGlobal.GFlag || subVar.ReferInfo != nil {
			continue
		}
		if a.completeCache.ExistStr(strName) {
			continue
		}

		a.completeCache.InsertCompleteVar(referFile.Name, strName, subVar)
	}
}

// 构造找到了变量第一次引用和跟踪引用地方，对变量进行补全
func (a *AllProject) varInfoDeepComplete(symbol *common.Symbol, symList []*common.Symbol,
	varStruct *common.DefineVarStruct, completeVar *common.CompleteVarStruct, comParam *CommonFuncParam) {
	strFind := completeVar.StrVec[0]
	fileName := symbol.FileName

	// 1) 没有追踪到最后的子项
	if len(symList) == 0 {
		a.expandNodefineMapComplete(fileName, strFind, comParam, completeVar, symbol.VarInfo)

		// 只在当前文件做代码提示高级功能
		//sufThreeStrVec := completeVar.StrVec[1:]

		// 尝试用协程进行高级功能展开
		//a.getFileCompleteExt(fileName, strFind, comParam, completeVar, symbol.VarInfo, sufThreeStrVec)
		return
	}

	// 获取最后一个切割是否为函数返回
	isLastFuncFlag := false
	if len(completeVar.IsFuncVec) >= len(completeVar.StrVec) {
		isLastFuncFlag = completeVar.IsFuncVec[len(completeVar.StrVec)-1]
	}

	// 不是直接像这种 a = import("a.lua") , 调用的 a.test()这样的补全
	// 获取局部变量或全局变量的所有成员函数补全
	for _, symbol := range symList {
		a.getVarInfoCompleteExt(symbol, completeVar.ColonFlag)

		if isLastFuncFlag {
			// 切割处理，最后一个是函数，获取这个函数的代码补全
			a.getFuncReturnCompleteExt(symbol, completeVar.ColonFlag, comParam)
		}
	}

	// 引用其他的文件，不做冒号语法
	lastSymbol := symList[len(symList)-1]
	if lastSymbol.VarInfo == nil {
		return
	}

	referInfo := lastSymbol.VarInfo.ReferInfo
	var referFile *results.FileResult
	if referInfo != nil {
		referFile = a.GetFirstReferFileResult(referInfo)
		if referFile == nil {
			log.Error("lspCodeComplete import file error, module=%s", completeVar.StrVec[0])
		}
	}

	// 高级功能展开的
	explanFlag := true
	if referFile != nil {
		referSubType := a.GetReferFrameType(referInfo)
		if referSubType == common.RtypeImport {
			// import引用一个文件，代码补全
			explanFlag = false
			if !completeVar.ColonFlag {
				a.getImportFileComlete(referFile)
			}
		} else if referSubType == common.RtypeRequire {
			find, returnExp := referFile.MainFunc.GetOneReturnExp()
			if !find {
				return
			}

			findExpList := []common.FindExpFile{}
			// 这里也需要做判断，函数返回的变量逐层跟踪，目前只跟踪了一层
			symList := a.FindDeepSymbolList(referFile.Name, returnExp, comParam, &findExpList, true, 1)
			for _, symbol := range symList {
				a.getVarInfoCompleteExt(symbol, completeVar.ColonFlag)
			}
		}
	}

	// 冒号语法不做高级提示功能
	if explanFlag {
		a.expandNodefineMapComplete(fileName, strFind, comParam, completeVar, symbol.VarInfo)

		// 只在当前文件做代码提示高级功能
		//sufThreeStrVec := completeVar.StrVec[1:]
		// 尝试用协程进行高级功能展开
		//a.getFileCompleteExt(fileName, strFind, comParam, completeVar, symbol.VarInfo, sufThreeStrVec)
	}
}

// 查找所有的协议前缀
func (a *AllProject) systemMoudleComplete(strModule string) bool {
	oneModule, ok := common.GConfig.SystemModuleTipsMap[strModule]
	if !ok {
		return false
	}

	// 提示该模块的所有函数
	for strName, oneFunc := range oneModule.ModuleFuncMap {
		a.completeCache.InsertCompleteSysModuleMem(strName, oneFunc.Detail,
			oneFunc.Documentation, common.IKFunction)
	}

	for _, oneVar := range oneModule.ModuleVarVec {
		a.completeCache.InsertCompleteSysModuleMem(oneVar.Label, oneVar.Detail,
			oneVar.Documentation, common.IKVariable)
	}
	return true
}

// 其他前缀代码补全
func (a *AllProject) otherPreComplete(comParam *CommonFuncParam, completeVar *common.CompleteVarStruct) {
	lenStrVec := len(completeVar.StrVec)
	// 最后的进行转换
	if lenStrVec >= 2 && !completeVar.LastEmptyFlag {
		completeVar.StrVec = completeVar.StrVec[0 : lenStrVec-1]
		lenStrVec = len(completeVar.StrVec)
		completeVar.LastEmptyFlag = true
	}

	strFind := completeVar.StrVec[0]

	// 1) 是否为协议前缀
	if lenStrVec == 1 && completeVar.LastEmptyFlag {
		if common.GConfig.IsStrProtocol(strFind) {
			// 为协议前缀，返回所有的返回
			a.protocolCodeComplete(strFind, comParam.secondProject, comParam.thirdStruct)
			return
		}
	}

	varStruct := common.DefineVarStruct{
		StrVec:    completeVar.StrVec,
		IsFuncVec: completeVar.IsFuncVec,
		PosLine:   completeVar.PosLine,
		PosCh:     completeVar.PosCh,
		ColonFlag: completeVar.ColonFlag,
		Exp:       completeVar.Exp,
	}

	symbol, symList := a.FindVarDefine(comParam.fileResult.Name, &varStruct)
	if symbol == nil {
		// 判断是否为系统模块函数提示
		if a.systemMoudleComplete(strFind) {
			return
		}

		// 判断是否在globalNodefineMaps中
		strName := completeVar.StrVec[0]
		if findVar, ok := comParam.fileResult.NodefineMaps[strName]; ok {
			// 对NodefineMap内的变量进行展开代码补全
			a.expandNodefineMapComplete(comParam.fileResult.Name, strName, comParam, completeVar, findVar)
		}

		// 遍历AST树，构造想要的代码补全数据
		//sufThreeStrVec := completeVar.StrVec[1:]
		//a.getFileCompleteExt(comParam.fileResult.Name, strName, comParam, completeVar, findVar, sufThreeStrVec)
		return
	}

	a.varInfoDeepComplete(symbol, symList, &varStruct, completeVar, comParam)
}

// 没有前缀的代码补全
func (a *AllProject) noPreComplete(comParam *CommonFuncParam, completeVar *common.CompleteVarStruct) {
	// 3.0) 先判断是否为函数参数的候选词代码补全
	if a.paramCandidateComplete(comParam, completeVar) {
		return
	}

	// 3) 单纯的文件范围内代码补全
	// 3.1) 先把文件的局部范围变量放进来
	fileName := comParam.fileResult.Name
	comParam.scope.GetCompleteVar(completeVar, fileName, comParam.loc, a.completeCache)

	// 3.2) 再把这个文件的全局变量放进来
	for strName, subVar := range comParam.fileResult.GlobalMaps {
		if !common.IsCompleteNeedShow(strName, completeVar) || a.completeCache.ExistStr(strName) {
			continue
		}

		a.completeCache.InsertCompleteVar(fileName, strName, subVar)
	}

	// 3.3) 把常用见的关键字放入进来
	for strName := range common.GConfig.CompKeyMap {
		if completeVar.IgnoreKeyWord {
			break
		}

		if !common.IsCompleteNeedShow(strName, completeVar) || a.completeCache.ExistStr(strName) {
			continue
		}

		a.completeCache.InsertCompleteKey(strName)
	}

	// 3.4) 把snippet放入进来
	for strName := range common.GConfig.CompSnippetMap {
		if completeVar.IgnoreKeyWord {
			break
		}

		if !common.IsCompleteNeedShow(strName, completeVar) || a.completeCache.ExistStr(strName) {
			continue
		}

		a.completeCache.InsertCompleteSnippet(strName)
	}
	// 3.4) 把_G的函数也包含进来
	// 默认只提示_G的函数，如果要提示_G的变量，需要配置打开，整体上会慢一点
	a.gValueComplete(comParam, completeVar, false, fileName)

	// 3.5) 把框架中引入的其他文件的方式，函数也包含进来
	referFrameFiles := common.GConfig.GetFrameReferFiles()
	for _, strOne := range referFrameFiles {
		if !common.IsCompleteNeedShow(strOne, completeVar) || a.completeCache.ExistStr(strOne) {
			continue
		}

		detail := "refer other file"
		a.completeCache.InsertCompleteNormal(strOne, detail, "", common.IKFunction)
	}

	// 3.6） 系统的全局函数放入进来
	for strName, oneSysFunc := range common.GConfig.SystemTipsMap {
		if !common.IsCompleteNeedShow(strName, completeVar) || a.completeCache.ExistStr(strName) {
			continue
		}

		a.completeCache.InsertCompleteSystemFunc(strName, oneSysFunc.Detail, oneSysFunc.Documentation)
	}

	// 3.7） 把系统的模块也加入进来
	for strName, oneMudule := range common.GConfig.SystemModuleTipsMap {
		if !common.IsCompleteNeedShow(strName, completeVar) || a.completeCache.ExistStr(strName) {
			continue
		}

		a.completeCache.InsertCompleteSystemModule(strName, oneMudule.Detail, oneMudule.Documentation)
	}

	// 3.68 把globalNodefineMaps中符合的变量也加入进来
	for strName := range comParam.fileResult.NodefineMaps {
		if !common.IsCompleteNeedShow(strName, completeVar) || a.completeCache.ExistStr(strName) {
			continue
		}

		subVar := comParam.fileResult.NodefineMaps[strName]
		a.completeCache.InsertCompleteVar(fileName, strName, subVar)
	}
}

// 代码补全进行的分发
func (a *AllProject) lspCodeComplete(comParam *CommonFuncParam, completeVar *common.CompleteVarStruct) {
	a.GetCompleteCache().SetCompleteVar(completeVar)

	// 1) 查找所有的_G符号
	if completeVar.StrVec[0] == "_G" {
		a.gPreComplete(comParam, completeVar)
		return
	}

	// 判断是否为注解table中的key值补全, 增强修复
	a.needAnnotateTableFieldRepair(comParam, completeVar)

	// 2） 没有前缀的代码补全, len(completeVar.StrVec) == 1
	if len(completeVar.StrVec) == 1 && !completeVar.LastEmptyFlag {
		a.noPreComplete(comParam, completeVar)
		return
	}

	// 3）模块变量的代码补全
	a.otherPreComplete(comParam, completeVar)
}

// GetCompleteCacheItems 获取所有的
func (a *AllProject) GetCompleteCacheItems() (items []common.OneCompleteData) {
	return a.completeCache.GetDataList()
}

// ClearCompleteCache 清除所有的代码补全缓冲
func (a *AllProject) ClearCompleteCache() {
	a.completeCache.ResertData()
}

func (a *AllProject) getVarCompleteData(varInfo *common.VarInfo, item *common.OneCompleteData) (luaFileStr string) {
	// 是否为特殊的冒号补全
	colonFlag := a.completeCache.GetColonFlag()
	dirManager := common.GConfig.GetDirManager()

	// 判断这个变量是否关联到了注解类型，如果有提取注解信息
	symbol := common.GetDefaultSymbol(item.LuaFile, varInfo)
	astType, strComment, strPreComment := a.getInfoFileAnnotateType(item.Label, symbol)
	if astType != nil {
		str := a.completeAnnotatTypeStr(astType, item.LuaFile, symbol.GetLine())
		if strPreComment == "" {
			item.Detail = str
		} else if strPreComment == "type" {
			item.Detail = item.Label + " : " + str
		} else {
			item.Detail = strPreComment + "  " + str
		}

		// 判断是否关联成number，如果是number类型尝试获取具体的值
		strType := varInfo.GetVarTypeDetail()
		if strings.HasPrefix(strType, "number: ") {
			item.Detail = item.Detail + ": " + strings.TrimPrefix(strType, "number: ")
		}

		if strComment != "" {
			item.Documentation = strComment
		}

		luaFileStr = dirManager.RemovePathDirPre(item.LuaFile)
		return
	}

	item.Detail = varInfo.GetVarTypeDetail()
	expandFlag := false
	var firstExistMap map[string]string = map[string]string{}
	if item.Detail == "table" || len(varInfo.SubMaps) > 0 || len(varInfo.ExpandStrMap) > 0 {
		item.Detail, firstExistMap = a.expandTableHover(symbol)
		expandFlag = true
	}

	if item.Detail == "any" || expandFlag {
		comParam := a.getCommFunc(item.LuaFile, varInfo.Loc.StartLine, varInfo.Loc.StartColumn)
		if comParam == nil {
			return
		}

		// 进行追踪出具体的信息
		findExpList := []common.FindExpFile{}
		// 这里也需要做判断，函数返回的变量逐层跟踪，目前只跟踪了一层
		symList := a.FindDeepSymbolList(item.LuaFile, item.VarInfo.ReferExp, comParam, &findExpList, true, 1)
		for _, symbolTmp := range symList {
			// 判断是否为注解类型
			if symbolTmp.AnnotateType != nil {
				secondStr, sendExistMap := a.expandTableHover(symbolTmp)
				item.Detail = item.Label + " : " + a.mergeTwoExistMap(symbol, item.Detail, firstExistMap, secondStr, sendExistMap)
				break
			}

			if symbolTmp.VarInfo != nil {
				strDetail := symbolTmp.VarInfo.GetVarTypeDetail()
				if strDetail == "table" || len(symbolTmp.VarInfo.SubMaps) > 0 || len(symbolTmp.VarInfo.ExpandStrMap) > 0 {
					secondStr, sendExistMap := a.expandTableHover(symbolTmp)
					strDetail = a.mergeTwoExistMap(symbol, item.Detail, firstExistMap, secondStr, sendExistMap)
				}

				if strDetail != "any" {
					item.Detail = strDetail
				}

				if symbolTmp.VarInfo.ReferFunc != nil {
					strFunc := a.getFuncShowStr(symbol.VarInfo, item.Label, true, colonFlag, true)
					item.Detail = "function " + strFunc
				}

				// 如果为引用的模块
				if symbolTmp.VarInfo.ReferInfo != nil {
					item.Detail = symbolTmp.VarInfo.ReferInfo.GetReferComment()
				}
			}
		}
	}

	if varInfo.ReferFunc != nil {
		strFunc := a.getFuncShowStr(symbol.VarInfo, item.Label, true, colonFlag, true)
		item.Detail = "function " + strFunc
	}

	// 如果为引用的模块
	if varInfo.ReferInfo != nil {
		item.Detail = varInfo.ReferInfo.GetReferComment()
	}

	line := varInfo.Loc.EndLine

	// 获取变量的注释
	strComment = a.GetLineComment(item.LuaFile, line)
	strComment = GetStrComment(strComment)
	item.Documentation = strComment
	luaFileStr = dirManager.RemovePathDirPre(item.LuaFile)
	return luaFileStr
}

// GetCompleteCacheIndexItem 获取单个缓存的结构
func (a *AllProject) GetCompleteCacheIndexItem(index int) (item common.OneCompleteData, luaFileStr string, flag bool) {
	item, flag = a.completeCache.GetIndexData(index)
	if !flag {
		return
	}

	// 1) 为普通的配置，直接返回
	if item.CacheKind == common.CKindNormal {
		if item.ExpandVarInfo == nil {
			return
		}

		luaFileStr = a.getCompleteExpandInfo(&item)
		return
	}

	dirManager := common.GConfig.GetDirManager()

	// 2) 配置的为变量
	if item.CacheKind == common.CKindVar {
		luaFileStr = a.getVarCompleteData(item.VarInfo, &item)
		return
	}

	// 3) 配置的为注解类型
	if item.CacheKind == common.CkindAnnotateType {
		typeOne := item.CreateTypeInfo
		if typeOne.AliasInfo != nil {
			item.Detail = "alias " + typeOne.AliasInfo.AliasState.Name

			strComment := typeOne.AliasInfo.AliasState.Comment
			if strComment != "" {
				item.Documentation = strComment
			}

			luaFileStr = dirManager.RemovePathDirPre(typeOne.AliasInfo.LuaFile)
		} else if typeOne.ClassInfo != nil {
			item.Detail = typeOne.ClassInfo.ClassState.Name
			if len(typeOne.ClassInfo.ClassState.ParentNameList) > 0 {
				item.Detail = item.Detail + " : " + strings.Join(typeOne.ClassInfo.ClassState.ParentNameList, " , ")
			} else {
				item.Detail = item.Detail + a.getClassFieldStr(typeOne.ClassInfo)
			}

			strComment := typeOne.ClassInfo.ClassState.Comment
			if strComment != "" {
				item.Documentation = strComment
			}

			luaFileStr = dirManager.RemovePathDirPre(typeOne.ClassInfo.LuaFile)
		}

		return
	}

	// 是否为特殊的冒号补全
	colonFlag := a.completeCache.GetColonFlag()
	// 4) 配置的为注解field信息
	if item.CacheKind == common.CkindClassField {
		if colonFlag {
			// 为冒号的补全
			if item.FieldColonFlag == annotateast.FieldColonHide {
				// 判断是否为隐藏式的，如果是隐藏式的，去掉第一个参数
				// ---@class ClassA
				// ---@field FunctionC fun(self:ClassA):void
				if oneFuncType := annotateast.GetAllFuncType(item.FieldState.FiledType); oneFuncType != nil {
					subFuncType, _ := oneFuncType.(*annotateast.FuncType)
					item.Detail = annotateast.FuncTypeConvertStr(subFuncType, 1)
				}
			} else {
				// 获取对应注解的类型
				item.Detail = annotateast.TypeConvertStr(item.FieldState.FiledType)
			}
		} else {
			// 为的补全
			// 判断是否为隐藏式的，如果是显示的，增加第一个参数self
			// ---@class ClassA
			// ---@field FunctionC : fun():void
			if item.FieldColonFlag == annotateast.FieldColonYes {
				if oneFuncType := annotateast.GetAllFuncType(item.FieldState.FiledType); oneFuncType != nil {
					subFuncType, _ := oneFuncType.(*annotateast.FuncType)
					item.Detail = annotateast.FuncTypeConvertStr(subFuncType, 2)
				}
			} else {
				// 获取对应注解的类型
				item.Detail = annotateast.TypeConvertStr(item.FieldState.FiledType)
			}
		}

		// 获取注释
		item.Documentation = item.FieldState.Comment
		luaFileStr = dirManager.RemovePathDirPre(item.LuaFile)
	}

	return
}
