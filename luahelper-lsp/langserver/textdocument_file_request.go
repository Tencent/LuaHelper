package langserver

import (
	"context"
	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
	"luahelper-lsp/langserver/strbytesconv"
	"time"
)

// TextDocumentDidOpen 打开了一个文件的请求
func (l *LspServer) TextDocumentDidOpen(ctx context.Context, vs lsp.DidOpenTextDocumentParams) error {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	// 判断打开的文件，是否是需要分析的文件
	strFile := pathpre.VscodeURIToString(string(vs.TextDocument.URI))
	project := l.getAllProject()
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return nil
	}

	// 判断是否已经当成lua处理
	if !common.GConfig.IsHandleAsLua(strFile) {
		log.Debug("not need to handle same as lua strFile=%s", strFile)
		return nil
	}

	l.setColorTime(0)

	fileCache := l.getFileCache()
	fileCache.SetFileContent(strFile, strbytesconv.StringToBytes(vs.TextDocument.Text))

	// 文件打开，删除cache的内容
	project.RemoveCacheContent(strFile)

	// 当成lua处理，先判断缓存中，是否有该文件，如果没有该文件，需要加载该文件
	if !project.IsInAllFilesMap(strFile) {
		var fileEventVec []check.FileEventStruct
		fileEventVec = append(fileEventVec, check.FileEventStruct{
			StrFile: strFile,
			Type:    check.FileEventCreated,
		})

		// 处理所有的文件变化
		if project.HandleFileEventChanges(fileEventVec) {
			// 再一次获取所有诊断信息
			l.pushAllDiagnosticsAgain(ctx)
		}
	}

	// 文件打开了，清除临时的错误显示
	l.ClearChangeFileErr(ctx, strFile)
	return nil
}

// TextDocumentDidChange 单个文件的内容变化了
func (l *LspServer) TextDocumentDidChange(ctx context.Context, vs lsp.DidChangeTextDocumentParams) error {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	strFile := pathpre.VscodeURIToString(string(vs.TextDocument.URI))
	project := l.getAllProject()
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return nil
	}

	fileCache := l.getFileCache()
	contents, found := fileCache.GetFileContent(strFile)
	if !found {
		log.Error("ApplyContentChanges get strFile=%s error", strFile)
		return nil
	}

	time1 := time.Now()
	changeContents, err := fileCache.ApplyContentChanges(strFile, contents, vs.ContentChanges)
	if err != nil {
		log.Error("ApplyContentChanges strFile=%s errInfo=%s", strFile, err.Error())
		return nil
	}
	ftime := time.Since(time1).Milliseconds()
	log.Debug("TextDocumentDidChang ApplyContentChanges, strFile=%s, time=%d", strFile, ftime)

	fileCache.SetFileContent(strFile, changeContents)
	contents, found = fileCache.GetFileContent(strFile)
	if !found {
		log.Error("ApplyContentChanges get strFile=%s error", strFile)
		return nil
	}

	errList := project.HandleFileChangeAnalysis(strFile, contents)
	if len(errList) > 0 {
		l.InsertChangeFileErr(ctx, strFile, errList)
		// 设置文件修改的时间
		l.setColorTime(0)
	} else {
		l.ClearChangeFileErr(ctx, strFile)
		l.ClearFileSyntaxErr(ctx, strFile)
		// 设置文件修改的时间
		l.setColorTime(time.Now().Unix())
	}

	log.Debug("TextDocumentDidChang end, strFile=%s", strFile)
	return nil
}

// WorkspaceChangeWatchedFiles 整个工程目录lua文件的变化
func (l *LspServer) WorkspaceChangeWatchedFiles(ctx context.Context, vs lsp.DidChangeWatchedFilesParams) error {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()
	log.Debug("WorkspaceChangeWatchedFiles..\n")

	project := l.getAllProject()
	var fileEventVec []check.FileEventStruct
	for _, fileEvents := range vs.Changes {
		strFile := pathpre.VscodeURIToString(string(fileEvents.URI))
		if !project.IsNeedHandle(strFile) {
			log.Debug("not need to handle strFile=%s", strFile)
			continue
		}

		if !common.GConfig.IsHandleAsLua(strFile) {
			log.Debug("not need to handle same as lua strFile=%s", strFile)
			continue
		}

		fileEventVec = append(fileEventVec, check.FileEventStruct{
			StrFile: strFile,
			Type:    (int)(fileEvents.Type),
		})

		// 文件有变更了，清除临时的错误显示
		l.ClearChangeFileErr(ctx, strFile)
	}

	if len(fileEventVec) > 0 {
		// 需要去处理文件的变化
		log.Debug("need to handle file venent changes num=%d", len(fileEventVec))
		// 处理所有的文件变化
		if !project.HandleFileEventChanges(fileEventVec) {
			return nil
		}

		// 再一次获取所有诊断信息
		l.pushAllDiagnosticsAgain(ctx)
	}

	// 更新下需要统计的信息
	l.SetLuaFileNumber(project.GetAllFileNumber())
	return nil
}

// TextDocumentDidClose 文件关闭了
func (l *LspServer) TextDocumentDidClose(ctx context.Context, vs lsp.DidCloseTextDocumentParams) error {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	strFile := pathpre.VscodeURIToString(string(vs.TextDocument.URI))
	project := l.getAllProject()
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return nil
	}

	fileCache := l.getFileCache()
	fileCache.DelFileContent(strFile)

	// 文件关闭，删除cache的内容
	project.RemoveCacheContent(strFile)

	// 文件关闭了，清除临时的错误显示
	l.ClearChangeFileErr(ctx, strFile)

	// 非工程文件夹下的文件，清除所有的诊断错误
	dirManager := common.GConfig.GetDirManager()
	if !dirManager.IsInDir(strFile) {
		l.ClearOneFileDiagnostic(ctx, strFile)
		l.RemoveFile(strFile)
		project.RemoveFile(strFile)
	}

	return nil
}

// TextDocumentDidSave 文件的内容进行保存
func (l *LspServer) TextDocumentDidSave(ctx context.Context, vs lsp.DidSaveTextDocumentParams) error {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	strFile := pathpre.VscodeURIToString(string(vs.TextDocument.URI))
	log.Debug("TextDocumentDidSave ..., strFile=%s", strFile)
	project := l.getAllProject()
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return nil
	}

	// 更新下保存的内容
	fileCache := l.getFileCache()
	fileCache.SetFileContent(strFile, strbytesconv.StringToBytes(*vs.Text))

	// 文件之前是全路径，转换成相对路径
	var fileEventVec []check.FileEventStruct
	fileEventVec = append(fileEventVec, check.FileEventStruct{
		StrFile: strFile,
		Type:    2,
	})

	// 处理所有的文件变化
	if !project.HandleFileEventChanges(fileEventVec) {
		// 文件保存了，清除临时的错误显示, 并且重新推送这个文件的错误信息
		l.SaveOneFilePushAgain(ctx, strFile)
		return nil
	}

	// 再一次获取所有诊断信息
	l.pushAllDiagnosticsAgain(ctx)

	// 文件保存了，清除临时的错误显示, 并且重新推送这个文件的错误信息
	l.SaveOneFilePushAgain(ctx, strFile)

	return nil
}
