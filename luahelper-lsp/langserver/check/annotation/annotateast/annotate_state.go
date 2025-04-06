package annotateast

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// AnnotateState 空接口， 一行注释表示的一个语句
type AnnotateState interface{}

// AnnotateAliasState alias
// ---@alias Handler fun(type: string, data: any):void
type AnnotateAliasState struct {
	Name       string         // alias名称，上面的例子为Handler
	NameLoc    lexer.Location // alias的名称位置
	AliasType  Type           // 具体对应的Type类型
	Comment    string         // 剩余的其他注释内容
	CommentLoc lexer.Location // 注释内容的位置信息
}

// AnnotateOverloadState 函数重载的类型
type AnnotateOverloadState struct {
	OverFunType *FuncType      // 重载具体的函数类型
	Comment     string         // 剩余的其他注释内容
	CommentLoc  lexer.Location // 注释内容的位置信息
}

// AnnotateTypeState 定义的类型
// 一行可能定义多个，例如下面的例子
// ---@type [const] number, [const] stirng
// local a, b
type AnnotateTypeState struct {
	ListType   []Type         // 多个类型，放在list里面
	Comment    string         // 其他所有的注释内容
	CommentLoc lexer.Location // 注释内容的位置信息
	ListConst  []bool         // 是否const
	ListEnum   []bool         // 是否为枚举的
}

// AnnotateClassState 定义的class
type AnnotateClassState struct {
	Name           string           // class的名称
	NameLoc        lexer.Location   // class的名称位置
	ParentNameList []string         // 可能存在多个父的对象的名称
	ParentLocList  []lexer.Location // 可能存在的多个父的对象的位置信息
	Comment        string           // 其他所有的注释内容
	CommentLoc     lexer.Location   // 注释内容的位置信息
}

// AnnotateFieldState 定义的成员结构
// ---@field [public|protected|private] field_name FIELD_TYPE[|OTHER_TYPE] [@comment]
type AnnotateFieldState struct {
	Name           string         // 成员结构的名称
	NameLoc        lexer.Location // field的名称位置
	FieldScopeType FieldScopeType // 属性的类型 public、protected、private
	FieldColonType FieldColonType // 属性是否为：
	FiledType      Type           // 成员对应属性
	IsOptional     bool           // 这个参数是否为可选的。例如 ---@field one? number ; 参数后面跟？表示参数是可选的
	Comment        string         // 其他所有的注释内容
	CommentLoc     lexer.Location // 注释的位置信息
}

// AnnotateParamState 函数参数
// ---@param param_name MY_TYPE[|other_type] [@comment]
type AnnotateParamState struct {
	Name       string         // 参数的名称
	NameLoc    lexer.Location // 参数的名称位置
	ParamType  Type           // 成员对应属性
	Comment    string         // 其他所有的注释内容
	IsOptional bool           // 这个参数是否为可选的。例如 ---@param one? number ; 参数后面跟？表示参数是可选的
	IsConst    bool           // 这个参数是否const。例如 ---@param cosnt one number
	CommentLoc lexer.Location // 注释内容的位置信息
}

// AnnotateReturnState 函数返回结构
type AnnotateReturnState struct {
	ReturnTypeList   []Type         // 成员对应属性
	ReturnOptionList []bool         // 判断返回是否为可选的，例如---@return integer?  code 表示是可选的
	Comment          string         // 其他所有的注释内容
	CommentLoc       lexer.Location // 注释内容的位置信息
}

// AnnotateGenericState 泛型的结构
// ---@generic T1 [: PARENT_TYPE] [, T2 [: PARENT_TYPE]] @comment @comment
// todo 泛型还没有处理
type AnnotateGenericState struct {
	NameList       []string         // 可能一行定义多个
	NameLocList    []lexer.Location // 所有的名称位置列表
	ParentNameList []string         // 可能包含父的结构
	ParentLocList  []lexer.Location // 所有的名称位置列表
	Comment        string           // 其他所有的注释内容
	CommentLoc     lexer.Location   // 注释内容的位置信息
}

// AnnotateGenericVarState 泛型变量的结构
type AnnotateGenericVarState struct {
	NormalStr   string // 可能的字符串
	IsBacktick  bool   // 是否字符串表达式变量
	IsArrayType bool   // 是否为数组类型
}

// AnnotateVarargState 可变参数的结构
type AnnotateVarargState struct {
	VarargType Type           // 定义的类型
	Comment    string         // 其他所有的注释内容
	CommentLoc lexer.Location // 注释内容的位置信息
}

// AnnotateEnumState 枚举的结构
type AnnotateEnumState struct {
	EnumLoc    lexer.Location // enum位置
	EnumType   EnumType       // 枚举的类型
	Comment    string         // 其他所有的注释内容
	CommentLoc lexer.Location // 注释内容的位置信息
}

// AnnotateEnumEndState 枚举的结构
type AnnotateEnumEndState struct {
	EnumLoc    lexer.Location // enum位置
	Comment    string         // 其他所有的注释内容
	CommentLoc lexer.Location // 注释内容的位置信息
}

// AnnotateNotValidState 无效的Stat
type AnnotateNotValidState struct {
}
