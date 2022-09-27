package langserver

import (
	"context"
	"fmt"

	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/codingconv"
	"luahelper-lsp/langserver/log"
	lsp "luahelper-lsp/langserver/protocol"
)

// TextDocumentSignatureHelp 补全函数的参数
func (l *LspServer) TextDocumentSignatureHelp(ctx context.Context, vs lsp.TextDocumentPositionParams) (signatureHelp lsp.SignatureHelp, err error) {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	comResult, activeParameter := l.doSignatureHelp(ctx, vs.TextDocument.URI, vs.Position)
	if !comResult.result {
		log.Debug("SignatureHelp return")
		return
	}

	pos := vs.Position
	varStruct := check.GetVarStruct(comResult.contents, comResult.offset, pos.Line, pos.Character)
	if !varStruct.ValidFlag {
		return
	}

	strFile := comResult.strFile
	project := l.getAllProject()
	flag, sinature, paramInfo := project.SignaturehelpFunc(strFile, &varStruct)
	sinature.Label = "function " + sinature.Label
	if !flag {
		log.Debug("SignatureHelp not func info.")
		return
	}

	str := codingconv.ConvertStrToUtf8(check.GetStrComment(sinature.Documentation))
	info := lsp.SignatureInformation{
		Label: codingconv.ConvertStrToUtf8(sinature.Label),
		Documentation: lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: str,
		},
	}

	for _, oneParamInfo := range paramInfo {
		str := codingconv.ConvertStrToUtf8(check.GetStrComment(oneParamInfo.Documentation))
		if oneParamInfo.AnnotateFlag {
			str = fmt.Sprintf("```%s\n%s\n```", "lua", codingconv.ConvertStrToUtf8(oneParamInfo.Documentation))
		}

		oneParam := lsp.ParameterInformation{
			Label: codingconv.ConvertStrToUtf8(oneParamInfo.Label),
			Documentation: lsp.MarkupContent{
				Kind:  lsp.Markdown,
				Value: str,
			},
		}
		info.Parameters = append(info.Parameters, oneParam)
	}

	signatureHelp.ActiveSignature = 0
	signatureHelp.ActiveParameter = (uint32)(activeParameter)
	signatureHelp.Signatures = []lsp.SignatureInformation{info}
	return
}

// doSignatureHelp 该文件为函数输入参数的时候，提示参数补全
func (l *LspServer) doSignatureHelp(ctx context.Context, url lsp.DocumentURI, pos lsp.Position) (comResult commFileRequest,
	activeParameter int) {
	// 判断打开的文件，是否是需要分析的文件
	comResult = l.beginFileRequest(url, pos)
	if !comResult.result {
		return
	}
	comResult.result = false
	if len(comResult.contents) == 0 || comResult.offset >= len(comResult.contents) {
		return
	}

	contents := comResult.contents
	offset := comResult.offset
	activeParameter = 0

	// If vscode auto-inserts closing ')' we will begin on ')' token in foo()
	// which will make the below algorithm think it's a nested call.
	if offset > 0 && offset < len(contents) && (contents[offset] == ')' || contents[offset] == ',') {
		offset--
	}

	if offset > 0 && offset == len(contents) {
		offset--
	}

	// Scan back out of call context.
	balance := 0
	for offset > 0 {
		c := contents[offset]
		if c == ')' {
			balance++
		} else if c == '(' {
			balance--
		}

		if balance == 0 && c == ',' {
			activeParameter++
		}

		offset--
		if balance == -1 {
			break
		}
	}

	if offset < 0 {
		return comResult, 0
	}

	comResult.offset = offset
	comResult.result = true
	return comResult, activeParameter
}

func (l *LspServer) getFuncParamCandidateType(ctx context.Context, url lsp.DocumentURI, pos lsp.Position) (annType annotateast.Type) {
	comResult, activeParameter := l.doSignatureHelp(ctx, url, pos)
	if !comResult.result {
		log.Debug("SignatureHelp return")
		return
	}

	varStruct := check.GetVarStruct(comResult.contents, comResult.offset, pos.Line, pos.Character)
	if !varStruct.ValidFlag {
		return
	}

	strFile := comResult.strFile
	project := l.getAllProject()
	flag, _, paramInfo := project.SignaturehelpFunc(strFile, &varStruct)

	if !flag {
		log.Debug("SignatureHelp not func info.")
		return
	}

	if activeParameter >= len(paramInfo) {
		return
	}

	annType = paramInfo[activeParameter].AnnType
	return
}
