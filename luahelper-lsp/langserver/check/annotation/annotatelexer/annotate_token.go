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
	ATokenKwEnum                         // enum 枚举段关键值
	ATokenKwEnumStart                    // start enum后面跟着的开始关键字，例如完整的为enum start
	ATokenKwEnumEnd                      // end enum后面跟着的结束关键字，例如完整的为enum end
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
	"enum":      ATokenKwEnum,
}
