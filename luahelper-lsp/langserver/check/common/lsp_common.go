package common

import (
	"fmt"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// ItemKind type
type ItemKind int

//
const (
	// IKVariable 变量
	IKVariable ItemKind = 1
	// IKFunction 函数
	IKFunction ItemKind = 2
	// IKKeyword 关键字
	IKKeyword ItemKind = 3
	// IKField 域
	IKField ItemKind = 4
	// IKAnnotateClass 注解的class类型
	IKAnnotateClass ItemKind = 5
	// IKAnnotateAlias 注解的alias类型
	IKAnnotateAlias ItemKind = 6
	// CIKSnippet 注释
	IKSnippet ItemKind = 15
)

// FileSymbolStruct 文件内的单个符号名
type FileSymbolStruct struct {
	Name          string             // 名称
	Kind          ItemKind           // 1表示变量，2表示函数
	Loc           lexer.Location     // 位置信息
	ContainerName string             // 包含者
	FileName      string             // 包含符号的lua文件名
	Children      []FileSymbolStruct // 当前block中包含的Symbol
}

// CompletionItemStruct 代码提示字段
type CompletionItemStruct struct {
	Label         string
	Kind          ItemKind // 1表示变量，2表示函数,  3为关键字， 4为模块CIKField
	InsetText     string
	Detail        string // 详细信息
	Documentation string
	CacheIndex    int // 对应的cacheIndex
	Line          int // 出现的行，用于代码获取注释使用
}

// SignatureHelpInfo SignatureHelp 时候提示的函数参数
type SignatureHelpInfo struct {
	Label         string
	Documentation string
}

// DefineVarStruct 查找变量定义的结构
type DefineVarStruct struct {
	PosLine      int      // 坐标的行, 从0开始
	PosCh        int      // 坐标的列, 从0开始
	ValidFlag    bool     // 是否为有效的
	Str          string   // 完整的切词
	ColonFlag    bool     // 最后一个变量是否为冒号的,例如为这样的ss.aa:bb 或是ss:bb
	StrVec       []string // 所有切分出来的数据
	IsFuncVec    []bool   // 切分出来的数据是否为函数, 与StrVec一一对应
	BracketsFlag bool     // 括号
	Exp          ast.Exp  // 表达式
}

// CompleteVarStruct 代码补全的定义结构
type CompleteVarStruct struct {
	PosLine             int      // 坐标的行, 从0开始
	PosCh               int      // 坐标的列, 从0开始
	StrVec              []string // 切分的是否为字符串数组
	IsFuncVec           []bool   // 切分出来的数据是否为函数, 与StrVec一一对应
	ColonFlag           bool     // 是否包含 :
	LastEmptyFlag       bool     // 最后一个字符是否为空白
	IgnoreKeyWord       bool     // 是否忽略关键字
	FilterCharacterFlag bool     // 查找的结果，是否过滤指定的字符
	FilterOneChar       rune     // 过滤的第一个字符（忽略大小写）
	FilterTwoChar       rune     // 过滤的第二个字符（忽略大小写）
	Exp                 ast.Exp  // 表达式
}

// RelateCheckInfo 相关联的告警信息
type RelateCheckInfo struct {
	LuaFile string         // 关联出错的lua文件
	ErrStr  string         // 关联出错的信息
	Loc     lexer.Location // 关联出错的位置信息
}

// CheckError 单个lua文件中，检测错误的信息
type CheckError struct {
	ErrType            CheckErrorType    // 检测错误的类型
	ErrStr             string            // 具体错误的信息
	Loc                lexer.Location    // 错误发生的行号
	EntryFile          string            // 入口文件
	RelateVec          []RelateCheckInfo // 关联告警的地方
	AnnotateSyntaxFlag bool              // 是否为注解语法分析错误
}

// ToString 错误信息，转成唯一的字符串
func (c CheckError) ToString() string {
	return fmt.Sprintf("%d:%d,%d,%d,%d:%s", c.ErrType, c.Loc.StartLine, c.Loc.StartColumn, c.Loc.EndLine,
		c.Loc.EndColumn, c.ErrStr)
}

// ColorType 获取颜色的类型
type ColorType int

const (
	// CTGlobalVar 全局变量类型
	CTGlobalVar ColorType = 0

	// CTGlobalFunc 全局函数类型
	CTGlobalFunc ColorType = 1

	// CTAnnotate 注解产生的颜色类型
	CTAnnotate ColorType = 2
)

// OneColorResut 一种颜色类型的返回的数据
type OneColorResut struct {
	LocVec []lexer.Location
}

// CheckReferenceSrc 查找引用的方式
type CheckReferenceSrc int

const (
	// CRSReference 普通的查找引用
	CRSReference CheckReferenceSrc = 1

	// CRSRename rename功能时候，调用的查找引用
	CRSRename CheckReferenceSrc = 2

	// CRSHighlight highligh功能时候，调用的查找引用
	CRSHighlight CheckReferenceSrc = 3
)
