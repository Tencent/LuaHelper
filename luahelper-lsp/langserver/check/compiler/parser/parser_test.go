package parser

import "testing"

func TestParseConst(t *testing.T) {
	parser := CreateParser([]byte("local a<const> = 1"), "test")
	_, _, err := parser.BeginAnalyze()
	if err != nil {
		t.Fatalf("parser const fatal, errstr=%s", err.Error())
	}

	parser1 := CreateParser([]byte("local a<consts> = 1"), "test")
	_, _, err1 := parser1.BeginAnalyze()
	if err1 == nil {
		t.Fatalf("parser cosnt fatal")
	}
}

func TestParseJitNumber(t *testing.T) {
	contentStr := "local a=555ULL; local a=555LL; local b = 2ll; local c = 34ull; local d = 42Ull;local e=333ULl; local f = 3Ll"
	parser := CreateParser([]byte(contentStr), "test")
	_, _, err := parser.BeginAnalyze()
	if err != nil {
		t.Fatalf("parser jitNumber fatal, errstr=%s", err.Error())
	}

	contentStr1 := "local a=0x2ll; local b = 0x3aULL; local c = 0x2aUll; local d = 0xacdULL"
	parser1 := CreateParser([]byte(contentStr1), "test")
	_, _, err1 := parser1.BeginAnalyze()
	if err1 != nil {
		t.Fatalf("parser jitNumber fatal, errstr=%s", err1.Error())
	}

	contentStrIf := `if banned and left_time > 0then
	print("ok")
	end`
	parserIf := CreateParser([]byte(contentStrIf), "test")
	_, _, errIf := parserIf.BeginAnalyze()
	if errIf != nil {
		t.Fatalf("parser jitNumber fatal, errstr=%s", errIf.Error())
	}

	contentArray := []string{"local a=555UL", " local a=555LLa", "local b = 2l", "local c = 34ullb", "local d = 42Uell",
		"local e=333ULlb", "local f = 3L2l", "local a=0x2gll", "local b = 0x3aULLl", "local c = 0x2aUfll", "local d = 0xacdUL"}
	for _, str := range contentArray {
		parser2 := CreateParser([]byte(str), "test")
		_, _, err2 := parser2.BeginAnalyze()
		if err2 == nil {
			t.Fatalf("parser jitNumber fatal")
		}
	}
}

func BenchmarkHello(b *testing.B) {

}
