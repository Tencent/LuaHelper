package results

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
)

// AnalysisThird 第三阶段分析的所有的单个lua文件，散落的包装的结构
type AnalysisThird struct {
	AllFile        map[string]bool                // 所有散落的lua文件
	AllIncludeFile map[string]bool                // 所有散落的lua文件， 包括引入的所有文件
	GlobalVarMaps  map[string]*common.VarInfoList // 所有文件忽略依赖关系，所有的文件的全局变量，包括_G和非_G的放入到map中
	ProtcolVarMaps map[string]*common.VarInfoList // 第一阶段进行完毕后, 协议前缀符号

	// 第一阶段扫描完毕后，引入其他文件require("one") 或dofile("one.lua") 会加载到_G全局符号表中
	// 当前已经执行的require加载的文件， 例如require("one") ，map中存放的是one.lua
	FileErrorMap map[string][]common.CheckError // 第三阶段所有的分析错误信息，转存在这里
}

// CreateAnalysisThirdAllStruct 创建第三阶段的结构
func CreateAnalysisThirdAllStruct() *AnalysisThird {
	return &AnalysisThird{
		AllFile:        map[string]bool{},
		AllIncludeFile: map[string]bool{},
		GlobalVarMaps:  map[string]*common.VarInfoList{},
		ProtcolVarMaps: map[string]*common.VarInfoList{},
		FileErrorMap:   map[string][]common.CheckError{},
	}
}

// AnalysisThirdFile 第三阶段分析的单个lua文件，该文件不属于工程
type AnalysisThirdFile struct {
	StrFile     string         // 文件的名称
	FileResult  *FileResult    // 单个文件分析的指针
	ThirdStruct *AnalysisThird // 散落的文件，公共的结果
	SuccessFlag bool           // 是否处理成功
}

// CreateAnalysisThirdFile 创建单个文件的指针
func CreateAnalysisThirdFile(strFile string) *AnalysisThirdFile {
	return &AnalysisThirdFile{
		StrFile:     strFile,
		FileResult:  nil,
		ThirdStruct: nil,
		SuccessFlag: true,
	}
}

// checkTerm 1 第一阶段进行完毕，第一阶段所有的全局符号表中插入一个新的全局变量信息
// checkTerm 2 表示第二阶段工程check中，每解析到一个_G的全局变量，放入到第二阶段全局map中
func (third *AnalysisThird) InsertThirdGlobalGMaps(strName string, varInfo *common.VarInfo) {
	globalGmaps := third.GlobalVarMaps
	if varInfo.ExtraGlobal.StrProPre != "" {
		globalGmaps = third.ProtcolVarMaps
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

// 判断是否应该把新的global信息插入进来， 考虑到多个文件有定义相同的全局符号，主要比较定义全局的函数层数或行数
func (third *AnalysisThird) JudgeShouldInsertGlobalInfo(strName string, varInfo *common.VarInfo) bool {
	globalGmaps := third.GlobalVarMaps

	// 如果是第二轮工程的check，_G的全局符号放入到工程的结构中
	varInfoList := globalGmaps[strName]
	if varInfoList == nil {
		// 之前没有出现应该直接插入
		return true
	}

	// 之前已经存在了，比较这次插入的是否应该
	for _, oneVar := range varInfoList.VarVec {
		if oneVar.ExtraGlobal.FileName == varInfo.ExtraGlobal.FileName {
			// 如果是在同一个文件忽略
			continue
		}

		if oneVar.ExtraGlobal.FuncLv < varInfo.ExtraGlobal.FuncLv {
			log.Debug("thirdStruct strName=%s, not insert, before file=%s, funcLv=%d,now file=%s, funcLv=%d",
				strName, oneVar.ExtraGlobal.FileName, oneVar.ExtraGlobal.FuncLv, varInfo.ExtraGlobal.FileName, varInfo.ExtraGlobal.FuncLv)
			return false
		}

		if oneVar.ExtraGlobal.ScopeLv < varInfo.ExtraGlobal.ScopeLv {
			log.Debug("thirdStruct strName=%s, not insert, before file=%s, ScopeLv=%d,now file=%s, ScopeLv=%d",
				strName, oneVar.ExtraGlobal.FileName, oneVar.ExtraGlobal.ScopeLv, varInfo.ExtraGlobal.FileName, varInfo.ExtraGlobal.ScopeLv)
			return false
		}

		if oneVar.Loc.StartLine <= varInfo.Loc.StartLine {
			// log.Debug("thirdStruct strName=%s, not insert, before file=%s, line=%d,now file=%s, line=%d",
			// 	strName, oneVar.ExtraGlobal.FileName, oneVar.Loc.StartLine, varInfo.ExtraGlobal.FileName,
			// 	varInfo.Loc.StartLine)
			return false
		}
	}

	return true
}

// 向第三阶段公共结构中，保存的全局指针，获取信息
// gFlag 表示是否只查找_G前缀变量的
func (third *AnalysisThird) FindThirdGlobalGInfo(gFlag bool, strName string, strProPre string) (bool, *common.VarInfo) {
	globalGmaps := third.GlobalVarMaps
	if strProPre != "" {
		globalGmaps = third.ProtcolVarMaps
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
