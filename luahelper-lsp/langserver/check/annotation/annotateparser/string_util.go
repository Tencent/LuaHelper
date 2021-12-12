package annotateparser

import "strings"

// 字符串去掉单引号或双引号
func splitStrQuotes(str string) (splitStr string, quetesFlag bool) {
	splitStr = str
	if len(str) <= 2 {
		return
	}

	if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") {
		splitStr = string(str[1 : len(str)-1])
		quetesFlag = true
		return
	}

	if strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"") {
		splitStr = string(str[1 : len(str)-1])
		quetesFlag = true
		return
	}

	return
}
