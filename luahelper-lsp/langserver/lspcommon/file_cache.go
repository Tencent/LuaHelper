package lspcommon

import (
	"bytes"
	"fmt"

	"luahelper-lsp/langserver/log"
	lsp "luahelper-lsp/langserver/protocol"
)

// FileCache 单个文件的cache管理
type FileCache struct {
	strName string // 文件名，文件全路径；例如g:\luaproject\src\tutorial.lua
	content []byte // 文件具体的内容
}

// FileMapCache cache map
type FileMapCache struct {
	m map[string]FileCache // 管理多个文件对象，key值为文件名，全路径
}

// CreateFileMapCache 创建文件cache管理对象
func CreateFileMapCache() *FileMapCache {
	return &FileMapCache{
		m: make(map[string]FileCache),
	}
}

// SetFileContent 设置某一个文件的内容
func (fileMapCache *FileMapCache) SetFileContent(strFile string, contents []byte) {
	fileMapCache.m[strFile] = FileCache{
		strName: strFile,
		content: contents,
	}
}

// DelFileContent 删除某一个文件的内存
func (fileMapCache *FileMapCache) DelFileContent(strFile string) {
	if _, ok := fileMapCache.m[strFile]; !ok {
		log.Error("DelFileContent err, not find strFile=%s", strFile)
	}
	delete(fileMapCache.m, strFile)
}

// GetFileContent 获取文件的内容
func (fileMapCache *FileMapCache) GetFileContent(strFile string) (contents []byte, found bool) {
	fileCache, ok := fileMapCache.m[strFile]
	if ok {
		contents = fileCache.content
		found = ok
	} else {
		found = false
		log.Error("GetFileContent error, strFile=%s", strFile)
	}

	return contents, found
}

//	Previously used bytes converted to rune.
//	Now use the high bit to determine how many bits the character occupies.
func offsetForStartAndEnd(contents []byte, startPos lsp.Position, endPos lsp.Position) (startOffset,
	endOffset int, err error) {
	line := uint32(0)
	col := uint32(0)
	offset := 0
	startFlag := false
	endFlag := false

	getCharBytes := func(b byte) int {
		num := 0
		for b&(1<<uint32(7-num)) != 0 {
			num++
		}
		return num
	}

	for index := 0; index < len(contents); index++ {
		if !startFlag {
			if line == startPos.Line && col == startPos.Character {
				startOffset = offset
				startFlag = true
			}
		}

		if !startFlag {
			if (line == startPos.Line && col > startPos.Character) || line > startPos.Line {
				return 0, 0, fmt.Errorf("character %d (zero-based) is beyond line %d boundary (zero-based)", startPos.Character, startPos.Line)
			}
		}

		if startFlag && !endFlag {
			if line == endPos.Line && col == endPos.Character {
				endOffset = offset
				endFlag = true

				// 两个都找到了
				return startOffset, endOffset, nil
			}

			if (line == endPos.Line && col > endPos.Character) || line > endPos.Line {
				return 0, 0, fmt.Errorf("character %d (zero-based) is beyond line %d boundary (zero-based)", endPos.Character, endPos.Line)
			}
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
	if startFlag && line == endPos.Line && col == endPos.Character {
		return startOffset, offset, nil
	}
	if !startFlag && line == startPos.Line && col == startPos.Character &&
		line == endPos.Line && col == endPos.Character {
		return offset, offset, nil
	}
	if line == 0 {
		return 0, 0, fmt.Errorf("character %d (zero-based) is beyond first line boundary", startPos.Character)
	}
	return 0, 0, fmt.Errorf("file only has %d lines", line+1)
}

// ApplyContentChanges updates `contents` based on `changes`
func (fileMapCache *FileMapCache) ApplyContentChanges(strFile string, contents []byte, changes []lsp.TextDocumentContentChangeEvent) ([]byte, error) {
	for _, change := range changes {
		if change.Range == nil && change.RangeLength == 0 {
			contents = []byte(change.Text) // new full content
			continue
		}

		start, end, err := offsetForStartAndEnd(contents, change.Range.Start, change.Range.End)
		if err != nil {
			return nil, fmt.Errorf("invalid position %q on %q: %s", change.Range.Start, strFile, err.Error())
		}

		if start < 0 || end > len(contents) || end < start {
			return nil, fmt.Errorf(" for out of range position %q on %q", change.Range, strFile)
		}
		// Try avoid doing too many allocations, so use bytes.Buffer
		b := &bytes.Buffer{}
		b.Grow(start + len(change.Text) + len(contents) - end)
		b.Write(contents[:start])
		b.WriteString(change.Text)
		b.Write(contents[end:])
		contents = b.Bytes()
	}

	return contents, nil
}
