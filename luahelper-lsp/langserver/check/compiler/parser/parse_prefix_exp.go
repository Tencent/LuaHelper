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
	aheadKind := l.LookAheadKind()
	if aheadKind == lexer.TkIdentifier {
		_, name := l.NextIdentifier() // Name
		loc := l.GetNowTokenLoc()
		exp = &ast.NameExp{
			Name: name,
			Loc:  loc,
		}
	} else if aheadKind == lexer.TkSepLparen { // ‘(’ exp ‘)’
		exp = p.parseParensExp()
	} else {
		l.NextToken()
		loc := l.GetNowTokenLoc()
		exp = &ast.BadExpr{
			Loc: loc,
		}
		p.insertParserErr(loc, "`%s` can not start", aheadKind.String())
	}
	return p.finishPrefixExp(exp, &beginLoc)
}

func (p *Parser) parseParensExp() ast.Exp {
	l := p.l
	l.NextTokenKind(lexer.TkSepLparen) // (
	beginLoc := l.GetNowTokenLoc()
	exp := p.parseExp() // exp

	l.NextTokenKind(lexer.TkSepRparen) // )
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
			l.NextToken()                      // ‘[’
			keyExp := p.parseExp()             // exp
			l.NextTokenKind(lexer.TkSepRbrack) // ‘]’
			endLoc := l.GetNowTokenLoc()
			loc := lexer.GetRangeLoc(beginLoc, &endLoc)
			exp = &ast.TableAccessExp{
				PrefixExp: exp,
				KeyExp:    keyExp,
				Loc:       loc,
			}
		case lexer.TkSepDot: // prefixexp ‘.’ Name
			l.NextToken() // ‘.’
			nextKind := l.LookAheadKind()
			var loc lexer.Location
			var filedName string
			if nextKind == lexer.TkIdentifier {
				_, filedName = l.NextIdentifier() // Name
				loc = l.GetNowTokenLoc()
			} else {
				filedName = ""
				loc = l.GetNowTokenLoc()
				p.insertParserErr(loc, "missing field or attribute names")
			}

			keyExp := &ast.StringExp{
				Str: filedName,
				Loc: loc,
			}
			endLoc := l.GetNowTokenLoc()
			tableLoc := lexer.GetRangeLoc(beginLoc, &endLoc)
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
		aheadKind := l.LookAheadKind()
		var filedName string
		if aheadKind == lexer.TkIdentifier {
			_, filedName = l.NextIdentifier()
		} else {
			filedName = ""
			loc := l.GetNowTokenLoc()
			p.insertParserErr(loc, "missing field or attribute names")
		}
		loc := l.GetNowTokenLoc()
		return &ast.StringExp{
			Str: filedName,
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
		l.NextTokenKind(lexer.TkSepRparen)
	case lexer.TkSepLcurly: // ‘{’ [fieldlist] ‘}’
		args = []ast.Exp{p.parseTableConstructorExp()}
	default: // LiteralString
		aheadKind := l.LookAheadKind()
		if aheadKind == lexer.TkString {
			_, str := l.NextTokenKind(lexer.TkString)
			loc := l.GetNowTokenLoc()
			args = []ast.Exp{&ast.StringExp{
				Str: str,
				Loc: loc,
			}}
		} else {
			loc := l.GetNowTokenLoc()
			p.insertParserErr(loc, "missing function call args")
		}
	}
	return
}
