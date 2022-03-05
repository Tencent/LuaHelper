package codingconv

import (
	"luahelper-lsp/langserver/strbytesconv"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// ConvertStrToUtf8 GBK change to utf8
func ConvertStrToUtf8(str string) string {
	if str == "" {
		return str
	}

	if isUtf8(strbytesconv.StringToBytes(str)) {
		return str
	}

	ret, ok := simplifiedchinese.GBK.NewDecoder().String(str)
	if ok != nil {
		return str
	}

	return ret
}

func preNUm(data byte) int {
	var i int = 0
	for data > 0 {
		if data%2 == 0 {
			i = 0
		} else {
			i++
		}
		data /= 2
	}

	return i
}

func isUtf8(data []byte) bool {
	for i := 0; i < len(data); {
		if data[i]&0x80 == 0x00 {
			// 0XXX_XXXX
			i++
			continue
		} else if num := preNUm(data[i]); num > 2 {
			// 110X_XXXX 10XX_XXXX
			// 1110_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_0XXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_10XX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_110X 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// preNUm() 返回首个字节的8个bits中首个0bit前面1bit的个数，该数量也是该字符所使用的字节数
			i++
			for j := 0; j < num-1; j++ {
				if i >= len(data) {
					return false
				}

				//判断后面的 num - 1 个字节是不是都是10开头
				if data[i]&0xc0 != 0x80 {
					return false
				}
				i++
			}
		} else {
			//其他情况说明不是utf-8
			return false
		}
	}
	return true
}
