package langserver

import (
	"context"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/lspcommon"
	protocol "luahelper-lsp/langserver/protocol"
	"luahelper-lsp/langserver/stringutil"
)

// TextDocumentReferences 文件中查找符合的所有的引用
func (l *LspServer) TextDocumentReferences(ctx context.Context, vs protocol.ReferenceParams) (locList []protocol.Location, err error) {
	comResult := l.beginFileRequest(vs.TextDocument.URI, vs.Position)
	if !comResult.result {
		return
	}

	if len(comResult.contents) == 0 || comResult.offset >= len(comResult.contents) {
		return
	}

	project := l.getAllProject()
	varStruct := stringutil.GetVarStruct(comResult.contents, comResult.offset, comResult.pos.Line, comResult.pos.Character)
	if !varStruct.ValidFlag {
		log.Error("TextDocumentReferences not valid")
		return
	}

	// subCtx, cancel := context.WithTimeout(ctx, 1 * time.Second)
	// defer cancel()

	// time.Sleep(time.Second * 2)

	// select {
	// case <-subCtx.Done():
	// 	fmt.Println("main", ctx.Err())
	// default:
	// 	log.Error("not hit")
	// }

	// 去掉前缀后的名字
	referenVecs := project.FindReferences(comResult.strFile, &varStruct, common.CRSReference)
	locList = make([]protocol.Location, 0, len(referenVecs))
	referenceNum := common.GConfig.ReferenceMaxNum
	for i, referVarInfo := range referenVecs {
		if i >= referenceNum {
			break
		}

		locList = append(locList, protocol.Location{
			URI:   stringutil.GetFileDocumentURI(referVarInfo.StrFile),
			Range: lspcommon.LocToRange(&referVarInfo.Loc),
		})
	}

	return locList, nil
}
