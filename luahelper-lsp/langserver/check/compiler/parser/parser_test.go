package parser

import (
	"io/ioutil"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseConst(t *testing.T) {
	parser := CreateParser([]byte("local a<const> = 1"), "test")
	block, _, errList := parser.BeginAnalyze()
	if len(errList) > 0 {
		t.Fatalf("parser const fatal, errstr=%s", "err.Error()")
	}

	if block == nil {
		t.Logf("is nil")
	}

	parser1 := CreateParser([]byte("local a<consts> = 1"), "test")
	_, _, errList1 := parser1.BeginAnalyze()
	if len(errList1) == 0 {
		t.Fatalf("parser cosnt fatal")
	}
}

func TestParseFunc(t *testing.T) {
	parser := CreateParser([]byte("aaa(11, a)"), "test")
	block, _, errList := parser.BeginAnalyze()
	if len(errList) > 0 {
		t.Fatalf("parser const fatal, errstr=%s", "err.Error()")
	}

	if block == nil {
		t.Logf("is nil")
	}

	parser1 := CreateParser([]byte("local a<consts> = 1"), "test")
	_, _, errList1 := parser1.BeginAnalyze()
	if len(errList1) == 0 {
		t.Fatalf("parser cosnt fatal")
	}
}

func TestParseJitNumber(t *testing.T) {
	contentStr := "local a=555ULL; local a=555LL; local b = 2ll; local c = 34ull; local d = 42Ull;local e=333ULl; local f = 3Ll"
	parser := CreateParser([]byte(contentStr), "test")
	_, _, errList := parser.BeginAnalyze()
	if len(errList) > 0 {
		t.Fatalf("parser jitNumber fatal, errstr=%s", "")
	}

	contentStr1 := "local a=0x2ll; local b = 0x3aULL; local c = 0x2aUll; local d = 0xacdULL"
	parser1 := CreateParser([]byte(contentStr1), "test")
	_, _, errList1 := parser1.BeginAnalyze()
	if len(errList1) > 0 {
		t.Fatalf("parser jitNumber fatal, errstr=%s", "")
	}

	contentStrIf := `if banned and left_time > 0then
	print("ok")
	end`
	parserIf := CreateParser([]byte(contentStrIf), "test")
	_, _, errList2 := parserIf.BeginAnalyze()
	if len(errList2) > 0 {
		t.Fatalf("parser jitNumber fatal, errstr=%s", "")
	}

	contentArray := []string{"local a=555UL", " local a=555LLa", "local b = 2l", "local c = 34ullb", "local d = 42Uell",
		"local e=333ULlb", "local f = 3L2l", "local a=0x2gll", "local b = 0x3aULLl", "local c = 0x2aUfll", "local d = 0xacdUL"}
	for _, str := range contentArray {
		parser2 := CreateParser([]byte(str), "test")
		_, _, errList3 := parser2.BeginAnalyze()
		if len(errList3) == 0 {
			t.Fatalf("parser jitNumber fatal")
		}
	}
}

func TestParseLuajitNum(t *testing.T) {
	n, ok:=parseLuajitNum("18446744073709551615ULL")
	if !ok {
		t.Fatalf("parser token err")
	}

	t.Logf("num:%v", n)
}

func TestParseIllegalToken(t *testing.T) {
	parser := CreateParser([]byte("local a = 1\n 尹飞 \n local b = 1"), "test")
	block, _, errList := parser.BeginAnalyze()
	if len(errList) == 0 {
		t.Fatalf("parser token err")
	}

	if block == nil {
		t.Logf("is nil")
	}
}

func TestParseIllegalIf(t *testing.T) {
	parser := CreateParser([]byte(" if \n local a = 1\n print(a)"), "test")
	block, _, errList := parser.BeginAnalyze()
	if len(errList) == 0 {
		t.Fatalf("parser if error")
	}

	if block == nil {
		t.Logf("is nil")
	}
}

func TestParseSpeDot(t *testing.T) {
	parser := CreateParser([]byte("."), "test")
	block, _, errList := parser.BeginAnalyze()
	if len(errList) == 0 {
		t.Fatalf("parser token err")
	}

	if block == nil {
		t.Logf("is nil")
	}
}

func TestParseIllegalExp(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)
	strRootPath := paths + "../../../../testdata/parse"
	strRootPath, _ = filepath.Abs(strRootPath)

	fileName := strRootPath + "/" + "test1.lua"
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("read file errr")
	}

	parser := CreateParser([]byte(data), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   7,
		StartColumn: 0,
		EndLine:     7,
		EndColumn:   2,
	})

	//1
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   6,
		StartColumn: 0,
		EndLine:     6,
		EndColumn:   7,
	})

	// 2
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   9,
		StartColumn: 0,
		EndLine:     9,
		EndColumn:   2,
	})

	//3
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   11,
		StartColumn: 0,
		EndLine:     11,
		EndColumn:   5,
	})

	// 4
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   13,
		StartColumn: 0,
		EndLine:     13,
		EndColumn:   6,
	})

	// 5
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   14,
		StartColumn: 0,
		EndLine:     14,
		EndColumn:   3,
	})

	// 6
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   15,
		StartColumn: 0,
		EndLine:     15,
		EndColumn:   3,
	})

	// 7
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   16,
		StartColumn: 0,
		EndLine:     16,
		EndColumn:   5,
	})

	// 8
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   17,
		StartColumn: 0,
		EndLine:     17,
		EndColumn:   2,
	})

	// 9
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   18,
		StartColumn: 0,
		EndLine:     18,
		EndColumn:   8,
	})

	// 10
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   29,
		StartColumn: 9,
		EndLine:     29,
		EndColumn:   15,
	})

	// 11
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   29,
		StartColumn: 19,
		EndLine:     29,
		EndColumn:   25,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}

	for i, oneErr := range errList {
		if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
			t.Fatalf("index=%d loc err", i)
		}
	}

	if block == nil {
		t.Logf("is nil")
	}
}

func TestParseLongStrIllegal(t *testing.T) {
	parser := CreateParser([]byte("sfasfasdf\n aa = [[sasfsf"), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   1,
		StartColumn: 0,
		EndLine:     1,
		EndColumn:   9,
	})

	// 1
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   2,
		StartColumn: 14,
		EndLine:     2,
		EndColumn:   14,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}

	for i, oneErr := range errList {
		if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
			t.Fatalf("index=%d loc err", i)
		}
	}

	if block == nil {
		t.Logf("is nil")
	}
}

func TestParseLongStrIllegal1(t *testing.T) {
	parser := CreateParser([]byte("sfasfasdf\n aa = [[sasfsf\nfadosfjasf"), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   1,
		StartColumn: 0,
		EndLine:     1,
		EndColumn:   9,
	})

	// 1
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   3,
		StartColumn: 10,
		EndLine:     3,
		EndColumn:   10,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}

	for i, oneErr := range errList {
		if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
			t.Fatalf("index=%d loc err", i)
		}
	}

	if block == nil {
		t.Logf("is nil")
	}
}

func TestParseShortStrIllegal1(t *testing.T) {
	parser := CreateParser([]byte("local a=\"fsf"), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   1,
		StartColumn: 11,
		EndLine:     1,
		EndColumn:   12,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}
	if block == nil {
		t.Logf("is nil")
	}

	for i, oneErr := range errList {
		if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
			t.Fatalf("index=%d loc err", i)
		}
	}
}

func TestParseShortStrIllegal2(t *testing.T) {
	parser := CreateParser([]byte("local a='fsf"), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   1,
		StartColumn: 11,
		EndLine:     1,
		EndColumn:   12,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}
	if block == nil {
		t.Logf("is nil")
	}

	for i, oneErr := range errList {
		if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
			t.Fatalf("index=%d loc err", i)
		}
	}
}

func TestParseShortStrIllegal3(t *testing.T) {
	parser := CreateParser([]byte("local a='fsf\nlocal b = 1\nprint(b)"), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   1,
		StartColumn: 12,
		EndLine:     1,
		EndColumn:   13,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}

	for i, oneErr := range errList {
		if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
			t.Fatalf("index=%d loc err", i)
		}
	}

	if block == nil {
		t.Logf("is nil")
	}
}

func TestParseExpectTokenIllega(t *testing.T) {
	parser := CreateParser([]byte("dfjsofjao\nsfjosjaf\nif faf fsf\nelseif fsf ffa\nend"), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   1,
		StartColumn: 0,
		EndLine:     1,
		EndColumn:   9,
	})
	// 1
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   2,
		StartColumn: 0,
		EndLine:     2,
		EndColumn:   8,
	})
	// 3
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   3,
		StartColumn: 7,
		EndLine:     3,
		EndColumn:   10,
	})
	// 4
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   4,
		StartColumn: 11,
		EndLine:     4,
		EndColumn:   14,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}

	for i, oneErr := range errList {
		if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
			t.Fatalf("index=%d loc err", i)
		}
	}

	if block == nil {
		t.Logf("is nil")
	}
}

// 解析之前panic例子=
func TestParseIllega1(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	paths, _ := filepath.Split(filename)
	strRootPath := paths + "../../../../testdata/parse"
	strRootPath, _ = filepath.Abs(strRootPath)

	fileName := strRootPath + "/" + "test2.lua"
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("read file errr")
	}

	parser := CreateParser([]byte(data), "test")
	block, _, errList := parser.BeginAnalyze()

	var expectLocList []lexer.Location

	// 0
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   1,
		StartColumn: 0,
		EndLine:     1,
		EndColumn:   9,
	})
	// 1
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   2,
		StartColumn: 0,
		EndLine:     2,
		EndColumn:   8,
	})
	// 3
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   3,
		StartColumn: 7,
		EndLine:     3,
		EndColumn:   10,
	})
	// 4
	expectLocList = append(expectLocList, lexer.Location{
		StartLine:   4,
		StartColumn: 11,
		EndLine:     4,
		EndColumn:   14,
	})

	if len(errList) != len(expectLocList) {
		t.Fatalf("parser errList len err")
	}

	// for i, oneErr := range errList {
	// 	if !lexer.CompareTwoLoc(&oneErr.Loc, &expectLocList[i]) {
	// 		t.Fatalf("index=%d loc err", i)
	// 	}
	// }

	if block == nil {
		t.Logf("is nil")
	}
}
