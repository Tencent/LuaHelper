package projects

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/results"
)

// 定义analysis访问check的接口

// Projects 接口
type Projects interface{
	// GetCacheFileStruct 获取第一阶段的文件结构，如果有cache，优先获取cache的内容，cache的会相对新
	GetCacheFileStruct(strFile string) (*results.FileStruct, bool)

	// 如果为自定义的引入其他lua文件的方式，获取真实的引入的子类型
	GetReferFrameType(referInfo *common.ReferInfo) (subReferType common.ReferFrameType)

	// getFirstFileStuct 获取第一阶段文件处理的结果
 	GetFirstFileStuct(strFile string) (*results.FileStruct, bool) 

	// getFirstReferFileResult 给一个引用关系，找第一阶段引用lua文件
 	GetFirstReferFileResult(referInfo *common.ReferInfo) *results.FileResult 

	// GetAllFilesMap 获取所有的文件map
	GetAllFilesMap() map[string]struct{}
}