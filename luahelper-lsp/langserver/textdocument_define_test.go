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

func TestProjectDefine(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/define"
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

	onePosition := lsp.Position{
		Line:      33,
		Character: 4,
	}
	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      29,
			Character: 6,
		},
		End: lsp.Position{
			Line:      29,
			Character: 9,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}
	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}

// define 跳转到文件，例如 require("one") 跳转到one/init.lua
func TestProjectDefineFile1(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/define"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test2.lua"
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

	onePosition := lsp.Position{
		Line:      0,
		Character: 27,
	}

	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      0,
			Character: 0,
		},
		End: lsp.Position{
			Line:      0,
			Character: 1,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}
	strSuffix := "init.lua"
	if !strings.HasSuffix(string(resLocationList[0].URI), strSuffix) {
		t.Fatalf("define file error")
	}

	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}

// define 跳转到文件，例如 require("one") 跳转到one.lua
func TestProjectDefineFile2(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/define"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test3.lua"
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

	onePosition := lsp.Position{
		Line:      0,
		Character: 27,
	}

	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      0,
			Character: 0,
		},
		End: lsp.Position{
			Line:      0,
			Character: 1,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}
	strSuffix := "be_define.lua"
	if !strings.HasSuffix(string(resLocationList[0].URI), strSuffix) {
		t.Fatalf("define file error")
	}

	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}

// define 跳转到文件，例如 import("one.lua") 跳转到one.lua
func TestProjectDefineFile3(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/define"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test4.lua"
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

	onePosition := lsp.Position{
		Line:      0,
		Character: 27,
	}

	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      0,
			Character: 0,
		},
		End: lsp.Position{
			Line:      0,
			Character: 1,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}
	strSuffix := "be_define.lua"
	if !strings.HasSuffix(string(resLocationList[0].URI), strSuffix) {
		t.Fatalf("define file error")
	}

	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}

// define 多个文件定义全局变量，跳转到不同的文件1
func TestProjectDefine4(t *testing.T) {
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

	onePosition := lsp.Position{
		Line:      1,
		Character: 13,
	}

	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      0,
			Character: 8,
		},
		End: lsp.Position{
			Line:      0,
			Character: 11,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}
	strSuffix := "two1.lua"
	if !strings.HasSuffix(string(resLocationList[0].URI), strSuffix) {
		t.Fatalf("define file error")
	}

	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}

// define 多个文件定义全局变量，跳转到不同的文件2
func TestProjectDefine5(t *testing.T) {
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

	onePosition := lsp.Position{
		Line:      0,
		Character: 13,
	}

	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      0,
			Character: 5,
		},
		End: lsp.Position{
			Line:      0,
			Character: 8,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}
	strSuffix := "two.lua"
	if !strings.HasSuffix(string(resLocationList[0].URI), strSuffix) {
		t.Fatalf("define file error")
	}

	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}

// define 当方法里local与self同名时，self指向正确的table
func TestProjectDefineFile6(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/define"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test5.lua"
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

	onePosition := lsp.Position{
		Line:      4,
		Character: 6,
	}

	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      0,
			Character: 6,
		},
		End: lsp.Position{
			Line:      0,
			Character: 9,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}

	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}


// 跳转指向self的函数
func TestProjectDefine7(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)

	strRootPath := paths + "../testdata/define"
	strRootPath, _ = filepath.Abs(strRootPath)

	strRootURI := "file://" + strRootPath
	lspServer := createLspTest(strRootPath, strRootURI)
	context := context.Background()

	fileName := strRootPath + "/" + "test5.lua"
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

	onePosition := lsp.Position{
		Line:      4,
		Character: 10,
	}

	resultRange := lsp.Range{
		Start: lsp.Position{
			Line:      7,
			Character: 13,
		},
		End: lsp.Position{
			Line:      7,
			Character: 16,
		},
	}

	defineParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: lsp.DocumentURI(fileName),
		},
		Position: onePosition,
	}

	resLocationList, err2 := lspServer.TextDocumentDefine(context, defineParams)
	if err2 != nil {
		t.Fatalf("define error")
	}
	if len(resLocationList) != 1 {
		t.Fatalf("location size error")
	}

	res0 := resLocationList[0].Range

	if res0.Start.Line != resultRange.Start.Line || res0.Start.Character != resultRange.Start.Character ||
		res0.End.Line != resultRange.End.Line || res0.End.Character != resultRange.End.Character {
		t.Fatalf("location error")
	}
}