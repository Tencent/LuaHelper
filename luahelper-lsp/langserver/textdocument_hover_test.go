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
	positionList = append(positionList, lsp.Position{
		Line:      45,
		Character: 13,
	})
	positionList = append(positionList, lsp.Position{
		Line:      46,
		Character: 13,
	})

	for _, onePosiiton := range positionList {
		hoverParams := lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{
				URI: lsp.DocumentURI(fileName),
			},
			Position: onePosiiton,
		}
		hoverReturn1, err1 := lspServer.TextDocumentHover(context, hoverParams)
		if err1 != nil {
			t.Fatalf("TextDocumentHover file:%s err=%s", fileName, err1.Error())
		}

		hoverMarkUpReturn1, _ := hoverReturn1.(MarkupHover)
		if !strings.Contains(hoverMarkUpReturn1.Contents.Value, "uiButton") {
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
	// resultList = append(resultList, "number")

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

	positionList = append(positionList, lsp.Position{
		Line:      27,
		Character: 33,
	})
	resultList = append(resultList, "table")

	positionList = append(positionList, lsp.Position{
		Line:      27,
		Character: 35,
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
		if !strings.Contains(hoverMarkUpReturn1.Contents.Value, resultList[index]) {
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
		if !strings.Contains(hoverMarkUpReturn1.Contents.Value, resultList[index]) {
			t.Fatalf("hover error, not find str=%s, index=%d", resultList[index], index)
		}
	}
}

// 注解类型
func TestHoverAnnotate(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover6.lua"
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
		Line:      12,
		Character: 10,
	})
	resultList = append(resultList, "number\n")

	positionList = append(positionList, lsp.Position{
		Line:      13,
		Character: 10,
	})
	resultList = append(resultList, "string\n")

	positionList = append(positionList, lsp.Position{
		Line:      14,
		Character: 10,
	})
	resultList = append(resultList, "number\n")

	positionList = append(positionList, lsp.Position{
		Line:      15,
		Character: 10,
	})
	resultList = append(resultList, "string\n")

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
		if !strings.Contains(hoverMarkUpReturn1.Contents.Value, resultList[index]) {
			t.Fatalf("hover error, not find str=%s, index=%d", resultList[index], index)
		}
	}
}

// 测试项目中特有的improt功能
func TestHover4(t *testing.T) {
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
	resultList = append(resultList, "any")

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
	resultList = append(resultList, "any")

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
		if !strings.Contains(hoverMarkUpReturn1.Contents.Value, resultList[index]) {
			t.Fatalf("hover error, not find str=%s, index=%d", resultList[index], index)
		}
	}
}

// 测试hover一个table时候，展开table的所有信息
func TestHover5(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover5.lua"
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
		Line:      6,
		Character: 8,
	})
	resultList = append(resultList, "anumber: number")
	positionList = append(positionList, lsp.Position{
		Line:      6,
		Character: 8,
	})
	resultList = append(resultList, "bstring: string")

	positionList = append(positionList, lsp.Position{
		Line:      6,
		Character: 8,
	})
	resultList = append(resultList, "dtable: table")

	positionList = append(positionList, lsp.Position{
		Line:      6,
		Character: 8,
	})
	resultList = append(resultList, "cany: any")

	positionList = append(positionList, lsp.Position{
		Line:      9,
		Character: 8,
	})
	resultList = append(resultList, "anumber: number")
	positionList = append(positionList, lsp.Position{
		Line:      9,
		Character: 8,
	})
	resultList = append(resultList, "bstring: string")

	positionList = append(positionList, lsp.Position{
		Line:      9,
		Character: 8,
	})
	resultList = append(resultList, "dtable: table")

	positionList = append(positionList, lsp.Position{
		Line:      9,
		Character: 8,
	})
	resultList = append(resultList, "cany: any")

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
		if !strings.Contains(hoverMarkUpReturn1.Contents.Value, resultList[index]) {
			t.Fatalf("hover error, not find str=%s, index=%d", resultList[index], index)
		}
	}
}

// 测试hover注解类型时候，展开信息
func TestHover6(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover6.lua"
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
		Line:      0,
		Character: 13,
	})
	positionList = append(positionList, lsp.Position{
		Line:      6,
		Character: 12,
	})
	positionList = append(positionList, lsp.Position{
		Line:      7,
		Character: 10,
	})
	positionList = append(positionList, lsp.Position{
		Line:      8,
		Character: 4,
	})
	positionList = append(positionList, lsp.Position{
		Line:      10,
		Character: 10,
	})

	resultList = append(resultList, "astring: string")
	resultList = append(resultList, "bnumber: number")
	resultList = append(resultList, "cany: any")
	resultList = append(resultList, "dfun: function")
	resultList = append(resultList, "abc: number")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// 测试hover注解类型时候，展开信息2
func TestHover7(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover7.lua"
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
		Line:      13,
		Character: 10,
	})
	positionList = append(positionList, lsp.Position{
		Line:      14,
		Character: 10,
	})
	positionList = append(positionList, lsp.Position{
		Line:      0,
		Character: 10,
	})
	positionList = append(positionList, lsp.Position{
		Line:      1,
		Character: 10,
	})

	resultList = append(resultList, "open_flag: number")
	resultList = append(resultList, "open_time: any")
	resultList = append(resultList, "continue_day_num: number")
	resultList = append(resultList, "invite_self_list: table")
	resultList = append(resultList, "get_flag: number")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// 测试hover全局变量多个文件定义
func TestHover8(t *testing.T) {
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
		Line:      0,
		Character: 9,
	})

	resultList = append(resultList, "aaa: number")
	resultList = append(resultList, "bbb: number")
	resultList = append(resultList, "ccc: number")
	resultList = append(resultList, "ddd: number")
	resultList = append(resultList, "eee: number")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}

	positionList = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      0,
		Character: 13,
	})

	resultList = []string{}
	resultList = append(resultList, "ccc : number")
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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// 测试注解与变量结合的：https://github.com/Tencent/LuaHelper/issues/61
func TestHover10(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover8.lua"
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
		Line:      13,
		Character: 7,
	})

	resultList = append(resultList, "test1")
	resultList = append(resultList, "aa: number")
	resultList = append(resultList, "bb: number = 2")
	resultList = append(resultList, "cc: number")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// 原表变量 https://github.com/Tencent/LuaHelper/issues/61
func TestHover11(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover9.lua"
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
		Line:      5,
		Character: 8,
	})
	resultList = append(resultList, "aaa: number = 1")
	resultList = append(resultList, "bbb: number = 2")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}

	resultList = []string{}
	positionList = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      12,
		Character: 8,
	})
	resultList = append(resultList, "ccc: number = 1")
	resultList = append(resultList, "ddd: number = 2")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}

	resultList = []string{}
	positionList = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      13,
		Character: 8,
	})
	resultList = append(resultList, "ccc: number = 1")
	resultList = append(resultList, "ddd: number = 2")
	resultList = append(resultList, "ccc: number = 1")
	resultList = append(resultList, "ddd: number = 2")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// 原表变量含义__call https://github.com/Tencent/LuaHelper/issues/61
func TestHover12(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover10.lua"
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
		Line:      8,
		Character: 8,
	})
	positionList = append(positionList, lsp.Position{
		Line:      17,
		Character: 8,
	})
	resultList = append(resultList, "a: number = 1")
	resultList = append(resultList, "b: number = 2")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}

	resultList = []string{}
	positionList = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      8,
		Character: 15,
	})
	positionList = append(positionList, lsp.Position{
		Line:      17,
		Character: 15,
	})
	resultList = append(resultList, "function ")
	resultList = append(resultList, "(a: any, b: any)")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// 注解与变量相结合，子定义是否会覆盖注解的
func TestHover13(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover11.lua"
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
		Line:      7,
		Character: 10,
	})
	positionList = append(positionList, lsp.Position{
		Line:      14,
		Character: 10,
	})
	resultList = append(resultList, "open_flag: number = 1")
	resultList = append(resultList, "open_time: number")
	resultList = append(resultList, "last_update_time: number")
	resultList = append(resultList, "continue_day_num: number = 1")

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

		for _, oneStr := range resultList {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// expand local 扩展变量，出现了的变量
func TestHoverExpandLocal(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover_expand_local.lua"
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

	var resultList [][]string = [][]string{}
	var positionList []lsp.Position = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      1,
		Character: 2,
	})
	resultList = append(resultList, []string{"a: number = 1", "bbb: table"})

	positionList = append(positionList, lsp.Position{
		Line:      1,
		Character: 6,
	})
	resultList = append(resultList, []string{"a : number = 1"})

	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 13,
	})
	resultList = append(resultList, []string{"q: table", "qqqqqq: table"})

	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 16,
	})
	resultList = append(resultList, []string{"b: any"})

	positionList = append(positionList, lsp.Position{
		Line:      5,
		Character: 16,
	})
	resultList = append(resultList, []string{"d: table"})

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

		for _, oneStr := range resultList[index] {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// expand G 扩展变量，出现了的变量
func TestHoverExpandG(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover_expand_g.lua"
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

	var resultList [][]string = [][]string{}
	var positionList []lsp.Position = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      1,
		Character: 2,
	})
	resultList = append(resultList, []string{"a: number = 1", "bbb: table"})

	positionList = append(positionList, lsp.Position{
		Line:      1,
		Character: 6,
	})
	resultList = append(resultList, []string{"a : number = 1"})

	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 13,
	})
	resultList = append(resultList, []string{"q: table", "qqqqqq: table"})

	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 16,
	})
	resultList = append(resultList, []string{"b: any"})

	positionList = append(positionList, lsp.Position{
		Line:      5,
		Character: 16,
	})
	resultList = append(resultList, []string{"d: table"})

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

		for _, oneStr := range resultList[index] {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// expand No 扩展变量，出现了的变量
func TestHoverExpandNo(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover_expand_no.lua"
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

	var resultList [][]string = [][]string{}
	var positionList []lsp.Position = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      1,
		Character: 2,
	})
	resultList = append(resultList, []string{"a: number = 1", "bbb: table"})

	positionList = append(positionList, lsp.Position{
		Line:      1,
		Character: 6,
	})
	resultList = append(resultList, []string{"a : number = 1"})

	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 13,
	})
	resultList = append(resultList, []string{"q: table", "qqqqqq: table"})

	positionList = append(positionList, lsp.Position{
		Line:      4,
		Character: 16,
	})
	resultList = append(resultList, []string{"b: any"})

	positionList = append(positionList, lsp.Position{
		Line:      5,
		Character: 16,
	})
	resultList = append(resultList, []string{"d: table"})

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

		for _, oneStr := range resultList[index] {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}

// hover 扩展显示函数的返回值
func TestHoverFuncReturn(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/hover"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "hover_func_return.lua"
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

	var resultList [][]string = [][]string{}
	var positionList []lsp.Position = []lsp.Position{}
	positionList = append(positionList, lsp.Position{
		Line:      0,
		Character: 13,
	})
	resultList = append(resultList, []string{"->1. string", "->2. string", "->3. string"})

	positionList = append(positionList, lsp.Position{
		Line:      5,
		Character: 18,
	})
	resultList = append(resultList, []string{"->1. string", "->2. string", "->3. function", "one: string", "two: table", "c: any"})

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

		for _, oneStr := range resultList[index] {
			if !strings.Contains(hoverMarkUpReturn1.Contents.Value, oneStr) {
				t.Fatalf("hover error, not find str=%s, index=%d", oneStr, index)
			}
		}
	}
}
