package check

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"luahelper-lsp/langserver/check/analysis"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/parser"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"reflect"
	"runtime"
	"time"
)

// FirstWorkChan 第一阶段协成的通信
type FirstWorkChan struct {
	strFile          string              // 需要处理的文件
	sendRunFlag      bool                // 是否继续运行
	allProject       *AllProject         // 发送方传递的全局对象指针
	returnFileStruct *results.FileStruct // 协程的结果
	returnChangeFlag bool                // 是否有变动
	saveContentFlag  bool                // 是否cache住文件的内容
}

// analysisFirstLuaFile 进行第一段的分析，主要进行词法和语法的分析，然后构造函数和全局变量信息
// content 为具体传入的内容, 若为nil表示没有传入内容，需要从文件中读取
// saveFlag 为是否cache住文件的内容
// realTimeFlag 实时分析标记，当实时敲入代码时候，该字段为true，不进行一些检查
// handleResult 返回0表示成功， 1表示读文件失败，2表示构造AST失败
// changeFlag 表示内容是否有改动
// beforeStruct 表示changeFlag如果没有改动，返回之前的FileStruct指针
func (allProject *AllProject) analysisFirstLuaFile(f *results.FileStruct, luaFile string, content []byte,
	saveFlag bool, realTimeFlag bool) (handleResult results.FileHandleResult, changeFlag bool, beforeStruct *results.FileStruct) {
	dirManager := common.GConfig.GetDirManager()
	if !dirManager.IsInWorkspace(luaFile) {
		f.IsCommonFile = false
	}

	handleResult = results.FileHandleOk
	f.Contents = nil

	time1 := time.Now()
	if content != nil {
		f.Contents = content
	} else {
		data, err1 := ioutil.ReadFile(luaFile)
		f.Contents = data
		if err1 != nil {
			errStr := fmt.Sprintf("read file=%s error", luaFile)
			log.Debug(errStr)
			handleResult = results.FileHandleReadErr
			changeFlag = true
			return
		}

		beforeStruct, _ = allProject.GetFirstFileStuct(luaFile)
		if beforeStruct != nil && bytes.Equal(beforeStruct.Contents, f.Contents) {
			log.Debug("strFile=%s content is the same", luaFile)
			return beforeStruct.HandleResult, false, beforeStruct
		}
	}

	ftime1 := time.Since(time1).Milliseconds()

	// 获取文件的修改时间
	time2 := time.Now()

	firstFile := results.CreateFileResult(luaFile, nil, results.CheckTermFirst, "")
	// 设置好指向的FileAnalysis
	f.FileResult = firstFile

	newParser := parser.CreateParser(f.Contents, luaFile)
	mainAst, commentMap, errList := newParser.BeginAnalyze()
	if len(errList) > 0 {
		for _, oneErr := range errList {
			firstFile.InsertError(common.CheckErrorSyntax, oneErr.ErrStr, oneErr.Loc)
		}
	}

	// 设置指向的AST
	firstFile.Block = mainAst
	firstFile.CommentMap = commentMap

	// 设置主函数的包含的位置信息
	firstFile.MainFunc.Loc = mainAst.Loc
	firstFile.MainFunc.MainScope.Loc = mainAst.Loc

	ftime2 := time.Since(time2).Milliseconds()
	time3 := time.Now()

	// 是否需要清空内容，节省空间，减少不必要的存储
	if !saveFlag {
		f.Contents = nil
	}

	firstFile.InertNewFunc(firstFile.MainFunc)

	analysis := analysis.CreateAnalysis(results.CheckTermFirst, "")
	analysis.Projects = allProject
	analysis.SetRealTimeFlag(realTimeFlag)

	// 进行第一阶段的遍历
	analysis.HandleFirstTraverseAST(firstFile)

	ftime3 := time.Since(time3).Milliseconds()

	time4 := time.Now()

	// 第一轮遍历完后，进行这个文件的所有注解解析
	f.AnnotateFile.AnalysisAllComment(commentMap)
	f.AnnotateFile.RelateTypeVarInfo(firstFile.GlobalMaps, firstFile.MainFunc.MainScope)
	ftime4 := time.Since(time4).Milliseconds()

	ftime5 := time.Since(time1).Milliseconds()
	log.Debug("handleFirstTraverseAST strFile=%s, readTime=%d, astTime=%d, firstTraTime=%d, annotatetime=%d, alltime=%d",
		luaFile, ftime1, ftime2, ftime3, ftime4, ftime5)
	return handleResult, true, nil
}

// GoRoutineFirstWork 第一轮分析lua的协程，接受主协程发送需要分析的文件
func GoRoutineFirstWork(ch chan FirstWorkChan) {
	for {
		request := <-ch
		if !request.sendRunFlag {
			// 协程停止运行
			break
		}

		fileStruct := results.CreateFileStruct(request.strFile)
		handleResult, changeFlag, beforeFileStruct := request.allProject.analysisFirstLuaFile(fileStruct,
			request.strFile, nil, request.saveContentFlag, false)

		// 如果没有改动，用之前的FileStruct
		if !changeFlag {
			beforeFileStruct.HandleResult = handleResult
		}
		fileStruct.HandleResult = handleResult

		chanResult := FirstWorkChan{
			strFile:          request.strFile,
			returnFileStruct: fileStruct,
			returnChangeFlag: changeFlag,
			saveContentFlag:  request.saveContentFlag,
		}

		ch <- chanResult
	}
}

func (a *AllProject) recvWorkChann(chanResult FirstWorkChan) (changeFlag bool) {
	// 第一阶段的结果插入进去
	changeFlag = chanResult.returnChangeFlag

	// 需要cache主内容
	if chanResult.saveContentFlag && chanResult.returnFileStruct.GetFileHandleErr() == results.FileHandleSyntaxErr {
		// 如果这次有语法错误，获取上一次的成功的文件，进行cache
		// 且cache中如果不存在
		oldStruct, flag := a.GetFirstFileStuct(chanResult.strFile)
		_, cacheFlag, _ := a.fileLRUMap.Get(chanResult.strFile)
		if flag && oldStruct.GetFileHandleErr() != results.FileHandleSyntaxErr && !cacheFlag {
			a.fileLRUMap.Set(chanResult.strFile, oldStruct)
		}
	}

	if changeFlag {
		a.insertFirstFileStruct(chanResult.strFile, chanResult.returnFileStruct)
	} else {
		log.Debug("recvWorkChann strFile=%s not change", chanResult.strFile)
	}

	return changeFlag
}

// firstCreateAndTraverseAst 协程分析所有的文件的一阶段（生成AST以及第一次遍历)
// saveContentFlag 为是否cache住文件的内容
func (a *AllProject) firstCreateAndTraverseAst(filesList []string, saveContentFlag bool) (changeFlag bool) {
	if len(filesList) == 0 {
		return false
	}

	// 是否有文件变化，重新进行了分析
	changeFlag = false
	//获取本机核心数
	corNum := runtime.NumCPU() + 2
	if len(filesList) < corNum {
		corNum = len(filesList)
	}
	// 协程组的通道
	chs := make([]chan FirstWorkChan, corNum)
	//创建反射对象集合，用于监听
	selectCase := make([]reflect.SelectCase, corNum)

	for i := 0; i < corNum; i++ {
		chs[i] = make(chan FirstWorkChan)
		selectCase[i].Dir = reflect.SelectRecv
		selectCase[i].Chan = reflect.ValueOf(chs[i])
		go GoRoutineFirstWork(chs[i])
	}

	//初始化协程
	for i := 0; i < corNum; i++ {
		chanRequest := FirstWorkChan{
			sendRunFlag:     true,
			strFile:         filesList[i],
			allProject:      a,
			saveContentFlag: saveContentFlag,
		}
		chs[i] <- chanRequest
	}

	//reflect接收数据
	taskDone := 0
	for recvNum := 0; recvNum < len(filesList); {
		chosen, recv, recvOK := reflect.Select(selectCase)
		if !recvOK {
			log.Error("ch%d error\n", chosen)
			//确保循环退出
			recvNum++
			continue
		}

		if a.recvWorkChann(recv.Interface().(FirstWorkChan)) {
			changeFlag = true
		}

		if recvNum+corNum < len(filesList) {
			chanRequest := FirstWorkChan{
				sendRunFlag:     true,
				strFile:         filesList[recvNum+corNum],
				allProject:      a,
				saveContentFlag: saveContentFlag,
			}
			chs[chosen] <- chanRequest
		} else {
			chanRequest := FirstWorkChan{
				sendRunFlag: false,
			}
			chs[chosen] <- chanRequest
		}
		taskDone++
		//确保循环退出
		recvNum++
	}
	if taskDone == len(filesList)-1 {
		log.Debug("all files has send work...")
	}
	//关闭所有的子协程
	for i := 0; i < corNum; i++ {
		close(chs[i])
	}
	return changeFlag
}

// HandleFirstAllProject 进行第一轮分析, 所有扫描的lua文件生成AST以及第一遍扫描
func (a *AllProject) HandleFirstAllProject() {
	log.Debug("begin firstCreateAndTraverseAst")
	time1 := time.Now()

	// 1) 协程生成所有文件的AST以及第一遍遍历AST
	var filesList = []string{}
	for fileName := range a.allFilesMap {
		filesList = append(filesList, fileName)
	}

	// 重建
	common.GConfig.RebuildSameFileNameVar(a.allFilesMap)
	// time2 := time.Now()
	// // 创建所有文件的目录结构
	// for strFile := range a.allFilesMap {
	// 	splitVec := strings.Split(strFile, "/")
	// 	vecLen := len(splitVec)

	// 	tmpFileDir := a.allFilesDirStruct
	// 	for index := range splitVec {
	// 		if index == vecLen - 1 {
	// 			if tmpFileDir.FilesMap == nil {
	// 				tmpFileDir.FilesMap = map[string]struct{}{}
	// 			}
	// 			tmpFileDir.FilesMap[strFile] = struct{}{}
	// 			break
	// 		}

	// 		subStr := strings.Join(splitVec[0:index + 1], "/")
	// 		if subFileDir, ok := tmpFileDir.SubDirList[subStr]; ok {
	// 			tmpFileDir = subFileDir
	// 		} else {
	// 			oneFileDir := &common.FileDirSturct{}
	// 			oneFileDir.CurDir = subStr
	// 			if tmpFileDir.SubDirList == nil {
	// 				tmpFileDir.SubDirList =  map[string]*common.FileDirSturct{}
	// 			}

	// 			tmpFileDir.SubDirList[subStr] = oneFileDir
	// 			tmpFileDir = oneFileDir
	// 		}
	// 	}
	// }
	// tc2 := time.Since(time2)
	// ftime2 := tc2.Milliseconds()
	// log.Debug("allFilesDirStruct cost time=%d(ms)", ftime2)

	a.firstCreateAndTraverseAst(filesList, false)

	tc := time.Since(time1)
	ftime := tc.Milliseconds()
	log.Debug("firstCreateAndTraverseAst cost time=%d(ms)", ftime)
}
