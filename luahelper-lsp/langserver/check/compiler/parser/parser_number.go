package parser

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var reHexFloat = regexp.MustCompile(`^([0-9a-f]+(\.[0-9a-f]*)?|([0-9a-f]*\.[0-9a-f]+))(p[+\-]?[0-9]+)?$`)

// 判断是否为简单整型, ^[+-]?[0-9]+$
func isSimpleInteger(str string) bool {
	fristCh := str[0]
	if len(str) == 1 && (fristCh == '+' || fristCh == '-') {
		return false
	}

	for i, ch := range str {
		if i == 0 && (ch == '+' || ch == '-') {
			continue
		}

		if !(ch >= '0' && ch <= '9') {
			return false
		}
	}

	return true
}

func isLuajitSimpleInterger(str string) bool {
	fristCh := str[0]
	if len(str) == 1 && (fristCh == '+' || fristCh == '-') {
		return false
	}

	loc := 0
	for i, ch := range str {
		if i == 0 && (ch == '+' || ch == '-') {
			continue
		}

		if !(ch >= '0' && ch <= '9') {
			loc = i
			break
		}
	}

	if loc != 0 {
		suffix := str[loc:]
		if suffix != "ull" && suffix != "ll" {
			return false
		}
	}

	return true
}

// 判断是否为十六进制整型, ^-?0x[0-9a-f]+$
func isHexInteger(str string) bool {
	strLen := len(str)
	if strLen <= 2 {
		return false
	}

	i := 0
	fristCh := str[i]
	if fristCh == '-' {
		i++
	}

	// 判断是否为以0x开头
	if str[i] != '0' {
		return false
	}
	i++

	if str[i] != 'x' {
		return false
	}
	i++

	if strLen == i {
		return false
	}

	str = str[i:]
	for _, ch := range str {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}

	return true
}

// 判断是否为十六进制Luajit整型
func isLuajitHexInteger(str string) bool {
	strLen := len(str)
	if strLen <= 2 {
		return false
	}

	i := 0
	fristCh := str[i]
	if fristCh == '-' {
		i++
	}

	// 判断是否为以0x开头
	if str[i] != '0' {
		return false
	}
	i++

	if str[i] != 'x' {
		return false
	}
	i++

	if strLen == i {
		return false
	}

	loc := 0
	str = str[i:]
	for i, ch := range str {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			loc = i
			break
		}
	}

	if loc != 0 {
		suffix := str[loc:]
		if suffix != "ull" && suffix != "ll" {
			return false
		}
	}

	return true
}

// parseInteger 解析整型
func parseInteger(str string) (int64, bool) {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)

	// 判断是否为整型, ^[+-]?[0-9]+$
	if len(str) == 0 {
		return 0, false
	}

	if !isSimpleInteger(str) && !isHexInteger(str) {
		return 0, false
	}

	if str[0] == '+' {
		str = str[1:]
	}
	if !strings.Contains(str, "0x") { // decimal
		i, err := strconv.ParseInt(str, 10, 64)
		return i, err == nil
	}

	// hex
	var sign int64 = 1
	if str[0] == '-' {
		sign = -1
		str = str[3:]
	} else {
		str = str[2:]
	}

	if len(str) > 16 {
		str = str[len(str)-16:] // cut long hex string
	}

	i, err := strconv.ParseUint(str, 16, 64)
	return sign * int64(i), err == nil
}

// parseFloat 解析浮点数
func parseFloat(str string) (float64, bool) {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)
	if strings.Contains(str, "nan") || strings.Contains(str, "inf") {
		return 0, false
	}
	if strings.HasPrefix(str, "0x") && len(str) > 2 {
		return parseHexFloat(str[2:])
	}
	if strings.HasPrefix(str, "+0x") && len(str) > 3 {
		return parseHexFloat(str[3:])
	}
	if strings.HasPrefix(str, "-0x") && len(str) > 3 {
		f, ok := parseHexFloat(str[3:])
		return -f, ok
	}
	f, err := strconv.ParseFloat(str, 64)
	return f, err == nil
}

// parseLuajitNum 解析Luajit数字
func parseLuajitNum(str string) (int64, bool) {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)

	// 判断是否为整型, ^[+-]?[0-9]+$
	if len(str) == 0 {
		return 0, false
	}

	if !isLuajitSimpleInterger(str) && !isLuajitHexInteger(str) {
		return 0, false
	}

	if str[0] == '+' {
		str = str[1:]
	}

	if str[len(str)-3] == 'u' {
		str = str[0 : len(str)-3]
	} else {
		str = str[0 : len(str)-2]
	}

	if !strings.Contains(str, "0x") { // decimal
		i, err := strconv.ParseUint(str, 10, 64)
		return int64(i), err == nil
	}

	// hex
	var sign int64 = 1
	if str[0] == '-' {
		sign = -1
		str = str[3:]
	} else {
		str = str[2:]
	}

	if len(str) > 16 {
		str = str[len(str)-16:] // cut long hex string
	}

	i, err := strconv.ParseUint(str, 16, 64)
	return sign * int64(i), err == nil
}

// (0x)ABC.DEFp10
func parseHexFloat(str string) (float64, bool) {
	var i16, f16, p10 float64 = 0, 0, 0

	if !reHexFloat.MatchString(str) {
		return 0, false
	}

	// decimal exponent
	if idxOfP := strings.Index(str, "p"); idxOfP > 0 {
		digits := str[idxOfP+1:]
		str = str[:idxOfP]

		var sign float64 = 1
		if digits[0] == '-' {
			sign = -1
		}
		if digits[0] == '-' || digits[0] == '+' {
			digits = digits[1:]
		}

		if len(str) == 0 || len(digits) == 0 {
			return 0, false
		}

		for i := 0; i < len(digits); i++ {
			if x, ok := parseDigit(digits[i], 10); ok {
				p10 = p10*10 + x
			} else {
				return 0, false
			}
		}

		p10 = sign * p10
	}

	// fractional part
	if idxOfDot := strings.Index(str, "."); idxOfDot >= 0 {
		digits := str[idxOfDot+1:]
		str = str[:idxOfDot]
		if len(str) == 0 && len(digits) == 0 {
			return 0, false
		}
		for i := len(digits) - 1; i >= 0; i-- {
			if x, ok := parseDigit(digits[i], 16); ok {
				f16 = (f16 + x) / 16
			} else {
				return 0, false
			}
		}
	}

	// integral part
	for i := 0; i < len(str); i++ {
		if x, ok := parseDigit(str[i], 16); ok {
			i16 = i16*16 + x
		} else {
			return 0, false
		}
	}

	// (i16 + f16) * 2^p10
	f := i16 + f16
	if p10 != 0 {
		f *= math.Pow(2, p10)
	}
	return f, true
}

func parseDigit(digit byte, base int) (float64, bool) {
	if base == 10 || base == 16 {
		switch digit {
		case '0':
			return 0, true
		case '1':
			return 1, true
		case '2':
			return 2, true
		case '3':
			return 3, true
		case '4':
			return 4, true
		case '5':
			return 5, true
		case '6':
			return 6, true
		case '7':
			return 7, true
		case '8':
			return 8, true
		case '9':
			return 9, true
		}
	}
	if base == 16 {
		switch digit {
		case 'a':
			return 10, true
		case 'b':
			return 11, true
		case 'c':
			return 12, true
		case 'd':
			return 13, true
		case 'e':
			return 14, true
		case 'f':
			return 15, true
		}
	}
	return -1, false
}
