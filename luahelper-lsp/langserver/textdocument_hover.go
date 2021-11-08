package langserver

import (
	"context"
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/codingconv"
	"luahelper-lsp/langserver/log"
	lsp "luahelper-lsp/langserver/protocol"
	"strings"
)

type MarkupHover struct {
	/**
	 * The hover's content
	 */
	Contents lsp.MarkupContent/*MarkupContent | MarkedString | MarkedString[]*/ `json:"contents"`
	//Contents MarkedString/*MarkupContent | MarkedString | MarkedString[]*/ `json:"contents"`
	/**
	 * An optional range
	 */
	Range lsp.Range `json:"range,omitempty"`
}

// TextDocumentHover 文件中查找变量的的定义
func (l *LspServer) TextDocumentHover(ctx context.Context, vs lsp.TextDocumentPositionParams) (hoverReturn interface{}, err error) {
	comResult := l.beginFileRequest(vs.TextDocument.URI, vs.Position)
	if !comResult.result {
		return nil, nil
	}

	if len(comResult.contents) == 0 {
		return nil, nil
	}

	strLable, strDoc, strLuaFile := l.getHoverStr(comResult)
	if strLable == "" && strDoc == "" && strLuaFile == "" {
		return nil, nil
	}

	var contents lsp.MarkupContent
	if strLable == "" {
		contents.Kind = lsp.PlainText
		contents.Value = strDoc
		if strLuaFile != "" {
			contents.Value = contents.Value + "\n" + strLuaFile
		}
	} else {
		contents.Kind = lsp.Markdown
		contents.Value = fmt.Sprintf("```%s\n%s\n```", "lua", strLable)

		if strDoc != "" || strLuaFile != "" {
			contents.Value = fmt.Sprintf("%s\n%s\n", contents.Value, "---")
		}

		strComment := strDoc

		// 紧凑的换行方式为：两个空格\n
		// 松散的换行方式为：\n\r
		if strLuaFile != "" {
			strComment = strComment + "\n\r" + strLuaFile
		}

		contents.Value = contents.Value + strComment
	}

	var hover = MarkupHover{
		Contents: contents,
		Range: lsp.Range{
			Start: vs.Position,
			End:   vs.Position,
		},
	}

	return hover, nil
}

// 获取鼠标悬停的字符串
func (l *LspServer) getHoverStr(comResult commFileRequest) (lableStr, docStr, luaFileStr string) {
	// 1) 判断是否普通悬停打开一个文件
	dirManager := common.GConfig.GetDirManager()
	if showOpenFile := l.hoverOpenFile(comResult); showOpenFile != "" {
		docStr = "lua file : " + dirManager.RemovePathDirPre(showOpenFile)
		return
	}

	// 2) 判断是否为注解类型的悬停提示
	if strLable, annotateStr, fileStr, flag := l.handleAnnotateTypeHover(comResult); flag {
		lableStr = strLable
		docStr = codingconv.ConvertStrToUtf8(annotateStr)
		luaFileStr = fileStr
		return
	}

	// 3) 普通查找定义悬浮
	varStruct := getVarStruct(comResult.contents, comResult.offset, comResult.pos.Line, comResult.pos.Character)
	if !varStruct.ValidFlag {
		log.Error("TextDocumentDefine not valid")
		return
	}
	project := l.getAllProject()
	lableStr, docStr, luaFileStr = project.GetLspHoverVarStr(comResult.strFile, &varStruct)
	docStr = codingconv.ConvertStrToUtf8(docStr)
	return
}

// 判断是否悬停提示打开一个文件
func (l *LspServer) hoverOpenFile(comResult commFileRequest) (fileName string) {
	strFirstFile, strSecondFile := getOpenFileStr(comResult.contents, comResult.offset, (int)(comResult.pos.Character))
	if strFirstFile == "" {
		return
	}

	log.Debug("strOpenFile=%s", strFirstFile)
	fileName = strFirstFile

	strFile := comResult.strFile
	project := l.getAllProject()
	// 如果require("one") 找到了one/init.lua hover显示的文件名为: one/init.lua
	defineVecs := project.FindOpenFileDefine(strFile, strFirstFile)
	if len(defineVecs) == 0 && strSecondFile != "" {
		defineVecs = project.FindOpenFileDefine(strFile, strSecondFile)
		if len(defineVecs) > 0 {
			fileName = strSecondFile
		}
	}

	return
}

// 判断是否为注解带来的悬停提示打开一个文件
// 处理注解系统带来的类型定义
func (l *LspServer) handleAnnotateTypeHover(comResult commFileRequest) (strLabl, strHover, strFile string, flag bool) {
	strLine := getCompeleteLineStr(comResult.contents, comResult.offset)
	if strLine == "" {
		return
	}

	beginIndex := strings.LastIndex(strLine, "---@")
	if beginIndex == -1 {
		return
	}

	// 首先切词
	strWord := spliteAnnotateStr(comResult.contents, comResult.offset)

	// 1) 判断是否为类似的 ---@param 注解
	strArea := "---@" + strWord
	if strings.Index(strLine, strArea) >= 0 {
		return "", getAnnotateAreaHoverStr(strWord), "", true
	}

	// 2)
	flag = true
	col := (int)(comResult.pos.Character) - (beginIndex + 2)
	annotateStr := strLine[beginIndex+2:]
	project := l.getAllProject()
	strLabl, strHover, strFile = project.AnnotateTypeHover(comResult.strFile, annotateStr, strWord, (int)(comResult.pos.Line), col)
	return
}

// 获取注解域的hoverstr
func getAnnotateAreaHoverStr(strWord string) (strHover string) {
	if strWord == "class" {
		return "---@class TYPE[: PARENT_TYPE {, PARENT_TYPE}]  [@comment]" +
			"\n\n" + "sample:\n---@class People @People class"
	} else if strWord == "field" {
		return "---@field [public|protected|private] field_name FIELD_TYPE{|OTHER_TYPE} [@comment]" +
			"\n\n" + "sample:\n---@field age number @age attr is number"
	} else if strWord == "type" {
		return "---@type TYPE{|OTHER_TYPE} [@comment]" +
			"\n\n" + "sample:\n---@field age number @age attr is number"
	} else if strWord == "param" {
		return "---@param param_name TYPE{|OTHER_TYPE} [@comment]" +
			"\n\n" + "sample:\n---@param param1 string @param1 is string"
	} else if strWord == "return" {
		return "---@return TYPE{|OTHER_TYPE} [@comment]" +
			"\n\n" + "sample:\n---@return string @return is string"
	} else if strWord == "alias" {
		return "---@alias new_type TYPE{|OTHER_TYPE}" +
			"\n\n" + "sample:\n---@alias Man People @Man is People"
	} else if strWord == "generic" {
		return "---@generic T [: PARENT_TYPE]" +
			"\n\n" + "sample:\n---@generic T1 @T1 is generic"
	} else if strWord == "overload" {
		return "---@overload fun(param_name : PARAM_TYPE) : RETURN_TYPE" +
			"\n\n" + "sample:\n---@overload fun(param1 : string) : number"
	}
	return
}

func spliteAnnotateStr(contents []byte, offset int) (str string) {
	conLen := len(contents)
	if offset == conLen {
		offset = offset - 1
	}

	// 判断查找的定义是否为
	// 向前找
	posCh := contents[offset]
	if offset > 0 && posCh != '_' && !IsDigit(posCh) && !IsLetter(posCh) {
		// 如果offset为非有效的字符，offset向前找一个字符
		offset--
	}

	beforeIndex := offset
	for index := offset; index >= 0; index-- {
		ch := contents[index]
		if ch == '_' || IsDigit(ch) || IsLetter(ch) || ch == '.' {
			beforeIndex = index
			continue
		}
		break
	}

	endIndex := offset
	for index := offset; index < len(contents); index++ {
		ch := contents[index]
		if ch == '_' || IsDigit(ch) || IsLetter(ch) || ch == '.' {
			endIndex = index
			continue
		}
		break
	}

	rangeConents := contents[beforeIndex : endIndex+1]
	str = string(rangeConents)
	return str
}
