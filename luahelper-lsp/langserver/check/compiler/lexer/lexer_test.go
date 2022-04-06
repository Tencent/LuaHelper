package lexer

import (
	"strings"
	"testing"
)

func setNowToken(l *Lexer) {
	l.nowToken.valid = true
	l.nowToken.line = 3
	l.nowToken.lineStartPos = 3423
	l.nowToken.rangeFromPos = 345452
	l.nowToken.rangeToPos = 1331
	l.nowToken.tokenKind = TkNumber
	l.nowToken.tokenStr = "fjsofjo"
}

func setAheardToken(l *Lexer) {
	l.aheadToken.valid = true
	l.aheadToken.tokenStr = "dfjosf"
	l.aheadToken.tokenKind = TkSepRparen
	l.aheadToken.line = 100
	l.aheadToken.lineStartPos = 23324
	l.aheadToken.rangeFromPos = 33423
	l.aheadToken.rangeToPos = 3003
}

func constructCopyone(l *Lexer) {
	backPreToken := l.preToken
	backNowToken := l.nowToken
	setNowToken(l)
	nextToken := l.nowToken
	l.preToken = backPreToken
	l.nowToken = backNowToken
	l.aheadToken = nextToken
}

var backPreToken Token
var backNowToken Token
var nextToken Token

func constructConstCopy(l *Lexer) {
	backPreToken = l.preToken
	backNowToken = l.nowToken
	setNowToken(l)
	nextToken = l.nowToken
	l.preToken = backPreToken
	l.nowToken = backNowToken
	l.aheadToken = nextToken
}

func constructEveryOneCopy(l *Lexer) {
	backPreToken.line = l.preToken.line
	backPreToken.lineStartPos = l.preToken.lineStartPos
	backPreToken.rangeFromPos = l.preToken.rangeFromPos
	backPreToken.rangeToPos = l.preToken.rangeToPos
	backPreToken.tokenKind = l.preToken.tokenKind
	backPreToken.tokenStr = l.preToken.tokenStr
	backPreToken.valid = l.preToken.valid

	backNowToken.line = l.nowToken.line
	backNowToken.lineStartPos = l.nowToken.lineStartPos
	backNowToken.rangeFromPos = l.nowToken.rangeFromPos
	backNowToken.rangeToPos = l.nowToken.rangeToPos
	backNowToken.tokenKind = l.nowToken.tokenKind
	backNowToken.tokenStr = l.nowToken.tokenStr
	backNowToken.valid = l.nowToken.valid

	setNowToken(l)

	nextToken.line = l.nowToken.line
	nextToken.lineStartPos = l.nowToken.lineStartPos
	nextToken.rangeFromPos = l.nowToken.rangeFromPos
	nextToken.rangeToPos = l.nowToken.rangeToPos
	nextToken.tokenKind = l.nowToken.tokenKind
	nextToken.tokenStr = l.nowToken.tokenStr
	nextToken.valid = l.nowToken.valid

	l.preToken.line = backPreToken.line
	l.preToken.lineStartPos = backPreToken.lineStartPos
	l.preToken.rangeFromPos = backPreToken.rangeFromPos
	l.preToken.rangeToPos = backPreToken.rangeToPos
	l.preToken.tokenKind = backPreToken.tokenKind
	l.preToken.tokenStr = backPreToken.tokenStr
	l.preToken.valid = backPreToken.valid

	l.nowToken.line = backNowToken.line
	l.nowToken.lineStartPos = backNowToken.lineStartPos
	l.nowToken.rangeFromPos = backNowToken.rangeFromPos
	l.nowToken.rangeToPos = backNowToken.rangeToPos
	l.nowToken.tokenKind = backNowToken.tokenKind
	l.nowToken.tokenStr = backNowToken.tokenStr
	l.nowToken.valid = backNowToken.valid

	l.aheadToken.line = nextToken.line
	l.aheadToken.lineStartPos = nextToken.lineStartPos
	l.aheadToken.rangeFromPos = nextToken.rangeFromPos
	l.aheadToken.rangeToPos = nextToken.rangeToPos
	l.aheadToken.tokenKind = nextToken.tokenKind
	l.aheadToken.tokenStr = nextToken.tokenStr
	l.aheadToken.valid = nextToken.valid
}

// BenchmarkTmpCopy 产生临时变量来赋值
func BenchmarkTmpCopy(b *testing.B) {
	b.ReportAllocs()
	l := NewLexer([]byte("local a = b"), "test.lua")

	setNowToken(l)
	setAheardToken(l)

	for i := 0; i < b.N; i++ {
		constructCopyone(l)
	}
}

// BenchmarkConstCopy 产生常量来替换
func BenchmarkConstCopy(b *testing.B) {
	b.ReportAllocs()
	l := NewLexer([]byte("local a = b"), "test.lua")

	setNowToken(l)
	setAheardToken(l)

	for i := 0; i < b.N; i++ {
		constructConstCopy(l)
	}
}

// BenchmarkEveryCopy copy每一个成员
func BenchmarkEveryCopy(b *testing.B) {
	b.ReportAllocs()
	l := NewLexer([]byte("local a = b"), "test.lua")

	setNowToken(l)
	setAheardToken(l)

	for i := 0; i < b.N; i++ {
		constructConstCopy(l)
	}
}

func isLetterTest(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func BenchmarkNormalIsLetter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isLetterTest('e')
	}
}

var a_one byte = 'a'
var z_one byte = 'z'
var A_one byte = 'A'
var Z_one byte = 'Z'

func isLettersTest(c byte) bool {
	return c >= a_one && c <= z_one || c >= A_one && c <= Z_one
}

func BenchmarkIsLetter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isLettersTest('e')
	}
}

type LexerOne struct {
	chunk string
}

func (l *LexerOne) test(s string) bool {
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

// BenchmarkLexerTest 改写的字符串匹配
func BenchmarkLexerTest(b *testing.B) {
	l := &LexerOne{
		chunk: "abdfsof",
	}
	for i := 0; i < b.N; i++ {
		l.test("ab")
	}
}

// BenchmarkLexerTestOld  利用字符串自带的方法匹配
func BenchmarkLexerTestOld(b *testing.B) {
	l := &LexerOne{
		chunk: "abdfsof",
	}
	for i := 0; i < b.N; i++ {
		strings.HasSuffix(l.chunk, "ab")
	}
}

func (l *LexerOne) testSpecial() bool {
	if len(l.chunk) < 2 {
		return false
	}

	if l.chunk[0] == 'a' && l.chunk[1] == 'b' {
		return true
	}

	return false
}

// BenchmarkLexerTestSpecial 匹配特殊的
func BenchmarkLexerTestSpecial(b *testing.B) {
	l := &LexerOne{
		chunk: "abdfsof",
	}
	for i := 0; i < b.N; i++ {
		l.testSpecial()
	}
}

type SplitStruct struct {
	LastLine    int      // 这块Type所在定义的最后行数
	TypeList    []int    //
	CommentList []string // 所有的类型的注释
	ConstList   []bool   // 是否标记常量
}

type InnerStruct struct {
	TypeList    int    //
	CommentList string // 所有的类型的注释
	ConstList   bool   // 是否标记常量
}
type CombineStruct struct {
	LastLine int // 这块Type所在定义的最后行数
	ArrList  []InnerStruct
}

// BenchmarkSplitTest1 测试分开的结构
func BenchmarkSplitTest1(b *testing.B) {

	splitArray := []SplitStruct{}
	for i := 0; i < b.N; i++ {
		splitOne := SplitStruct{}
		splitOne.CommentList = append(splitOne.CommentList, "one")
		splitOne.TypeList = append(splitOne.TypeList, 1)
		splitOne.ConstList = append(splitOne.ConstList, false)

		splitOne.CommentList = append(splitOne.CommentList, "one1")
		splitOne.TypeList = append(splitOne.TypeList, 2)
		splitOne.ConstList = append(splitOne.ConstList, false)

		splitOne.CommentList = append(splitOne.CommentList, "one2")
		splitOne.TypeList = append(splitOne.TypeList, 3)
		splitOne.ConstList = append(splitOne.ConstList, false)

		splitOne.CommentList = append(splitOne.CommentList, "one3")
		splitOne.TypeList = append(splitOne.TypeList, 3)
		splitOne.ConstList = append(splitOne.ConstList, false)

		splitArray = append(splitArray, splitOne)
	}

	b.Logf("len=%d", len(splitArray))
}

func BenchmarkCombineTest1(b *testing.B) {

	combineArray := []CombineStruct{}
	for i := 0; i < b.N; i++ {
		combineOne := CombineStruct{}
		combineOne.ArrList = append(combineOne.ArrList, InnerStruct{
			TypeList:    1,
			CommentList: "one",
			ConstList:   false,
		})

		combineOne.ArrList = append(combineOne.ArrList, InnerStruct{
			TypeList:    2,
			CommentList: "one1",
			ConstList:   false,
		})

		combineOne.ArrList = append(combineOne.ArrList, InnerStruct{
			TypeList:    3,
			CommentList: "one2",
			ConstList:   false,
		})

		combineOne.ArrList = append(combineOne.ArrList, InnerStruct{
			TypeList:    3,
			CommentList: "one3",
			ConstList:   false,
		})

		combineArray = append(combineArray, combineOne)
	}

	b.Logf("len=%d", len(combineArray))
}
