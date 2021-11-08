package langserver

import (
	"context"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/lspcommon"
	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
)

// GetColorParams 全局颜色配置
type GetColorParams struct {
	Uri string `json:"uri"`
}

// IAnnotator 单个颜色
type IAnnotator struct {
	Uri           string      `json:"uri"`
	Ranges        []lsp.Range `json:"ranges"`
	AnnotatorType int         `json:"annotatorType"`
}

// TextDocumentGetVarColor 获取文档中变量的颜色
func (l *LspServer)TextDocumentGetVarColor(ctx context.Context, vs GetColorParams) (annolist []IAnnotator, err error) {
	project := l.getAllProject()

	// 判断打开的文件，是否是需要分析的文件
	strFile := pathpre.VscodeURIToString(string(vs.Uri))
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return
	}

	colorResult := project.FindAllColorVar(strFile)
	allLen := 0
	for _, oneColor := range colorResult {
		allLen = allLen + len(oneColor.LocVec)
	}

	annolist = make([]IAnnotator, 0, allLen)
	for colorType, oneColor := range colorResult {
		oneAnno := IAnnotator{
			Uri:           vs.Uri,
			AnnotatorType: (int)(colorType),
		}

		for _, oneLoc := range oneColor.LocVec {
			oneRange := lspcommon.LocToRange(&oneLoc)
			oneAnno.Ranges = append(oneAnno.Ranges, oneRange)
		}
		annolist = append(annolist, oneAnno)
	}

	return
}

func (l *LspServer)TextDocumentColor(ctx context.Context,colorParams lsp.DocumentColorParams) (colorList []lsp.ColorInformation,err error) {
	log.Debug("not need to handle strFile=%s", colorParams.TextDocument.URI)

	project := l.getAllProject()

	// 判断打开的文件，是否是需要分析的文件
	strFile := pathpre.VscodeURIToString(string(colorParams.TextDocument.URI))
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return
	}

	colorResult := project.FindAllColorVar(strFile)
	allLen := 0
	for _, oneColor := range colorResult {
		allLen = allLen + len(oneColor.LocVec)
	}

	colorList = make([]lsp.ColorInformation, 0, allLen)
	for _, oneColor := range colorResult {
		for _, oneLoc := range oneColor.LocVec {
			oneAnno := lsp.ColorInformation{
				Range: lspcommon.LocToRange(&oneLoc),
				Color: lsp.Color {
					Red:0.5,
					Green:0.5,
					Blue:0.5,
				},
			}
			colorList = append(colorList, oneAnno)
		}
	}

	return colorList, err
}