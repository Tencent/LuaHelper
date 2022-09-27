package langserver

import (
	"context"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/lspcommon"
	lsp "luahelper-lsp/langserver/protocol"
)

// WorkspaceSymbolRequest 全工程符合查找提示，返回多个符合
func (l *LspServer) WorkspaceSymbolRequest(ctx context.Context, vs lsp.WorkspaceSymbolParams) (items []lsp.SymbolInformation, err error) {
	project := l.getAllProject()
	fileSymbolVec := project.FindWorkspaceAllSymbol(vs.Query)

	vecLen := len(fileSymbolVec)
	items = make([]lsp.SymbolInformation, 0, vecLen)

	for _, oneSymbol := range fileSymbolVec {
		loc := lsp.Location{
			URI:   lspcommon.GetFileDocumentURI(oneSymbol.FileName),
			Range: lspcommon.LocToRange(&oneSymbol.Loc),
		}

		item := lsp.SymbolInformation{
			Name:          oneSymbol.Name,
			Kind:          lsp.Variable,
			Location:      loc,
			ContainerName: oneSymbol.ContainerName,
		}

		if oneSymbol.Kind == common.IKFunction {
			item.Kind = lsp.Function
		} else if oneSymbol.Kind == common.IKAnnotateClass {
			item.Kind = lsp.Interface
		} else if oneSymbol.Kind == common.IKAnnotateAlias {
			item.Kind = lsp.Interface
		}
		items = append(items, item)
	}

	return
}
