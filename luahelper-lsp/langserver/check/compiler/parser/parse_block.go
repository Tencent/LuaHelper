package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// block ::= {stat} [retstat]
func (p *Parser) parseBlock() *ast.Block {
	return &ast.Block{
		Stats:   p.parseStats(),
		RetExps: p.parseRetExps(),
	}
}

func (p *Parser) parseStats() []ast.Stat {
	stats := make([]ast.Stat, 0, 1)
	for !isReturnOrBlockEnd(p.l.LookAheadKind()) {
		stat := p.parseStat()
		if _, ok := stat.(*ast.EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

// retstat ::= return [explist] [‘;’]
// explist ::= exp {‘,’ exp}
func (p *Parser) parseRetExps() []ast.Exp {
	l := p.l
	if l.LookAheadKind() != lexer.TkKwReturn {
		return nil
	}

	l.NextToken()
	switch l.LookAheadKind() {
	case lexer.TkEof, lexer.TkKwEnd,
		lexer.TkKwElse, lexer.TkKwElseif, lexer.TkKwUntil:
		return []ast.Exp{}
	case lexer.TkSepSemi:
		l.NextToken()
		return []ast.Exp{}
	default:
		exps := p.parseExpList()
		if l.LookAheadKind() == lexer.TkSepSemi {
			l.NextToken()
		}
		return exps
	}
}

func isReturnOrBlockEnd(tokenKind lexer.TkKind) bool {
	switch tokenKind {
	case lexer.TkKwReturn, lexer.TkEof, lexer.TkKwEnd,
		lexer.TkKwElse, lexer.TkKwElseif, lexer.TkKwUntil:
		return true
	}
	return false
}
