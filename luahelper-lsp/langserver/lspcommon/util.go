package lspcommon

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/lexer"

	"luahelper-lsp/langserver/pathpre"
	lsp "luahelper-lsp/langserver/protocol"
)

// LocToRange luacheck里面的LocStruct转换为Range结构
func LocToRange(loc *lexer.Location) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      uint32(loc.StartLine) - 1,
			Character: uint32(loc.StartColumn),
		},
		End: lsp.Position{
			Line:      uint32(loc.EndLine) - 1,
			Character: (uint32)(loc.EndColumn),
		},
	}
}

func IsSameErrList(oldErrList []common.CheckError, newErrList []common.CheckError) bool {
	oldLen := len(oldErrList)
	newLen := len(newErrList)
	if oldLen != newLen {
		// 如果长度不相等，肯定不一致
		return false
	}

	for i := 0; i < oldLen; i++ {
		oldOne := oldErrList[i]
		newOne := newErrList[i]

		if oldOne.ToString() != newOne.ToString() {
			return false
		}
	}

	return true
}

// 获取文件名转换为的DocumentURI路径
func GetFileDocumentURI(strFile string) lsp.DocumentURI {
	return lsp.DocumentURI(pathpre.StringToVscodeURI(strFile))
}

// OffsetForPosition Previously used bytes converted to rune.
// Now use the high bit to determine how many bits the character occupies.
// posLine (zero-based, from 0)
func OffsetForPosition(contents []byte, posLine, posCh int) (int, error) {
	line := 0
	col := 0
	offset := 0

	getCharBytes := func(b byte) int {
		num := 0
		for b&(1<<uint32(7-num)) != 0 {
			num++
		}
		return num
	}

	for index := 0; index < len(contents); index++ {
		if line == posLine && col == posCh {
			return offset, nil
		}

		if (line == posLine && col > posCh) || line > posLine {
			return 0, fmt.Errorf("character %d (zero-based) is beyond line %d boundary (zero-based)", posCh, posLine)
		}

		curChar := contents[index]
		if curChar > 127 {
			curCharBytes := getCharBytes(curChar)
			index += curCharBytes - 1
			offset += curCharBytes - 1
		}
		offset++
		if curChar == '\n' {
			line++
			col = 0
		} else {
			col++
		}

	}
	if line == posLine && col == posCh {
		return offset, nil
	}
	if line == 0 {
		return 0, fmt.Errorf("character %d (zero-based) is beyond first line boundary", posCh)
	}
	return 0, fmt.Errorf("file only has %d lines", line+1)
}
