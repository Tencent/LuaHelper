package lexer

import "strconv"

//TkKind token kind
type TkKind int

// The list of tokens.
const (
	IKIllegal    TkKind = iota      // illegal
	TkEOF                           // end-of-file
	TkVararg                        // ...
	TkSepSemi                       // ;
	TkSepComma                      // ,
	TkSepDot                        // .
	TkSepColon                      // :
	TkSepLabel                      // ::
	TkSepLparen                     // (
	TkSepRparen                     // )
	TkSepLbrack                     // [
	TkSepRbrack                     // ]
	TkSepLcurly                     // {
	TkSepRcurly                     // }
	TkOpAssign                      // =
	TkOpMinus                       // - (sub or unm)
	TkOpWave                        // ~ (bnot or bxor)
	TkOpAdd                         // +
	TkOpMul                         // *
	TkOpDiv                         // /
	TkOpIdiv                        // //
	TkOpPow                         // ^
	TkOpMod                         // %
	TkOpBand                        // &
	TkOpBor                         // |
	TkOpShr                         // >>
	TkOpShl                         // <<
	TkOpConcat                      // ..
	TkOpLt                          // <
	TkOpLe                          // <=
	TkOpGt                          // >
	TkOpGe                          // >=
	TkOpEq                          // ==
	TkOpNe                          // ~=
	TkOpNen                         // #
	TkOpAnd                         // and
	TkOpOr                          // or
	TkOpNot                         // not
	TkKwBreak                       // break
	TkKwDo                          // do
	TkKwElse                        // else
	TkKwElseif                      // elseif
	TkKwEnd                         // end
	TkKwFalse                       // false
	TkKwFor                         // for
	TkKwFunction                    // function
	TkKwGoto                        // goto
	TkKwIf                          // if
	TkKwIn                          // in
	TkKwLocal                       // local
	TkKwNil                         // nil
	TkKwRepeat                      // repeat
	TkKwReturn                      // return
	TkKwThen                        // then
	TkKwTrue                        // true
	TkKwUntil                       // until
	TkKwWhile                       // while
	TkIdentifier                    // identifier
	TkNumber                        // number literal
	TkString                        // string literal
	TkOpUnm      TkKind = TkOpMinus // unary minus
	TkOpSub      TkKind = TkOpMinus
	TkOpBnot     TkKind = TkOpWave
	TkOpBxor     TkKind = TkOpWave
)

var tokenKinds = [...]string{
	IKIllegal: "ILLEGAL",

	TkEOF:        "EOF",            // end-of-file
	TkVararg:     "...",            // ...
	TkSepSemi:    ";",              // ;
	TkSepComma:   ",",              // ,
	TkSepDot:     ".",              // .
	TkSepColon:   ":",              // :
	TkSepLabel:   "::",             // ::
	TkSepLparen:  "(",              // (
	TkSepRparen:  ")",              // )
	TkSepLbrack:  "[",              // [
	TkSepRbrack:  "]",              // ]
	TkSepLcurly:  "{",              // {
	TkSepRcurly:  "}",              // }
	TkOpAssign:   "=",              // =
	TkOpMinus:    "-",              // - (sub or unm)
	TkOpWave:     "~",              // ~ (bnot or bxor)
	TkOpAdd:      "+",              // +
	TkOpMul:      "*",              // *
	TkOpDiv:      "/",              // /
	TkOpIdiv:     "//",             // //
	TkOpPow:      "^",              // ^
	TkOpMod:      "%",              // %
	TkOpBand:     "&",              // &
	TkOpBor:      "|",              // |
	TkOpShr:      ">>",             // >>
	TkOpShl:      "<<",             // <<
	TkOpConcat:   "..",             // ..
	TkOpLt:       "<",              // <
	TkOpLe:       "<=",             // <=
	TkOpGt:       ">",              // >
	TkOpGe:       ">=",             // >=
	TkOpEq:       "==",             // ==
	TkOpNe:       "~=",             // ~=
	TkOpNen:      "#",              // #
	TkOpAnd:      "and",            // and
	TkOpOr:       "or",             // or
	TkOpNot:      "not",            // not
	TkKwBreak:    "break",          // break
	TkKwDo:       "do",             // do
	TkKwElse:     "else",           // else
	TkKwElseif:   "elseif",         // elseif
	TkKwEnd:      "end",            // end
	TkKwFalse:    "false",          // false
	TkKwFor:      "for",            // for
	TkKwFunction: "function",       // function
	TkKwGoto:     "goto",           // goto
	TkKwIf:       "if",             // if
	TkKwIn:       "in",             // in
	TkKwLocal:    "local",          // local
	TkKwNil:      "nil",            // nil
	TkKwRepeat:   "repeat",         // repeat
	TkKwReturn:   "return",         // return
	TkKwThen:     "then",           // then
	TkKwTrue:     "true",           // true
	TkKwUntil:    "until",          // until
	TkKwWhile:    "while",          // while
	TkIdentifier: "identifier",     // identifier
	TkNumber:     "number literal", // number literal
	TkString:     "string literal", // string literal
}

func (tok TkKind) String() string {
	s := ""
	if 0 <= tok && tok < TkKind(len(tokenKinds)) {
		s = tokenKinds[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

var keywords = map[string]TkKind{
	"and":      TkOpAnd,
	"break":    TkKwBreak,
	"do":       TkKwDo,
	"else":     TkKwElse,
	"elseif":   TkKwElseif,
	"end":      TkKwEnd,
	"false":    TkKwFalse,
	"for":      TkKwFor,
	"function": TkKwFunction,
	"goto":     TkKwGoto,
	"if":       TkKwIf,
	"in":       TkKwIn,
	"local":    TkKwLocal,
	"nil":      TkKwNil,
	"not":      TkOpNot,
	"or":       TkOpOr,
	"repeat":   TkKwRepeat,
	"return":   TkKwReturn,
	"then":     TkKwThen,
	"true":     TkKwTrue,
	"until":    TkKwUntil,
	"while":    TkKwWhile,
}
