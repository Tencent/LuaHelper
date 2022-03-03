package check

import (
	"luahelper-lsp/langserver/check/analysis"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	"reflect"
	"runtime"
	"time"
)

// thirdFileChan 第三阶段协成的检测工程的通信
type thirdFileChan struct {
	sendRunFlag     bool   // 是否继续运行
	strFile         string // 需要处理的文件
	sendThirdStruct *results.AnalysisThird
	allProject      *AllProject
	returnThirdFile *results.AnalysisThirdFile // 协程的结果
}

// goThirdFile 第三轮分析lua单个文件，接受主协程发送需要分析的文件
func goThirdFile(ch chan thirdFileChan) {
	for {
		request := <-ch
		if !request.sendRunFlag {
			// 协程停止运行
			break
		}

		analysisThird := results.CreateAnalysisThirdFile(request.strFile)
		request.allProject.handleOneFile(request.sendThirdStruct, analysisThird)

		chanResult := thirdFileChan{
			strFile:         request.strFile,
			returnThirdFile: analysisThird,
		}

		ch <- chanResult
	}
}

func recvThirdFile(third *results.AnalysisThird, thirdChan thirdFileChan) {
	// 第三阶段的结果插入进去
	if thirdChan.returnThirdFile.FileResult == nil {
		return
	}

	if len(thirdChan.returnThirdFile.FileResult.CheckErrVec) == 0 {
		return
	}

	third.FileErrorMap[thirdChan.strFile] = thirdChan.returnThirdFile.FileResult.CheckErrVec
}

//  进行第三轮分析，散落的文件
func (a *AllProject) handleFiles(third *results.AnalysisThird) {
	var fileList []string
	for strFile := range third.AllFile {
		fileList = append(fileList, strFile)
	}
	listLen := len(fileList)
	if listLen == 0 {
		return
	}

	//获取本机核心数
	corNum := runtime.NumCPU() + 2
	if listLen < corNum {
		corNum = listLen
	}
	// 协程组的通道
	chs := make([]chan thirdFileChan, corNum)
	//创建反射对象集合，用于监听
	selectCase := make([]reflect.SelectCase, corNum)

	for i := 0; i < corNum; i++ {
		chs[i] = make(chan thirdFileChan)
		selectCase[i].Dir = reflect.SelectRecv
		selectCase[i].Chan = reflect.ValueOf(chs[i])
		go goThirdFile(chs[i])
	}

	//初始化协程
	for i := 0; i < corNum; i++ {
		chanRequest := thirdFileChan{
			sendRunFlag:     true,
			strFile:         fileList[i],
			sendThirdStruct: third,
			allProject:      a,
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

		recvThirdFile(third, recv.Interface().(thirdFileChan))

		if recvNum+corNum < listLen {
			chanRequest := thirdFileChan{
				sendRunFlag:     true,
				strFile:         fileList[recvNum+corNum],
				sendThirdStruct: third,
				allProject:      a,
			}
			chs[chosen] <- chanRequest
		} else {
			chanRequest := thirdFileChan{
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

// 扫描散落的文件，包含的所有文件
func (a *AllProject) scanFileIncludeAllFiles(third *results.AnalysisThird, strFile string) {
	strFile = pathpre.GetRemovePreStr(strFile)
	third.AllIncludeFile[strFile] = true

	fileStruct, _ := a.GetFirstFileStuct(strFile)
	if fileStruct == nil {
		log.Error("third file include luafile:%s not exist", strFile)
		return
	}

	if fileStruct.HandleResult != results.FileHandleOk {
		log.Error("second checkOneProject entryluafile:%s check error=%d", strFile,
			fileStruct.HandleResult)
		return
	}

	fileResult := fileStruct.FileResult
	for _, referInfo := range fileResult.ReferVec {
		if !referInfo.Valid || referInfo.ReferValidStr == "" {
			continue
		}

		strFileValid := referInfo.ReferValidStr
		if third.AllIncludeFile[strFileValid] {
			continue
		}

		a.scanFileIncludeAllFiles(third, strFileValid)
	}
}

func (a *AllProject) handleOneFile(thirdStruct *results.AnalysisThird, third *results.AnalysisThirdFile) {
	// 设置散落文件全局对象指针
	third.ThirdStruct = thirdStruct
	log.Debug("thirdfile=%s", third.StrFile)

	strFile := pathpre.GetRemovePreStr(third.StrFile)
	fileStruct, _ := a.GetFirstFileStuct(strFile)
	if fileStruct == nil {
		log.Error("handleOneFile luafile:%s not exist", strFile)
		third.SuccessFlag = false
		return
	}

	if fileStruct.HandleResult != results.FileHandleOk {
		log.Error("handleOneFile luafile:%s check error=%d", strFile, fileStruct.HandleResult)
		third.SuccessFlag = false
		return
	}

	fileResult := fileStruct.FileResult
	if fileResult == nil {
		log.Error("handleOneFile luafile:%s get fileResult error", strFile)
		third.SuccessFlag = false
		return
	}

	time1 := time.Now()

	// 创建第三轮遍历的包裹对象
	analysis := analysis.CreateAnalysis(results.CheckTermThird, third.StrFile)
	analysis.AnalysisThird = third
	analysis.Projects = a

	analysis.HandleTermTraverseAST(results.CheckTermThird, fileResult, nil)
	tc := time.Since(time1)
	ftime := tc.Milliseconds()
	log.Debug("handleOneFile %s, cost time=%d(ms)", strFile, ftime)
}

// HandleNotCheckThirdFile 处理特殊的第三轮非特殊校验的
func (a *AllProject) HandleNotCheckThirdFile() {
	log.Debug("begin HandleNotCheckThirdFile")

	// 清空一些东西
	a.thirdStruct = nil
	time1 := time.Now()
	thirdStruct := results.CreateAnalysisThirdAllStruct()
	a.thirdStruct = thirdStruct

	if len(a.allFilesMap) == 0 {
		log.Debug("allNotIncludeFile num is zore, return")
		return
	}

	for strFile := range a.allFilesMap {
		thirdStruct.AllFile[strFile] = true
		thirdStruct.AllIncludeFile[strFile] = true
	}

	log.Debug("allIncludeFile num: %d", len(thirdStruct.AllFile))

	// 生成第三轮的所有符号
	a.generateAllGlobalMaps(thirdStruct)

	tc := time.Since(time1)
	ftime := tc.Milliseconds()
	log.Debug("handlethirdFiles, filenum:%d, cost time=%d(ms)", len(thirdStruct.AllFile), ftime)
}

// HandleAllThirdFile 进行第三轮分析，主要分析哪些不在工程目录的文件
func (a *AllProject) HandleAllThirdFile() {
	log.Debug("begin HandleAllThirdFile")

	// 清空一些东西
	a.thirdStruct = nil
	time1 := time.Now()

	// 首先找出哪些不在任何工程分析中的文件
	// 所有已经被工程分析过的文件
	allProjectIncludeFile := map[string]bool{}
	for _, analysisSecond := range a.analysisSecondMap {
		for strFile := range analysisSecond.AllFiles {
			allProjectIncludeFile[strFile] = true
		}
	}

	thirdStruct := results.CreateAnalysisThirdAllStruct()
	a.thirdStruct = thirdStruct

	for strFile := range a.allFilesMap {
		if allProjectIncludeFile[strFile] {
			continue
		}

		thirdStruct.AllFile[strFile] = true
	}

	for strFile := range a.clientExpFileMap {
		thirdStruct.AllFile[strFile] = true
	}

	log.Debug("allIncludeFile num: %d, allNotIncludeFile: %d", len(allProjectIncludeFile), len(thirdStruct.AllFile))
	if len(thirdStruct.AllFile) == 0 {
		log.Debug("allNotIncludeFile num is zore, return")
		return
	}

	// 所有散落的文件，组成一个新的符号表，_G的符号表
	for strFile := range thirdStruct.AllFile {
		a.scanFileIncludeAllFiles(thirdStruct, strFile)
	}

	log.Debug("allfiles num:%d, scanFileIncludeAllFiles num:%d", len(thirdStruct.AllFile), len(thirdStruct.AllIncludeFile))
	// 散落的文件，所有隐含加载的所有文件

	// 遍历包括所有的引用加载文件，处理require("one") 放入到全局变量中
	// 只有当为服务器强依赖校验时候，才需要加入进来
	a.generateAllGlobalMaps(thirdStruct)

	// 处理所有散落的文件,多协程的方式
	a.handleFiles(thirdStruct)

	tc := time.Since(time1)
	ftime := tc.Milliseconds()
	log.Debug("handlethirdFiles, filenum:%d, cost time=%d(ms)", len(thirdStruct.AllFile), ftime)
}

// 客户端的模式，文件没有依赖关系，把所有文件的全局符号，导入进来(包括_G的符号)
func (a *AllProject) generateAllGlobalMaps(third *results.AnalysisThird) {
	for strFile := range third.AllIncludeFile {
		fileStruct := a.fileStructMap[strFile]
		if fileStruct == nil {
			log.Debug("fileStruct error, strFile=%s", strFile)
			continue
		}

		// 文件加载失败的忽略
		if fileStruct.HandleResult != results.FileHandleOk {
			continue
		}

		fileResult := fileStruct.FileResult
		for strName, oneVar := range fileResult.GlobalMaps {
			// 先判断之前是否已经存在的该全局信息
			if flag := third.JudgeShouldInsertGlobalInfo(strName, oneVar); !flag {
				continue
			}

			third.InsertThirdGlobalGMaps(strName, oneVar)
		}

		// 查找所有的协议前缀符号表
		for strName, oneVar := range fileResult.ProtocolMaps {
			for {
				if oneVar.ExtraGlobal.GFlag || oneVar.ExtraGlobal.StrProPre != "" {
					third.InsertThirdGlobalGMaps(strName, oneVar)
				}

				if oneVar.ExtraGlobal.Prev == nil {
					break
				}

				oneVar = oneVar.ExtraGlobal.Prev
			}
		}

	}

	for strFile := range third.AllIncludeFile {
		fileStruct := a.fileStructMap[strFile]
		if fileStruct == nil {
			continue
		}

		// 文件加载失败的忽略
		if fileStruct.HandleResult != results.FileHandleOk {
			continue
		}

		fileResult := fileStruct.FileResult
		if fileResult.NodefineMaps == nil {
			continue
		}

		// 查找所有未定义信息
		for strName, oneVar := range fileResult.NodefineMaps {
			if oneVar.SubMaps == nil {
				continue
			}

			ok, gVar := third.FindThirdGlobalGInfo(false, strName, "")
			if !ok {
				continue
			}

			for subName, subVar := range oneVar.SubMaps {
				if !gVar.IsExistMember(subName) {
					gVar.InsertSubMember(subName, subVar)
				}
			}
		}
	}
}
