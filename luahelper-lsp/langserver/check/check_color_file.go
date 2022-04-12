package check

import (
	"luahelper-lsp/langserver/check/analysis"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	"time"
)

// FindAllColorVar 查找文件中的所有全局变量
func (a *AllProject) FindAllColorVar(strFile string) (colorResult map[common.ColorType]*common.OneColorResut) {
	// 如果为局部变量，只在当前文件查找
	color := results.CreateColorFileInfo(strFile)

	// 1) 获取变量的着色
	a.handleFindVarColor(color)
	if color.FileResult == nil {
		return
	}

	// 当前文件的所有全局变量定义，也插入其中
	for _, oneVar := range color.FileResult.GlobalMaps {
		color.InsertOneGlobalColor(oneVar, &(oneVar.Loc))
	}

	// 2) 获取这个文件静态的注解产生的着色, 临时把这个颜色放入到全局变量着色里面
	// anntotateLocVec := a.getAnnotateColor(strFile)
	// if len(anntotateLocVec) > 0 {
	// 	oneColor := color.ColorResult[common.CTAnnotate]
	// 	if oneColor == nil {
	// 		oneColor := &common.OneColorResut{
	// 			LocVec: make([]lexer.LocInfo, 0, 1),
	// 		}
	// 		oneColor.LocVec = anntotateLocVec
	// 		color.ColorResult[common.CTAnnotate] = oneColor
	// 	} else {
	// 		oneColor.LocVec = append(oneColor.LocVec, anntotateLocVec...)
	// 	}
	// }

	colorResult = color.ColorResult
	return
}

func (a *AllProject) handleFindVarColor(color *results.ColorFileResult) {
	strFile := pathpre.GetRemovePreStr(color.StrFile)
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("handleOneFile file not valid file=%s", strFile)
		return
	}
	fileResult := fileStruct.FileResult

	time1 := time.Now()

	// 创建第三轮遍历的包裹对象
	analysis := analysis.CreateAnalysis(results.CheckTermFive, color.StrFile)
	analysis.ColorResult = color
	analysis.Projects = a
	analysis.HandleTermTraverseAST(results.CheckTermFive, fileResult, nil)

	ftime := time.Since(time1).Milliseconds()
	log.Debug("handleFindVarColor handleOneFile %s, cost time=%d(ms)", strFile, ftime)
}
