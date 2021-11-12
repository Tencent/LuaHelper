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

func TestHover1(t *testing.T) {
	// 修复 https://github.com/Tencent/LuaHelper/issues/42
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover1.lua"
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

	var positionList []lsp.Position = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      21,
		Character: 8,
	})
	positionList = append(positionList, lsp.Position{
		Line:      23,
		Character: 8,
	})
	positionList = append(positionList, lsp.Position{
		Line:      25,
		Character: 8,
	})
	positionList = append(positionList, lsp.Position{
		Line:      28,
		Character: 8,
	})
	positionList = append(positionList, lsp.Position{
		Line:      30,
		Character: 8,
	})
	positionList = append(positionList, lsp.Position{
		Line:      39,
		Character: 8,
	})
	positionList = append(positionList, lsp.Position{
		Line:      40,
		Character: 8,
	})

	for _, onePoisiton := range positionList {
		hoverParams := lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
			Position: onePoisiton,
		}
		hoverReturn1, err1 := lspServer.TextDocumentHover(context, hoverParams)
		if err1 != nil {
			t.Fatalf("TextDocumentHover file:%s err=%s", fileName, err1.Error())
		}

		hoverMarkUpReturn1, _ := hoverReturn1.(MarkupHover)
		if strings.Index(hoverMarkUpReturn1.Contents.Value, "uiButton") < 0 {
			t.Fatalf("hover error, not find uiButton")
		}
	}
}

func TestHover2(t *testing.T) {
	// 测试require另外一个文件
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover2.lua"
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

	var resultList []string = []string{}

	var positionList []lsp.Position = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      2,
		Character: 13,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      3,
		Character: 13,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 13,
	})
	resultList = append(resultList, "table")

	positionList = append(positionList, lsp.Position{
		Line:      5,
		Character: 15,
	})
	resultList = append(resultList, "number")

	
	positionList = append(positionList, lsp.Position{
		Line:      6,
		Character: 13,
	})
	resultList = append(resultList, "table")

	positionList = append(positionList, lsp.Position{
		Line:      7,
		Character: 15,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      10,
		Character: 13,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      13,
		Character: 13,
	})
	resultList = append(resultList, "number")

	/** hover2_requrie2 **/
	positionList = append(positionList, lsp.Position{
		Line:      19,
		Character: 13,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      20,
		Character: 13,
	})
	resultList = append(resultList, "table")

	// 这条用例暂时过不了
	// positionList = append(positionList, lsp.Position{
	// 	Line:      21,
	// 	Character: 15,
	// })
	//resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      22,
		Character: 13,
	})
	resultList = append(resultList, "table")


	positionList = append(positionList, lsp.Position{
		Line:      25,
		Character: 11,
	})
	resultList = append(resultList, "table")
	
	positionList = append(positionList, lsp.Position{
		Line:      25,
		Character: 13,
	})
	resultList = append(resultList, "number")


	for index, onePoisiton := range positionList {
		hoverParams := lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
			Position: onePoisiton,
		}
		hoverReturn1, err1 := lspServer.TextDocumentHover(context, hoverParams)
		if err1 != nil {
			t.Fatalf("TextDocumentHover file:%s err=%s", fileName, err1.Error())
		}

		hoverMarkUpReturn1, _ := hoverReturn1.(MarkupHover)
		if strings.Index(hoverMarkUpReturn1.Contents.Value, resultList[index]) < 0 {
			t.Fatalf("hover error, not find str=%s, index=%d", resultList[index], index)
		}
	}
}

func TestHover3(t *testing.T) {
	// 简单的hover测试
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover3.lua"
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

	var resultList []string = []string{}

	var positionList []lsp.Position = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      18,
		Character: 13,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      19,
		Character: 13,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      20,
		Character: 13,
	})
	resultList = append(resultList, "table")

	positionList = append(positionList, lsp.Position{
		Line:      21,
		Character: 15,
	})
	resultList = append(resultList, "number")

	positionList = append(positionList, lsp.Position{
		Line:      22,
		Character: 13,
	})
	resultList = append(resultList, "table")

	
	for index, onePoisiton := range positionList {
		hoverParams := lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
			Position: onePoisiton,
		}
		hoverReturn1, err1 := lspServer.TextDocumentHover(context, hoverParams)
		if err1 != nil {
			t.Fatalf("TextDocumentHover file:%s err=%s", fileName, err1.Error())
		}

		hoverMarkUpReturn1, _ := hoverReturn1.(MarkupHover)
		if strings.Index(hoverMarkUpReturn1.Contents.Value, resultList[index]) < 0 {
			t.Fatalf("hover error, not find str=%s, index=%d", resultList[index], index)
		}
	}
}

func TestHover4(t *testing.T) {
	// 测试项目中特有的improt功能
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover4.lua"
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

	var resultList []string = []string{}
	var positionList []lsp.Position = []lsp.Position{}

	positionList = append(positionList, lsp.Position{
		Line:      2,
		Character: 13,
	})
	resultList = append(resultList, "import")

	positionList = append(positionList, lsp.Position{
		Line:      3,
		Character: 13,
	})
	resultList = append(resultList, "number")
	
	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 13,
	})
	resultList = append(resultList, "function")

	positionList = append(positionList, lsp.Position{
		Line:      5,
		Character: 13,
	})
	resultList = append(resultList, "import")


	for index, onePoisiton := range positionList {
		hoverParams := lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
			Position: onePoisiton,
		}
		hoverReturn1, err1 := lspServer.TextDocumentHover(context, hoverParams)
		if err1 != nil {
			t.Fatalf("TextDocumentHover file:%s err=%s", fileName, err1.Error())
		}

		hoverMarkUpReturn1, _ := hoverReturn1.(MarkupHover)
		if strings.Index(hoverMarkUpReturn1.Contents.Value, resultList[index]) < 0 {
			t.Fatalf("hover error, not find str=%s, index=%d", resultList[index], index)
		}
	}
}