package langserver

import (
	"context"
	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/lspcommon"
	lsp "luahelper-lsp/langserver/protocol"
	"strings"
)

// TextDocumentDefine 文件中查找变量的的定义
func (l *LspServer) TextDocumentDefine(ctx context.Context, vs lsp.TextDocumentPositionParams) (locList []lsp.Location, err error) {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	fileRequest := l.beginFileRequest(vs.TextDocument.URI, vs.Position)
	if !fileRequest.result {
		log.Error("TextDocumentDefine beginFileRequest false, uri=%s", vs.TextDocument.URI)
		return
	}

	if len(fileRequest.contents) == 0 || fileRequest.offset >= len(fileRequest.contents) {
		return
	}

	strFile := fileRequest.strFile
	project := l.getAllProject()

	// 1）判断查找的定义是否为打开一个文件
	fileList := getOpenFileStr(fileRequest.contents, fileRequest.offset, (int)(fileRequest.pos.Character))
	var openDefineVecs []check.DefineStruct
	for _, strItem := range fileList {
		openDefineVecs = project.FindOpenFileDefine(strFile, strItem)
		if len(openDefineVecs) > 0 {
			break
		}
		locList = defineVecConvert(openDefineVecs)
		return locList, nil
	}
	
	// 2) 判断是否为---@ 注解引入的查找里面的类型定义
	defineAnnotateVecs, flag := l.handleAnnotateTypeDefine(strFile, fileRequest.contents, fileRequest.offset,
		(int)(fileRequest.pos.Line), (int)(fileRequest.pos.Character))
	if flag {
		locList = defineVecConvert(defineAnnotateVecs)
		return locList, nil
	}

	// 3) 其他的查找定义
	varStruct := getVarStruct(fileRequest.contents, fileRequest.offset, fileRequest.pos.Line, fileRequest.pos.Character)
	if !varStruct.ValidFlag || len(varStruct.StrVec) == 0 {
		log.Error("TextDocumentDefine not valid")
		return
	}

	defineVecs := project.FindVarDefineInfo(strFile, &varStruct)
	if len(defineVecs) > 0 {
		locList = defineVecConvert(defineVecs)
		return locList, nil
	}

	return locList, nil
}

// handleAnnotateTypeDefine 处理注解系统带来的类型定义
func (l *LspServer) handleAnnotateTypeDefine(strFile string, contents []byte, offset int,
	posLine int, posCharacter int) (defineVecs []check.DefineStruct, flag bool) {
	strLine := getCompeleteLineStr(contents, offset)
	if strLine == "" {
		return
	}

	posCh := contents[offset]
	if offset > 0 && posCh != '_' && !IsDigit(posCh) && !IsLetter(posCh) {
		// 如果offset为非有效的字符，offset向前找一个字符
		posCharacter--
	}

	beginIndex := strings.LastIndex(strLine, "---@")
	if beginIndex == -1 {
		return
	}

	flag = true
	col := posCharacter - (beginIndex + 2)
	annotateStr := strLine[beginIndex+2:]
	project := l.getAllProject()
	defineVecs = project.AnnotateTypeDefine(strFile, annotateStr, posLine, col)
	return
}

// defineVecConvert 转换为返回的定义结构
func defineVecConvert(defineVecs []check.DefineStruct) (locList []lsp.Location) {
	for _, defineVarInfo := range defineVecs {
		locList = append(locList, lsp.Location{
			URI:   getFileDocumentURI(defineVarInfo.StrFile),
			Range: lspcommon.LocToRange(&defineVarInfo.Loc),
		})
	}

	return locList
}
