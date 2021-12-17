package parser

import (
	"fmt"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// Parser 语法分析器，为把Lua源码文件解析成AST抽象语法树
type Parser struct {
	// 词法分析器对象
	l *lexer.Lexer

	parseErrs []lexer.ParseError
}

// CreateParser 创建一个分析对象
func CreateParser(chunk []byte, chunkName string) *Parser {
	parser := &Parser{}
	errHandler := parser.insertErr
	parser.l = lexer.NewLexer(chunk, chunkName)
	parser.l.SetErrHandler(errHandler)

	return parser
}

// BeginAnalyze 开始分析
func (p *Parser) BeginAnalyze() (block *ast.Block, commentMap map[int]*lexer.CommentInfo, errList []lexer.ParseError) {
	defer func() {
		if err1 := recover(); err1 != nil {
			block = &ast.Block{}
			commentMap = p.l.GetCommentMap()
			errList = p.parseErrs
			return
		}
	}()

	p.l.SkipFirstLineComment()

	blockBeginLoc := p.l.GetHeardTokenLoc()
	block = p.parseBlock() // block
	blockEndLoc := p.l.GetNowTokenLoc()
	block.Loc = lexer.GetRangeLoc(&blockBeginLoc, &blockEndLoc)

	p.l.NextTokenKind(lexer.TkEOF)
	p.l.SetEnd()
	return block, p.l.GetCommentMap(), p.parseErrs
}

// BeginAnalyzeExp ParseExp single exp
func (p *Parser) BeginAnalyzeExp() (exp ast.Exp) {
	defer func() {
		if err2 := recover(); err2 != nil {
			exp = nil
		}
	}()

	exp = p.parseSubExp(0)
	return exp
}

// GetErrList get parse error list
func (p *Parser) GetErrList() (errList []lexer.ParseError) {
	return p.parseErrs
}

// insert now token info
func (p *Parser) insertParserErr(loc lexer.Location, f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	paseError := lexer.ParseError{
		ErrStr:      err,
		Loc:         loc,
		ReadFileErr: false,
	}

	p.insertErr(paseError)
}

func (p *Parser) insertErr(oneErr lexer.ParseError) {
	if len(p.parseErrs) < 30 {
		p.parseErrs = append(p.parseErrs, oneErr)
	} else {
		oneErr.ErrStr = oneErr.ErrStr + "(too many err...)"
		p.parseErrs = append(p.parseErrs, oneErr)
		manyError := &lexer.TooManyErr{
			ErrNum: 30,
		}

		panic(manyError)
	}
}
