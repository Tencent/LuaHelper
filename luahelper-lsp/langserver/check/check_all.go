package check

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"sync"
	"time"
)

// 整体检查的入口函数

// AllProject 所有工程包含的内容
type AllProject struct {
	// 所有需要分析的文件map
	allFilesMap map[string]struct{}

	// entryFile string
	// 所有的工程入口分析文件
	entryFilesList []string

	// 插件客户端路径额外扩展的Lua文件夹文件map
	clientExpFileMap map[string]struct{}

	// 第一阶段所有文件的结果
	fileStructMap   map[string]*results.FileStruct
	fileStructMutex sync.Mutex //  第一阶段所有文件的结果的互斥锁

	// 保存vscode客户端实时输入产生无法语法错误的文件结果 *FileStruct，代码提示会优先查找这个结构，这个是最新的
	fileLRUMap *common.LRUCache

	// 第二阶段分析的工程
	analysisSecondMap map[string]*results.SingleProjectResult

	// 第三阶段分析的全局结果
	thirdStruct *results.AnalysisThird

	// 管理所有的注释创建的type类型，key值为名称，value是这个类型的列表，允许多个存在
	createTypeMap map[string]common.CreateTypeList

	// 代码补全cache
	completeCache *common.CompleteCache

	// 整体分析的阶段数
	checkTerm results.CheckTerm
}


// CreateAllProject 创建整个检查工程
func CreateAllProject(allFilesList []string, entryFileArr []string, clientExpPathList []string) *AllProject {
	// 第一阶段（生成AST，第一次遍历AST），用多协程分析所有的扫描出来的文件
	allProject := &AllProject{
		allFilesMap:       map[string]struct{}{},
		entryFilesList:    entryFileArr,
		clientExpFileMap:  map[string]struct{}{},
		fileStructMap:     map[string]*results.FileStruct{},
		analysisSecondMap: map[string]*results.SingleProjectResult{},
		thirdStruct:       nil,
		createTypeMap:     map[string]common.CreateTypeList{},
		checkTerm:         results.CheckTermFirst,
		completeCache:     common.CreateCompleteCache(),
		fileLRUMap:        common.NewLRUCache(20),
	}

	// 传人的所有文件列表转换成map
	for _, fileName := range allFilesList {
		allProject.allFilesMap[fileName] = struct{}{}
	}

	// 传人的插件客户端额外的文件列表转换成map
	for _, fileName := range clientExpPathList {
		allProject.clientExpFileMap[fileName] = struct{}{}
	}

	return allProject
}

// HandleCheck 进行分析检查
func (a *AllProject) HandleCheck() {
	time1 := time.Now()

	// 0) 清除掉文件的cache
	common.GConfig.ClearCacheFileMap()

	a.setCheckTerm(results.CheckTermFirst)
	// 1) 进行第一轮分析, 所有扫描的lua文件生成AST以及第一遍扫描
	a.HandleFirstAllProject()

	ftime1 := time.Since(time1).Milliseconds()

	var ftime2 int64 = 0
	var ftime3 int64 = 0

	dirManager := common.GConfig.GetDirManager()
	mainDir := dirManager.GetMainDir()

	// 判断是否要进行特殊的检测
	if len(a.entryFilesList) == 0 && !common.GConfig.IsSpecialCheck() {
		if mainDir != "" {
			// 进行特殊的第三轮非检查，处理生成符号
			time3 := time.Now()
			a.setCheckTerm(results.CheckTermThird)
			a.HandleNotCheckThirdFile()
			ftime3 = time.Since(time3).Milliseconds()
		}
	} else {
		// 若没有打开目录，则不进行第二阶段和第三阶段分析
		if mainDir != "" {
			time2 := time.Now()
			a.setCheckTerm(results.CheckTermSecond)
			// 2) 进行第二轮分析，以入口工程的文件进行分析
			a.HandleAllSecondProject()
			ftime2 = time.Since(time2).Milliseconds()

			time3 := time.Now()
			a.setCheckTerm(results.CheckTermThird)
			// 3) 进行第三轮分析，主要分析不在工程中的散落文件
			a.HandleAllThirdFile()
			ftime3 = time.Since(time3).Milliseconds()
		}
	}

	// 4) 重新创建所有的createTypeMap 注释类型
	a.rebuidCreateTypeMap()
	a.checkAllAnnotate()

	ftime := time.Since(time1).Milliseconds()
	log.Debug("HandleCheck,  all time=%d, first=%d, second=%d, third=%d", ftime, ftime1, ftime2, ftime3)
}

//  重新创建所有的createTypeMap 注释类型
func (a *AllProject) rebuidCreateTypeMap() {
	time1 := time.Now()

	// 清空之前的
	a.createTypeMap = map[string]common.CreateTypeList{}

	// 遍历所有文件的注释类型，整合成一个整体
	for _, fileStruct := range a.fileStructMap {
		for strName, createTypeList := range fileStruct.AnnotateFile.CreateTypeMap {
			typeList, ok := a.createTypeMap[strName]
			if ok {
				typeList.List = append(typeList.List, createTypeList.List...)
				a.createTypeMap[strName] = typeList
			} else {
				a.createTypeMap[strName] = createTypeList
			}
		}
	}

	tc := time.Since(time1)
	ftime := tc.Milliseconds()
	log.Debug("rebuidCreateTypeMap time:%d", ftime)
}

// GetAllFilesMap 获取分析的文件map列表
func (a *AllProject) GetAllFilesMap() (allFilesMap map[string]struct{}) {
	return a.allFilesMap
}

// IsInAllFilesMap 指定的文件，是否在已经分析的map列表中
func (a *AllProject) IsInAllFilesMap(strFile string) bool {
	_, flag := a.allFilesMap[strFile]
	return flag
}

// GetAllFileNumber 获取工程中所有lua文件的数量
func (a *AllProject) GetAllFileNumber() int {
	return len(a.allFilesMap)
}

// RemoveFile 删除一个文件
func (a *AllProject) RemoveFile(strFile string) {
	// 1) 所有需要分析的文件，删除它
	delete(a.allFilesMap, strFile)

	// 2)删除第一阶段的文件指针
	_, beforeExitFlag := a.GetFirstFileStuct(strFile)
	delete(a.fileStructMap, strFile)
	_, endExitFlag := a.GetFirstFileStuct(strFile)

	// 3) 删除cache的内容
	a.RemoveCacheContent(strFile)
	log.Debug("delete strFile=%s, beforeFlag=%t, endFlag=%t", strFile, beforeExitFlag, endExitFlag)
}
