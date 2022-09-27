package stringutil

import (
	"regexp"
	"strings"
)

// IsDigit 是否是数字
func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// IsLetter 是否为英文字符
func IsLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// GetBeforeIndex 切分
func GetBeforeIndex(contents []byte, offset int) (beforeIndex int) {
	beforeIndex = offset
	roundBracketsBalance := 0
	squareBracketsBalance := 0
	for index := offset; index >= 0; index-- {
		ch := contents[index]
		if ch == '\r' || ch == '\n' {
			break
		}

		if ch == '_' || ch == '.' ||
			ch == ':' || IsDigit(ch) ||
			IsLetter(ch) || ch == ')' ||
			ch == '(' || ch == '[' ||
			ch == ']' {

			if ch == ')' {
				roundBracketsBalance++
			} else if ch == ']' {
				squareBracketsBalance++
			} else if ch == '(' {
				roundBracketsBalance--
				if roundBracketsBalance < 0 {
					break
				}
			} else if ch == '[' {
				squareBracketsBalance--
				if squareBracketsBalance < 0 {
					break
				}
			}

			beforeIndex = index
			continue
		}

		if roundBracketsBalance > 0 || squareBracketsBalance > 0 {
			continue
		}

		break
	}

	if beforeIndex < 0 {
		beforeIndex = 0
	}

	return beforeIndex
}

// SplitVec 将vec里的字符串再切分 传入切分符号 返回一个新的vec
func SplitVec(srcVec []string, sep string) (retVec []string) {
	for _, oneStr := range srcVec {
		tmpVec := strings.Split(oneStr, sep)
		retVec = append(retVec, tmpVec...)
	}

	return retVec
}

// 获取当前输入点这行的起始offset
func GetLineOffset(offset int, contents []byte) int {
	lineIndex := offset - 1
	for ; lineIndex >= 0; lineIndex-- {
		ch := contents[lineIndex]
		if ch == '\r' || ch == '\n' {
			break
		}
	}

	if lineIndex+1 >= offset {
		return -1
	}

	return lineIndex + 1
}

// 获取截止到当前输入点的这行内容
func GetPreLineStr(offset int, contents []byte) (strLine string) {
	lineIndex := offset - 1
	for ; lineIndex >= 0; lineIndex-- {
		ch := contents[lineIndex]
		if ch == '\r' || ch == '\n' {
			break
		}
	}

	if lineIndex+1 >= offset {
		return ""
	}

	strLine = string(contents[lineIndex+1 : offset])
	return strLine
}

// 获取一行完整的内容
func GetCompeleteLineStr(contents []byte, offset int) (lineStr string) {
	conLen := len(contents)
	if offset == conLen {
		offset = offset - 1
	}
	beforeLinePos := offset
	for index := offset - 1; index >= 0; index-- {
		ch := contents[index]
		if ch == '\r' || ch == '\n' {
			break
		}
		beforeLinePos = index
	}
	endLinePos := offset
	for index := offset; index < conLen; index++ {
		ch := contents[index]
		if ch == '\r' || ch == '\n' {
			break
		}
		endLinePos = index
	}
	lineStr = string(contents[beforeLinePos : endLinePos+1])
	return lineStr
}

// stringutil.GetOpenFileStr 判断是否为打开的文件，返回文件名
// 当为require时候，可能返回两个文件名， 因为require("one") 含义可能是包含one.lua 也有可能是one/init.lua
// firstStr 为优先级高的文件名，secondStr为优先级次之的文件名
func GetOpenFileStr(contents []byte, offset int, character int, referFiles []string) []string {
	// 获取当前行的所有内容
	lineContents := GetCompeleteLineStr(contents, offset)

	// 1) 引入其他文件的正则
	regDofile := regexp.MustCompile(`dofile *?\( *?\"[0-9a-zA-Z_/\-]+.lua\" *?\)`)
	regRequire := regexp.MustCompile(`require *?(\()? *?[\"|\'][0-9a-zA-Z_/\-|.]+[\"|\'] *?(\))?`)

	// ""内的内容
	regFen := regexp.MustCompile(`[\"|\'][0-9a-zA-Z_/\.\-]+[\"|\']`)

	// 是否需要.lua后缀
	needLuaSuffix := false
	requireFlag := false

	// 匹配的表达式
	importVec := regDofile.FindAllString(lineContents, -1)
	if len(importVec) == 0 {
		importVec = regRequire.FindAllString(lineContents, -1)
		if len(importVec) > 0 {
			requireFlag = true
		}
	} else {
		needLuaSuffix = true
	}

	if len(importVec) == 0 {
		for _, strOne := range referFiles {
			regImport1 := regexp.MustCompile(strOne + ` *?(\()? *?[\"|\'][0-9a-zA-Z_/|.\-]+.lua+[\"|\'] *?(\))?`)
			importVec = regImport1.FindAllString(lineContents, -1)
			if len(importVec) > 0 {
				needLuaSuffix = true
				break
			}

			regImport2 := regexp.MustCompile(strOne + ` *?(\()? *?[\"|\'][0-9a-zA-Z_/|.\-]+[\"|\'] *?(\))?`)
			importVec = regImport2.FindAllString(lineContents, -1)
			if len(importVec) > 0 {
				needLuaSuffix = false
				break
			}
		}
	}

	strOpenFile := ""
	for _, importStr := range importVec {
		findIndex := strings.Index(lineContents, importStr)
		if findIndex == -1 {
			continue
		}

		regStrFen := regFen.FindAllString(importStr, -1)
		if len(regStrFen) == 0 {
			continue
		}

		strFileTemp := regStrFen[0]
		if len(strFileTemp) < 2 {
			continue
		}

		strFileTemp = strFileTemp[1 : len(strFileTemp)-1]
		findTempIndex := strings.Index(importStr, strFileTemp)
		if findTempIndex == -1 {
			continue
		}

		importBeginIndex := findIndex + findTempIndex
		importEndIndex := findIndex + findTempIndex + len(strFileTemp)

		if character >= importBeginIndex && character <= importEndIndex {
			if needLuaSuffix && strings.HasSuffix(strFileTemp, ".lua") {
				strFileTemp = strFileTemp[0 : len(strFileTemp)-4]
			}

			strFileTemp = strings.Replace(strFileTemp, ".", "/", -1)
			strOpenFile = strFileTemp
			break
		}
	}

	if strOpenFile == "" {
		return make([]string, 0)
	}

	strModName := strings.TrimSuffix(strOpenFile, ".lua")
	result := []string{strModName + ".lua", strModName + ".so"}
	if requireFlag {
		result = append(result, strModName+"/init.lua")
	}
	return result
}

// 如果输入的为 b["c"]，当对c查找定义时候，offset移动到]后面，再统一处理
func matchSpecialBracketsStr(contents []byte, offset int) (flag bool, beforeIndex int, endIndex int) {
	rightI := -1
	rightFlag := 0
	for index := offset; index < len(contents); index++ {
		ch := contents[index]
		if ch == '\r' || ch == '\n' {
			break
		}

		if rightFlag == 0 && (ch == '"' || ch == '\'') {
			rightFlag = 1
		}

		if rightFlag == 1 && ch == ']' {
			rightI = index
			break
		}
	}

	if rightI == -1 {
		return
	}

	leftI := -1
	leftFlag := 0
	for index := offset; index >= 0; index-- {
		ch := contents[index]
		if ch == '\r' || ch == '\n' {
			break
		}

		if leftFlag == 0 && (ch == '"' || ch == '\'') {
			leftFlag = 1
		}

		if leftFlag == 1 && ch == '[' {
			leftI = index
			break
		}
	}

	if leftI == -1 {
		return
	}

	flag = true
	beforeIndex = leftI
	endIndex = rightI
	return
}

// 判断是否有括号
func GetContentBracketsFlag(contents []byte, offset int) (beforeIndex int, endIndex int, bracketsFlag bool) {
	// 判断光标是否在特殊的[""]里面
	flag, _, rightI := matchSpecialBracketsStr(contents, offset)
	if flag {
		offset = rightI
	}

	beforeIndex = GetBeforeIndex(contents, offset)
	leftBracketsFlag := strings.Contains(string(contents[beforeIndex:offset]), "[")

	endIndex = offset
	rightBracketsFlag := false

	for index := offset; index < len(contents); index++ {
		ch := contents[index]
		if ch == '\r' || ch == '\n' {
			break
		}

		if ch == '_' || IsDigit(ch) || IsLetter(ch) {
			endIndex = index
			continue
		}

		if ch == ']' {
			rightBracketsFlag = true
		}

		break
	}
	bracketsFlag = false
	if leftBracketsFlag && rightBracketsFlag {
		bracketsFlag = true
	}

	return beforeIndex, endIndex, bracketsFlag
}
