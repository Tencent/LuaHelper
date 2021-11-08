package lexer

//TkKind token kind
type TkKind int

const (
	TkEof        TkKind = iota      // end-of-file
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
