package langserver

import (
	"context"
	"luahelper-lsp/langserver/log"
)

// ProgressState 进度的状态
type ProgressState int

const (
	// PSDirtory 加载目录
	PSDirtory ProgressState = 0
	// PSFile 加载文件中
	PSFile ProgressState = 1
	// PSFinish 加载完毕
	PSFinish ProgressState = 2
)

// IProgressReport 插件后台给插件前端推送加载工程的进度状态
type IProgressReport struct {
	State int    `json:"State"` //  通知的状态，0为加载目录, 1为加载文件中，2为加载完毕
	Text  string `json:"Text"`  // 显示的字符串
}

// PushProgressReport 给客户端推送加载的进度消息
func  (l *LspServer)PushProgressReport(ctx context.Context, state int, strContent string) {
	err := l.server.Notify(ctx, "luahelper/progressReport", IProgressReport{
		State: state,
		Text:  strContent,
	})

	if err != nil {
		log.Debug("PushProgressReport error=%v", err)
	}
}
