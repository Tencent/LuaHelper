package check

import (
	"luahelper-lsp/langserver/check/analysis"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	"reflect"
	"runtime"
	"time"
)

func (a *AllProject) handleFindferences(r *results.ReferenceFileResult) {
	strFile := pathpre.GetRemovePreStr(r.StrFile)
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("handleOneFile luafile:%s not exist", strFile)
		return
	}
	fileResult := fileStruct.FileResult
	time1 := time.Now()

	// 创建第四轮遍历的包裹对象
	analysis := analysis.CreateAnalysis(results.CheckTermFour, r.StrFile)
	analysis.ReferenceResult = r
	analysis.Projects = a
	analysis.HandleTermTraverseAST(results.CheckTermFour, fileResult, nil)

	ftime := time.Since(time1).Milliseconds()
	log.Debug("handleFindferences handleOneFile %s, cost time=%d(ms)", strFile, ftime)
}

// 获取到有效的FileAnalysis
func (a *AllProject) getFileAnalysis(strFile string) *results.FileResult {
	// 1）先查找该文件是否存在
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("getFileAnalysis luafile:%s not exist", strFile)
		return nil
	}

	return fileStruct.FileResult
}

// FindReferences 查找引用 1111
func (a *AllProject) FindReferences(strFile string, varStruct *common.DefineVarStruct,
	checkSrc common.CheckReferenceSrc) (findVecs []DefineStruct) {
	lastDefine, oldInfoFlie, isWhole := a.FindReferenceVarDefine(strFile, varStruct)
	if oldInfoFlie == nil || oldInfoFlie.FileName == "" || oldInfoFlie.VarInfo == nil {
		return
	}

	luaInFile := oldInfoFlie.FileName
	defineLoc := lastDefine
	oldLocVar := oldInfoFlie.VarInfo

	ignoreDefineLoc := defineLoc
	if (checkSrc == common.CRSReference && common.GConfig.ReferenceDefineFlag) ||
		(checkSrc == common.CRSHighlight && strFile == luaInFile) ||
		(checkSrc == common.CRSRename && isWhole) {
		findVecs = append(findVecs, DefineStruct{
			StrFile: luaInFile,
			Loc:     defineLoc,
		})
	}

	// 如果该文件属于多个第二阶段指针，返回第二阶段工程项目最多的文件
	secondProjecFileNum := 0
	// 引用的文件所需要处理的第二阶段工程名
	var secondProjectVec []*results.SingleProjectResult
	for strEntry, tempSecondProject := range a.analysisSecondMap {
		if _, ok := tempSecondProject.AllFiles[strFile]; !ok {
			continue
		}

		secondProjectVec = append(secondProjectVec, tempSecondProject)
		if len(tempSecondProject.AllFiles) > secondProjecFileNum {
			//secondProject = tempSecondProject
			secondProjecFileNum = len(tempSecondProject.AllFiles)
			log.Error("FindVarDefine, belong secondProject file=%s, entryFile=%s, filesNum=%d, entry=%s", strFile,
				tempSecondProject.EntryFile, secondProjecFileNum, strEntry)
		}
	}

	// 5) 文件属于的第三阶段的指针, 代码作用域已经函数域
	thirdStruct := a.thirdStruct

	// 判断是否在所有文件中查找引用
	allFileFind := false
	if oldLocVar.IsGlobal() {
		allFileFind = true
	}

	if strFile != luaInFile {
		allFileFind = true
	}

	// 获取到这个文件返回，如果这个文件有返回这个变量的父变量，需要全局引用查找
	if !allFileFind {
		fileResult := a.getFileAnalysis(luaInFile)
		if fileResult != nil {
			locVar := fileResult.MainFunc.GetOneReturnVar()
			if oldLocVar.HasSubVarInfo(locVar) {
				allFileFind = true
			}
		}
	}

	// 查找的为局部变量，只在当前文件查找，不需要利用多协程。高亮同理
	dirManager := common.GConfig.GetDirManager()
	if !allFileFind || checkSrc == common.CRSHighlight || !dirManager.IsInDir(strFile) {
		// 如果为局部变量，只在当前文件查找
		inFile := luaInFile //变量定义的文件
		if checkSrc == common.CRSHighlight {
			inFile = strFile //变量当前的文件
		}

		analysisFour := results.CreateReferenceFileResult(inFile)
		var sendThirdStruct *results.AnalysisThird
		if thirdStruct != nil && thirdStruct.AllIncludeFile[inFile] {
			sendThirdStruct = thirdStruct
		}
		analysisFour.SetFindReferenceInfo(luaInFile, oldLocVar, secondProjectVec, varStruct.StrVec, ignoreDefineLoc,
			sendThirdStruct)

		a.handleFindferences(analysisFour)
		for _, oneLoc := range analysisFour.FindLocVec {
			if !ignoreDefineLoc.IsInitialLoc() && ignoreDefineLoc.IsInLocStruct(oneLoc.StartLine, oneLoc.StartColumn) &&
				ignoreDefineLoc.IsInLocStruct(oneLoc.EndLine, oneLoc.EndColumn) {
				continue
			}

			findVecs = append(findVecs, DefineStruct{
				StrFile: inFile,
				Loc:     oneLoc,
			})
		}
	} else {
		// 所有需要处理的文件，需要多协程分析
		allFileMap := map[string]struct{}{}
		var secondProjectVec []*results.SingleProjectResult
		for _, secondProject := range a.analysisSecondMap {
			if _, ok := secondProject.AllFiles[luaInFile]; !ok {
				continue
			}

			secondProjectVec = append(secondProjectVec, secondProject)
			for strFile := range secondProject.AllFiles {
				allFileMap[strFile] = struct{}{}
			}
		}

		if thirdStruct.AllIncludeFile[luaInFile] {
			for strFile := range thirdStruct.AllIncludeFile {
				allFileMap[strFile] = struct{}{}
			}
		} else {
			// 第三阶段不需要查找该文件
			thirdStruct = nil
		}

		referenceParam := ReferenceParam{
			analysisThird:    thirdStruct,
			secondProjectVec: secondProjectVec,
			findSymbol:       oldLocVar,
			fileName:         luaInFile,
			suffStrVec:       varStruct.StrVec,
			ignoreDefineLoc:  ignoreDefineLoc,
		}

		delete(allFileMap, luaInFile)
		var fileList []string
		// 定义的文件，放在对首
		fileList = append(fileList, luaInFile)
		for strFile := range allFileMap {
			fileList = append(fileList, strFile)
		}
		handleAllFilesReference(fileList, a, referenceParam, &findVecs)
	}

	return findVecs
}

// ReferenceParam 封装传递的结果
type ReferenceParam struct {
	analysisThird    *results.AnalysisThird
	secondProjectVec []*results.SingleProjectResult
	findSymbol       *common.VarInfo
	fileName         string
	suffStrVec       []string
	ignoreDefineLoc  lexer.Location
}

// FourFileChan 第四阶段协成的检测工程的通信
type FourFileChan struct {
	sendRunFlag    bool   // 是否继续运行
	strFile        string // 需要处理的文件
	allProject     *AllProject
	referenceParam ReferenceParam   // 引用传递的参数
	findLocVec     []lexer.Location // 找到的引用关系位置
}

// GoRoutineFourFile 第四轮分析lua单个文件，接受主协程发送需要分析的文件
func GoRoutineFourFile(ch chan FourFileChan) {
	for {
		request := <-ch
		if !request.sendRunFlag {
			// 协程停止运行
			break
		}

		analysisFour := results.CreateReferenceFileResult(request.strFile)
		referParam := request.referenceParam
		analysisFour.SetFindReferenceInfo(referParam.fileName, referParam.findSymbol, referParam.secondProjectVec,
			referParam.suffStrVec, referParam.ignoreDefineLoc, referParam.analysisThird)

		request.allProject.handleFindferences(analysisFour)

		chanResult := FourFileChan{
			strFile:    request.strFile,
			findLocVec: analysisFour.FindLocVec,
		}

		ch <- chanResult
	}
}

func recvFourFile(defineVecs *[]DefineStruct, fourFileChan FourFileChan, ignoreDefineLoc lexer.Location) {
	// 所有结构放入进去
	for _, oneLoc := range fourFileChan.findLocVec {
		if !ignoreDefineLoc.IsInitialLoc() && ignoreDefineLoc.IsInLocStruct(oneLoc.StartLine, oneLoc.StartColumn) &&
			ignoreDefineLoc.IsInLocStruct(oneLoc.EndLine, oneLoc.EndColumn) {
			continue
		}

		*defineVecs = append(*defineVecs, DefineStruct{
			StrFile: fourFileChan.strFile,
			Loc:     oneLoc,
		})
	}
}

//  多协程分析所有的文件
func handleAllFilesReference(fileList []string, allProject *AllProject, referenceParam ReferenceParam,
	defineVecs *[]DefineStruct) {
	listLen := len(fileList)
	if listLen == 0 {
		return
	}

	//获取本机核心数
	corNum := runtime.NumCPU() + 2
	if listLen < corNum {
		corNum = listLen
	}

	//协程组的通道
	chs := make([]chan FourFileChan, corNum)
	//创建反射对象集合，用于监听
	selectCase := make([]reflect.SelectCase, corNum)
	//开启协程池
	for i := 0; i < corNum; i++ {
		chs[i] = make(chan FourFileChan)
		selectCase[i].Dir = reflect.SelectRecv
		selectCase[i].Chan = reflect.ValueOf(chs[i])
		go GoRoutineFourFile(chs[i])
	}
	//初始化协程
	for i := 0; i < corNum; i++ {
		chanRequest := FourFileChan{
			sendRunFlag:    true,
			allProject:     allProject,
			strFile:        fileList[i],
			referenceParam: referenceParam,
		}
		chs[i] <- chanRequest
	}

	//reflect接收数据
	taskDone := 0
	for recvNum := 0; recvNum < listLen; {
		chosen, recv, recvOK := reflect.Select(selectCase)
		if !recvOK {
			log.Error("ch%d error\n", chosen)
			//确保循环退出
			recvNum++
			continue
		}
		recvFourFile(defineVecs, recv.Interface().(FourFileChan), referenceParam.ignoreDefineLoc)

		if recvNum+corNum < listLen {
			chanRequest := FourFileChan{
				sendRunFlag:    true,
				allProject:     allProject,
				strFile:        fileList[recvNum+corNum],
				referenceParam: referenceParam,
			}
			chs[chosen] <- chanRequest
		} else {
			chanRequest := FourFileChan{
				sendRunFlag: false,
			}
			chs[chosen] <- chanRequest
		}
		//任务完成数
		taskDone++
		//确保循环退出
		recvNum++
	}
	if taskDone == listLen-1 {
		log.Debug("all files has send work...")
	}
	//关闭所有的子协程
	for i := 0; i < corNum; i++ {
		close(chs[i])
	}
}
