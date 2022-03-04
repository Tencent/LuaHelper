package langserver

import (
	"context"
	"io/ioutil"
	lsp "luahelper-lsp/langserver/protocol"
	"path/filepath"
	"runtime"
	"strings"
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
	changeRange  lsp.Range    // 代码修改的Range坐标
	changText    string       // 修改的内容
	compLoc      lsp.Position // 代码补全的坐标信息
	resultList   []string     // 所有的提示的结构字符串
	noResultList []string     // 不应该出现的结果列表
	resolveList  []string     // 当选择某一个子项时，具体显示的内容
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
	//5)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      1,
				Character: 35,
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

// # table的长度代码补全
func TestComplete4(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test4.lua"
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
				Line:      3,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "#"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"#abcdefg"}
		oneComplete.noResultList = []string{"#and", "#if", "#while"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	//2
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      3,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "#a"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"#abcdefg"}
		oneComplete.noResultList = []string{"#and", "#if", "#while"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	//2
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      3,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "#abcdefg."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"abc"}
		oneComplete.noResultList = []string{"#abcdefg.abc", "#abc", "#while"}
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

		for _, noStr := range oneComplete.noResultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if noStr == oneCompReturn.Label {
					findFlag = true
				}
			}

			if findFlag {
				t.Fatalf("noResultList find complete index=%d, str=%s", index, noStr)
			}
		}
	}
}

// 测试5，返回的table有嵌套的table代码补全
func TestComplete5(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test5.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	// 0)
	oneComplete.changeRange = lsp.Range{
		Start: lsp.Position{
			Line:      23,
			Character: 0,
		},
	}
	oneComplete.changeRange.End = oneComplete.changeRange.Start
	oneComplete.changText = "f1.b."
	oneComplete.compLoc = lsp.Position{
		Line:      oneComplete.changeRange.Start.Line,
		Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
	}
	oneComplete.resultList = []string{"b1"}
	testCompleteList = append(testCompleteList, oneComplete)
	// 1)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      23,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "f1.b.b1."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"b2"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 2)
	oneComplete.changeRange = lsp.Range{
		Start: lsp.Position{
			Line:      23,
			Character: 0,
		},
	}
	oneComplete.changeRange.End = oneComplete.changeRange.Start
	oneComplete.changText = "c2.b."
	oneComplete.compLoc = lsp.Position{
		Line:      oneComplete.changeRange.Start.Line,
		Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
	}
	oneComplete.resultList = []string{"b1"}
	testCompleteList = append(testCompleteList, oneComplete)

	// 3)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      23,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "c2.b.b1."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"b2"}
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

		for _, noStr := range oneComplete.noResultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if noStr == oneCompReturn.Label {
					findFlag = true
				}
			}

			if findFlag {
				t.Fatalf("noResultList find complete index=%d, str=%s", index, noStr)
			}
		}
	}
}

// 测试6，补全table的时候，展开table详细的信息
func TestComplete6(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test6.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	var resultList []string = []string{}
	resultList = append(resultList, "anumber: number")
	resultList = append(resultList, "bstring: string")
	resultList = append(resultList, "cany: any")
	resultList = append(resultList, "dtable: table")

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      11,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcde"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"abcdef"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 1)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      11,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "cdef12"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"cdef123"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}
		for _, oneStr := range resultList {
			if !strings.Contains(resultItem.Documentation.Value, oneStr) {
				t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d, not str=%s", fileName, index, oneStr)
			}
		}
	}
}

// 测试，注解class，展开table详细的信息
func TestComplete7(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test7.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	var resultList []string = []string{}
	resultList = append(resultList, "astring: string")
	resultList = append(resultList, "bnumber: number")
	resultList = append(resultList, "cany: any")
	resultList = append(resultList, "dfun: function")
	resultList = append(resultList, "abc: number")

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      12,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "aaaa"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"aaaa1"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 1)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      12,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "aaaa"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"aaaaa2"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 2)
	{
		var oneComplete TestCompleteInfo
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      12,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "---@type one"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"one11"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}
		for _, oneStr := range resultList {
			if !strings.Contains(resultItem.Documentation.Value, oneStr) {
				t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d, not str=%s", fileName, index, oneStr)
			}
		}
	}
}

// 测试，注解class，展开table详细的信息2
func TestComplete8(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test8.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	var resultList []string = []string{}
	resultList = append(resultList, "open_flag: number")
	resultList = append(resultList, "open_time: any")
	resultList = append(resultList, "continue_day_num: number")
	resultList = append(resultList, "invite_self_list: table")
	resultList = append(resultList, "get_flag: number")

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      16,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "lose_label_dat"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"lose_label_data"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}
		for _, oneStr := range resultList {
			if !strings.Contains(resultItem.Documentation.Value, oneStr) {
				t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d, not str=%s", fileName, index, oneStr)
			}
		}
	}
}

func TestComplete9(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	pluginPath := paths + "../../luahelper-vscode"
	pluginPath, _ = filepath.Abs(pluginPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTestWithPlugin(strRootPath, strRootURI, pluginPath)
	context := context.Background()

	fileName := strRootPath + "/" + "test9.lua"
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
				Line:      4,
				Character: 19,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"isyieldable", "resume", "status", "yield", "create", "close", "running", "wrap", "start"}
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

// 测试全局table多个文件定义，对它的代码补全
func TestComplete10(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/project/globalVar"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "three.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	var resultList []string = []string{}
	resultList = append(resultList, "number")

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      2,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"aaa", "bbb", "ccc", "ddd", "eee"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}
		for _, oneStr := range resultList {
			if !strings.Contains(resultItem.Documentation.Value, oneStr) {
				t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d, not str=%s", fileName, index, oneStr)
			}
		}
	}
}

// 测试alias的代码补全
func TestCompleteAlias1(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test_alias1.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	var resultList []string = []string{}
	resultList = append(resultList, "Read data")
	resultList = append(resultList, "Write data to this program by `file`")
	resultList = append(resultList, "acddd")

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      13,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "popen4(\"1\",  \""
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"\"w\"", "\"r\""}
		testCompleteList = append(testCompleteList, oneComplete)
	}
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      13,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "popen4(\"1\",  "
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"\"w\"", "\"r\"", "fsfsfsf"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}

		resolveFlag := false
		for _, oneStr := range resultList {
			if strings.Contains(resultItem.Documentation.Value, oneStr) {
				resolveFlag = true

			}
		}
		if !resolveFlag {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d", fileName, index)
		}
	}
}

// 测试alias的代码补全, 返回的内容不包含 " 或是 '
func TestCompleteAlias12(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test_alias1.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	var resultList []string = []string{}
	resultList = append(resultList, "Read data")
	resultList = append(resultList, "Write data to this program by `file`")
	resultList = append(resultList, "acddd")

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      13,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "popen4(\"1\",  \""
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"\"w\"", "\"r\""}
		testCompleteList = append(testCompleteList, oneComplete)
	}
	// 1)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      13,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "popen4(\"1\",  '"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"'w'", "'r'"}
		testCompleteList = append(testCompleteList, oneComplete)
	}
	// 2)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      13,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "popen4(\"1\",  f"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"fsfsfsf"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}

		resolveFlag := false
		for _, oneStr := range resultList {
			if strings.Contains(resultItem.Documentation.Value, oneStr) {
				resolveFlag = true
			}
		}
		if !resolveFlag {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d", fileName, index)
		}
		if strings.Contains(resultItem.InsertText, "\"") || strings.Contains(resultItem.InsertText, "'") {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d contains err", fileName, index)
		}
	}
}

// 测试注解table中的key值补全
func TestCompleteTableKey(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test_table_key.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	var resultList []string = []string{}
	resultList = append(resultList, "dsfosf")
	resultList = append(resultList, "fdsfos")
	resultList = append(resultList, "acddd")

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      7,
				Character: 9,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "n"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"one", "two"}
		testCompleteList = append(testCompleteList, oneComplete)
	}
	// 1)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      8,
				Character: 9,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "o"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"one", "two"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}

		resolveFlag := false
		for _, oneStr := range resultList {
			if strings.Contains(resultItem.Documentation.Value, oneStr) {
				resolveFlag = true
			}
		}
		if !resolveFlag {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d", fileName, index)
		}
		if strings.Contains(resultItem.InsertText, "\"") || strings.Contains(resultItem.InsertText, "'") {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d contains err", fileName, index)
		}
	}
}

// 历史记录智能补全，全局变量
func TestCompleteIntelligent1(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test_intelligent1.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcd"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"abcdd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 1)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"addd", "a", "eff", "dd()", "ccc"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 2)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd:dd()."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"dd1"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 3)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd()."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"ddd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 4)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd().ddd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"a", "dd", "dddp"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 5)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd().ddd.dd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"ddd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 6)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.ccc(one, one)."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"a"}
		oneComplete.resolveList = []string{"table = {"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 6)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.ccc(one, one).a."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"b"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}

		if len(oneComplete.resolveList) == 0 {
			continue
		}

		resolveFlag := false
		for _, oneStr := range oneComplete.resolveList {
			if strings.Contains(resultItem.Documentation.Value, oneStr) {
				resolveFlag = true
			}
		}
		if !resolveFlag {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d", fileName, index)
		}
		if strings.Contains(resultItem.InsertText, "\"") || strings.Contains(resultItem.InsertText, "'") {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d contains err", fileName, index)
		}
	}
}

// 历史记录智能补全, local变量
func TestCompleteIntelligent2(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test_intelligent2.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcd"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"abcdd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 1)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"addd", "a", "eff", "dd()", "ccc"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 2)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd:dd()."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"dd1"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 3)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd()."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"ddd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 4)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd().ddd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"a", "dd", "dddp"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 5)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd().ddd.dd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"ddd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 6)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.ccc(one, one)."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"a"}
		oneComplete.resolveList = []string{"table = {"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 6)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.ccc(one, one).a."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"b"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}

		if len(oneComplete.resolveList) == 0 {
			continue
		}

		resolveFlag := false
		for _, oneStr := range oneComplete.resolveList {
			if strings.Contains(resultItem.Documentation.Value, oneStr) {
				resolveFlag = true
			}
		}
		if !resolveFlag {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d", fileName, index)
		}
		if strings.Contains(resultItem.InsertText, "\"") || strings.Contains(resultItem.InsertText, "'") {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d contains err", fileName, index)
		}
	}
}

// 历史记录智能补全, 未定义的变量
func TestCompleteIntelligent3(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/complete/intelligent3"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test_intelligent3.lua"
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Fatalf("read file:%s err=%s", fileName, err.Error())
	}

	var testCompleteList []TestCompleteInfo = []TestCompleteInfo{}
	var oneComplete TestCompleteInfo

	// 0)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcd"
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"abcdd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 1)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"addd", "a", "eff", "dd()", "ccc"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 2)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd:dd()."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"dd1"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 3)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd()."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"ddd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 4)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd().ddd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"a", "dd", "dddp"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 5)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.addd().ddd.dd."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"ddd"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 6)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.ccc(one, one)."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"a"}
		oneComplete.resolveList = []string{"table = {"}
		testCompleteList = append(testCompleteList, oneComplete)
	}

	// 6)
	{
		oneComplete.changeRange = lsp.Range{
			Start: lsp.Position{
				Line:      15,
				Character: 0,
			},
		}
		oneComplete.changeRange.End = oneComplete.changeRange.Start
		oneComplete.changText = "abcdd.ccc(one, one).a."
		oneComplete.compLoc = lsp.Position{
			Line:      oneComplete.changeRange.Start.Line,
			Character: oneComplete.changeRange.Start.Character + (uint32)(len(oneComplete.changText)),
		}
		oneComplete.resultList = []string{"b"}
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
		var oneItem CompletionItemTmp
		for _, resultStr := range oneComplete.resultList {
			findFlag := false
			for _, oneCompReturn := range completionListTmp.Items {
				if resultStr == oneCompReturn.Label {
					findFlag = true
					oneItem = oneCompReturn
				}
			}

			if !findFlag {
				t.Fatalf("not find complete index=%d, str=%s", index, resultStr)
			}
		}

		var inputItem lsp.CompletionItem = lsp.CompletionItem{
			Label: oneItem.Label,
			Kind:  oneItem.Kind,
			Data:  oneItem.Data,
		}

		resultItem, err3 := lspServer.TextDocumentCompleteResolve(context, inputItem)
		if err3 != nil {
			t.Fatalf("TextDocumentCompleteResolve err, index=%d", index)
		}

		if len(oneComplete.resolveList) == 0 {
			continue
		}

		resolveFlag := false
		for _, oneStr := range oneComplete.resolveList {
			if strings.Contains(resultItem.Documentation.Value, oneStr) {
				resolveFlag = true
			}
		}
		if !resolveFlag {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d", fileName, index)
		}
		if strings.Contains(resultItem.InsertText, "\"") || strings.Contains(resultItem.InsertText, "'") {
			t.Fatalf("TextDocumentCompleteResolve file=%s, index=%d contains err", fileName, index)
		}
	}
}
