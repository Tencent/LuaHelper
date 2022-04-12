package results

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// ColorFileResult 第5阶段，分析单个文件，获取全局变量着色的功能
type ColorFileResult struct {
	StrFile     string                                     // 文件的名称
	FileResult  *FileResult                                // 单个文件分析的指针
	ColorResult map[common.ColorType]*common.OneColorResut // 保存找到的文件中，所有的颜色数据
}

// CreateColorFileInfo 创建第五阶段的文件分析指针
func CreateColorFileInfo(strFile string) *ColorFileResult {
	return &ColorFileResult{
		StrFile:     strFile,
		FileResult:  nil,
		ColorResult: map[common.ColorType]*common.OneColorResut{},
	}
}

// InsertOneColorElem 插入一个找到的全局信息
func (c *ColorFileResult) InsertOneColorElem(color common.ColorType, loc *lexer.Location) {
	// 如果是第二轮工程的check，_G的全局符号放入到工程的结构中
	oneColor := c.ColorResult[color]
	if oneColor == nil {
		oneColor := &common.OneColorResut{
			LocVec: make([]lexer.Location, 0, 1),
		}
		oneColor.LocVec = append(oneColor.LocVec, *loc)
		c.ColorResult[color] = oneColor
	} else {
		oneColor.LocVec = append(oneColor.LocVec, *loc)
	}
}

// 插入一个全局变量，进行判断
func (color *ColorFileResult) InsertOneGlobalColor(varInfo *common.VarInfo, loc *lexer.Location) {
	if varInfo == nil {
		return
	}

	if varInfo.ReferFunc != nil {
		color.InsertOneColorElem(common.CTGlobalFunc, loc)
	} else {
		color.InsertOneColorElem(common.CTGlobalVar, loc)
	}
}
