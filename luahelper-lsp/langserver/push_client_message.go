package langserver

import (
	"context"

	lsp "luahelper-lsp/langserver/protocol"

	"luahelper-lsp/langserver/log"
)

// // PushShowMessage 给客户端推送提示消息
// func PushShowMessage(ctx context.Context, messageType lsp.MessageType, strContent string) {
// 	err := GLspGlobal.GServer.Notify(ctx, "window/showMessage", lsp.ShowMessageParams{
// 		Type:    messageType,
// 		Message: strContent,
// 	})

// 	if err != nil {
// 		log.Debug("PushShowMessage error=%v", err)
// 	}
// }

// sendDiagnostics 给客户端推送错误诊断消息
func (l *LspServer)sendDiagnostics(ctx context.Context, diagnostics lsp.PublishDiagnosticsParams) {
	err := l.server.Notify(ctx, "textDocument/publishDiagnostics", diagnostics)
	if err != nil {
		log.Debug("PushShowMessage error=%v", err)
	}
}
