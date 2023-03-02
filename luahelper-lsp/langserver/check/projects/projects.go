package projects

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/results"
)

// 定义analysis访问check的接口

// Projects 接口
type Projects interface {
	// GetCacheFileStruct 获取第一阶段的文件结构，如果有cache，优先获取cache的内容，cache的会相对新
	GetCacheFileStruct(strFile string) (*results.FileStruct, bool)

	// 如果为自定义的引入其他lua文件的方式，获取真实的引入的子类型
	GetReferFrameType(referInfo *common.ReferInfo) (subReferType common.ReferFrameType)

	// getFirstFileStuct 获取第一阶段文件处理的结果
	GetFirstFileStuct(strFile string) (*results.FileStruct, bool)

	// getFirstReferFileResult 给一个引用关系，找第一阶段引用lua文件
	GetFirstReferFileResult(referInfo *common.ReferInfo) *results.FileResult

	// GetAllFilesMap 获取所有的文件map
	GetAllFilesMap() map[string]string

	// GetFileIndexInfo 获取文件的缓存结构
	GetFileIndexInfo() *common.FileIndexInfo

	// GetAllFilesPreStrMap 获取所有文件的前置路径map
	//GetAllFilesPreStrMap() map[string]string

	// GetFuncDefaultParamInfo 在函数注解中获取默认参数标记
	GetFuncDefaultParamInfo(fileName string, lastLine int, paramNameList []string) (paramDefaultNum int)

	IsMemberOfAnnotateClassByVar(strMemName string, strVarName string, varInfo *common.VarInfo) (isMember bool, className string)

	IsMemberOfAnnotateClassByLoc(strFile string, strFieldNamelist []string, lineForGetAnnotate int) (isMemberMap map[string]bool, className string)

	IsAnnotateTypeConst(name string, varInfo *common.VarInfo) (isConst bool)

	GetAnnotateTypeString(varInfo *common.VarInfo, varName string, keyName string, idx int) (retVec []string)

	GetFuncParamType(fileName string, lastLine int) (retMap map[string][]annotateast.Type)

	GetFuncParamTypeByClass(className string, funcName string) (retMap map[string][]string)
	GetFuncReturnTypeByClass(className string, funcName string) (retVec [][]string)

	GetFuncReturnInfo(fileName string, lastLine int) (paramInfo *common.FragementReturnInfo)

	GetFuncReturnType(fileName string, lastLine int) (retVec [][]annotateast.Type)

	GetFuncReturnTypeVec(fileName string, lastLine int) (retVec [][]string)

	GetAnnClassInfo(className string) *common.CreateTypeInfo

	IsFieldOfClass(className string, fieldName string) bool

	GetVarAnnType(fileName string, lastLine int) (string, bool)
}
