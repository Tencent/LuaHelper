package langserver

import (
	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/lspcommon"
	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
	"sync"
	"time"

	"github.com/yinfei8/jrpc2"
	"github.com/yinfei8/jrpc2/handler"
)

// 插件定义的版本号
var clientVerStr string = "0.2.20"

type serverState int

const (
	serverCreated      = serverState(iota)
	serverInitializing //  initialize request
	serverInitialized  // initialized request
	serverShutDown
)

// LspServer lsp调用的全局对象
type LspServer struct {
	// 与客户端json rpc2通信的对象
	server *jrpc2.Server

	// 管理所有Lua工程的对象
	project *check.AllProject

	// 打开文件的缓冲
	fileCache *lspcommon.FileMapCache

	// 所有文件的诊断错误信息, 静态的
	fileErrorMap map[string][]common.CheckError

	// 所有文件的诊断错误信息, 动态的，文件实时修改了，但是没有保存的错误
	fileChangeErrorMap map[string][]common.CheckError

	// 请求互斥锁
	requestMutex sync.Mutex

	// 向中心服务器，需要上报统计的信息
	onlineReport OnlineReport

	// 最后一次获取文档着色功能的时间
	colorTime int64

	// 是否处理过ChangeConfiguration 标记
	changeConfFlag bool

	// 是否允许向中心服务器上报自己的使用
	enableReport bool

	stateMu sync.Mutex
	state   serverState
}

// 管理指针
var lspServer *LspServer = nil

// CreateLspServer 创建全局Glsp管理对象
func CreateLspServer() *LspServer {
	lspServer = &LspServer{
		server:             nil,
		project:            nil,
		fileErrorMap:       map[string][]common.CheckError{},
		fileChangeErrorMap: map[string][]common.CheckError{},
		fileCache:          lspcommon.CreateFileMapCache(),
		onlineReport: OnlineReport{
			ClientType:  "vsc",
			FileNumber:  0,
			OsType:      "vsc",
			ClientVer:   clientVerStr,
			FirstReport: 1,
		},
		colorTime:      0,
		changeConfFlag: false,
	}

	return lspServer
}

// CreateServer 创建server
func CreateServer() *jrpc2.Server {
	lspServer := CreateLspServer()

	lspServer.server = jrpc2.NewServer(handler.Map{
		"initialize":                          handler.New(lspServer.Initialize),
		"initialized":                         handler.New(lspServer.Initialized),
		"textDocument/didChange":              handler.New(lspServer.TextDocumentDidChange),
		"textDocument/didSave":                handler.New(lspServer.TextDocumentDidSave),
		"textDocument/didOpen":                handler.New(lspServer.TextDocumentDidOpen),
		"textDocument/didClose":               handler.New(lspServer.TextDocumentDidClose),
		"textDocument/definition":             handler.New(lspServer.TextDocumentDefine),
		"textDocument/hover":                  handler.New(lspServer.TextDocumentHover),
		"textDocument/references":             handler.New(lspServer.TextDocumentReferences),
		"textDocument/documentSymbol":         handler.New(lspServer.TextDocumentSymbol),
		"textDocument/rename":                 handler.New(lspServer.TextDocumentRename),
		"textDocument/documentHighlight":      handler.New(lspServer.TextDocumentHighlight),
		"textDocument/signatureHelp":          handler.New(lspServer.TextDocumentSignatureHelp),
		"textDocument/documentColor":          handler.New(lspServer.TextDocumentColor),
		"textDocument/codeLens":               handler.New(lspServer.TextDocumentCodeLens),
		"textDocument/documentLink":           handler.New(lspServer.TextDocumentdocumentLink),
		"textDocument/completion":             handler.New(lspServer.TextDocumentComplete),
		"completionItem/resolve":              handler.New(lspServer.TextDocumentCompleteResolve),
		"workspace/didChangeConfiguration":    handler.New(lspServer.ChangeConfiguration),
		"workspace/didChangeWorkspaceFolders": handler.New(lspServer.WorkspaceChangeWorkspaceFolders),
		"workspace/didChangeWatchedFiles":     handler.New(lspServer.WorkspaceChangeWatchedFiles),
		"workspace/symbol":                    handler.New(lspServer.WorkspaceSymbolRequest),
		"luahelper/getVarColor":               handler.New(lspServer.TextDocumentGetVarColor),
		"luahelper/getOnlineReq":              handler.New(lspServer.GetOnlineReq),
		"$/cancelRequest":                     handler.New(lspServer.CancelRequest),
		"shutdown":                            handler.New(lspServer.Shutdown),
		"exit":                                handler.New(lspServer.Exit),
	}, &jrpc2.ServerOptions{
		AllowPush:   true,
		Concurrency: 4,
		Logger:      log.LspLog,
	})

	return lspServer.server
}

//getAllProject 获取CheckProject
func (g *LspServer) getAllProject() *check.AllProject {
	return g.project
}

//getFileCache 获取文件缓冲map
func (g *LspServer) getFileCache() *lspcommon.FileMapCache {
	return g.fileCache
}

// setColorTime 设置获取color着色的时间
func (g *LspServer) setColorTime(timeValue int64) {
	g.colorTime = timeValue
}

// isCanHighlight 判断是否可以对变量着色功能, 防止修改文件过程中频繁调用着色功能
func (g *LspServer) isCanHighlight() bool {
	// 如果修改文件的时间太频繁，返回false
	nowTime := time.Now().Unix()
	return nowTime-g.colorTime >= 3
}

// setConfigSet 设置配置
func setConfigSet(referenceMaxNum int, referenceDefine bool) {
	common.GConfig.ReferenceDefineFlag = referenceDefine
	if referenceMaxNum > 0 {
		common.GConfig.ReferenceMaxNum = referenceMaxNum
	}
}

// commFileRequest 通用的文件处理请求
type commFileRequest struct {
	result   bool         // 处理结果
	strFile  string       // 文件名
	contents []byte       // 文件具体的内容
	offset   int          // 光标的偏移
	pos      lsp.Position // 请求的行与列
}

// beginFileRequest 通用的文件处理请求预处理
func (l *LspServer) beginFileRequest(url lsp.DocumentURI, pos lsp.Position) (fileRequest commFileRequest) {
	fileRequest.result = false

	strFile := pathpre.VscodeURIToString(string(url))
	project := l.getAllProject()
	if !project.IsNeedHandle(strFile) {
		log.Debug("not need to handle strFile=%s", strFile)
		return
	}

	fileCache := l.getFileCache()
	contents, found := fileCache.GetFileContent(strFile)
	if !found {
		log.Error("file %s not find contents", strFile)
		return
	}

	offset, err := lspcommon.OffsetForPosition(contents, (int)(pos.Line), (int)(pos.Character))
	if err != nil {
		log.Error("file position error=%s", err.Error())
		return
	}

	fileRequest.result = true
	fileRequest.strFile = strFile
	fileRequest.contents = contents
	fileRequest.offset = offset
	fileRequest.pos = pos
	return
}
