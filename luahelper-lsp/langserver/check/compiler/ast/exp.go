package ast

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp

prefixexp ::= var | functioncall | ‘(’ exp ‘)’

var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name

functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
*/

// Exp 表达式空接口
type Exp interface{}

// NilExp nil
type NilExp struct {
	Loc lexer.Location
}

// A BadExpr node is a placeholder for an expression containing
// syntax errors for which a correct expression node cannot be
// created.
//
type BadExpr struct {
	Loc lexer.Location
}

// TrueExp true
type TrueExp struct {
	Loc lexer.Location
}

// FalseExp false
type FalseExp struct {
	Loc lexer.Location
}

// VarargExp ...
type VarargExp struct {
	Loc lexer.Location
}

// IntegerExp 整数
type IntegerExp struct {
	Val int64
	Loc lexer.Location
}

// FloatExp 浮点数
type FloatExp struct {
	Val float64
	Loc lexer.Location
}

// Luajit 整数
type LuajitNum struct {
	Val int64
	Loc lexer.Location
}

// StringExp 字符串
type StringExp struct {
	Str string
	Loc lexer.Location
}

// UnopExp 一元表达式 unop exp
type UnopExp struct {
	Op  lexer.TkKind // operator
	Exp Exp
	Loc lexer.Location
}

// BinopExp 二元表达式 exp1 op exp2
type BinopExp struct {
	Op   lexer.TkKind // operator
	Exp1 Exp
	Exp2 Exp
	Loc  lexer.Location
}

// ConcatExp 删除了，转成了BinopExp， 其中 op为：TkOpConcat
type ConcatExp struct {
	Exp1 Exp
	Exp2 Exp
	Loc  lexer.Location
}

// TableConstructorExp table构造
// tableconstructor ::= ‘{’ [fieldlist] ‘}’
// fieldlist ::= field {fieldsep field} [fieldsep]
// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
// fieldsep ::= ‘,’ | ‘;’
type TableConstructorExp struct {
	KeyExps []Exp
	ValExps []Exp
	Loc     lexer.Location
}

// FuncDefExp 函数定义
// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
type FuncDefExp struct {
	ClassName  string // 例如 function table.func() end // table即ClassName
	FuncName   string // 例如 function table.func() end // func即FuncName
	ParList    []string
	ParLocList []lexer.Location // 所有参数的位置信息
	Block      *Block
	Loc        lexer.Location
	IsVararg   bool // 是否是...可变参数
	IsColon    bool // 是否为: 这样的函数
}

/*
prefixexp ::= Name |
              ‘(’ exp ‘)’ |
              prefixexp ‘[’ exp ‘]’ |
              prefixexp ‘.’ Name |
              prefixexp ‘:’ Name args |
              prefixexp args
*/

// NameExp 引用其他变量
type NameExp struct {
	Name string
	Loc  lexer.Location
}

// ParensExp 括号包含表达式或值
type ParensExp struct {
	Exp Exp
	Loc lexer.Location
}

// TableAccessExp 成员变量获取
type TableAccessExp struct {
	PrefixExp Exp
	KeyExp    Exp
	Loc       lexer.Location
}

// FuncCallExp 函数调用
// 当调用这样的函数时 aaa:bb("1", "2")
// 其中aaa 为 PrefixExp， bb 为NameExp，括号内的为参数
type FuncCallExp struct {
	PrefixExp Exp
	NameExp   *StringExp
	Args      []Exp
	Loc       lexer.Location
}
