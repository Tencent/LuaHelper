package annotateast

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// 自带的type类型有下面的
// nil
// boolean
// number
// function
// userdata
// thread
// table
// any
// void

// type的所有用法
// 用法的完整格式
// ---@type MY_TYPE{|OTHER_TYPE]} [@comment]

// Type 注解类型的空接口
type Type interface{}

// NormalType 普通的类型
type NormalType struct {
	StrName   string         // 关联的类型的字符串名字（或是alias的名字)
	NameLoc   lexer.Location // 简单类型的位置信息
	ShowColor bool           // 着色的时候，显示位置
}

// MultiType 多种类型，选择其中一种都可以
type MultiType struct {
	Loc      lexer.Location // 整个func包含的位置信息
	TypeList []Type         // 多种类型的list
}

// ArrayType 关联的array 列表类型
type ArrayType struct {
	Loc      lexer.Location // 整个包含的位置信息
	ItemType Type           // 里面的类型，又是Type的一种
}

// TableType 关联的table map类型
type TableType struct {
	Loc         lexer.Location // 整个包含的位置信息
	TableStrLoc lexer.Location // table这个字符串的位置信息
	EmptyFlag   bool           // 是否为空的table，不包含key或value
	KeyType     Type           // 里面的key类型，又是Type的一种
	ValueType   Type           // 里面的Value类型，又是Type的一种
}

// FuncType 函数的类型
type FuncType struct {
	Loc              lexer.Location   // 整个func包含的位置信息
	FunLoc           lexer.Location   // fun关键字的位置信息，需要特殊着色
	ParamNameList    []string         // 函数参数名称的列表
	ParamNameLocList []lexer.Location // 整个参数列表的的位置信息
	ParamTypeList    []Type           // 函数参数类型的列表
	ParamOptionList  []bool           // 参数是否为可选的，如果为这样的 one? : string 表示参数one是可选的
	ReturnTypeList   []Type           // 函数返回值的列表
}

// ConstType 常量类型，例如 "aa" | "bb"
type ConstType struct {
	Loc        lexer.Location // 位置信息
	Name       string         // 名称
	QuotesFlag bool           // 是否为字符串双引号标记 String double quotes
	Comment    string         // 额外注释说明
}

// NotValidType 不是有效的
type NotValidType struct {
}
