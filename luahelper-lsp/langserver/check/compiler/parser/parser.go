package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// Parser 语法分析器，为把Lua源码文件解析成AST抽象语法树
type Parser struct {
	// 词法分析器对象
	l *lexer.Lexer
}

// CreateParser 创建一个分析对象
func CreateParser(chunk []byte, chunkName string) *Parser {
	return &Parser{
		l: lexer.NewLexer(chunk, chunkName),
	}
}

// BeginAnalyze 开始分析
func (p *Parser) BeginAnalyze() (block *ast.Block,commentMap map[int]*lexer.CommentInfo, err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			luaParseErr := recoverErr.(lexer.LuaParseError)
			err = luaParseErr
		}
	}()

	p.l.SkipFirstLineComment()

	blockBeginLoc := p.l.GetHeardTokenLoc()
	block = p.parseBlock() // block
	blockEndLoc := p.l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	p.l.NextTokenOfKind(lexer.TkEof)
	p.l.SetEnd()
	return block, p.l.GetCommentMap(), nil
}

// ParseExp single exp
func (p *Parser) BeginAnalyzeExp() (exp ast.Exp) {
	defer func() {
		if err2 := recover(); err2 != nil {
			exp = nil
		}
	}()

	exp = p.parseSubExp(0)
	return exp
}
