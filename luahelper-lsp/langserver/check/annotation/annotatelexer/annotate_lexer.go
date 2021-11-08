package annotatelexer

import (
	"fmt"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"strings"
)

// 注解模块的词法分析

// AnnotateToken 注解功能解析出来的每个单词
type AnnotateToken struct {
	valid         bool       // valid or not
	tokenStr      string     // token string
	tokenKind     ATokenType // token kind
	col           int        // 列号
	tokenStartCol int        // 切词的开始位置
}

// GetTokenKind 获取tokenkind
func (t *AnnotateToken) GetTokenKind() ATokenType {
	return t.tokenKind
}

// AnnotateLexer 注解模块词法分析的结构
type AnnotateLexer struct {
	chunk             string         // 这一行注释的内容
	line              int            // 这行注释的起始行号
	col               int            // 这行注释的起始列号
	tokenStartCol     int            // 切词的开始位置
	lastNormalTypeLoc lexer.Location // 最新annotateast.NormalType的位置信息
	preToken          AnnotateToken
	nowToken          AnnotateToken
	aheadToken        AnnotateToken
}

// CreateAnnotateLexer 创建一个注释的词法分析器
func CreateAnnotateLexer(chunk *string, line, col int) *AnnotateLexer {
	return &AnnotateLexer{
		chunk: *chunk,
		preToken: AnnotateToken{
			valid: false,
		},
		nowToken: AnnotateToken{
			valid: false,
		},
		aheadToken: AnnotateToken{
			valid: false,
		},
		line:          line,
		col:           col,
		tokenStartCol: col,
	}
}

// CheckHeardValid 校验头部是否有效，头部只能以-@开头
func (l *AnnotateLexer) CheckHeardValid() bool {
	// 判断这行内容是否以-@开头
	if l.test("-@") {
		l.next(2)
		return true
	}

	return false
}

func (l *AnnotateLexer) test(s string) bool {
	sLen := len(s)
	if len(l.chunk) < sLen {
		return false
	}

	for i := 0; i < sLen; i++ {
		if l.chunk[i] != s[i] {
			return false
		}
	}

	return true
}

func (l *AnnotateLexer) next(n int) {
	l.chunk = l.chunk[n:]
	l.col = l.col + n
}

// setNowToken 设置当前的单词
func (l *AnnotateLexer) setNowToken(kind ATokenType, tokenStr string) {
	l.preToken = l.nowToken
	l.nowToken.valid = true
	l.nowToken.col = l.col
	l.nowToken.tokenStartCol = l.tokenStartCol
	l.nowToken.tokenKind = kind
	l.nowToken.tokenStr = tokenStr
}

// skipWhiteSpaces 跳过空格
func (l *AnnotateLexer) skipWhiteSpaces() {
	for len(l.chunk) > 0 {
		if isWhiteSpace(l.chunk[0]) {
			l.next(1)
		} else {
			break
		}
	}
}

// NextTokenStruct 获取下一个单词结构
func (l *AnnotateLexer) NextTokenStruct() {
	if l.aheadToken.valid {
		l.preToken = l.nowToken
		l.nowToken = l.aheadToken
		l.aheadToken.valid = false
		return
	}

	l.skipWhiteSpaces()
	l.tokenStartCol = l.col

	if len(l.chunk) == 0 {
		l.setNowToken(ATokenEOF, "EOF")
		return
	}

	switch l.chunk[0] {
	case ',':
		l.next(1)
		l.setNowToken(ATokenSepComma, ",")
		return
	case '(':
		l.next(1)
		l.setNowToken(ATokenVSepLparen, "(")
		return
	case ')':
		l.next(1)
		l.setNowToken(ATokenVSepRparen, ")")
		return
	case '[':
		l.next(1)
		l.setNowToken(ATokenVSepLbrack, "[")
		return
	case ']':
		l.next(1)
		l.setNowToken(ATokenVSepRbrack, "]")
		return
	case '|':
		l.next(1)
		l.setNowToken(ATokenBor, "|")
		return
	case '<':
		l.next(1)
		l.setNowToken(ATokenLt, "<")
		return
	case '>':
		l.next(1)
		l.setNowToken(ATokenGt, ">")
		return
	case '@':
		l.next(1)
		l.setNowToken(ATokenAt, "@")
		return
	case '?':
		l.next(1)
		l.setNowToken(ATokenOption, "?")
		return
	case '.':
		if l.test("...") {
			l.next(3)
			l.setNowToken(ATokenVararg, "...")
			return
		}
	case ':':
		l.next(1)
		l.setNowToken(ATokenSepColon, ":")
		return
	case '\'', '"':
		l.setNowToken(ATokenString, l.scanShortString())
		return
	}

	c := l.chunk[0]
	if c == '_' || isLetter(c) || isDigit(c) {
		token := l.scanIdentifier()
		if kind, found := keywords[token]; found {
			l.setNowToken(kind, token)
		} else {
			l.setNowToken(ATokenKwIdentifier, token)
		}
		return
	}

	remianStr := string(l.chunk)
	l.setNowToken(ATokenKwOther, remianStr)
	return
}

// NextToken 下一个单词
func (l *AnnotateLexer) NextToken() (kind ATokenType, token string) {
	l.NextTokenStruct()
	kind = l.nowToken.tokenKind
	token = l.nowToken.tokenStr
	return
}

// NextTokenOfKind 检查下一个单词的类型，匹配
func (l *AnnotateLexer) NextTokenOfKind(kind ATokenType) (token string) {
	_kind, token := l.NextToken()
	if kind != _kind {
		l.ErrorPrint(AErrorKind, kind, "syntax error near '%s'", token)
	}
	return token
}

// NextIdentifier 下一个标识
func (l *AnnotateLexer) NextIdentifier() (token string) {
	return l.NextTokenOfKind(ATokenKwIdentifier)
}

// NextFieldName 下一个是标识符或是关键字名称, 可以用好field域的名称
func (l *AnnotateLexer) NextFieldName() string {
	kind, token := l.NextToken()
	if kind == ATokenKwIdentifier {
		return token
	}

	// 判断是否为特定的关键字
	inKind, flag := keywords[token]
	if !flag || inKind != kind {
		l.ErrorPrint(AErrorKind, inKind, "syntax error near '%s'", token)
	}

	return token
}

// NextTypeIdentifier 注解类型的标识符
func (l *AnnotateLexer) NextTypeIdentifier() (token string) {
	_kind, token := l.NextToken()
	if ATokenKwIdentifier != _kind {
		l.ErrorPrint(AErrorTypeIdentifier, ATokenKwIdentifier, "syntax error near '%s'", token)
	}
	return token
}

// NextParamName 下一个是标识符或是关键字名称，可以用于函数的参数名
func (l *AnnotateLexer) NextParamName() string {
	kind, token := l.NextToken()
	if kind == ATokenKwIdentifier || kind == ATokenVararg {
		return token
	}

	// 判断是否为特定的关键字
	inKind, flag := keywords[token]
	if !flag || inKind != kind {
		l.ErrorPrint(AErrorKind, inKind, "syntax error near '%s'", token)
	}

	return token
}

func (l *AnnotateLexer) scanShortString() string {
	delimiter := l.chunk[0]
	stringStart := 1

	str := ""
	i := 1
	for i < len(l.chunk) {
		ch := l.chunk[i]
		i++
		if delimiter == ch {
			// 找到了匹配的
			break
		}
	}

	if i-1 >= 1 {
		str += l.chunk[stringStart : i-1]
	}
	l.next(i)

	return str
}

// scanIdentifier 获取下个标识符
func (l *AnnotateLexer) scanIdentifier() string {
	i := 1
	for ; i < len(l.chunk); i++ {
		c := l.chunk[i]
		if isLetter(c) || isDigit(c) || c == '_' || c == '.' {
			continue
		}

		break
	}

	str := l.chunk[0:i]
	l.next(i)

	return str
}

func (l *AnnotateLexer) lookAheardToken() {
	if l.aheadToken.valid {
		return
	}

	backPreToken := l.preToken
	backNowToken := l.nowToken

	l.NextTokenStruct()
	nextToken := l.nowToken
	l.preToken = backPreToken
	l.nowToken = backNowToken
	l.aheadToken = nextToken
}

// LookAheadKind 获取头部的token
func (l *AnnotateLexer) LookAheadKind() ATokenType {
	if l.aheadToken.valid {
		return l.aheadToken.tokenKind
	}

	l.lookAheardToken()
	if !l.aheadToken.valid {
		l.ErrorPrint(AErrorKind, ATokenEOF, "syntax error near '%s'", l.aheadToken.tokenStr)
	}
	return l.aheadToken.tokenKind
}

// GetRemainComment 获取这行剩余的内容
func (l *AnnotateLexer) GetRemainComment() (str string, loc lexer.Location) {
	if l.aheadToken.valid {
		if l.aheadToken.tokenKind == ATokenEOF || l.aheadToken.tokenKind == ATokenKwOther {
			str = l.chunk
		} else {
			str = l.aheadToken.tokenStr + l.chunk
		}

		loc = l.GetHeardLoc()
		loc.EndColumn = loc.EndColumn + len(l.chunk)
	} else {
		str = l.chunk

		loc = l.GetNowLoc()
		loc.StartColumn = loc.StartColumn + 1
		loc.EndColumn = loc.EndColumn + len(l.chunk)
	}

	str = strings.TrimPrefix(str, "@")
	return str, loc
}

// ErrorPrint 错误打印，词法分析报异常，终止后面的分析
func (l *AnnotateLexer) ErrorPrint(errType AnnotateErrType, needKind ATokenType, f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	errShow := fmt.Sprintf("%d: %s", l.preToken.col, err)
	paseError := ParseAnnotateErr{
		ErrType:  errType,
		NeedKind: needKind,
		ErrStr:   "annotate warn : " + err,
		ShowStr:  errShow,
		ErrToken: l.preToken,
		NowToken: l.nowToken,
		ErrLoc: lexer.Location{
			StartLine:   l.line,
			StartColumn: l.col,
			EndLine:     l.line,
			EndColumn:   l.col + len(l.chunk),
		},
	}
	panic(paseError)
}

// GetHeardLoc get look aheard Token location
func (l *AnnotateLexer) GetHeardLoc() lexer.Location {
	l.lookAheardToken()

	return lexer.Location{
		StartLine:   l.line,
		StartColumn: l.aheadToken.tokenStartCol,
		EndLine:     l.line,
		EndColumn:   l.aheadToken.col,
	}
}

// GetNowLoc get current Token location
func (l *AnnotateLexer) GetNowLoc() lexer.Location {
	if !l.nowToken.valid {
		return l.GetHeardLoc()
	}

	return lexer.Location{
		StartLine:   l.line,
		StartColumn: l.nowToken.tokenStartCol,
		EndLine:     l.line,
		EndColumn:   l.nowToken.col,
	}
}

// GetPreLoc get before Token location
func (l *AnnotateLexer) GetPreLoc() lexer.Location {
	if !l.preToken.valid {
		return lexer.Location{
			StartLine:   1,
			StartColumn: 0,
			EndLine:     1,
			EndColumn:   0,
		}
	}

	return lexer.Location{
		StartLine:   l.line,
		StartColumn: l.preToken.tokenStartCol,
		EndLine:     l.line,
		EndColumn:   l.preToken.col,
	}
}

// SetLastNormalTypeLoc 更新最新annotateast.NormalType的位置信息
func (l *AnnotateLexer) SetLastNormalTypeLoc(loc lexer.Location) {
	l.lastNormalTypeLoc = loc
}

// GetLastNormalTypeLoc 获取最新的annotateast.NormalType的位置信息
func (l *AnnotateLexer) GetLastNormalTypeLoc() (loc lexer.Location) {
	return l.lastNormalTypeLoc
}

func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
