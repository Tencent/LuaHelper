package parser

import (
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

var _statEmpty = &ast.EmptyStat{}

/*
stat ::=  ‘;’
	| break
	| ‘::’ Name ‘::’
	| goto Name
	| do block end
	| while exp do block end
	| repeat block until exp
	| if exp then block {elseif exp then block} [else block] end
	| for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
	| for namelist in explist do block end
	| function funcname funcbody
	| local function Name funcbody
	| local namelist [‘=’ explist]
	| varlist ‘=’ explist
	| functioncall
*/
func (p *Parser) parseStat() ast.Stat {
	switch p.l.LookAheadKind() {
	case lexer.TkSepSemi:
		return p.parseEmptyStat()
	case lexer.TkKwBreak:
		return p.parseBreakStat()
	case lexer.TkSepLabel:
		return p.parseLabelStat()
	case lexer.TkKwGoto:
		return p.parseGotoStat()
	case lexer.TkKwDo:
		return p.parseDoStat()
	case lexer.TkKwWhile:
		return p.parseWhileStat()
	case lexer.TkKwRepeat:
		return p.parseRepeatStat()
	case lexer.TkKwIf:
		return p.parseIfStat()
	case lexer.TkKwFor:
		return p.parseForStat()
	case lexer.TkKwFunction:
		return p.parseFuncDefStat()
	case lexer.TkKwLocal:
		return p.parseLocalAssignOrFuncDefStat()
	case lexer.IKIllegal:
		return p.parseIKIllegalStat()
	default:
		return p.parseAssignOrFuncCallStat()
	}
}

// ;
func (p *Parser) parseEmptyStat() *ast.EmptyStat {
	p.l.NextTokenKind(lexer.TkSepSemi)
	return _statEmpty
}

// break
func (p *Parser) parseBreakStat() *ast.BreakStat {
	p.l.NextTokenKind(lexer.TkKwBreak)

	return &ast.BreakStat{
		//Loc: l.GetNowTokenLoc(),
	}
}

// ‘::’ Name ‘::’
func (p *Parser) parseLabelStat() *ast.LabelStat {
	p.l.NextTokenKind(lexer.TkSepLabel) // ::
	_, name := p.l.NextIdentifier()     // name
	loc := p.l.GetNowTokenLoc()
	p.l.NextTokenKind(lexer.TkSepLabel) // ::
	return &ast.LabelStat{
		Name: name,
		Loc:  loc,
	}
}

// goto Name
func (p *Parser) parseGotoStat() *ast.GotoStat {
	p.l.NextTokenKind(lexer.TkKwGoto) // goto
	_, name := p.l.NextIdentifier()   // name
	return &ast.GotoStat{
		Name: name,
		Loc:  p.l.GetNowTokenLoc(),
	}
}

// do block end
func (p *Parser) parseDoStat() *ast.DoStat {
	l := p.l
	l.NextTokenKind(lexer.TkKwDo) // do
	beginLoc := l.GetNowTokenLoc()

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end
	endLoc := l.GetNowTokenLoc()

	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.DoStat{
		Block: block,
		Loc:   loc,
	}
}

// while exp do block end
func (p *Parser) parseWhileStat() *ast.WhileStat {
	l := p.l
	l.NextTokenKind(lexer.TkKwWhile) // while
	beginLoc := l.GetNowTokenLoc()
	exp := p.parseExp()           // exp
	l.NextTokenKind(lexer.TkKwDo) // do

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)

	return &ast.WhileStat{
		Exp:   exp,
		Block: block,
		Loc:   loc,
	}
}

// repeat block until exp
func (p *Parser) parseRepeatStat() *ast.RepeatStat {
	l := p.l
	l.NextTokenKind(lexer.TkKwRepeat) // repeat
	beginLoc := l.GetNowTokenLoc()

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwUntil) // until
	exp := p.parseExp()              // exp
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)

	return &ast.RepeatStat{
		Block: block,
		Exp:   exp,
		Loc:   loc,
	}
}

// if exp then block {elseif exp then block} [else block] end
func (p *Parser) parseIfStat() *ast.IfStat {
	l := p.l
	exps := make([]ast.Exp, 0, 1)
	blocks := make([]*ast.Block, 0, 1)

	l.NextTokenKind(lexer.TkKwIf) // if
	beginLoc := l.GetNowTokenLoc()

	exps = append(exps, p.parseExp()) // exp
	l.NextTokenKind(lexer.TkKwThen)   // then

	thenBlockBeginLoc := l.GetHeardTokenLoc()
	thenBlock := p.parseBlock()
	thenBlockEndLoc := l.GetHeardTokenLoc()
	thenBlock.Loc = lexer.GetRangeLocExcludeEnd(&thenBlockBeginLoc, &thenBlockEndLoc)
	blocks = append(blocks, thenBlock) // block

	for l.LookAheadKind() == lexer.TkKwElseif {
		l.NextToken()                     // elseif
		exps = append(exps, p.parseExp()) // exp
		l.NextTokenKind(lexer.TkKwThen)   // then

		elseifBlockBeginLoc := l.GetHeardTokenLoc()
		elseifBlock := p.parseBlock()
		elseifBlockEndLoc := l.GetHeardTokenLoc()
		elseifBlock.Loc = lexer.GetRangeLocExcludeEnd(&elseifBlockBeginLoc, &elseifBlockEndLoc)

		blocks = append(blocks, elseifBlock) // block
	}

	// else block => elseif true then block
	if l.LookAheadKind() == lexer.TkKwElse {
		l.NextToken() // else
		exps = append(exps, &ast.TrueExp{
			Loc: l.GetNowTokenLoc(),
		})

		elseBlockBeginLoc := l.GetHeardTokenLoc()
		elseBlock := p.parseBlock()

		elseBlockEndLoc := l.GetHeardTokenLoc()
		elseBlock.Loc = lexer.GetRangeLocExcludeEnd(&elseBlockBeginLoc, &elseBlockEndLoc)

		blocks = append(blocks, elseBlock) // block
	}

	l.NextTokenKind(lexer.TkKwEnd) // end

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.IfStat{
		Exps:   exps,
		Blocks: blocks,
		Loc:    loc,
	}
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
// for namelist in explist do block end
func (p *Parser) parseForStat() ast.Stat {
	l := p.l
	lineOfFor, _ := l.NextTokenKind(lexer.TkKwFor)
	beginLoc := l.GetNowTokenLoc()

	_, name := l.NextIdentifier()
	if l.LookAheadKind() == lexer.TkOpAssign {
		return p.finishForNumStat(lineOfFor, name, &beginLoc)
	}
	return p.finishForInStat(name, &beginLoc)
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
func (p *Parser) finishForNumStat(lineOfFor int, varName string, beginLoc *lexer.Location) *ast.ForNumStat {
	l := p.l
	varNameLoc := l.GetNowTokenLoc()
	l.NextTokenKind(lexer.TkOpAssign) // for name =
	initExp := p.parseExp()           // exp
	l.NextTokenKind(lexer.TkSepComma) // ,
	limitExp := p.parseExp()          // exp

	var stepExp ast.Exp
	if l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()          // ,
		stepExp = p.parseExp() // exp
	} else {
		// 这里的位置值可能不太准确
		stepExp = &ast.IntegerExp{
			Val: 1,
			//Loc: l.GetNowTokenLoc(),
		}
	}

	l.NextTokenKind(lexer.TkKwDo) // do

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(beginLoc, &endLoc)

	return &ast.ForNumStat{
		VarName:  varName,
		VarLoc:   varNameLoc,
		InitExp:  initExp,
		LimitExp: limitExp,
		StepExp:  stepExp,
		Block:    block,
		Loc:      loc,
	}
}

// for namelist in explist do block end
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
func (p *Parser) finishForInStat(name0 string, beginLoc *lexer.Location) *ast.ForInStat {
	l := p.l
	varLoc0 := l.GetNowTokenLoc()
	nameList, nameLocList := p.finishNameList(name0, varLoc0) // for namelist
	l.NextTokenKind(lexer.TkKwIn)                             // in
	expList := p.parseExpList()                               // explist
	l.NextTokenKind(lexer.TkKwDo)                             // do

	blockBeginLoc := l.GetHeardTokenLoc()
	block := p.parseBlock() // block
	blockEndLoc := l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	l.NextTokenKind(lexer.TkKwEnd) // end

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(beginLoc, &endLoc)

	return &ast.ForInStat{
		NameList:    nameList,
		NameLocList: nameLocList,
		ExpList:     expList,
		Block:       block,
		Loc:         loc,
	}
}

// namelist ::= Name {‘,’ Name}
func (p *Parser) finishNameList(name0 string, varLoc0 lexer.Location) ([]string, []lexer.Location) {
	l := p.l
	names := []string{name0}
	locs := []lexer.Location{varLoc0}
	for l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()                 // ,
		_, name := l.NextIdentifier() // Name
		loc := l.GetNowTokenLoc()
		locs = append(locs, loc)
		names = append(names, name)
	}
	return names, locs
}

// get local var attribute, add by guochuliang 2020-08-20
func (p *Parser) getLocalAttribute() ast.LocalAttr {
	l := p.l
	if l.LookAheadKind() == lexer.TkOpLt {
		l.NextToken()
		_, attr := l.NextIdentifier()

		if attr == "close" {
			l.NextTokenKind(lexer.TkOpGt)
			return ast.RDKTOCLOSE
		} else if attr == "const" {
			l.NextTokenKind(lexer.TkOpGt)
			return ast.RDKCONST
		} else {
			p.insertParserErr(l.GetNowTokenLoc(), "unrecognized local varible attribute '%s' ", attr)
			l.NextTokenKind(lexer.TkOpGt)
		}
	}
	return ast.VDKREG
}

//namelist for loacl var after lua 5.4 add support for local var attribute, added by guochuliang 2020-08-20
//5.3
// namelist ::= Name {‘,’ Name}
// 5.4
// namelist ::= Name attrib {‘,’ Name attrib}
// attrib ::= [ '<' Name '>' ] 5.4
func (p *Parser) finishLocalNameList(name0 string, varLoc0 lexer.Location, kind ast.LocalAttr) ([]string,
	[]lexer.Location, []ast.LocalAttr) {
	l := p.l
	index := -1
	if kind == ast.RDKTOCLOSE {
		index++
	}
	names := []string{name0}
	kinds := []ast.LocalAttr{kind}
	locs := []lexer.Location{varLoc0}
	for l.LookAheadKind() == lexer.TkSepComma {
		l.NextToken()                 // ,
		_, name := l.NextIdentifier() // Name
		loc := l.GetNowTokenLoc()
		kind := p.getLocalAttribute()
		if kind == ast.RDKTOCLOSE {
			if index != -1 {
				p.insertParserErr(l.GetPreTokenLoc(), "more than one to_be_close variables found in local list")
			} else {
				index++
			}
		}
		locs = append(locs, loc)
		kinds = append(kinds, kind)
		names = append(names, name)
	}
	return names, locs, kinds
}

// local function Name funcbody
// local namelist [‘=’ explist]
func (p *Parser) parseLocalAssignOrFuncDefStat() ast.Stat {
	l := p.l
	l.NextTokenKind(lexer.TkKwLocal)
	if l.LookAheadKind() == lexer.TkKwFunction {
		return p.finishLocalFuncDefStat()
	}

	return p.finishLocalVarDeclStat()
}

/*
http://www.lua.org/manual/5.3/manual.html#3.4.11

function f() end          =>  f = function() end
function t.a.b.c.f() end  =>  t.a.b.c.f = function() end
function t.a.b.c:f() end  =>  t.a.b.c.f = function(self) end
local function f() end    =>  local f; f = function() end

The statement `local function f () body end`
translates to `local f; f = function () body end`
not to `local f = function () body end`
(This only makes a difference when the body of the function
 contains references to f.)
*/
// local function Name funcbody
func (p *Parser) finishLocalFuncDefStat() *ast.LocalFuncDefStat {
	l := p.l
	beginLoc := l.GetNowTokenLoc()

	l.NextTokenKind(lexer.TkKwFunction) // local function
	_, name := l.NextIdentifier()       // name
	nameLoc := l.GetNowTokenLoc()
	fdExp := p.parseFuncDefExp(&beginLoc) // funcbody

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.LocalFuncDefStat{
		Name:    name,
		NameLoc: nameLoc,
		Exp:     fdExp,
		Loc:     loc,
	}
}

// local namelist [‘=’ explist]
func (p *Parser) finishLocalVarDeclStat() *ast.LocalVarDeclStat {
	l := p.l
	beginLoc := l.GetNowTokenLoc()
	_, name0 := l.NextIdentifier() // local Name
	varLoc0 := l.GetNowTokenLoc()
	kind0 := p.getLocalAttribute()                                              // added to support lua5.4
	nameList, locList, attrList := p.finishLocalNameList(name0, varLoc0, kind0) // { , Name attrib}
	var expList []ast.Exp
	if l.LookAheadKind() == lexer.TkOpAssign {
		l.NextToken()              // ==
		expList = p.parseExpList() // explist
	}

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.LocalVarDeclStat{
		NameList:   nameList,
		VarLocList: locList,
		AttrList:   attrList,
		ExpList:    expList,
		Loc:        loc,
	}
}

// varlist ‘=’ explist
// functioncall
func (p *Parser) parseAssignOrFuncCallStat() ast.Stat {
	l := p.l
	beginLoc := l.GetHeardTokenLoc()
	prefixExp := p.parsePrefixExp()
	if _, ok := prefixExp.(*ast.BadExpr); ok {
		return &ast.EmptyStat{}
	}

	if fc, ok := prefixExp.(*ast.FuncCallExp); ok {
		endLoc := l.GetNowTokenLoc()
		fc.Loc = lexer.GetRangeLoc(&beginLoc, &endLoc)
		return fc
	}

	assignStat := p.parseAssignStat(beginLoc, prefixExp)

	return assignStat
}

// varlist ‘=’ explist |
func (p *Parser) parseAssignStat(preLoc lexer.Location, var0 ast.Exp) ast.Stat {
	l := p.l
	symList := p.finishVarList(var0) // varlist

	aheadKind := l.LookAheadKind()
	if aheadKind != lexer.TkOpAssign {
		nowLoc := l.GetNowTokenLoc()
		loc := lexer.GetRangeLoc(&preLoc, &nowLoc)
		p.insertParserErr(loc, "expression cannot be used as a statement")
		return &ast.EmptyStat{}
	}

	l.NextTokenKind(lexer.TkOpAssign) // =
	expList := p.parseExpList()       // explist
	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&preLoc, &endLoc)

	return &ast.AssignStat{
		VarList: symList,
		ExpList: expList,
		Loc:     loc,
	}
}

// varlist ::= var {‘,’ var}
func (p *Parser) finishVarList(var0 ast.Exp) []ast.Exp {
	l := p.l
	vars := []ast.Exp{p.checkVar(var0)}         // var := p
	for l.LookAheadKind() == lexer.TkSepComma { // {
		l.NextToken()                        // ,
		exp := p.parsePrefixExp()            // var
		vars = append(vars, p.checkVar(exp)) //
	} // }
	return vars
}

// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
func (p *Parser) checkVar(exp ast.Exp) ast.Exp {
	l := p.l
	switch exp.(type) {
	case *ast.NameExp, *ast.TableAccessExp, *ast.BadExpr:
		return exp
	}

	loc := l.GetNowTokenLoc()
	return &ast.BadExpr{
		Loc: loc,
	}
	// l.NextTokenKind(-1) // trigger error
	// panic("unreachable!")
}

// function funcname funcbody
// funcname ::= Name {‘.’ Name} [‘:’ Name]
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
func (p *Parser) parseFuncDefStat() *ast.AssignStat {
	l := p.l
	l.NextTokenKind(lexer.TkKwFunction) // function
	beginLoc := l.GetNowTokenLoc()
	fnExp, hasColon, className, funcName := p.parseFuncName() // funcname
	selfLoc := l.GetNowTokenLoc()
	fdExp := p.parseFuncDefExp(&beginLoc) // funcbody
	fdExp.ClassName = className
	fdExp.FuncName = funcName
	if hasColon { // insert self
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
		fdExp.IsColon = true

		fdExp.ParLocList = append(fdExp.ParLocList, lexer.Location{})
		copy(fdExp.ParLocList[1:], fdExp.ParLocList)
		fdExp.ParLocList[0] = selfLoc
	}

	endLoc := l.GetNowTokenLoc()
	loc := lexer.GetRangeLoc(&beginLoc, &endLoc)
	return &ast.AssignStat{
		VarList: []ast.Exp{fnExp},
		ExpList: []ast.Exp{fdExp},
		Loc:     loc,
	}
}

// funcname ::= Name {‘.’ Name} [‘:’ Name]
// 设定：最后一个是函数名 倒数第二个是类名
func (p *Parser) parseFuncName() (exp ast.Exp, hasColon bool, className string, funcName string) {
	l := p.l
	_, name := l.NextIdentifier()
	loc := l.GetNowTokenLoc()

	beginTableLoc := l.GetNowTokenLoc()
	exp = &ast.NameExp{
		Name: name,
		Loc:  loc,
	}

	funcName = name
	for l.LookAheadKind() == lexer.TkSepDot {
		className = name
		l.NextToken()
		_, name := l.NextIdentifier()
		funcName = name
		loc := l.GetNowTokenLoc()
		idx := &ast.StringExp{
			Str: name,
			Loc: loc,
		}

		endTableLoc := l.GetNowTokenLoc()
		tableLoc := lexer.GetRangeLoc(&beginTableLoc, &endTableLoc)

		exp = &ast.TableAccessExp{
			PrefixExp: exp,
			KeyExp:    idx,
			Loc:       tableLoc,
		}
	}

	if l.LookAheadKind() == lexer.TkSepColon {
		className = name
		l.NextToken()
		_, name := l.NextIdentifier()
		funcName = name
		loc := l.GetNowTokenLoc()
		idx := &ast.StringExp{
			Str: name,
			Loc: loc,
		}

		endTableLoc := l.GetNowTokenLoc()
		tableLoc := lexer.GetRangeLoc(&beginTableLoc, &endTableLoc)

		exp = &ast.TableAccessExp{
			PrefixExp: exp,
			KeyExp:    idx,
			Loc:       tableLoc,
		}
		hasColon = true
	}

	return
}

// func (p *Parser) parseIKIllegalStat() *ast.IllegalStat{
// 	l := p.l
// 	loc := l.GetNowTokenLoc()
// 	l.NextToken()
// 	return &ast.IllegalStat{
// 		Name: "",
// 		Loc:  loc,
// 	}
// }

func (p *Parser) parseIKIllegalStat() *ast.EmptyStat {
	p.l.NextToken()
	return _statEmpty
}
