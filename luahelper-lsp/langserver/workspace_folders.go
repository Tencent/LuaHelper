package langserver

import (
	"context"

	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/check/common"

	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
)

// WorkspaceChangeWorkspaceFolders 文件夹有变化
func (l *LspServer)WorkspaceChangeWorkspaceFolders(ctx context.Context, params lsp.DidChangeWorkspaceFoldersParams) error {
	for _, oneFolder := range params.Event.Added {
		folderPath := pathpre.VscodeURIToString(oneFolder.URI)
		l.addWorkspaceFolder(ctx, folderPath)
	}
	
	for _, oneFolder := range params.Event.Removed {
		folderPath := pathpre.VscodeURIToString(oneFolder.URI)
		l.removeWorkspaceFolder(ctx, folderPath)
	}
	return nil
}

// addWorkspaceFolder 处理增加文件夹
func (l *LspServer)addWorkspaceFolder(ctx context.Context, dirpath string) {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()
	dirManager := common.GConfig.GetDirManager()

	// 若增加的是当前workspace 文件夹中包含的子文件夹， 则不需要做任何处理
	if dirManager.IsDirExistWorkspace(dirpath) {
		log.Debug("current added dir=%s has existed in the workspaceFolder, not need analysis", dirpath)
		return
	}
	dirManager.PushOneSubDir(dirpath)

	allProject := l.getAllProject()

	// 获取当前文件夹下所有lua文件
	addFiles := dirManager.GetDirFileList(dirpath, false)

	// 设置所有增加文件的事件
	addFileEvent := make([]check.FileEventStruct, 0, len(addFiles))
	for _, file := range addFiles {
		addFileEvent = append(addFileEvent, check.FileEventStruct{
			StrFile: file,
			Type:    1,
		})
		// 文件有变更了，清除临时的错误显示
		l.ClearChangeFileErr(ctx, file)
	}

	if len(addFileEvent) > 0 {
		// 需要去处理文件的变化
		log.Debug("need to handle file venent changes num=%d", len(addFileEvent))
		// 处理所有的文件变化
		allProject.HandleFileEventChanges(addFileEvent) 

		// 再一次获取所有诊断信息
		l.pushAllDiagnosticsAgain(ctx)
	}

	// 更新下需要统计的信息
	l.SetLuaFileNumber(allProject.GetAllFileNumber())
}

// removeWorkspaceFolder 处理移除文件夹
func (l *LspServer)removeWorkspaceFolder(ctx context.Context, dirpath string) {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()
	dirManager := common.GConfig.GetDirManager()

	if !dirManager.RemoveOneSubDir(dirpath) {
		log.Debug("remove error :%s", dirpath)	
	}

	// 若移除的是subDirs 中的文件夹， 则直接进行所有文件删除处理
	allProject := l.getAllProject()
	allFilesMap := allProject.GetAllFilesMap()

	// 获取当前文件夹下所有lua文件
	removedFiles := dirManager.GetDirFileList(dirpath, false)

	delFileEvent := make([]check.FileEventStruct, 0, len(removedFiles))
	for _, file := range removedFiles {
		if _, ok := allFilesMap[file]; !ok {
			continue
		}
		delFileEvent = append(delFileEvent, check.FileEventStruct{
			StrFile: file,
			Type:    3,
		})
		// 文件有变更了，清除临时的错误显示
		l.ClearChangeFileErr(ctx, file)
		l.ClearOneFileDiagnostic(ctx, file)
		l.RemoveFile(file)
	}

	if len(delFileEvent) > 0 {
		// 需要去处理文件的变化
		log.Debug("need to handle file venent changes num=%d", len(delFileEvent))
		// 处理所有的文件变化
		allProject.HandleFileEventChanges(delFileEvent)

		// 再一次获取所有诊断信息
		l.pushAllDiagnosticsAgain(ctx)
	}

	// 更新下需要统计的信息
	l.SetLuaFileNumber(allProject.GetAllFileNumber())
}
