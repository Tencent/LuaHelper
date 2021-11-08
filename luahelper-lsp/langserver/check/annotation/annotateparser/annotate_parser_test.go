package annotateparser

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
	"testing"
)

func TestAnnotateParserType(t *testing.T) {
	commentInfo1 := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@type string",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@type string @fjosfjo",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-@type ...",
				Line: 3,
				Col:  0,
			},
		},
	}
	fragent1, errVec := ParseCommentFragment(commentInfo1)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate type fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent1.Stats) == 0 {
		t.Fatalf("parser annotate type stats len is 0")
	}

	//-- 
}

func TestAnnotateParserFun(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@type fun(f: integer|function, what?: string):debuginfo",
				Line: 1,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate fun fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) == 0 {
		t.Fatalf("parser annotate type stats len is 0")
	}
}


func TestAnnotateParserRun(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@return ...",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@return integer? code",
				Line: 2,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) == 0 {
		t.Fatalf("parser annotate type stats len is 0")
	}
}
