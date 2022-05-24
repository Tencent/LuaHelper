package annotatelexer

// ATokenType 类型
type ATokenType int8

const (
	ATokenEOF          ATokenType = iota // end-of-file
	ATokenSepComma                       // ,
	ATokenSepColon                       // :
	ATokenVararg                         // ... 函数的可变参数
	ATokenVSepLparen                     // (
	ATokenVSepRparen                     // )
	ATokenVSepLbrack                     // [
	ATokenVSepRbrack                     // ]
	ATokenBor                            // |
	ATokenLt                             // <
	ATokenGt                             // >
	ATokenAt                             // @
	ATokenOption                         // ?
	ATokenString                         // 定义的其他字符串
	ATokenKwFun                          // fun
	ATokenKwTable                        // table
	ATokenKwType                         // type
	ATokenKwParam                        // param
	ATokenKwField                        // field
	ATokenKwClass                        // class
	ATokenKwReturn                       // return
	ATokenKwOverload                     // overload
	ATokenKwAlias                        // alias
	ATokenKwGeneric                      // generic
	ATokenKwPubic                        // public
	ATokenKwProtected                    // protected
	ATokenKwPrivate                      // private
	ATokenKwVararg                       // vararg
	ATokenKwIdentifier                   // identifier
	ATokenKwConst                        // const
	ATokenKwOther                        // other token， not valid
)

var keywords = map[string]ATokenType{
	"fun":       ATokenKwFun,
	"table":     ATokenKwTable,
	"type":      ATokenKwType,
	"param":     ATokenKwParam,
	"field":     ATokenKwField,
	"class":     ATokenKwClass,
	"return":    ATokenKwReturn,
	"overload":  ATokenKwOverload,
	"alias":     ATokenKwAlias,
	"generic":   ATokenKwGeneric,
	"public":    ATokenKwPubic,
	"protected": ATokenKwProtected,
	"private":   ATokenKwPrivate,
	"vararg":    ATokenKwVararg,
	"const":     ATokenKwConst,
}
