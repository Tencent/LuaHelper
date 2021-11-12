package langserver

import (
	"context"
	"io/ioutil"
	lsp "luahelper-lsp/langserver/protocol"
	"path/filepath"
	"runtime"
	"testing"
)

func TestComplete1 (t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)
 
	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootUri := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootUri)
	context := context.Background()

	fileName := strRootPath + "/" + "test1.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil { 
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	openParams := lsp.DidOpenTextDocumentParams{
		TextDocument :lsp.TextDocumentItem{
			URI: lsp.DocumentURI(fileName),
			Text: string(data),
		},
	}
	err1 := lspServer.TextDocumentDidOpen(context, openParams)
	if err1 != nil {
		t.Fatalf("didopen file:%s err=%s", fileName, err1.Error())
	}
	
	changParams := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier {
			TextDocumentIdentifier: lsp.TextDocumentIdentifier {
				URI: lsp.DocumentURI(fileName),
			},
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent {
			{
				Range: &lsp.Range{
					Start: lsp.Position{
						Line: 0,
						Character: 14,
					},
					End: lsp.Position{
						Line: 0,
						Character: 14,
					},
				},
				RangeLength: 0,
				Text: "ab",
			},
		},
	}

	lspServer.TextDocumentDidChange(context, changParams)


	completionParams := lsp.CompletionParams{
		TextDocumentPositionParams:lsp.TextDocumentPositionParams{
			TextDocument:lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
			Position: lsp.Position{
				Line: 0,
				Character: 16,
			},
		},
		Context: lsp.CompletionContext{
			TriggerKind: lsp.CompletionTriggerKind(1),
		},
	}

	completionReturn, err2 := lspServer.TextDocumentComplete(context, completionParams)
	if err2 != nil {
		t.Fatalf("complete file:%s err=%s", fileName, err2.Error())
	}

	completionListTmp, _:= completionReturn.(CompletionListTmp)

	t.Logf("comlete file:%s, completionLen=%d", fileName, len(completionListTmp.Items))
}