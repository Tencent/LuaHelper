package pathpre

import (
	"net/url"
	"strings"
)

// preFixStr 工程项目的前缀，windows平台的前缀是 file:///， mac平台的前缀是file://
var preFixStr string = "file:///"

// InitialRootURIAndPath 初始化前缀
func InitialRootURIAndPath(rootURI, rootPath string) {
	rootURI, _ = url.QueryUnescape(rootURI)
	rootURI = strings.Replace(rootURI, "\\", "/", -1)
	
	rootPath, _ = url.QueryUnescape(rootPath)
	rootPath = strings.Replace(rootPath, "\\", "/", -1)

	if len(rootURI) < 8 {
		return
	}

	subRootURI := rootURI[7:]
	if subRootURI == rootPath {
		preFixStr = "file://"
	}
}

// GetRemovePreStr 如果字符串以 ./为前缀，去除掉前缀
func GetRemovePreStr(str string) string {
	if strings.HasPrefix(str, "./") {
		str = str[2:]
	}

	return str
}

// VscodeURIToString 插件传入的路径转换
// vscode 传来的路径： file:///g%3A/luaproject
// 统一转换为：g%3A/luaproject，去掉前缀的file:///，并且都是这样的/../
func VscodeURIToString(strURL string) string {
	fileURL := strings.Replace(strURL, preFixStr, "", 1)
	fileURL, _ = url.QueryUnescape(fileURL)
	fileURL = strings.Replace(string(fileURL), "\\", "/", -1)

	return fileURL
}

// StringToVscodeURI 文件真实路径转换成类似的 file:///g%3A/luaproject/src/tutorial.lua"
func StringToVscodeURI(strPath string) string {
	strPath = strings.Replace(string(strPath), "\\", "/", -1)
	strEncode := strPath
	strURI := preFixStr + strEncode
	return strURI
}

// GeConvertPathFormat 文件路径统一为
func GeConvertPathFormat(strPath string) string {
	strDir := strings.Replace(string(strPath), "\\", "/", -1)
	return strDir
}
