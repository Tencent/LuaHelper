package langserver

import (
	"context"
	"luahelper-lsp/langserver/check/common"
	lsp "luahelper-lsp/langserver/protocol"

	"github.com/yinfei8/jrpc2"
	"github.com/yinfei8/jrpc2/handler"
)

func createLspTest(strRootPath string, strRootUri string) *LspServer{
	common.GlobalConfigDefautInit()
	common.GConfig.IntialGlobalVar()
	
	lspServer := CreateLspServer()
	lspServer.server = jrpc2.NewServer(handler.Map{}, &jrpc2.ServerOptions{
		AllowPush:   false,
		Concurrency: 1,
	})

	context := context.Background()
	initializeParams := InitializeParams{
		InitializeParams : lsp.InitializeParams {
			InnerInitializeParams : lsp.InnerInitializeParams{
				RootPath : strRootPath,
				RootURI: lsp.DocumentURI(strRootUri),
			},
		},
	}
	lspServer.Initialize(context, initializeParams)
	return lspServer
}
