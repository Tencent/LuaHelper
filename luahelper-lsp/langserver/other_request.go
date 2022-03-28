package langserver

import (
	"context"

	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/pathpre"

	"luahelper-lsp/langserver/log"
	lsp "luahelper-lsp/langserver/protocol"
)

// CancelRequest 取消一个请求
func (l *LspServer) CancelRequest(ctx context.Context, vs lsp.CancelParams) error {
	log.Debug("CancelRequest, id=%v", vs.ID)
	return nil
	// var strId string
	// switch id := vs.ID.(type) {
	// case float64:
	// 	strId = fmt.Sprintf("%.f", id)
	// 	break
	// case string:
	// 	strId = id
	// 	break
	// default:
	// 	return nil
	// }

	// l.server.CancelRequest(strId)
	// return nil
}

// TextDocumentCodeLens 请求
func (l *LspServer) TextDocumentCodeLens(ctx context.Context, vs lsp.CodeLensParams) (edit []lsp.CodeLens, err error) {
	return
}

// TextDocumentdocumentLink 请求
func (l *LspServer) TextDocumentdocumentLink(ctx context.Context, vs lsp.DocumentLinkParams) (edit []lsp.DocumentLink, err error) {
	return
}

// SourceParams 请求的原参数
type SourceParams struct {
	Roots []string `json:"roots,omitempty"`
}

// ComletionParams 代码提示设置
type ComletionParams struct {
	Setting string `json:"setting,omitempty"`
}

// ProjectParams 工程的设置
type BaseParams struct {
	Rootdir              string   `json:"Rootdir,omitempty"`
	IgnoreFileOrDir      []string `json:"IgnoreFileOrDir,omitempty"`
	IgnoreFileOrDirError []string `json:"IgnoreFileOrDirError,omitempty"`
	RequirePathSeparator string   `json:"RequirePathSeparator,omitempty"`
	ReferenceMaxNum      int      `json:"ReferenceMaxNum,omitempty"`
	ReferenceDefineFlag  bool     `json:"ReferenceIncudeDefine,omitempty"`
	PreviewFieldsNum     int      `json:"PreviewFieldsNum,omitempty"`
	EnableReport         bool     `json:"Report,omitempty"`
}

// WarnParams 引用的设置
type WarnParams struct {
	AllEnable                      bool `json:"AllEnable,omitempty"`
	CheckSyntax                    bool `json:"CheckSyntax,omitempty"`
	CheckNoDefine                  bool `json:"CheckNoDefine,omitempty"`
	CheckAfterDefine               bool `json:"CheckAfterDefine,omitempty"`
	CheckLocalNoUse                bool `json:"CheckLocalNoUse,omitempty"`
	CheckTableDuplicateKey         bool `json:"CheckTableDuplicateKey,omitempty"`
	CheckReferNoFile               bool `json:"CheckReferNoFile,omitempty"`
	CheckAssignParamNum            bool `json:"CheckAssignParamNum,omitempty"`
	CheckLocalDefineParamNum       bool `json:"CheckLocalDefineParamNum,omitempty"`
	CheckGotoLable                 bool `json:"CheckGotoLable,omitempty"`
	CheckFuncParam                 bool `json:"CheckFuncParam,omitempty"`
	CheckImportModuleVar           bool `json:"CheckImportModuleVar,omitempty"`
	CheckIfNotVar                  bool `json:"CheckIfNotVar,omitempty"`
	CheckFunctionDuplicateParam    bool `json:"CheckFunctionDuplicateParam,omitempty"`
	CheckBinaryExpressionDuplicate bool `json:"CheckBinaryExpressionDuplicate,omitempty"`
	CheckErrorOrAlwaysTrue         bool `json:"CheckErrorOrAlwaysTrue,omitempty"`
	CheckErrorAndAlwaysFalse       bool `json:"CheckErrorAndAlwaysFalse,omitempty"`
	CheckNoUseAssign               bool `json:"CheckNoUseAssign,omitempty"`
	CheckAnnotateType              bool `json:"CheckAnnotateType,omitempty"`
	CheckDuplicateIf               bool `json:"CheckDuplicateIf,omitempty"`
	CheckSelfAssign                bool `json:"CheckSelfAssign,omitempty"`
	CheckFloatEq                   bool `json:"CheckFloatEq,omitempty"`
	CheckClassField                bool `json:"CheckClassField,omitempty"`
	CheckConstAssign               bool `json:"CheckConstAssign,omitempty"`
}

// LuahelperParams 整体的设置
type LuahelperParams struct {
	Base      BaseParams `json:"base,omitempty"`
	WarnParam WarnParams `json:"Warn,omitempty"`
}

// SettingsParam 设置参数
type SettingsParam struct {
	Luahelper LuahelperParams `json:"luahelper,omitempty"`
	Files     interface{}     `json:"files,omitempty"`
}

// ChangeConfigurationParams 变更配置
type ChangeConfigurationParams struct {
	Settings SettingsParam `json:"settings,omitempty"`
}

// Registration 单个后台动态注册能力
type Registration struct {
	ID              string                                   `json:"id,omitempty"`
	Method          string                                   `json:"method,omitempty"`
	RegisterOptions DidChangeWatchedFilesRegistrationOptions `json:"registerOptions,omitempty"`
}

// RegistrationParams 动态注册能力
type RegistrationParams struct {
	Registrations []Registration `json:"registrations,omitempty"`
}

// FileSystemWatcher 动态注册修改文件监控
type FileSystemWatcher struct {
	GlobPattern string `json:"globPattern,omitempty"`
	Kind        int    `json:"kind,omitempty"`
}

// DidChangeWatchedFilesRegistrationOptions 所有动态注册修改文件监控
type DidChangeWatchedFilesRegistrationOptions struct {
	/**
	 * The watchers to register.
	 */
	Watchers []FileSystemWatcher `json:"Watchers,omitempty"`
}

// 获取其他后缀关联到的lua 类型  (associalList string[])
func getAssociationList(associationData interface{}) (associalList []string) {
	oneFilesData, ok := associationData.(map[string]interface{})
	if !ok {
		return
	}

	associations, flag := oneFilesData["associations"]
	if !flag {
		return
	}

	return getInitAssociationList(associations)
}

// ChangeConfiguration 修改配置请求
func (l *LspServer) ChangeConfiguration(ctx context.Context, vs ChangeConfigurationParams) error {
	base := vs.Settings.Luahelper.Base
	setConfigSet(base.ReferenceMaxNum, base.ReferenceDefineFlag)
	l.enableReport = base.EnableReport

	// 设置预览table成员的数量
	common.GConfig.SetPreviewFieldsNum(base.PreviewFieldsNum)
	if !l.changeConfFlag {
		l.changeConfFlag = true
		return nil
	}

	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	if common.GConfig.ReadJSONFlag {
		return nil
	}

	associalList := getAssociationList(vs.Settings.Files)
	common.GConfig.SetAssocialList(associalList)

	// 按顺序插入
	checkFlagList := getWarnCheckList(&vs.Settings.Luahelper.WarnParam)

	common.GConfig.HandleChangeCheckList(checkFlagList, base.IgnoreFileOrDir, base.IgnoreFileOrDirError)

	// 设置require其他lua文件的路径分割
	common.GConfig.SetRequirePathSeparator(base.RequirePathSeparator)

	return l.handleChange(ctx)
}

// clearLspServer 清空存在的信息
func (l *LspServer) clearLspServer(ctx context.Context) {
	for strFile := range l.fileErrorMap {
		l.ClearOneFileDiagnostic(ctx, strFile)
	}

	l.fileErrorMap = map[string][]common.CheckError{}
	l.fileChangeErrorMap = map[string][]common.CheckError{}
	//l.fileCache = lspcommon.CreateFileMapCache()
}

func (l *LspServer) handleChange(ctx context.Context) error {
	l.clearLspServer(ctx)

	dirManager := common.GConfig.GetDirManager()
	mainDir := dirManager.GetMainDir()

	var checkList []string
	checkList = dirManager.GetMainDirFileList()

	subDirCheckList := dirManager.GetSubDirsFileList()
	checkList = append(checkList, subDirCheckList...)

	clientExpPathList := dirManager.GetPathFileList(dirManager.GetClientExtLuaPath())
	checkList = append(checkList, clientExpPathList...)

	// 补全所有的入口文件
	var entryFileList []string
	for _, luaFile := range common.GConfig.ProjectFiles {
		luaFile = pathpre.GetRemovePreStr(luaFile)
		entryFileList = append(entryFileList, dirManager.GetCompletePath(mainDir, luaFile))
	}
	allProject := check.CreateAllProject(checkList, entryFileList, clientExpPathList)
	allProject.HandleCheck()

	// 工程路径变量设置到Glsp侧
	l.project = allProject

	// 再一次获取所有诊断信息
	l.pushAllDiagnosticsAgain(ctx)

	return nil
}

// InitializedParams 初始化参数
type InitializedParams struct {
	Settings interface{} `json:"settings"`
}

// Shutdown lsp 关闭
func (l *LspServer) Shutdown(ctx context.Context) error {
	log.Debug("Shutdown")
	return nil
}

// Exit 退出了
func (l *LspServer) Exit(ctx context.Context) error {
	log.Debug("Exit")
	return nil
}

// 获取配置的告警列表
func getWarnCheckList(warnParam *WarnParams) (checkFlagList []bool) {
	checkFlagList = []bool{
		warnParam.AllEnable,
		warnParam.CheckSyntax,
		warnParam.CheckNoDefine,
		warnParam.CheckAfterDefine,
		warnParam.CheckLocalNoUse,
		warnParam.CheckTableDuplicateKey,
		warnParam.CheckReferNoFile,
		warnParam.CheckAssignParamNum,
		warnParam.CheckLocalDefineParamNum,
		warnParam.CheckGotoLable,
		warnParam.CheckFuncParam,
		warnParam.CheckImportModuleVar,
		warnParam.CheckIfNotVar,
		warnParam.CheckFunctionDuplicateParam,
		warnParam.CheckBinaryExpressionDuplicate,
		warnParam.CheckErrorOrAlwaysTrue,
		warnParam.CheckErrorAndAlwaysFalse,
		warnParam.CheckNoUseAssign,
		warnParam.CheckAnnotateType,
		warnParam.CheckDuplicateIf,
		warnParam.CheckSelfAssign,
		warnParam.CheckFloatEq,
		warnParam.CheckClassField,
		warnParam.CheckConstAssign,
	}

	return checkFlagList
}
