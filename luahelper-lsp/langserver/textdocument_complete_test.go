package langserver

import (
	"context"
	"io/ioutil"
	lsp "luahelper-lsp/langserver/protocol"
	"path/filepath"
	"runtime"
	"testing"
)

func TestComplete1(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
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

	changParams := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{
				Range: &lsp.Range{
					Start: lsp.Position{
						Line:      0,
						Character: 14,
					},
					End: lsp.Position{
						Line:      0,
						Character: 14,
					},
				},
				RangeLength: 0,
				Text:        "ab",
			},
		},
	}

	lspServer.TextDocumentDidChange(context, changParams)

	completionParams := lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
			Position: lsp.Position{
				Line:      0,
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

	completionListTmp, _ := completionReturn.(CompletionListTmp)

	t.Logf("comlete file:%s, completionLen=%d", fileName, len(completionListTmp.Items))
}

type TestCompleteInfo struct {
	changeRange lsp.Range    // 代码修改的Range坐标
	changText   string       // 修改的内容
	compLoc     lsp.Position // 代码补全的坐标信息
	resultList  []string     // 所有的提示的结构字符串
}

func TestComplete2(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test2.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}

	// 1)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      5,
				Character: 15,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 2)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      5,
				Character: 15,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ":"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 3)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      17,
				Character: 36,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 4
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      17,
				Character: 36,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ":"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 5)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      21,
				Character: 27,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 6)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      21,
				Character: 27,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ":"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 7)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      25,
				Character: 45,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 8)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      25,
				Character: 45,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ":"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 9)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      28,
				Character: 54,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 10)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      28,
				Character: 54,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ":"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 11)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      45,
				Character: 28,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 12)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      45,
				Character: 28,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ":"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 13)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      46,
				Character: 28,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 14)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      46,
				Character: 28,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ":"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 15)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      47,
				Character: 4,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn8."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 16)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      47,
				Character: 4,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn8:"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 17)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn6."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 18)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn6:"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 19)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn3."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 20)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn3:"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 21)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn4."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 22)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn4:"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 23)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn1."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 24)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn1:"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 25)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn7."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY", "x_", "y_"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 26)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      50,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "btn7:"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"new", "setX", "setY"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	for index, oneComplete := range testCompleteList {
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

		changParams := lsp.DidChangeTextDocumentParams{
			TextDocument: lsp.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI(fileName),
				},
			},
			ContentChanges: []lsp.TextDocumentContentChangeEvent{
				{
					Range:       &oneComplete.changeRange,
					RangeLength: 0,
					Text:        oneComplete.changText,
				},
			},
		}

		lspServer.TextDocumentDidChange(context, changParams)

		completionParams := lsp.CompletionParams{
			TextDocumentPositionParams: lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI(fileName),
				},
				Position: oneComplete.compLoc,
			},
			Context: lsp.CompletionContext{
				TriggerKind: lsp.CompletionTriggerKind(1),
			},
		}

		completionReturn, err2 := lspServer.TextDocumentComplete(context, completionParams)
		if err2 != nil {
			t.Fatalf("complete file:%s err=%s", fileName, err2.Error())
		}

		completionListTmp, _ := completionReturn.(CompletionListTmp)

		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}
	}
}

func TestComplete3(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test3.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}

	// 1)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 34,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"b", "c", "f", "d", "e"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	//2
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 34,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ".f."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"d", "g"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	//3
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 34,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = ".d."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"a", "b"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 4)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      1,
				Character: 35,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"b", "c", "f", "d", "e"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 下面这个没有完成
	// 5)
	// {
	// 	var oneComplete TestCompleteInfo
	// 	oneComplete.changeRange = lsp.Range{
	// 		Start: lsp.Position{
	// 			Line:      1,
	// 			Character: 35,
	// 		},
	// 	}
	// 	oneComplete.changeRange.End = oneComplete.changeRange.Start
	// 	oneComplete.changText = ".d."
	// 	oneComplete.compLoc = lsp.Position{
	// 		Line:      oneComplete.changeRange.Start.Line,
	// 		Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
	// 	}
	// 	oneComplete.resultList = []string{"a", "b"}
	// 	testCompleteList = append(testCompleteList, oneComplete)
	// }

	for index, oneComplete := range testCompleteList {
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

		changParams := lsp.DidChangeTextDocumentParams{
			TextDocument: lsp.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI(fileName),
				},
			},
			ContentChanges: []lsp.TextDocumentContentChangeEvent{
				{
					Range:       &oneComplete.changeRange,
					RangeLength: 0,
					Text:        oneComplete.changText,
				},
			},
		}

		lspServer.TextDocumentDidChange(context, changParams)

		completionParams := lsp.CompletionParams{
			TextDocumentPositionParams: lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{
					URI: lsp.DocumentURI(fileName),
				},
				Position: oneComplete.compLoc,
			},
			Context: lsp.CompletionContext{
				TriggerKind: lsp.CompletionTriggerKind(1),
			},
		}

		completionReturn, err2 := lspServer.TextDocumentComplete(context, completionParams)
		if err2 != nil {
			t.Fatalf("complete file:%s err=%s", fileName, err2.Error())
		}

		completionListTmp, _ := completionReturn.(CompletionListTmp)

		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}
	}
}
