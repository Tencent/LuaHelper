package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args

// prefixexp ::= Name
// 	| ‘(’ exp ‘)’
// 	| prefixexp ‘[’ exp ‘]’
// 	| prefixexp ‘.’ Name
// 	| prefixexp [‘:’ Name] args

func (p *Parser) parsePrefixExp() ast.Exp {
	l := p.l
	var exp ast.Exp

	beginLoc := l.GetHeardTokenLoc()
	if l.LookAheadKind() == lexer.TkIdentifier {
		_, name := l.NextIdentifier() // Name
		loc := l.GetNowTokenLoc()
		exp = &ast.NameExp{
			Name: name,
			Loc:  loc,
		}
	} else { // ‘(’ exp ‘)’
		exp = p.parseParensExp()
	}
	return p.finishPrefixExp(exp, &beginLoc)
}

func (p *Parser) parseParensExp() ast.Exp {
	l := p.l
	l.NextTokenOfKind(lexer.TkSepLparen) // (
	beginLoc := l.GetNowTokenLoc()
	exp := p.parseExp() // exp

	l.NextTokenOfKind(lexer.TkSepRparen) // )
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)

	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp, *ast.NameExp, *ast.TableAccessExp:
		return &ast.ParensExp{
			Exp: exp,
			Loc: loc,
		}
	}

	// no need to keep parens
	return exp
}

func (p *Parser) finishPrefixExp(exp ast.Exp, beginLoc *lexer.Location) ast.Exp {
	l := p.l
	for {
		switch l.LookAheadKind() {
		case lexer.TkSepLbrack: // prefixexp ‘[’ exp ‘]’
			l.NextToken()                        // ‘[’
			keyExp := p.parseExp()               // exp
			l.NextTokenOfKind(lexer.TkSepRbrack) // ‘]’
			endLoc := l.GetNowTokenLoc()
			loc := lexer.GetRangeLoc(beginLoc, &endLoc)
			exp = &ast.TableAccessExp{
				PrefixExp: exp,
				KeyExp:    keyExp,
				Loc:       loc,
			}
		case lexer.TkSepDot: // prefixexp ‘.’ Name
			l.NextToken()                 // ‘.’
			_, name := l.NextIdentifier() // Name
			loc := l.GetNowTokenLoc()
			endLoc := l.GetNowTokenLoc()
			tableLoc := lexer.GetRangeLoc(beginLoc, &endLoc)
			keyExp := &ast.StringExp{
				Str: name,
				Loc: loc,
			}
			exp = &ast.TableAccessExp{
				PrefixExp: exp,
				KeyExp:    keyExp,
				Loc:       tableLoc,
			}
		case lexer.TkSepColon, // prefixexp ‘:’ Name args
			lexer.TkSepLparen, lexer.TkSepLcurly, lexer.TkString: // prefixexp args
			exp = p.finishFuncCallExp(exp)
		default:
			return exp
		}
	}
	//return exp
}

// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
func (p *Parser) finishFuncCallExp(prefixExp ast.Exp) *ast.FuncCallExp {
	l := p.l
	beginLoc := l.GetNowTokenLoc()
	nameExp := p.parseNameExp()
	args := p.parseArgs()
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.FuncCallExp{
		PrefixExp: prefixExp,
		NameExp:   nameExp,
		Args:      args,
		Loc:       loc,
	}
}

func (p *Parser) parseNameExp() *ast.StringExp {
	l := p.l
	if l.LookAheadKind() == lexer.TkSepColon {
		l.NextToken()
		_, name := l.NextIdentifier()
		loc := l.GetNowTokenLoc()
		return &ast.StringExp{
			Str: name,
			Loc: loc,
		}
	}
	return nil
}

// args ::=  ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func (p *Parser) parseArgs() (args []ast.Exp) {
	l := p.l
	switch l.LookAheadKind() {
	case lexer.TkSepLparen: // ‘(’ [explist] ‘)’
		l.NextToken() // lexer.TkSepLparen
		if l.LookAheadKind() != lexer.TkSepRparen {
			args = p.parseExpList()
		}
		l.NextTokenOfKind(lexer.TkSepRparen)
	case lexer.TkSepLcurly: // ‘{’ [fieldlist] ‘}’
		args = []ast.Exp{p.parseTableConstructorExp()}
	default: // LiteralString
		_, str := l.NextTokenOfKind(lexer.TkString)
		loc := l.GetNowTokenLoc()
		args = []ast.Exp{&ast.StringExp{
			Str: str,
			Loc: loc,
		}}
	}
	return
}
