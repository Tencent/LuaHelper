package ast

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// chunk ::= block
// type Chunk *Block

// Block code block
// block ::= {stat} [retstat]
// retstat ::= return [explist] [';']
// explist ::= exp {',' exp}}
type Block struct {
	Stats   []Stat
	RetExps []Exp
	Loc     lexer.Location
}
