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

// SecondProjectChan 第二阶段协成的检测工程的通信
type SecondProjectChan struct {
	sendRunFlag  bool                         // 是否继续运行
	strFile      string                       // 需要处理的工程入口文件
	allProject   *AllProject                  // 全局处理指针，方便所有到第一阶段的数据
	returnSecond *results.SingleProjectResult // 协程的结果
}

// goSecondProject 第二轮分析lua工程的协程，接受主协程发送需要分析的文件
func goSecondProject(ch chan SecondProjectChan) {
	for {
		request := <-ch
		if !request.sendRunFlag {
			// 协程停止运行
			break
		}

		analysisSecond := results.CreateAnalysisSecondProject(request.strFile)
		request.allProject.checkOneProject(analysisSecond)

		chanResult := SecondProjectChan{
			strFile:      request.strFile,
			returnSecond: analysisSecond,
		}

		ch <- chanResult
	}
}

func (a *AllProject) recvSecondProject(projectChan SecondProjectChan) {
	log.Debug("recvSecondProject file=%s", projectChan.strFile)
	// 第二阶段的结果插入进去
	// 先删除之前的
	delete(a.analysisSecondMap, projectChan.strFile)
	projectChan.returnSecond.FinishSecondProject()
	a.analysisSecondMap[projectChan.strFile] = projectChan.returnSecond
}

// 传入工程入口文件vec，对这些工程入口文件进行分析
func (a *AllProject) handleProjectEntryFileVec(projectVec []string) {
	vecLen := len(projectVec)
	if vecLen == 0 {
		return
	}

	//获取本机核心数
	corNum := runtime.NumCPU() + 2
	if vecLen < corNum {
		corNum = vecLen
	}
	// 协程组的通道
	chs := make([]chan SecondProjectChan, corNum)
	//创建反射对象集合，用于监听
	selectCase := make([]reflect.SelectCase, corNum)

	for i := 0; i < corNum; i++ {
		chs[i] = make(chan SecondProjectChan)
		selectCase[i].Dir = reflect.SelectRecv
		selectCase[i].Chan = reflect.ValueOf(chs[i])
		go goSecondProject(chs[i])
	}

	//初始化协程
	for i := 0; i < corNum; i++ {
		chanRequest := SecondProjectChan{
			sendRunFlag: true,
			strFile:     projectVec[i],
			allProject:  a,
		}
		chs[i] <- chanRequest
	}

	//reflect接收数据
	taskDone := 0
	for recvNum := 0; recvNum < vecLen; {
		chosen, recv, recvOK := reflect.Select(selectCase)
		if !recvOK {
			log.Error("ch%d error\n", chosen)
			//确保循环退出
			recvNum++
			continue
		}

		a.recvSecondProject(recv.Interface().(SecondProjectChan))

		if recvNum+corNum < vecLen {
			chanRequest := SecondProjectChan{
				sendRunFlag: true,
				strFile:     projectVec[recvNum+corNum],
				allProject:  a,
			}
			chs[chosen] <- chanRequest
		} else {
			chanRequest := SecondProjectChan{
				sendRunFlag: false,
			}
			chs[chosen] <- chanRequest
		}
		//任务完成数
		taskDone++
		//确保循环退出
		recvNum++
	}
	if taskDone == vecLen-1 {
		log.Debug("all files has send work...")
	}
	//关闭所有的子协程
	for i := 0; i < corNum; i++ {
		close(chs[i])
	}
}

// HandleAllSecondProject 进行第二轮分析，以入口工程的文件进行分析
func (a *AllProject) HandleAllSecondProject() {
	log.Debug("begin handleAllSecondProject")
	time1 := time.Now()

	// 对所有的工程进行分析
	a.handleProjectEntryFileVec(a.entryFilesList)

	tc := time.Since(time1)
	ftime := tc.Milliseconds()
	log.Debug("handleAllSecondProject, projectnum:%d, cost time=%d(ms)", len(a.entryFilesList), ftime)
}

// 第一阶段进行完毕后，扫描所有文件第一阶段结构文件，一次性生成完整的_G的全局符号表
func (a *AllProject) generateAllFristGlobalGMaps(second *results.SingleProjectResult) {
	for strFile := range second.AllFiles {
		fileStruct := a.fileStructMap[strFile]
		if fileStruct == nil {
			log.Debug("fileStruct error, strFile=%s", strFile)
			continue
		}

		_, cientExpFlag := a.clientExpFileMap[strFile]

		// 文件加载失败的忽略
		if fileStruct.HandleResult != results.FileHandleOk {
			continue
		}

		fileResult := fileStruct.FileResult

		for strName, varInfo := range fileResult.GlobalMaps {
			if !cientExpFlag && !varInfo.ExtraGlobal.GFlag && varInfo.ExtraGlobal.StrProPre == "" {
				continue
			}

			second.InsertGlobalGMaps(strName, varInfo, results.CheckTermFirst)
		}

		// 查找所有的协议前缀符号表
		for strName, varInfo := range fileResult.ProtocolMaps {
			for {
				if varInfo.ExtraGlobal.GFlag || varInfo.ExtraGlobal.StrProPre != "" {
					second.InsertGlobalGMaps(strName, varInfo, results.CheckTermFirst)
				}

				if varInfo.ExtraGlobal.Prev == nil {
					break
				}

				varInfo = varInfo.ExtraGlobal.Prev
			}
		}
	}
}

// 第一阶段扫描完毕后，扫描所有的require("one") --one为one.lua
// 所有的one.lua全局符号表加载到_G符号表中
func (a *AllProject) generateRequireFileGlobalGmaps(second *results.SingleProjectResult, analysis *analysis.Analysis) {
	for strFile := range second.AllFiles {
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

		// 遍历该文件的所有引用，判断是否要把引用到的其他文件全局符号表加入进来
		for _, referInfo := range fileResult.ReferVec {
			// 第一轮中require加载的_G符号
			analysis.InsertRequireInfoGlobalVars(referInfo, results.CheckTermFirst)
		}
	}
}

// 插入由于插件客户端提供的额外文件夹引入的全局符号, 插入到第二轮全局符号表中
func (a *AllProject) generateClientExpLuaSecond(second *results.SingleProjectResult, analysis *analysis.Analysis) {
	for strFile := range a.clientExpFileMap {
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

		for strName, varInfo := range fileResult.GlobalMaps {
			// if !varInfo.ExtraGlobal.GFlag && varInfo.ExtraGlobal.StrProPre == "" {
			// 	continue
			// }

			second.InsertGlobalGMaps(strName, varInfo, results.CheckTermSecond)
		}
	}
}

// 加入插件客户端额外提供的lua文件夹文件
func (a *AllProject) addClientExpLuaFile(second *results.SingleProjectResult) {
	for strFile := range a.clientExpFileMap {
		second.AllFiles[strFile] = struct{}{}
	}
}

// 只扫描工程中所有加载的文件
func (a *AllProject) scanProjectAllFiles(second *results.SingleProjectResult, strFile string) {
	strFile = pathpre.GetRemovePreStr(strFile)
	second.AllFiles[strFile] = struct{}{}

	fileStruct, _ := a.GetFirstFileStuct(strFile)
	if fileStruct == nil {
		log.Error("second checkOneProject entryluafile:%s not exist", second.EntryFile)
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
		if _, ok := second.AllFiles[strFileValid]; ok {
			continue
		}

		a.scanProjectAllFiles(second, strFileValid)
	}
}

func (a *AllProject) checkOneProject(second *results.SingleProjectResult) {
	log.Debug(" entryfile=%s", second.EntryFile)

	time1 := time.Now()
	strFile := pathpre.GetRemovePreStr(second.EntryFile)
	fileStruct, _ := a.GetFirstFileStuct(strFile)
	if fileStruct == nil {
		log.Error("second checkOneProject entryluafile:%s not exist", strFile)
		return
	}

	if fileStruct.HandleResult != results.FileHandleOk {
		log.Error("second checkOneProject entryluafile:%s check error=%d", strFile,
			fileStruct.HandleResult)
		return
	}

	fileResult := fileStruct.FileResult
	if fileResult == nil {
		log.Error("second checkOneProject entryluafile:%s get fileResult error", strFile)
		return
	}

	// 加入插件客户端额外提供的文件夹文件
	a.addClientExpLuaFile(second)

	// 扫描工程中加载的所有文件
	a.scanProjectAllFiles(second, strFile)

	//第一阶段进行完毕后，扫描所有文件第一阶段结构文件，一次性生成完整的_G的全局符号表
	a.generateAllFristGlobalGMaps(second)

	// 创建第二轮遍历的包裹对象
	analysis := analysis.CreateAnalysis(results.CheckTermSecond, strFile)
	analysis.SingleProjectResult = second
	analysis.Projects = a

	// 插入第一阶段的全局符号表
	a.generateRequireFileGlobalGmaps(second, analysis)

	// 插入由于插件客户端提供的额外文件夹引入的全局符号, 插入到第二轮全局符号表中
	a.generateClientExpLuaSecond(second, analysis)

	analysis.HandleSecondProjectTraverseAST(fileResult, nil)

	a.handleOtherFileInsertSub(second)

	ftime := time.Since(time1).Milliseconds()
	log.Debug("checkOneProject %s, filenum=%d, cost time=%d(ms)", second.EntryFile, len(second.AllFiles), ftime)
}

// 处理其他文件给全局变量增加的子成员
func (a *AllProject) handleOtherFileInsertSub(second *results.SingleProjectResult) {
	for strFile := range second.AllFiles {
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

			ok, gVar := second.FindGlobalGInfo(strName, results.CheckTermFirst, "")
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
