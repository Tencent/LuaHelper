package check

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"time"
)

// HandleFileEventChanges 项目工程文件的变化
// 返回值表示诊断信息是否有变化
func (a *AllProject) HandleFileEventChanges(fileEventVec []FileEventStruct) (changeDiagnostic bool) {
	// 0) 清除掉文件的cache
	common.GConfig.ClearCacheFileMap()

	// 判断该文件是属于哪些第二阶段的工程
	belongProjectMap := map[string]struct{}{}

	// 需要重新分析的文件列表(包括第一遍生成AST)
	needAgainFileVec := []string{}

	// 需要重新分析的引用关系，其他文件引用了这个文件，都需要重新分析引用关系；文件新增或删除，都会导致原有的引用关系改变
	needReferFileMap := map[string]struct{}{}

	// 是否处理所有的文件
	handleAllFlag := false
	thirdFlag := false

	// 是否有变化
	changeFlag := false
	changeDiagnostic = false
	dirManager := common.GConfig.GetDirManager()

	// 删除文件的列表
	deleteFileMap := map[string]struct{}{}

	// 1) 处理有变化的文件
	time0 := time.Now()
	for _, fileEvents := range fileEventVec {
		strFile := fileEvents.StrFile
		if fileEvents.Type == FileEventCreated {
			a.allFilesMap[strFile] = common.CompleteFilePathToPreStr(strFile)
			a.fileIndexInfo.InsertOneFile(strFile)
			
			needAgainFileVec = append(needAgainFileVec, strFile)
			if dirManager.IsInDir(strFile) {
				needReferFileMap[strFile] = struct{}{}
				// 新创建的文件，全量的工程分析
				handleAllFlag = true
				thirdFlag = true
				log.Debug("add strFile=%s need complete handle all files", strFile)
			}
		} else if fileEvents.Type == FileEventChanged {
			needAgainFileVec = append(needAgainFileVec, strFile)
		} else if fileEvents.Type == FileEventDeleted {
			a.RemoveFile(strFile)
			deleteFileMap[strFile] = struct{}{}

			if dirManager.IsInDir(strFile) {
				needReferFileMap[strFile] = struct{}{}
				// 删除了文件，尽量全量的工程分析下
				handleAllFlag = true
				thirdFlag = true
			}
		}

		for strEntryFile, analysisSecond := range a.analysisSecondMap {
			if _, ok := analysisSecond.AllFiles[strFile]; ok {
				//log.Debug("strFile=%s is belong project=%s", strFile, strEntryFile)
				belongProjectMap[strEntryFile] = struct{}{}
			}
		}

		if a.thirdStruct != nil {
			if _, ok := a.thirdStruct.AllIncludeFile[strFile]; ok {
				//log.Debug("strFile=%s is belong third files", strFile)
				thirdFlag = true
			}
		}
	}

	// 有文件新增或删除，重建
	if handleAllFlag {
		common.GConfig.RebuildSameFileNameVar(a.allFilesMap)
	}

	// 2.1) 如果之前的文件有包含错误码为6的错误，这次有新的文件增加，之前的文件，还需要进行扫描下
	// needAgainFileVec = a.handleNeedAgainFileVec(needAgainFileVec, deleteFileMap)

	// 2) 判断是否有必要重新进行一阶段分析的文件
	time1 := time.Now()
	if len(needAgainFileVec) > 0 {
		// 设置第一轮标记
		a.setCheckTerm(results.CheckTermFirst)

		changeFlag = a.firstCreateAndTraverseAst(needAgainFileVec, true)

		// 保存后进行的分析，判断是否要删除cache中的内容，创建AST没有问题的结果，需要删除cache的内容
		for _, strFile := range needAgainFileVec {
			fileStruct, _ := a.GetFirstFileStuct(strFile)
			if fileStruct != nil && fileStruct.GetFileHandleErr() == results.FileHandleOk {
				// 分析成功的，删除cache的内容
				a.RemoveCacheContent(strFile)
			}
		}

		a.rebuidCreateTypeMap()
	}

	time2 := time.Now()
	log.Debug("needAgainFileVec len=%d, checkAstTime=%d", len(needAgainFileVec), time.Since(time1).Milliseconds())

	// 3) 判断引用关系的是否有变化的
	if len(needReferFileMap) > 0 {
		// 全量处理所有的引用关系
		for _, fileStruct := range a.fileStructMap {
			if fileStruct.FileResult == nil {
				continue
			}

			fileStruct.FileResult.ReanalyseReferInfo(needReferFileMap, a.allFilesMap, a.fileIndexInfo)
		}
		// 引用关系变了，诊断信息也要跟着改变
		changeDiagnostic = true
	}

	log.Debug("needReferFileMap len=%d, costTime=%d, changeFlag=%t", len(needReferFileMap),
		time.Since(time2).Milliseconds(), changeFlag)
	if !changeFlag && !handleAllFlag {
		log.Debug("HandleFileEventChanges change false, changeDiagnostic=%t", changeDiagnostic)
		return changeDiagnostic
	}

	changeDiagnostic = true

	// 判断是否要进行特殊的校验
	if len(a.entryFilesList) == 0 && !common.GConfig.IsSpecialCheck() {
		time5 := time.Now()

		if thirdFlag {
			// 不需要进行特殊的校验
			a.setCheckTerm(results.CheckTermThird)
			a.HandleNotCheckThirdFile()
		}

		log.Debug("HandleAllThirdFile thirdFlag=%t, third costTime=%d, all_time=%d", thirdFlag,
			time.Since(time5).Milliseconds(), time.Since(time0).Milliseconds())
	} else {
		// 5) 对有需要的工程做第二遍工程遍历
		handleProjectLen := 0
		time4 := time.Now()
		if handleAllFlag {
			// 设置第二轮标记
			a.setCheckTerm(results.CheckTermSecond)
			a.handleProjectEntryFileVec(a.entryFilesList)
			handleProjectLen = len(a.entryFilesList)
		} else {
			// 设置第二轮标记
			a.setCheckTerm(results.CheckTermSecond)

			var belongProjectVec []string
			for belongFile := range belongProjectMap {
				belongProjectVec = append(belongProjectVec, belongFile)
			}
			if len(belongProjectVec) > 0 {
				a.handleProjectEntryFileVec(belongProjectVec)
			}
			handleProjectLen = len(belongProjectVec)
		}

		log.Debug("handleProject len=%d, costTime=%d", handleProjectLen, time.Since(time4).Milliseconds())
		time5 := time.Now()

		// 6) 对有需要的散落文件做第三遍的散落文件遍历
		if thirdFlag {
			// 设置第三轮标记
			a.setCheckTerm(results.CheckTermThird)
			a.HandleAllThirdFile()
		}

		log.Debug("HandleAllThirdFile flag=%t, costTime=%d, all_time=%d", thirdFlag,
			time.Since(time5).Milliseconds(), time.Since(time0).Milliseconds())
	}

	// 7) 重新创建所有的createTypeMap 注释类型
	a.rebuidCreateTypeMap()

	a.checkAllAnnotate()
	return true
}

// RemoveCacheContent 删除cache的结构
func (a *AllProject) RemoveCacheContent(strFile string) {
	if ok := a.fileLRUMap.Remove(strFile); ok {
		log.Debug("RemoveCacheContent ok strFile=%s", strFile)
	}
}

// HandleFileChangeAnalysis 代码实时变化时候，进行分析判断是否要保存到cache中
func (a *AllProject) HandleFileChangeAnalysis(strFile string, content []byte) (errList []common.CheckError) {
	fileStruct := results.CreateFileStruct(strFile)
	handleResult, _, _ := a.analysisFirstLuaFile(fileStruct, strFile, content, false, true)
	fileStruct.HandleResult = handleResult

	// 分析成功, 插入到cache中
	if handleResult == results.FileHandleOk {
		// 实时分析成功了，保存在cache中
		a.fileLRUMap.Set(strFile, fileStruct)
		log.Debug("HandleFileChangeAnalysis ok strFile=%s", strFile)
	}

	if fileStruct.FileResult != nil {
		errList = fileStruct.FileResult.GetAstCheckError()
	}

	return errList
}
