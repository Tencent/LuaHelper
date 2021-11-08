package results

import "luahelper-lsp/langserver/check/common"

// SingleProjectResult 第二阶段分析的工程结构
type SingleProjectResult struct {
	EntryFile         string                         // 工程的入口文件
	AllFiles          map[string]struct{}            // 该工程包含的所有的lua文件
	FirstGlobalGMaps  map[string]*common.VarInfoList // 第一阶段进行完毕后，扫描所有文件第一阶段结构文件，一次性生成完整的_G的全局符号表
	FirstProtcolMaps  map[string]*common.VarInfoList // 第一阶段进行完毕后, 扫描所有的第一阶段结构文件，协议前缀符号表
	AnalysisFileMap   map[string]*FileResult         // 第二轮单个文件分析完的结果
	secondProtcolMaps map[string]*common.VarInfoList // 第二阶段所有的全局信息,第二阶段所有的全局信息,
	secondGlobalGMaps map[string]*common.VarInfoList // 第二阶段所有的全局信息, 有_G前缀;同一个key的全局信息可能有多个,全局map信息按执行流动态增加
	FileErrorMap      map[string][]common.CheckError // 第二阶段所有的分析错误信息，转存在这里

	// 第一阶段扫描完毕后，引入其他文件require("one") 或dofile("one.lua") 会加载到_G全局符号表中
	// 当前已经执行的require加载的文件， 例如require("one") ，map中存放的是one.lua
	FirstRequireFileMap  map[string]bool
	SecondRequireFileMap map[string]bool
}

// CreateAnalysisSecondProject 创建第二阶段分析的结果指针
func CreateAnalysisSecondProject(entryFile string) *SingleProjectResult {
	return &SingleProjectResult{
		EntryFile:            entryFile,
		AllFiles:             map[string]struct{}{},
		FirstGlobalGMaps:     map[string]*common.VarInfoList{},
		FirstProtcolMaps:     map[string]*common.VarInfoList{},
		AnalysisFileMap:      map[string]*FileResult{},
		secondGlobalGMaps:    map[string]*common.VarInfoList{},
		secondProtcolMaps:    map[string]*common.VarInfoList{},
		FileErrorMap:         map[string][]common.CheckError{},
		FirstRequireFileMap:  map[string]bool{},
		SecondRequireFileMap: map[string]bool{},
	}
}

// checkTerm 1 第一阶段进行完毕，第一阶段所有的全局符号表中插入一个新的全局变量信息
// checkTerm 2 表示第二阶段工程check中，每解析到一个_G的全局变量，放入到第二阶段全局map中
func (s *SingleProjectResult) InsertGlobalGMaps(strName string, varInfo *common.VarInfo, checkTerm CheckTerm) {
	var globalGmaps map[string]*common.VarInfoList
	strProPre := varInfo.ExtraGlobal.StrProPre
	if checkTerm == CheckTermFirst {
		globalGmaps = s.FirstGlobalGMaps
		if strProPre != "" {
			globalGmaps = s.FirstProtcolMaps
		}
	} else if checkTerm == CheckTermSecond {
		globalGmaps = s.secondGlobalGMaps
		if strProPre != "" {
			globalGmaps = s.secondProtcolMaps
		}
	} else {
		return
	}

	// 如果是第二轮工程的check，_G的全局符号放入到工程的结构中
	varInfoList := globalGmaps[strName]
	if varInfoList == nil {
		varInfoList := &common.VarInfoList{
			VarVec: make([]*common.VarInfo, 0, 1),
		}
		// 之前这个strName _G的符号不存在，创建
		varInfoList.VarVec = append(varInfoList.VarVec, varInfo)
		globalGmaps[strName] = varInfoList
	} else {
		// 之前这个strName _G的符号已经存在，插入到vec中
		varInfoList.VarVec = append(varInfoList.VarVec, varInfo)
	}
}

// 在第一阶段或第二阶段_G全局map变量中，查找变量的定义以及指向
// checkTerm 1表示第一阶段，checkTerm 2表示第二阶段
// strProPre 为协议的前缀，例如c2s. s2s
func (s *SingleProjectResult) FindGlobalGInfo(strName string, checkTerm CheckTerm, strProPre string) (bool, *common.VarInfo) {
	var globalGmaps map[string]*common.VarInfoList
	if checkTerm == CheckTermFirst {
		globalGmaps = s.FirstGlobalGMaps
		if strProPre != "" {
			globalGmaps = s.FirstProtcolMaps
		}
	} else if checkTerm == CheckTermSecond {
		globalGmaps = s.secondGlobalGMaps
		if strProPre != "" {
			globalGmaps = s.secondProtcolMaps
		}
	} else {
		return false, nil
	}

	varInfoList := globalGmaps[strName]
	if varInfoList == nil {
		return false, nil
	}

	if len(varInfoList.VarVec) == 0 {
		return false, nil
	}

	for i := len(varInfoList.VarVec) - 1; i >= 0; i-- {
		oneVar := varInfoList.VarVec[i]
		if oneVar.ExtraGlobal.StrProPre == strProPre {
			return true, oneVar
		}
	}

	return false, nil
}

// 第二阶段分析完了，把文件分析的错误转存在专用结构中，销毁不用的数据
func (s *SingleProjectResult) FinishSecondProject() {
	// 1) 转成所有的错误
	for strFile, fileResult := range s.AnalysisFileMap {
		if len(fileResult.CheckErrVec) == 0 {
			continue
		}

		s.FileErrorMap[strFile] = fileResult.CheckErrVec
	}

	s.AnalysisFileMap = map[string]*FileResult{}
	s.FirstRequireFileMap = map[string]bool{}
	s.SecondRequireFileMap = map[string]bool{}
}
