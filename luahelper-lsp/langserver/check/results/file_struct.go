package results

import "luahelper-lsp/langserver/check/common"

// FileHandleResult 表示文件处理的结果
type FileHandleResult int

const (
	// results.FileHandleOk 表示成功, 构造AST成功
	FileHandleOk FileHandleResult = 1
	// results.FileHandleReadErr 读文件失败
	FileHandleReadErr FileHandleResult = 2
	// results.FileHandleSyntaxErr 构造AST失败（lua语法分析错误）
	FileHandleSyntaxErr FileHandleResult = 3
)

// FileStruct 以文件为单位，存放第一阶段单个文件分析的所有结构
type FileStruct struct {
	StrFile      string               // 文件的名称
	HandleResult FileHandleResult     // 第一阶段处理的结果，是否有错误
	FileResult   *FileResult          // 第一遍单个lua文件扫描AST的结果
	AnnotateFile *common.AnnotateFile // 这个文件对应注解信息
	IsCommonFile bool                 // 是否是常规工程下的文件
	Contents     []byte               // 文件内容的切片，用于比较文件是否有变动
}

// CreateFileStruct 创建一个新的FileStruct
func CreateFileStruct(strFile string) *FileStruct {
	return &FileStruct{
		StrFile:      strFile,
		HandleResult: FileHandleOk,
		FileResult:   nil,
		AnnotateFile: common.CreateAnnotateFile(strFile),
		IsCommonFile: true,
	}
}

// GetFileHandleErr 获取是否有错误
func (f *FileStruct) GetFileHandleErr() FileHandleResult {
	return f.HandleResult
}