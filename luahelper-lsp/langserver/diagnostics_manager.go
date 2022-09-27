package langserver

import (
	"context"
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/lspcommon"

	lsp "luahelper-lsp/langserver/protocol"
)

// pushFileErrList 给指定的文件推送指定的错误列表
func (l *LspServer) pushFileErrList(ctx context.Context, strFile string, fileErrVec []common.CheckError) {
	var diagnostics lsp.PublishDiagnosticsParams
	diagnostics.URI = lspcommon.GetFileDocumentURI(strFile)
	diagnostics.Diagnostics = []lsp.Diagnostic{}
	for _, oneErr := range fileErrVec {
		diagnostics.Diagnostics = append(diagnostics.Diagnostics, changeErrToDiagnostic(&oneErr))
	}

	// 发送单个文件的诊断信息
	l.sendDiagnostics(ctx, diagnostics)
}

// GetAllDiagnostics 获取所有的诊断错误
func (l *LspServer) GetAllDiagnostics(ctx context.Context) {
	project := l.getAllProject()
	if project == nil {
		log.Error("CheckProject is nil")
		return
	}

	// 保存全局的错误诊断信息
	l.fileErrorMap = project.GetAllFileErrorInfo()
	if len(l.fileErrorMap) == 0 {
		log.Debug("GetAllFileErrorInfo is empty..")
		return
	}

	// 真正推送所有的诊断错误
	for strFile, fileErrVec := range l.fileErrorMap {
		l.pushFileErrList(ctx, strFile, fileErrVec)
	}
}

// pushAllDiagnosticsAgain 再次全量获取诊断信息，增量推送诊断信息给客户端
func (l *LspServer) pushAllDiagnosticsAgain(ctx context.Context) {
	project := l.getAllProject()
	fileErrorMap := project.GetAllFileErrorInfo()

	// 遍历之前的诊断错误，看是否要清除部分诊断错误
	// 如果之前存在告警，但是现在修复了，清除这个文件的错误
	for strFile := range l.fileErrorMap {
		if _, ok := fileErrorMap[strFile]; !ok {
			// 之前有该文件诊断错误，但是修复了; 清除这个文件的错误
			l.ClearOneFileDiagnostic(ctx, strFile)
		}
	}

	// 判断所有的新的告警，是否要增量推送
	// 与老的告警对比，判断是否要改动，改动的重新推送
	for strFile, newErrList := range fileErrorMap {
		oldErrList, ok := l.fileErrorMap[strFile]
		if !ok {
			// 新的告警在老的中不存在，说明是全新的，推送这个文件的告警
			l.pushFileErrList(ctx, strFile, newErrList)
			continue
		}

		if lspcommon.IsSameErrList(oldErrList, newErrList) {
			// 告警是一样的，不用推送
			log.Debug("strFile=%s warn is the same.", strFile)
			continue
		}

		// 新的告警，在老的中存在，需要对比分析下，是否跟之前完全一样，如果完全一样，不进行推送
		l.pushFileErrList(ctx, strFile, newErrList)
	}

	l.fileErrorMap = fileErrorMap
	if len(fileErrorMap) == 0 {
		log.Debug("GetAllFileErrorInfo is empty..")
		l.pushAllChangeFileDiagnosticErr(ctx)
		return
	}
}

// ClearOneFileDiagnostic 清空某个文件的所有诊断错误
func (l *LspServer) ClearOneFileDiagnostic(ctx context.Context, strFile string) {
	var diagnostics lsp.PublishDiagnosticsParams
	diagnostics.URI = lspcommon.GetFileDocumentURI(strFile)
	diagnostics.Diagnostics = []lsp.Diagnostic{}
	l.sendDiagnostics(ctx, diagnostics)
}

// RemoveFile 非工程文件被删除，清除错误
func (l *LspServer) RemoveFile(strFile string) {
	delete(l.fileErrorMap, strFile)
}

// pushFileChangeDiagnostic 推送文件临时变化的诊断错误
func (l *LspServer) pushFileChangeDiagnostic(ctx context.Context, strFile string) {
	errList, ok := l.fileChangeErrorMap[strFile]
	if !ok {
		return
	}

	var diagnostics lsp.PublishDiagnosticsParams
	diagnostics.URI = lspcommon.GetFileDocumentURI(strFile)
	diagnostics.Diagnostics = []lsp.Diagnostic{}
	for _, oneErr := range errList {
		diagnostics.Diagnostics = append(diagnostics.Diagnostics, changeErrToDiagnostic(&oneErr))
	}

	// 发送单个文件的诊断信息
	l.sendDiagnostics(ctx, diagnostics)
}

// pushFileDiagnostic 再次推送某个文件的诊断错误
// ignoreSyntax 表示是否忽略语法错误
func (l *LspServer) pushFileDiagnostic(ctx context.Context, strFile string, ignoreSyntax bool) {
	fileErrVec, ok := l.fileErrorMap[strFile]
	if !ok {
		// 如果之前不存在任何的错误，退出
		return
	}

	var diagnostics lsp.PublishDiagnosticsParams
	diagnostics.URI = lspcommon.GetFileDocumentURI(strFile)
	diagnostics.Diagnostics = []lsp.Diagnostic{}
	for _, oneErr := range fileErrVec {
		if oneErr.ErrType == common.CheckErrorSyntax && ignoreSyntax {
			continue
		}
		diagnostics.Diagnostics = append(diagnostics.Diagnostics, changeErrToDiagnostic(&oneErr))
	}

	// 发送单个文件的诊断信息
	l.sendDiagnostics(ctx, diagnostics)
}

// InsertChangeFileErr 文件实时变化，但是没有保存时候，插入实时分析的语法错误
func (l *LspServer) InsertChangeFileErr(ctx context.Context, strFile string, errList []common.CheckError) {
	l.fileChangeErrorMap[strFile] = errList

	// 推送新的诊断错误
	l.pushFileChangeDiagnostic(ctx, strFile)
}

// ClearChangeFileErr 清除一个文件实时变化产生的AST错误
func (l *LspServer) ClearChangeFileErr(ctx context.Context, strFile string) {
	if _, ok := l.fileChangeErrorMap[strFile]; ok {
		// 之前存在，现在没有语法错误了
		delete(l.fileChangeErrorMap, strFile)

		// 删除这个文件诊断错误
		l.ClearOneFileDiagnostic(ctx, strFile)

		// 恢复原来的诊断问题
		l.pushFileDiagnostic(ctx, strFile, true)
	}
}

// SaveOneFilePushAgain 文件保存的操作，重新推送这个文件的错误信息
func (l *LspServer) SaveOneFilePushAgain(ctx context.Context, strFile string) {
	// 删除临时保存的错误
	delete(l.fileChangeErrorMap, strFile)

	if _, ok := l.fileErrorMap[strFile]; ok {
		// 恢复原来的诊断问题
		l.pushFileDiagnostic(ctx, strFile, false)
	} else {
		// 删除这个文件诊断错误
		l.ClearOneFileDiagnostic(ctx, strFile)
	}
}

// ClearFileSyntaxErr 清除文件的语法错误
func (l *LspServer) ClearFileSyntaxErr(ctx context.Context, strFile string) {
	_, ok := l.fileErrorMap[strFile]
	if !ok {
		// 如果之前不存在任何的错误，退出
		return
	}

	// 删除这个文件诊断错误
	l.ClearOneFileDiagnostic(ctx, strFile)
	l.pushFileDiagnostic(ctx, strFile, true)
}

// pushAllChangeFileDiagnosticErr 推送所有的临时错误
func (l *LspServer) pushAllChangeFileDiagnosticErr(ctx context.Context) {
	for strFile := range l.fileChangeErrorMap {
		// 删除这个文件诊断错误
		l.ClearOneFileDiagnostic(ctx, strFile)

		// 推送新的诊断错误
		l.pushFileChangeDiagnostic(ctx, strFile)
	}
}

// changeErrToDiagnostic 该文件为所有分析文件的诊断管理
func changeErrToDiagnostic(checkErr *common.CheckError) lsp.Diagnostic {
	var diagnostic lsp.Diagnostic
	if checkErr.ErrType == common.CheckErrorSyntax {
		diagnostic.Severity = lsp.SeverityError
	} else if checkErr.ErrType == common.CheckErrorAnnotate {
		diagnostic.Severity = lsp.SeverityInformation
	} else {
		diagnostic.Severity = lsp.SeverityWarning
	}
	strPre := ""
	if checkErr.ErrType == common.CheckErrorSyntax {
		strPre = fmt.Sprintf("[Warn type:%d], ", checkErr.ErrType)
	} else {
		strPre = fmt.Sprintf("[Warn type:%d], ", checkErr.ErrType)
	}

	diagnostic.Message = strPre + checkErr.ErrStr

	if checkErr.EntryFile != "" && !common.GConfig.IsHasProjectEntryFile() {
		if checkErr.EntryFile == "common project" {
			diagnostic.Message = strPre + checkErr.ErrStr + ". <" + checkErr.EntryFile + ">"
		} else {
			diagnostic.Message = strPre + checkErr.ErrStr + ". <process entry file: " + checkErr.EntryFile + ">"
		}
	}

	for _, oneRelate := range checkErr.RelateVec {
		oneRelateLsp := lsp.DiagnosticRelatedInformation{
			Location: lsp.Location{
				URI:   lspcommon.GetFileDocumentURI(oneRelate.LuaFile),
				Range: lspcommon.LocToRange(&oneRelate.Loc),
			},
			Message: oneRelate.ErrStr,
		}
		diagnostic.RelatedInformation = append(diagnostic.RelatedInformation, oneRelateLsp)
	}

	diagnostic.Range = lspcommon.LocToRange(&checkErr.Loc)
	return diagnostic
}
