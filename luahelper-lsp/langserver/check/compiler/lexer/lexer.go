package lexer

import (
	"fmt"
	"luahelper-lsp/langserver/codingconv"
	"luahelper-lsp/langserver/strbytesconv"
	"regexp"
	"strings"
	"unicode/utf8"
)

const newLine = "\n"

var newLineReplacer = strings.NewReplacer(
	"\r\n", newLine,
	"\n\r", newLine,
	"\r", newLine,
	"\n", newLine,
)

// TokenStruct 词法分析出来的每个单词
type TokenStruct struct {
	valid        bool   // valid or not
	tokenStr     string // token string
	tokenKind    TkKind // token kind
	line         int    // line number
	lineStartPos int    // this line start in all pos
	rangeFromPos int    // token start in all pos
	rangeToPos   int    // token end in all pos
}

// ErrorHandler 词法分析上报错误
type ErrorHandler func(oneErr ParseError)

// Lexer 词法分析的结构
type Lexer struct {
	chunk         string // source code
	chunkName     string // source name
	line          int    // current line number
	lineStartPos  int    // this line start in all pos
	tokenStartPos int    // token start in all pos
	currentPos    int

	preToken   TokenStruct
	nowToken   TokenStruct
	aheadToken TokenStruct

	commentMap map[int]*CommentInfo // 保存所有的注释信息, key值为行号，从1开始。如果该注释有多行，为最后一行的行号。

	errHandler ErrorHandler // error reporting; or nil
}

// NewLexer 创建一个词法分析器
func NewLexer(chunk []byte, chunkName string) *Lexer {
	return &Lexer{
		chunk:     strbytesconv.BytesToString(chunk),
		chunkName: chunkName,
		line:      1,
		preToken: TokenStruct{
			valid: false,
		},
		nowToken: TokenStruct{
			valid: false,
		},
		aheadToken: TokenStruct{
			valid: false,
		},
		lineStartPos:  0,
		tokenStartPos: 0,
		currentPos:    0,
		commentMap:    map[int]*CommentInfo{},
	}
}

// SetErrHandler 设置分析错误的处理函数
func (l *Lexer) SetErrHandler(errHandler ErrorHandler) {
	l.errHandler = errHandler
}

// GetCommentMap 获取所有的注释map
func (l *Lexer) GetCommentMap() map[int]*CommentInfo {
	return l.commentMap
}

// SkipFirstLineComment 跳过首行可能的注释
func (l *Lexer) SkipFirstLineComment() {
	if len(l.chunk) < 1 {
		return
	}

	if len(l.chunk) >= 3 && l.chunk[0] == 239 && l.chunk[1] == 187 && l.chunk[2] == 191 {
		l.chunk = l.chunk[3:]
	}

	if len(l.chunk) < 1 {
		return
	}

	if l.chunk[0] != '#' {
		return
	}

	l.next(1)
	for len(l.chunk) > 0 && !isNewLine(l.chunk[0]) {
		l.next(1)
	}
}

func (l *Lexer) lookAheardToken() {
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
func (l *Lexer) LookAheadKind() TkKind {
	if l.aheadToken.valid {
		return l.aheadToken.tokenKind
	}

	l.lookAheardToken()

	if !l.aheadToken.valid {
		l.ErrorPrint("syntax error near '%s'", l.aheadToken.tokenStr)
	}
	return l.aheadToken.tokenKind
}

// GetPreTokenLoc get before Token location
func (l *Lexer) GetPreTokenLoc() Location {
	if !l.preToken.valid {
		return Location{
			StartLine:   1,
			StartColumn: 0,
			EndLine:     1,
			EndColumn:   0,
		}
	}

	return Location{
		StartLine:   l.preToken.line,
		StartColumn: l.preToken.rangeFromPos - l.preToken.lineStartPos,
		EndLine:     l.preToken.line,
		EndColumn:   l.preToken.rangeToPos - l.preToken.lineStartPos,
	}
}

// GetNowTokenLoc get current Token location
func (l *Lexer) GetNowTokenLoc() Location {
	if !l.nowToken.valid {
		return l.GetHeardTokenLoc()
	}
	if l.nowToken.lineStartPos > l.nowToken.rangeFromPos {
		return Location{
			StartLine:   l.preToken.line,
			StartColumn: l.preToken.rangeToPos - l.preToken.lineStartPos + 1,
			EndLine:     l.nowToken.line,
			EndColumn:   l.nowToken.rangeToPos - l.nowToken.lineStartPos,
		}
	}

	return Location{
		StartLine:   l.nowToken.line,
		StartColumn: l.nowToken.rangeFromPos - l.nowToken.lineStartPos,
		EndLine:     l.nowToken.line,
		EndColumn:   l.nowToken.rangeToPos - l.nowToken.lineStartPos,
	}
}

// GetHeardTokenLoc get look aheard Token location
func (l *Lexer) GetHeardTokenLoc() Location {
	l.lookAheardToken()

	if l.aheadToken.lineStartPos > l.aheadToken.rangeFromPos {
		return Location{
			StartLine:   l.nowToken.line,
			StartColumn: l.nowToken.rangeToPos - l.nowToken.lineStartPos + 1,
			EndLine:     l.aheadToken.line,
			EndColumn:   l.aheadToken.rangeToPos - l.aheadToken.lineStartPos,
		}
	}

	return Location{
		StartLine:   l.aheadToken.line,
		StartColumn: l.aheadToken.rangeFromPos - l.aheadToken.lineStartPos,
		EndLine:     l.aheadToken.line,
		EndColumn:   l.aheadToken.rangeToPos - l.aheadToken.lineStartPos,
	}
}

// NextIdentifier 下一个标识
func (l *Lexer) NextIdentifier() (line int, token string) {
	return l.NextTokenOfKind(TkIdentifier)
}

// NextTokenOfKind 检查下一个单词的类型，匹配
func (l *Lexer) NextTokenOfKind(kind TkKind) (line int, token string) {
	line, _kind, token := l.NextToken()
	if kind != _kind {
		l.ErrorPrint("syntax error near '%s'", token)
	}
	return line, token
}

// NextToken 下一个单词
func (l *Lexer) NextToken() (line int, kind TkKind, token string) {
	l.NextTokenStruct()
	line = l.nowToken.line
	kind = l.nowToken.tokenKind
	token = l.nowToken.tokenStr
	return
}

// setNowToken 设置当前的单词
func (l *Lexer) setNowToken(kind TkKind, tokenStr string) {
	l.preToken = l.nowToken
	l.nowToken.valid = true
	l.nowToken.line = l.line
	l.nowToken.lineStartPos = l.lineStartPos
	l.nowToken.rangeFromPos = l.tokenStartPos
	l.nowToken.rangeToPos = l.currentPos
	l.nowToken.tokenKind = kind
	l.nowToken.tokenStr = tokenStr
}

// NextTokenStruct 获取下一个单词结构
func (l *Lexer) NextTokenStruct() {
	if l.aheadToken.valid {
		l.preToken = l.nowToken
		l.nowToken = l.aheadToken
		l.aheadToken.valid = false
		return
	}

	l.skipWhiteSpaces()
	l.tokenStartPos = l.currentPos

	if len(l.chunk) == 0 {
		// file end
		l.setNowToken(TkEOF, "EOF")
		return
	}

	switch l.chunk[0] {
	case ';':
		l.next(1)
		l.setNowToken(TkSepSemi, ";")
		return
	case ',':
		l.next(1)
		l.setNowToken(TkSepComma, ",")
		return
	case '(':
		l.next(1)
		l.setNowToken(TkSepLparen, "(")
		return
	case ')':
		l.next(1)
		l.setNowToken(TkSepRparen, ")")
		return
	case ']':
		l.next(1)
		l.setNowToken(TkSepRbrack, "]")
		return
	case '{':
		l.next(1)
		l.setNowToken(TkSepLcurly, "{")
		return
	case '}':
		l.next(1)
		l.setNowToken(TkSepRcurly, "}")
		return
	case '+':
		l.next(1)
		l.setNowToken(TkOpAdd, "+")
		return
	case '-':
		l.next(1)
		l.setNowToken(TkOpMinus, "-")
		return
	case '*':
		l.next(1)
		l.setNowToken(TkOpMul, "*")
		return
	case '^':
		l.next(1)
		l.setNowToken(TkOpPow, "^")
		return
	case '%':
		l.next(1)
		l.setNowToken(TkOpMod, "%")
		return
	case '&':
		l.next(1)
		l.setNowToken(TkOpBand, "&")
		return
	case '|':
		l.next(1)
		l.setNowToken(TkOpBor, "|")
		return
	case '#':
		l.next(1)
		l.setNowToken(TkOpNen, "#")
		return
	case ':':
		if l.test("::") {
			l.next(2)
			l.setNowToken(TkSepLabel, "::")
		} else {
			l.next(1)
			l.setNowToken(TkSepColon, ":")
		}
		return
	case '/':
		if l.test("//") {
			l.next(2)
			l.setNowToken(TkOpIdiv, "//")
		} else {
			l.next(1)
			l.setNowToken(TkOpDiv, "/")
		}
		return
	case '~':
		if l.test("~=") {
			l.next(2)
			l.setNowToken(TkOpNe, "~=")
		} else {
			l.next(1)
			l.setNowToken(TkOpWave, "~")
		}
		return
	case '=':
		if l.test("==") {
			l.next(2)
			l.setNowToken(TkOpEq, "==")
		} else {
			l.next(1)
			l.setNowToken(TkOpAssign, "=")
		}
		return
	case '<':
		if l.test("<<") {
			l.next(2)
			l.setNowToken(TkOpShl, "<<")
		} else if l.test("<=") {
			l.next(2)
			l.setNowToken(TkOpLe, "<=")
		} else {
			l.next(1)
			l.setNowToken(TkOpLt, "<")
		}
		return
	case '>':
		if l.test(">>") {
			l.next(2)
			l.setNowToken(TkOpShr, ">>")
		} else if l.test(">=") {
			l.next(2)
			l.setNowToken(TkOpGe, ">=")
		} else {
			l.next(1)
			l.setNowToken(TkOpGt, ">")
		}
		return
	case '.':
		if l.test("...") {
			l.next(3)
			l.setNowToken(TkVararg, "...")
			return
		} else if l.test("..") {
			l.next(2)
			l.setNowToken(TkOpConcat, "..")
			return
		} else if len(l.chunk) == 1 || !isDigit(l.chunk[1]) {
			l.next(1)
			l.setNowToken(TkSepDot, ".")
			return
		}
	case '[':
		if l.test("[[") || l.test("[=") {
			l.setNowToken(TkString, l.scanLongString())
		} else {
			l.next(1)
			l.setNowToken(TkSepLbrack, "[")
		}
		return
	case '\'', '"':
		l.setNowToken(TkString, l.scanShortString())
		return
	}

	c := l.chunk[0]
	if c == '.' || isDigit(c) {
		token := l.scanNumber()
		l.setNowToken(TkNumber, token)
		return
	}

	if c == '_' || isLetter(c) {
		token := l.scanIdentifier()
		if kind, ok := keywords[token]; ok {
			l.setNowToken(kind, token)
		} else {
			l.setNowToken(TkIdentifier, token)
		}
		return
	}

	illegalStr :=  l.scanIllegalToken()
	l.ErrorPrint("unexpected Unicode-name:%s", illegalStr)
	l.setNowToken(IKIllegal, illegalStr)
	return
}

func (l *Lexer) scanIllegalToken() string {
	str := ""
	i := 0
	for i < len(l.chunk) {
		ch := l.chunk[i]
		i++
		if ch == ' ' || ch == '\n' {
			break
		}
	}

	str += l.chunk[0 : i - 1]

	//l.next(i)
	// 转换为字符的个数，不在是utf8字节数
	l.chunk = l.chunk[i:]
	strTemp := codingconv.ConvertStrToUtf8(str)
	l.currentPos = l.currentPos + utf8.RuneCountInString(strTemp)

	return str
}

func (l *Lexer) next(n int) {
	l.chunk = l.chunk[n:]
	l.currentPos = l.currentPos + n
}

func (l *Lexer) test(s string) bool {
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

// ErrorPrint 错误打印，词法分析报异常，终止后面的分析
func (l *Lexer) ErrorPrint(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	errShow := fmt.Sprintf("%s:%d: %s", l.chunkName, l.preToken.line, err)
	paseError := ParseError{
		ErrStr:      err,
		ShowStr:     errShow,
		ErrToken:    l.preToken,
		Loc:         l.GetPreTokenLoc(),
		ReadFileErr: false,
	}

	if l.errHandler != nil {
		l.errHandler(paseError)
	}

	//panic(paseError)
}

func (l *Lexer) isEnterWrap() bool {
	if len(l.chunk) < 2 {
		return false
	}

	if (l.chunk[0] == '\r' && l.chunk[1] == '\n') || (l.chunk[0] == '\n' && l.chunk[1] == '\r') {
		return true
	}

	return false
}

// isPreComment 判断是否为 --开头
func (l *Lexer) isPreComment() bool {
	if len(l.chunk) < 2 {
		return false
	}

	if l.chunk[0] == '-' && l.chunk[1] == '-' {
		return true
	}

	return false
}

// skipWhiteSpaces 跳过空格
func (l *Lexer) skipWhiteSpaces() {
	var commentInfo *CommentInfo
	lastLine := 0
	for len(l.chunk) > 0 {
		//if l.test("\r\n") || l.test("\n\r") {
		if l.isEnterWrap() {
			l.next(2)
			l.line++
			l.lineStartPos = l.currentPos
			continue
		} else if isNewLine(l.chunk[0]) {
			l.next(1)
			l.line++
			l.lineStartPos = l.currentPos
			continue
		} else if isWhiteSpace(l.chunk[0]) {
			l.next(1)
			continue
		} else if !l.isPreComment() {
			break
		}

		// 下面的逻辑是--开头的注释
		loc := Location{}
		if l.nowToken.valid {
			loc = l.GetNowTokenLoc()
		}

		// headFlag 表示是否为首行注释还是尾部注释
		headFlag := true
		if loc.EndLine == l.line {
			headFlag = false
		}

		startCol := l.currentPos - l.lineStartPos + 2
		shortFlag, skipComment := l.skipComment()

		// 剔除掉首行的注释 \n-- 当为[[ ]] 这样的注释是，会存在
		skipComment = strings.TrimSuffix(skipComment, "\n--")
		if commentInfo == nil {
			commentInfo = &CommentInfo{
				ShortFlag: shortFlag,
				HeadFlag:  headFlag,
			}

			lastLine = l.line
			if shortFlag {
				commentInfo.LineVec = append(commentInfo.LineVec, CommentLine{
					Str:  skipComment,
					Line: lastLine,
					Col:  startCol,
				})
			}

			if !headFlag {
				// 保存尾部注释
				l.commentMap[lastLine] = commentInfo
				commentInfo = nil
			}
			continue
		}

		// commentInfo之前存在，判断是否要拼接
		if (commentInfo.ShortFlag != shortFlag) || (!commentInfo.ShortFlag && !shortFlag) ||
			(l.line != lastLine+1) {
			// 以前注释保留
			l.commentMap[lastLine] = commentInfo

			// 新的注释信息
			commentInfo = &CommentInfo{
				ShortFlag: shortFlag,
				HeadFlag:  headFlag,
			}
		}

		lastLine = l.line
		if shortFlag {
			commentInfo.LineVec = append(commentInfo.LineVec, CommentLine{
				Str:  skipComment,
				Line: lastLine,
				Col:  startCol,
			})
		}
	}

	if commentInfo != nil {
		// 保存这个注释
		l.commentMap[lastLine] = commentInfo
	}
}

// skipComment 跳过注释, 获取注释内容
// shortFlag true表示长注释，false表示短注释
// strComment 为注释的内容
func (l *Lexer) skipComment() (shortFlag bool, strComment string) {
	l.next(2) // skip --

	// long comment ?
	if l.test("[") {
		if l.matchLongStringBacket() != "" {
			shortFlag = false
			strComment = l.scanLongString()
			return
		}
	}

	index := 0
	lenChunk := len(l.chunk)
	for {
		if lenChunk <= index {
			break
		}
		ch := l.chunk[index]
		if isNewLine(ch) {
			break
		}
		index++
	}

	shortFlag = true
	strComment = l.chunk[0:index]
	l.next(index)
	return
}

// scanIdentifier 获取下个标识符
func (l *Lexer) scanIdentifier() string {
	i := 1
	for ; i < len(l.chunk); i++ {
		c := l.chunk[i]
		if isLetter(c) || isDigit(c) || c == '_' {
			continue
		}

		break
	}

	str := l.chunk[0:i]
	l.next(i)

	return str
}

// 判断字符串中，是否有指定的字符
func checkHasChar(ch byte, strStr string) bool {
	for _, one := range strStr {
		if ch == (byte)(one) {
			return true
		}
	}

	return false
}

// scanNumber 扫描数字
func (l *Lexer) scanNumber() string {
	beginCh := l.chunk[0]
	i := 0
	i++
	if beginCh == '.' {
		oneFlag, oneChar := l.getIndexChar(i)
		if !oneFlag {
			l.ErrorPrint("malformed number")
		}

		beginCh = oneChar
		i++
	}

	strExpo := "Ee"

	nextFlag, nextCh := l.getIndexChar(i)
	if nextFlag {
		if beginCh == '0' && checkHasChar(nextCh, "xX") {
			strExpo = "Pp"
			i++
		}

		for {
			nextFlag2, nextCh2 := l.getIndexChar(i)
			if !nextFlag2 {
				break
			}

			if checkHasChar(nextCh2, strExpo) {
				i++
				nextFlag3, nextCh3 := l.getIndexChar(i)
				if nextFlag3 && checkHasChar(nextCh3, "-+") {
					i++
				}
			}

			nextFlag4, nextCh4 := l.getIndexChar(i)
			if !nextFlag4 {
				break
			}

			if isDigit(nextCh4) || (nextCh4 >= 'a' && nextCh4 <= 'f') || (nextCh4 >= 'A' && nextCh4 <= 'F') ||
				(nextCh4 == 'u' || nextCh4 == 'U' || nextCh4 == 'l' || nextCh4 == 'L') {
				i++
			} else if nextCh4 == '.' {
				i++
			} else {
				break
			}
		}
	}

	// 切分出来的字符串
	str := l.chunk[0:i]
	l.next(i)
	return str
}

// 匹配一个正则表达式
func (l *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(l.chunk); token != "" {
		l.next(len(token))
		return token
	}
	l.ErrorPrint("unreachable!")
	return ""
}

// scanLongString 扫描长字符串
func (l *Lexer) scanLongString() string {
	longBracket := l.matchLongStringBacket()
	if longBracket == "" {
		l.ErrorPrint("invalid long string delimiter near '%s'",
			l.chunk[0:2])
	}

	longBracketEnd := strings.Replace(longBracket, "[", "]", -1)
	longBracketIdx := strings.Index(l.chunk, longBracketEnd)
	if longBracketIdx < 0 {
		l.ErrorPrint("unfinished long string or comment")
	}

	str := l.chunk[len(longBracket):longBracketIdx]
	l.next(longBracketIdx + len(longBracketEnd))

	str = newLineReplacer.Replace(str)
	l.line += strings.Count(str, "\n")
	l.lineStartPos = l.currentPos

	if len(str) > 0 && str[0] == '\n' {
		str = str[1:]
	}

	return str
}

// 获取长字符串的前缀
func (l *Lexer) matchLongStringBacket() string {
	if !l.test("[") {
		return ""
	}

	if l.test("[[") {
		return "[["
	}

	for index := 1; index < len(l.chunk); index++ {
		if l.chunk[index] == '=' {
			continue
		}
		if l.chunk[index] != '[' {
			break
		}
		return l.chunk[:index+1]
	}
	return ""
}

// isNewWhiteSpace 判断是否空白
func (l *Lexer) isNewWhiteSpace(ch byte) bool {
	switch ch {
	case '\t', '\v', '\f', ' ':
		return true
	}
	return false
}

func (l *Lexer) getIndexChar(i int) (flag bool, ch byte) {
	if i >= len(l.chunk) {
		return false, ' '
	}

	return true, l.chunk[i]
}

func (l *Lexer) consumeEOL(i *int) bool { ///--- EOL = End Of Line
	charCode := l.chunk[*i]
	flag, peekCharCode := l.getIndexChar(*i + 1)
	if !flag {
		return false
	}

	if charCode == '\r' || charCode == '\n' {
		// Count \n\r and \r\n as one newline.
		if charCode == '\r' && peekCharCode == '\n' {
			*i = *i + 1
		}
		if peekCharCode == '\r' && charCode == '\n' {
			*i = *i + 1
		}
		l.line++
		*i = *i + 1
		l.lineStartPos = l.currentPos + *i
		return true
	}

	return false
}

func (l *Lexer) readEscapeSequence(i *int) (str string) {
	nowFlag, nowCh := l.getIndexChar(*i)
	if !nowFlag {
		l.ErrorPrint("unfinished string")
	}

	var sequenceStart = *i
	switch nowCh {
	// Lua allow the following escape sequences.
	// We don't escape the bell sequence.
	case 'a':
		(*i) = (*i) + 1
		return string('\a')
	case 'b':
		(*i) = (*i) + 1
		return string('\b')
	case 'f':
		(*i) = (*i) + 1
		return string('\f')
	case 'n':
		(*i) = (*i) + 1
		return string('\n')
	case 'r':
		(*i) = (*i) + 1
		return string('\r')
	case 't':
		(*i) = (*i) + 1
		return string('\t')
	case 'v':
		(*i) = (*i) + 1
		return string('\x0B')
	case 'x':
		// \xXX, where XX is a sequence of exactly two hexadecimal digits
		oneFlag, oneChar := l.getIndexChar(*i + 1)
		twoFlag, twoChar := l.getIndexChar(*i + 2)
		if oneFlag && twoFlag && isHexDigit(oneChar) && isHexDigit(twoChar) {
			*i = *i + 3
			return string('\\') + l.chunk[sequenceStart:*i]
		}
		(*i) = (*i) + 1

		return string('\\') + string('x')
	//case 'u': // todo，这里减少了，参考lua5.3源码的实现
	case '\n', '\r':
		if !l.consumeEOL(i) {
			l.ErrorPrint("unfinished string")
		}
		return string('\n')
	case '\\':
		*i = *i + 1
		return string(nowCh)
	case '\'':
		*i = *i + 1
		return string(nowCh)
	case 34:
		// '\"'
		*i = *i + 1
		return string(nowCh)

	// Skips the following span of white-space.
	case 'z':
		(*i) = (*i) + 1
		for *i < len(l.chunk) {
			charCode := l.chunk[*i]

			if l.isNewWhiteSpace(charCode) {
				(*i) = (*i) + 1
			} else if !l.consumeEOL(i) {
				break
			}
		}
		//skipWhiteSpace();
		return ""
	// Byte representation should for now be returned as is.

	default:
		// \ddd, where ddd is a sequence of up to three decimal digits.
		oneFlag, oneChar := l.getIndexChar(*i)
		if oneFlag && isDigit(oneChar) {
			for {
				(*i) = (*i) + 1
				tmpFlag, tmpChar := l.getIndexChar(*i)
				if tmpFlag && isDigit(tmpChar) {
					continue
				}
				break
			}

			return string('\\') + l.chunk[sequenceStart:*i]
		}

		(*i) = (*i) + 1
		// Simply return the \ as is, it's not escaping any sequence.
		return string(oneChar)
	}
}

func (l *Lexer) scanShortString() string {
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

		if i >= len(l.chunk) || (ch == '\r' || ch == '\n') {
			str += l.chunk[stringStart : i-1]
			l.next(i)
			l.ErrorPrint("unfinished string")
		}

		if ch != '\\' {
			continue
		}

		str += l.chunk[stringStart:i-1] + l.readEscapeSequence(&i)
		stringStart = i

	}

	if stringStart >= len(l.chunk) {
		l.ErrorPrint("unfinished string")
	}

	str += l.chunk[stringStart : i-1]

	//l.next(i)
	// 转换为字符的个数，不在是utf8字节数
	l.chunk = l.chunk[i:]

	strTemp := codingconv.ConvertStrToUtf8(str)

	l.currentPos = l.currentPos + utf8.RuneCountInString(strTemp) + 2

	return str
}

// SetEnd 设置结尾
func (l *Lexer) SetEnd() {
	l.chunk = ""
}

func isWhiteSpace(c byte) bool {
	if c == ' ' || c == '\t' || c == '\v' || c == '\f' {
		return true
	}

	return false
}

func isNewLine(c byte) bool {
	return c == '\r' || c == '\n'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isHexDigit(charCode byte) bool {
	return (charCode >= 48 && charCode <= 57) || (charCode >= 97 && charCode <= 102) || (charCode >= 65 && charCode <= 70)
}
