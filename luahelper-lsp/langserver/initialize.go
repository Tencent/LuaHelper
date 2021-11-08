package langserver

import (
	"context"
	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
	"strings"
	"time"
)

// InitializationOptions 初始化选项
type InitializationOptions struct {
	Client                 string      `json:"client,omitempty"`
	PluginPath             string      `json:"PluginPath,omitempty"`
	ReferenceMaxNum        int         `json:"referenceMaxNum,omitempty"`
	ReferenceDefineFlag    bool        `json:"referenceDefineFlag,omitempty"`
	LocalRun               bool        `json:"LocalRun,omitempty"`
	FileAssociationsConfig interface{} `json:"FileAssociationsConfig,omitempty"`

	AllEnable                      bool     `json:"AllEnable,omitempty"`
	CheckSyntax                    bool     `json:"CheckSyntax,omitempty"`
	CheckNoDefine                  bool     `json:"CheckNoDefine,omitempty"`
	CheckAfterDefine               bool     `json:"CheckAfterDefine,omitempty"`
	CheckLocalNoUse                bool     `json:"CheckLocalNoUse,omitempty"`
	CheckTableDuplicateKey         bool     `json:"CheckTableDuplicateKey,omitempty"`
	CheckReferNoFile               bool     `json:"CheckReferNoFile,omitempty"`
	CheckAssignParamNum            bool     `json:"CheckAssignParamNum,omitempty"`
	CheckLocalDefineParamNum       bool     `json:"CheckLocalDefineParamNum,omitempty"`
	CheckGotoLable                 bool     `json:"CheckGotoLable,omitempty"`
	CheckFuncParam                 bool     `json:"CheckFuncParam,omitempty"`
	CheckImportModuleVar           bool     `json:"CheckImportModuleVar,omitempty"`
	CheckIfNotVar                  bool     `json:"CheckIfNotVar,omitempty"`
	CheckFunctionDuplicateParam    bool     `json:"CheckFunctionDuplicateParam,omitempty"`
	CheckBinaryExpressionDuplicate bool     `json:"CheckBinaryExpressionDuplicate,omitempty"`
	CheckErrorOrAlwaysTrue         bool     `json:"CheckErrorOrAlwaysTrue,omitempty"`
	CheckErrorAndAlwaysFalse       bool     `json:"CheckErrorAndAlwaysFalse,omitempty"`
	CheckNoUseAssign               bool     `json:"CheckNoUseAssign,omitempty"`
	CheckAnnotateType              bool     `json:"CheckAnnotateType,omitempty"`
	IgnoreFileOrDir                []string `json:"IgnoreFileOrDir,omitempty"`
	IgnoreFileOrDirError           []string `json:"IgnoreFileOrDirError,omitempty"`
}

// InitializeParams 初始化参数
type InitializeParams struct {
	lsp.InitializeParams
	InitializationOptions *InitializationOptions `json:"initializationOptions,omitempty"`
}

// Initialize lsp初始化函数
func (l *LspServer) Initialize(ctx context.Context, vs InitializeParams) (lsp.InitializeResult, error) {
	pathpre.InitialRootURIAndPath(string(vs.RootURI), string(vs.RootPath))

	log.Debug("Initialize ..., rootDir=%s, rooturl=%s", vs.RootPath, vs.RootURI)
	vscodeRoot := pathpre.VscodeURIToString(string(vs.RootURI))
	dirManager := common.GConfig.GetDirManager()
	dirManager.SetVSRootDir(vscodeRoot)

	workspaceFolderNum := len(vs.WorkspaceFolders)

	initOptions := vs.InitializationOptions
	if initOptions == nil {
		initOptions = getDefaultIntialOptions()
	}
	dirManager.SetClientPluginPath(initOptions.PluginPath)

	setConfigSet(initOptions.ReferenceMaxNum, initOptions.ReferenceDefineFlag)

	// 初始化时获取其他后缀关联到的lua
	associalList := getInitAssociationList(initOptions.FileAssociationsConfig)
	log.Debug("associalList len:%d", len(associalList))
	common.GConfig.SetAssocialList(associalList)

	// 按顺序插入
	checkFlagList := getCheckFlagList(initOptions)

	initErr := l.initialCheckProject(ctx, checkFlagList, initOptions.Client, workspaceFolderNum, vs.WorkspaceFolders,
		initOptions.LocalRun, initOptions.IgnoreFileOrDir, initOptions.IgnoreFileOrDirError)
	if initErr != nil {
		log.Error("initial luahelper err: " + initErr.Error())
		return lsp.InitializeResult{}, initErr
	}
	log.Debug("initial luahelper ok")

	return lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			InnerServerCapabilities: lsp.InnerServerCapabilities{
				TextDocumentSync: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.Incremental,
					Save: lsp.SaveOptions{
						IncludeText: true,
					},
				},
				CompletionProvider: lsp.CompletionOptions{
					ResolveProvider: true,
					//TriggerCharacters: strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:", ""),
					TriggerCharacters: []string{".", "\"", "'", ":", "-", "@"},
				},
				ColorProvider:           false,
				HoverProvider:           true,
				WorkspaceSymbolProvider: true,
				DefinitionProvider:      true,
				ReferencesProvider:      true,
				DocumentSymbolProvider:  true,
				SignatureHelpProvider: lsp.SignatureHelpOptions{
					TriggerCharacters: []string{"(", ","},
				},
				CodeLensProvider: lsp.CodeLensOptions{
					ResolveProvider: false,
				},
				DocumentLinkProvider: lsp.DocumentLinkOptions{
					ResolveProvider: false,
				},
				RenameProvider:            true,
				DocumentHighlightProvider: true,
				Workspace: lsp.WorkspaceGn{
					WorkspaceFolders: lsp.WorkspaceFoldersGn{
						Supported:           true,
						ChangeNotifications: "workspace/didChangeWorkspaceFolders",
					},
				},
			},
		},
	}, nil
}

// Initialized 初始化
func (l *LspServer) Initialized(ctx context.Context, initialParam InitializedParams) error {
	log.Debug("Initialized")
	// 获取所有的诊断错误
	l.GetAllDiagnostics(ctx)
	return nil
}

// 初始化CheckProject
// checkList 为检查关卡的切片
func (l *LspServer) initialCheckProject(ctx context.Context, checkFlagList []bool, clientType string,
	workspaceFolderNum int, workspaceFolder []lsp.WorkspaceFolder, isLocal bool, ignoreFileOrDir []string,
	ignoreFileOrDirErr []string) error {
	// 目录管理统一设置vscodeRoot目录
	dirManager := common.GConfig.GetDirManager()
	vscodeRoot := dirManager.GetVsRootDir()
	timeBegin := time.Now()

	// 尝试读取vscode目录下面的ylua.json工程配置文件
	if readErr := common.GConfig.ReadConfig(vscodeRoot, "luahelper.json", checkFlagList, ignoreFileOrDir,
		ignoreFileOrDirErr); readErr != nil {
		return readErr
	}

	if isLocal {
		// 为本地运行，没有插件前端，插件前端无法传递额外的Lua文件夹，忽略系统的模块和变量
		common.GConfig.InsertIngoreSystemModule()

		// 为本地运行，没有插件前端，插件前端无法传递额外的Lua文件夹，忽略系统的注解类型
		common.GConfig.InsertIngoreSystemAnnotateType()
	}

	dirManager.InitMainDir()
	for _, oneFloder := range workspaceFolder {
		log.Debug("floder=%s", oneFloder.URI)

		folderPath := pathpre.VscodeURIToString(oneFloder.URI)
		// 若增加的是当前workspace 文件夹中包含的子文件夹， 则不需要做任何处理
		if dirManager.IsDirExistWorkspace(folderPath) {
			log.Debug("current added dir=%s has existed in the workspaceFolder, not need analysis", folderPath)
			continue
		}
		dirManager.PushOneSubDir(folderPath)
	}

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
	costMsTime := time.Since(timeBegin).Milliseconds()

	// 设置向中心服需要统计的信息
	l.SetOnlineReportParam(clientType, allProject.GetAllFileNumber(), (int)(costMsTime), workspaceFolderNum)

	// 启动一个协程来向中心服务器上报信息
	go l.UDPReportOnline()
	return nil
}

// 初始化时，获取其他后缀关联到的lua 类型  (associalList string[])
func getInitAssociationList(associationInitData interface{}) (associalList []string) {
	assMap, flag := associationInitData.(map[string]interface{})
	if !flag {
		return
	}

	for strType, valueInterface := range assMap {
		strValue, flag := valueInterface.(string)
		if !flag {
			continue
		}

		strValue = strings.ToLower(strValue)
		if strValue != "lua" {
			continue
		}

		associalList = append(associalList, strType)
	}

	return
}

// getDefaultIntialOptions 获取默认的初始化参数
func getDefaultIntialOptions() (initOptions *InitializationOptions) {
	initOptions = &InitializationOptions{
		Client:                         "vsc",
		ReferenceMaxNum:                3000,
		ReferenceDefineFlag:            true,
		LocalRun:                       true,
		AllEnable:                      true,
		CheckSyntax:                    true,
		CheckNoDefine:                  false,
		CheckAfterDefine:               false,
		CheckLocalNoUse:                false,
		CheckTableDuplicateKey:         false,
		CheckReferNoFile:               false,
		CheckAssignParamNum:            false,
		CheckLocalDefineParamNum:       false,
		CheckGotoLable:                 false,
		CheckFuncParam:                 false,
		CheckImportModuleVar:           false,
		CheckIfNotVar:                  false,
		CheckFunctionDuplicateParam:    false,
		CheckBinaryExpressionDuplicate: false,
		CheckErrorOrAlwaysTrue:         false,
		CheckErrorAndAlwaysFalse:       false,
		CheckNoUseAssign:               false,
		CheckAnnotateType:              false,
	}

	return initOptions
}

// getCheckFlagList 获取check的列表
func getCheckFlagList(initOptions *InitializationOptions) (checkFlagList []bool) {
	// 按顺序插入
	checkFlagList = []bool{
		initOptions.AllEnable,
		initOptions.CheckSyntax,
		initOptions.CheckNoDefine,
		initOptions.CheckAfterDefine,
		initOptions.CheckLocalNoUse,
		initOptions.CheckTableDuplicateKey,
		initOptions.CheckReferNoFile,
		initOptions.CheckAssignParamNum,
		initOptions.CheckLocalDefineParamNum,
		initOptions.CheckGotoLable,
		initOptions.CheckFuncParam,
		initOptions.CheckImportModuleVar,
		initOptions.CheckIfNotVar,
		initOptions.CheckFunctionDuplicateParam,
		initOptions.CheckBinaryExpressionDuplicate,
		initOptions.CheckErrorOrAlwaysTrue,
		initOptions.CheckErrorAndAlwaysFalse,
		initOptions.CheckNoUseAssign,
		initOptions.CheckAnnotateType,
	}

	return checkFlagList
}
