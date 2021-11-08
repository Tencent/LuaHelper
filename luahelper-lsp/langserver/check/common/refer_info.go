package common

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// ReferType  引用文件的类型
type ReferType int

const (
	_ ReferType = iota
	// ReferNotValid 为无效的
	ReferNotValid = 0

	// ReferTypeFrame 框架自定义引入的方式
	ReferTypeFrame = 1

	// ReferTypeDofile dofile
	ReferTypeDofile = 2

	//ReferTypeLoadfile loadfile
	ReferTypeLoadfile = 3

	// ReferTypeRequire require
	ReferTypeRequire = 4
)

// ReferFrameType 当为框架引入一个文件时候，具体的引入方式
type ReferFrameType int

const (
	_ ReferFrameType = iota

	// RtypeImport 为自定义的import引入方式
	RtypeImport = 0

	// RtypeRequire 为类似require
	RtypeRequire = 1

	// RtypeAuto import和require都有可能。看引入文件的是否有返回，如果返回是一个table，那就是require，否则是import
	RtypeAuto = 2

	// RtypeNotValid 无效的
	RtypeNotValid = 3
)

// ReferInfo 单个lua文件中，显示引用了其他的lua文件或是库
// 例如 import("one.lua")
type ReferInfo struct {
	ReferTypeStr  string         // 引用类型的真实名称，例如用户定制的import或是import1
	ReferType     ReferType      // 引用的类型 import, dofile, loadfile, require
	ReferStr      string         // 引用指向的名称,例如 one.lua
	ReferValidStr string         // 引用指向的有效lua文件，例如require("two"), 实际上是引用的two.lua, 如果为空表示忽略引入的，可能是引入的.so
	Loc           lexer.Location // 具体的位置信息
	ReferVarLocal bool           // 引用如果赋值给了变量，true表示赋值的变量是否为local变量
	Valid         bool           // 第一遍check AST是否有效，如果无效，引用不用跟入进去分析，默认为true
}

// CreateOneReferInfo 创建一个引用信息
func CreateOneReferInfo(referTypeStr, referStr string, loc lexer.Location) *ReferInfo {
	referType := StrToReferType(referTypeStr)
	if referType == ReferNotValid {
		return nil
	}

	return &ReferInfo{
		ReferTypeStr:  referTypeStr,
		ReferType:     referType,
		ReferStr:      referStr,
		ReferValidStr: "",
		ReferVarLocal: false,
		Valid:         true,
		Loc:           loc,
	}
}

// StrToReferType 字符串转换为对应的ReferType
func StrToReferType(referTypeStr string) (referType ReferType) {
	if referTypeStr == "dofile" {
		referType = ReferTypeDofile
	} else if referTypeStr == "loadfile" {
		referType = ReferTypeLoadfile
	} else if referTypeStr == "require" {
		referType = ReferTypeRequire
	} else {
		if GConfig.IsFrameReferOtherFile(referTypeStr) {
			referType = ReferTypeFrame
		} else {
			referType = ReferNotValid
		}
	}

	return
}

// GetReferComment 获取引用信息的字符串，用于代码提示用
func (refer *ReferInfo) GetReferComment() string {
	return refer.ReferTypeStr + "(\"" + refer.ReferStr + "\")"
}
