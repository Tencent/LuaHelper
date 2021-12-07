package langserver

import (
	"context"
	"io/ioutil"
	lsp "luahelper-lsp/langserver/protocol"
	"path/filepath"
	"runtime"
	"testing"
)

func TestProjectOpen(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/project/test1"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test1.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	openParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:  lsp.DocumentURI(fileName),
			Text: string(data),
		},
	}
	err1 := lspServer.TextDocumentDidOpen(context, openParams)
	if err1 != nil {
		t.Fatalf("didopen file:%s err=%s", fileName, err1.Error())
	}
}
