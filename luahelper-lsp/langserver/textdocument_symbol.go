package langserver

import (
	"context"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/lspcommon"
	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
)

// TextDocumentSymbol 提示文件中生成所有的符合 @使用
func (l *LspServer)TextDocumentSymbol(ctx context.Context, vs lsp.DocumentSymbolParams) (itemsResult []lsp.DocumentSymbol, err error) {
	strFile := pathpre.VscodeURIToString(string(vs.TextDocument.URI))
	project := l.getAllProject()
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return
	}

	// 文件截断后的名字
	fileSymbolVec := project.FindFileAllSymbol(strFile)

	// 将filesymbols 转换为 lsp document
	itemsResult = transferSymbolVec(fileSymbolVec)
	return
}

// transferSymbolVec 转换fileSybmols为DocumentSymbol
func transferSymbolVec(fileSymbolVec []common.FileSymbolStruct) (items []lsp.DocumentSymbol) {
	vecLen := len(fileSymbolVec)
	items = make([]lsp.DocumentSymbol, 0, vecLen)

	for _, oneSymbol := range fileSymbolVec {
		ra := lspcommon.LocToRange(&oneSymbol.Loc)

		fullName := oneSymbol.Name
		if oneSymbol.ContainerName != "" {
			if oneSymbol.ContainerName == "local" {
				fullName = oneSymbol.ContainerName + " " + fullName
			} else {
				fullName = oneSymbol.ContainerName + "." + fullName
			}
		}

		symbol := lsp.DocumentSymbol{
			Name:           fullName,
			Kind:           lsp.Variable,
			Range:          ra,
			SelectionRange: ra,
		}

		if oneSymbol.Children != nil {
			symbol.Children = transferSymbolVec(oneSymbol.Children)
		}
		if oneSymbol.Kind == common.IKAnnotateAlias {
			symbol.Kind = lsp.Interface
			symbol.Detail = "annotate alias"
		} else if oneSymbol.Kind == common.IKAnnotateClass {
			symbol.Kind = lsp.Interface
			symbol.Detail = "annotate class"
		} else if oneSymbol.Kind == common.IKFunction {
			symbol.Kind = lsp.Function
			symbol.Detail = "function"
		} else if len(oneSymbol.Children) != 0 {
			symbol.Kind = lsp.Class
			symbol.Detail = "table"
		} else {
			symbol.Detail = "variable"
		}

		items = append(items, symbol)
	}
	return
}
