package results

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// CompleteFileResult 第五阶段分析的单个lua文件，查找对变量.， 例如ss.data.one 构造出来层级table，做代码提示
type CompleteFileResult struct {
	StrFile        string              // 文件的名称
	FileName       string              // 引用变量所在的lua文件，如果是本文件的变量，fileName = strFile
	StrName        string              // 查找引用的变量的名称，如果是table为table的名称
	SufVec         []string            // 剩下的前缀
	FindVar        *common.VarInfo     // 查找的变量值
	FindContentMap map[string]struct{} // 找到的所有.符号，查找的结果
	PosLine        int                 // 输入光标的行
	PosCh          int                 // 输入光标的列
}

// CreateAnalysisFiveFile 创建单个引用的指针
func CreateAnalysisFiveFile(strFile string) *CompleteFileResult {
	return &CompleteFileResult{
		StrFile:        strFile,
		FileName:       "",
		FindVar:        nil,
		FindContentMap: map[string]struct{}{},
	}
}

// 匹配全局变量是否是自己要找的
func (c *CompleteFileResult) HandAccessFindInfo(strName string, varInfo *common.VarInfo) bool {
	if c.FindVar == nil || varInfo == nil {
		return false
	}

	if c.StrName != strName {
		return false
	}

	if !lexer.CompareTwoLoc(&c.FindVar.Loc, &varInfo.Loc) {
		return false
	}

	return true
}

// InsertAccessStr 插入匹配到的access
func (c *CompleteFileResult) InsertAccessStr(str string) {
	if str == "" {
		return
	}

	c.FindContentMap[str] = struct{}{}
}
